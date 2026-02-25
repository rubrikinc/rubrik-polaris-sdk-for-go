// Copyright 2025 Rubrik, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package aws

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AddAccountWithCFT adds the AWS account, with specified RSC features, to RSC
// using an AWS CloudFormation template workflow. Returns the RSC cloud account
// ID of the added account. If adding the account fails due to permission
// problems when creating the CloudFormation stack, it's safe to do call again
// with the same parameters after the permission problems have been resolved.
// If name isn't given as an option it's derived from information in the cloud.
// The result can vary slightly depending on AWS permissions.
func (a API) AddAccountWithCFT(ctx context.Context, account AccountFunc, features []core.Feature, opts ...OptionFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if account == nil {
		return uuid.Nil, errors.New("account is not allowed to be nil")
	}
	if len(features) == 0 {
		return uuid.Nil, errors.New("no features specified")
	}

	config, err := account(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup account: %s", err)
	}
	if config.config == nil {
		return uuid.Nil, errors.New("account config is required by the CloudFormation workflow")
	}
	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return uuid.Nil, fmt.Errorf("failed to lookup option: %s", err)
		}
	}
	if options.name != "" {
		config.name = options.name
	}

	// If there already is an RSC cloud account for the given AWS account we use
	// the same account name when adding the feature. RSC does not allow the
	// name to change between features.
	cloudAccount, err := a.AccountByNativeID(ctx, config.NativeID)
	if err != nil && !errors.Is(err, graphql.ErrNotFound) {
		return uuid.Nil, fmt.Errorf("failed to get account: %s", err)
	}
	if err == nil {
		config.name = cloudAccount.Name
	}

	// Onboard features. Outpost account needs to exist prior to adding the
	// features since the account will be referenced in the CloudFormation
	// template.
	if feature, ok := core.LookupFeature(features, core.FeatureOutpost); ok {
		features = slices.DeleteFunc(features, func(feature core.Feature) bool {
			return feature.Equal(core.FeatureOutpost)
		})

		separateOutpostAccount, err := a.addOutpostWithCFT(ctx, feature, config, options)
		if err != nil {
			return uuid.Nil, err
		}

		// When the cloud account doesn't exist and only the outpost feature is
		// added, for a separate outpost account, the lookup of the RSC cloud
		// account ID for the main account will fail.
		if separateOutpostAccount && cloudAccount.ID == uuid.Nil && len(features) == 0 {
			return uuid.Nil, nil
		}
	}
	if len(features) != 0 {
		if err = a.addAccountWithCFT(ctx, features, config, options); err != nil {
			return uuid.Nil, err
		}
	}

	// If the RSC cloud account did not exist prior, we retrieve the RSC cloud
	// account ID.
	if cloudAccount.ID == uuid.Nil {
		cloudAccount, err = a.AccountByNativeID(ctx, config.NativeID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to get account: %s", err)
		}
	}

	return cloudAccount.ID, nil
}

// RemoveAccountWithCFT removes the RSC features from the account with the
// specified ID, onboarded with the AWS CloudFormation template workflow.
// If a Cloud Native Protection feature is being removed and deleteSnapshots is
// true, the snapshots are deleted otherwise they are kept.
func (a API) RemoveAccountWithCFT(ctx context.Context, account AccountFunc, features []core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	if account == nil {
		return errors.New("account is not allowed to be nil")
	}
	if len(features) == 0 {
		return errors.New("no features specified")
	}

	config, err := account(ctx)
	if err != nil {
		return fmt.Errorf("failed to lookup account: %s", err)
	}
	if config.config == nil {
		return errors.New("account config is required by the CloudFormation workflow")
	}

	cloudAccount, err := a.AccountByNativeID(ctx, config.NativeID)
	if err != nil {
		return fmt.Errorf("failed to get account: %s", err)
	}

	// Check that the account has all the features that are going to be removed.
	for _, feature := range features {
		if _, ok := cloudAccount.Feature(feature); !ok {
			return fmt.Errorf("feature %s %w", feature, graphql.ErrNotFound)
		}
	}

	// The Cloud Discovery feature must be removed after all protection
	// features.
	if _, ok := core.LookupFeature(features, core.FeatureCloudDiscovery); ok {
		features = slices.DeleteFunc(features, func(feature core.Feature) bool {
			return feature.Equal(core.FeatureCloudDiscovery)
		})
		features = append(features, core.FeatureCloudDiscovery)
	}

	for _, feature := range features {
		if err := a.removeAccountWithCFT(ctx, config, cloudAccount, feature, deleteSnapshots); err != nil {
			return err
		}
	}

	return nil
}

func (a API) addAccountWithCFT(ctx context.Context, features []core.Feature, config account, options options) error {
	a.log.Print(log.Trace)

	accountInit, err := aws.Wrap(a.client).ValidateAndCreateCloudAccount(ctx, config.cloud, config.NativeID, config.name, features)
	if err != nil {
		return fmt.Errorf("failed to validate account: %s", err)
	}

	err = aws.Wrap(a.client).FinalizeCloudAccountProtection(ctx, config.cloud, config.NativeID, config.name, features, options.regions, accountInit)
	if err != nil {
		return fmt.Errorf("failed to add account: %s", err)
	}

	err = awsUpdateStack(ctx, a.client.Log(), *config.config, accountInit.StackName, accountInit.TemplateURL)
	if err != nil {
		return fmt.Errorf("failed to update CloudFormation stack: %s", err)
	}

	return nil
}

// addOutpostWithCFT adds the Outpost feature to the account specified in the
// options. If no account is specified in the options, the Outpost feature is
// added to the account specified in the config. Returns true if the Outpost
// account is a separate account, which is the case when a separate outpost
// account ID is specified in the options.
func (a API) addOutpostWithCFT(ctx context.Context, feature core.Feature, config account, options options) (bool, error) {
	// If no outpost account ID is given, we use the same account as specified
	// by the config.
	outpostAccountID := options.outpostAccountID
	if outpostAccountID == "" {
		outpostAccountID = config.NativeID
	}

	var separateOutpostAccount bool
	if outpostAccountID != config.NativeID {
		separateOutpostAccount = true
	}

	if options.outpostAccountProfile != nil {
		var err error
		config, err = options.outpostAccountProfile(ctx)
		if err != nil {
			return separateOutpostAccount, fmt.Errorf("failed to get outpost account config: %s", err)
		}
	}
	config.NativeID = outpostAccountID

	return separateOutpostAccount, a.addAccountWithCFT(ctx, []core.Feature{feature}, config, options)
}

func (a API) removeAccountWithCFT(ctx context.Context, config account, account CloudAccount, feature core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	if err := a.disableFeature(ctx, account, feature, deleteSnapshots); err != nil {
		return fmt.Errorf("failed to disable feature %s: %s", feature, err)
	}

	cfmURL, err := aws.Wrap(a.client).PrepareCloudAccountDeletion(ctx, account.ID, feature)
	if err != nil {
		return fmt.Errorf("failed to prepare to delete account: %s", err)
	}

	if cfmURL != "" {
		if strings.Contains(cfmURL, "#/stack/update") {
			i := strings.LastIndex(cfmURL, "#/stack/update") + 1
			if i == 0 {
				return errors.New("CloudFormation url does not contain #/stack/update")
			}

			u, err := url.Parse(cfmURL[i:])
			if err != nil {
				return fmt.Errorf("failed to parse CloudFormation url: %s", err)
			}
			stackID := u.Query().Get("stackId")
			tmplURL := u.Query().Get("templateURL")

			err = awsUpdateStack(ctx, a.client.Log(), *config.config, stackID, tmplURL)
			if err != nil {
				return fmt.Errorf("failed to update CloudFormation stack: %s", err)
			}
		} else {
			i := strings.LastIndex(cfmURL, "#/stack/detail") + 1
			if i == 0 {
				return errors.New("CloudFormation url does not contain #/stack/detail")
			}

			u, err := url.Parse(cfmURL[i:])
			if err != nil {
				return fmt.Errorf("failed to parse CloudFormation url: %s", err)
			}
			stackID := u.Query().Get("stackId")

			err = awsDeleteStack(ctx, a.client.Log(), *config.config, stackID)
			if err != nil {
				return fmt.Errorf("failed to delete CloudFormation stack: %s", err)
			}
		}
	}

	err = aws.Wrap(a.client).FinalizeCloudAccountDeletion(ctx, account.ID, feature)
	if err != nil {
		return fmt.Errorf("failed to delete account: %s", err)
	}

	return nil
}
