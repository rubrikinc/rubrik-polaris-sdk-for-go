// Copyright 2021 Rubrik, Inc.
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
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for Amazon Web Services.
type API struct {
	Version string
	gql     *graphql.Client
}

// NewAPI returns a new API instance. Note that this is a very cheap call to
// make.
func NewAPI(gql *graphql.Client, version string) API {
	return API{Version: version, gql: gql}
}

// CloudAccount for Amazon Web Services accounts.
type CloudAccount struct {
	ID       uuid.UUID
	NativeID string
	Name     string
	Features []Feature
}

// Feature returns the specified feature from the CloudAccount's features.
func (c CloudAccount) Feature(feature core.CloudAccountFeature) (Feature, bool) {
	for _, f := range c.Features {
		if f.Name == feature {
			return f, true
		}
	}

	return Feature{}, false
}

// Feature for Amazon Web Services accounts.
type Feature struct {
	Name     core.CloudAccountFeature
	Regions  []string
	RoleArn  string
	StackArn string
	Status   core.CloudAccountStatus
}

// HasRegion returns true if the feature is enabled for the specified region.
func (f Feature) HasRegion(region string) bool {
	for _, r := range f.Regions {
		if r == region {
			return true
		}
	}

	return false
}

// toCloudAccountID returns the Polaris cloud account id for the specified
// identity. If the identity is a Polaris cloud account id no remote endpoint
// is called.
func (a API) toCloudAccountID(ctx context.Context, id IdentityFunc) (uuid.UUID, error) {
	a.gql.Log().Print(log.Trace, "polaris/aws.toCloudAccountID")

	if id == nil {
		return uuid.Nil, errors.New("polaris: id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	if identity.internal {
		id, err := uuid.Parse(identity.id)
		if err != nil {
			return uuid.Nil, err
		}

		return id, nil
	}

	accountsWithFeatures, err := aws.Wrap(a.gql).CloudAccountsWithFeatures(ctx, core.CloudNativeProtection, identity.id)
	if err != nil {
		return uuid.Nil, err
	}
	if len(accountsWithFeatures) < 1 {
		return uuid.Nil, fmt.Errorf("polaris: account %w", graphql.ErrNotFound)
	}
	if len(accountsWithFeatures) > 1 {
		return uuid.Nil, fmt.Errorf("polaris: account %w", graphql.ErrNotUnique)
	}

	return accountsWithFeatures[0].Account.ID, nil
}

// toNativeID returns the AWS account id for the specified identity. If the
// identity is an AWS account id no remote endpoint is called.
func (a API) toNativeID(ctx context.Context, id IdentityFunc) (string, error) {
	a.gql.Log().Print(log.Trace, "polaris/aws.toNativeID")

	if id == nil {
		return "", errors.New("polaris: id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return "", err
	}

	if !identity.internal {
		return identity.id, nil
	}

	uid, err := uuid.Parse(identity.id)
	if err != nil {
		return "", nil
	}

	accountWithFeatures, err := aws.Wrap(a.gql).CloudAccountWithFeatures(ctx, uid, core.CloudNativeProtection)
	if err != nil {
		return "", err
	}

	return accountWithFeatures.Account.NativeID, nil
}

// toCloudAccount converts a polaris/graphql/aws CloudAccountWithFeatures to a
// polaris/aws CloudAccount.
func toCloudAccount(accountWithFeatures aws.CloudAccountWithFeatures) CloudAccount {
	features := make([]Feature, 0, len(accountWithFeatures.Features))
	for _, feature := range accountWithFeatures.Features {
		features = append(features, Feature{
			Name:     feature.Name,
			Regions:  aws.FormatRegions(feature.Regions),
			RoleArn:  feature.RoleArn,
			StackArn: feature.StackArn,
			Status:   feature.Status,
		})
	}

	return CloudAccount{
		ID:       accountWithFeatures.Account.ID,
		NativeID: accountWithFeatures.Account.NativeID,
		Name:     accountWithFeatures.Account.Name,
		Features: features,
	}
}

// Account returns the account with specified id and feature.
func (a API) Account(ctx context.Context, id IdentityFunc, feature core.CloudAccountFeature) (CloudAccount, error) {
	a.gql.Log().Print(log.Trace, "polaris/aws.Account")

	if id == nil {
		return CloudAccount{}, errors.New("polaris: id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return CloudAccount{}, err
	}

	if identity.internal {
		cloudAccountID, err := uuid.Parse(identity.id)
		if err != nil {
			return CloudAccount{}, err
		}

		accountsWithFeatures, err := aws.Wrap(a.gql).CloudAccountsWithFeatures(ctx, feature, "")
		if err != nil {
			return CloudAccount{}, err
		}

		for _, accountWithFeatures := range accountsWithFeatures {
			if accountWithFeatures.Account.ID == cloudAccountID {
				return toCloudAccount(accountWithFeatures), nil
			}
		}
	} else {
		accountsWithFeatures, err := aws.Wrap(a.gql).CloudAccountsWithFeatures(ctx, feature, identity.id)
		if err != nil {
			return CloudAccount{}, err
		}
		if len(accountsWithFeatures) == 1 {
			return toCloudAccount(accountsWithFeatures[0]), nil
		}
		if len(accountsWithFeatures) > 1 {
			return CloudAccount{}, fmt.Errorf("polaris: account %w", graphql.ErrNotUnique)
		}
	}

	return CloudAccount{}, fmt.Errorf("polaris: account %w", graphql.ErrNotFound)
}

// Accounts return all accounts with the specified feature matching the filter.
// The filter can be used to search for account id, account name and role arn.
func (a API) Accounts(ctx context.Context, feature core.CloudAccountFeature, filter string) ([]CloudAccount, error) {
	a.gql.Log().Print(log.Trace, "polaris/aws.Accounts")

	accountsWithFeatures, err := aws.Wrap(a.gql).CloudAccountsWithFeatures(ctx, feature, filter)
	if err != nil {
		return nil, err
	}

	accounts := make([]CloudAccount, 0, len(accountsWithFeatures))
	for _, accountWithFeatures := range accountsWithFeatures {
		accounts = append(accounts, toCloudAccount(accountWithFeatures))
	}

	return accounts, nil
}

// AddAccount adds the specified project to Polaris. If name isn't given as an
// option it's derived from information in the cloud. The result can vary
// slightly depending on permissions. Returns the Polaris cloud account id of
// the added account.
func (a API) AddAccount(ctx context.Context, account AccountFunc, opts ...OptionFunc) (uuid.UUID, error) {
	a.gql.Log().Print(log.Trace, "polaris/aws.AddAccount")

	if account == nil {
		return uuid.Nil, errors.New("polaris: account is not allowed to be nil")
	}
	config, err := account(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return uuid.Nil, err
		}
	}
	if options.name != "" {
		config.name = options.name
	}

	accountInit, err := aws.Wrap(a.gql).ValidateAndCreateCloudAccount(ctx, config.id, config.name,
		core.CloudNativeProtection)
	if err != nil {
		return uuid.Nil, err
	}

	err = aws.Wrap(a.gql).FinalizeCloudAccountProtection(ctx, config.id, config.name, core.CloudNativeProtection,
		options.regions, accountInit)
	if err != nil {
		return uuid.Nil, err
	}

	// Retrieve the Polaris cloud account id of the newly added account.
	akkount, err := a.Account(ctx, AccountID(config.id), core.CloudNativeProtection)
	if err != nil {
		return uuid.Nil, err
	}

	a.gql.Log().Printf(log.Debug, "creating CloudFormation stack: %v", accountInit.StackName)

	err = awsUpdateStack(ctx, config.config, accountInit.StackName, accountInit.TemplateURL)
	if err != nil {
		return uuid.Nil, err
	}

	return akkount.ID, nil
}

// RemoveAccount removes the account with the specified id from Polaris. If
// deleteSnapshots is true the snapshots are deleted otherwise they are kept.
func (a API) RemoveAccount(ctx context.Context, account AccountFunc, deleteSnapshots bool) error {
	a.gql.Log().Print(log.Trace, "polaris/aws.RemoveAccount")

	if account == nil {
		return errors.New("polaris: account is not allowed to be nil")
	}
	config, err := account(ctx)
	if err != nil {
		return err
	}

	akkount, err := a.Account(ctx, AccountID(config.id), core.AllFeatures)
	if err != nil {
		return err
	}

	cnpFeature, ok := akkount.Feature(core.CloudNativeProtection)
	if !ok {
		return errors.New("polaris: account does not have the cloud native protection feature")
	}

	if cnpFeature.Status != core.Disabled {
		jobID, err := aws.Wrap(a.gql).StartNativeAccountDisableJob(ctx, akkount.ID, aws.EC2, deleteSnapshots)
		if err != nil {
			return err
		}

		state, err := core.Wrap(a.gql).WaitForTaskChain(ctx, jobID, 10*time.Second)
		if err != nil {
			return err
		}
		if state != core.TaskChainSucceeded {
			return fmt.Errorf("polaris: taskchain failed: jobID=%v, state=%v", jobID, state)
		}
	}

	cfmURL, err := aws.Wrap(a.gql).PrepareCloudAccountDeletion(ctx, akkount.ID, core.CloudNativeProtection)
	if err != nil {
		return err
	}

	// For now we don't downgrade the stack, we just remove it if Cloud Native
	// Protection is the last feature being removed.
	features := len(akkount.Features)
	if _, ok := akkount.Feature(core.CloudAccounts); ok {
		features--
	}
	if _, ok := akkount.Feature(core.CloudNativeProtection); ok {
		features--
	}
	if _, ok := akkount.Feature(core.Exocompute); ok {
		features--
	}
	if features == 0 {
		i := strings.LastIndex(cfmURL, "#/stack/detail") + 1
		if i == 0 {
			return errors.New("polaris: CloudFormation url does not contain #/stack/detail")
		}

		u, err := url.Parse(cfmURL[i:])
		if err != nil {
			return err
		}
		stackID := u.Query().Get("stackId")

		a.gql.Log().Printf(log.Debug, "deleting CloudFormation stack: %s", stackID)
		err = awsDeleteStack(ctx, config.config, stackID)
		if err != nil {
			return err
		}
	}

	err = aws.Wrap(a.gql).FinalizeCloudAccountDeletion(ctx, akkount.ID, core.CloudNativeProtection)
	if err != nil {
		return err
	}

	return nil
}

// UpdateAccount updates the account with the specied id and feature. It's
// currently not possible to update the account name.
func (a API) UpdateAccount(ctx context.Context, id IdentityFunc, feature core.CloudAccountFeature, opts ...OptionFunc) error {
	a.gql.Log().Print(log.Trace, "polaris/aws.UpdateAccount")

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return err
		}
	}
	if len(options.regions) == 0 {
		return errors.New("polaris: nothing to update")
	}

	account, err := a.Account(ctx, id, feature)
	if err != nil {
		return err
	}

	err = aws.Wrap(a.gql).UpdateCloudAccount(ctx, core.UpdateRegions, account.ID, feature, options.regions)
	if err != nil {
		return err
	}

	return nil
}
