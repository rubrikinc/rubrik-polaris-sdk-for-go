// Copyright 2026 Rubrik, Inc.
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
	"encoding/json"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CloudAccountServiceType identifies the RSC service type of an AWS cloud
// account. It corresponds to the AwsCloudAccountServiceType GraphQL enum.
type CloudAccountServiceType string

const (
	// ServiceTypeUnspecified is the zero value. The RSC backend does not accept
	// an unspecified service type for account onboarding, so the validate and
	// finalize calls treat it as ServiceTypeNonBaaS.
	ServiceTypeUnspecified CloudAccountServiceType = ""

	// ServiceTypeBaaS identifies a Backup as a Service (BaaS), Rubrik-hosted,
	// AWS cloud account.
	ServiceTypeBaaS CloudAccountServiceType = "AWS_CLOUD_ACCOUNT_SERVICE_TYPE_BAAS"

	// ServiceTypeNonBaaS identifies a standard (non-BaaS) AWS cloud account.
	ServiceTypeNonBaaS CloudAccountServiceType = "AWS_CLOUD_ACCOUNT_SERVICE_TYPE_NON_BAAS"
)

// TriggerCftStatusPolling asks RSC to start a background task that polls the
// CloudFormation stack status in AWS for the specified account and features.
// The accountID is the RSC cloud account ID. This call is best-effort:
// callers should log and continue on error.
func (a API) TriggerCftStatusPolling(ctx context.Context, accountID uuid.UUID, features []core.Feature) error {
	a.log.Print(log.Trace)

	query := awsTriggerCftStatusPollingQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		AwsCloudAccountID uuid.UUID `json:"awsCloudAccountId"`
		Features          []string  `json:"features"`
	}{AwsCloudAccountID: accountID, Features: core.FeatureNames(features)})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result any `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}

// baasCloudAccount is a single cloud account in a CompleteBaasOnboarding
// request (CloudAccountDetailsInput).
type baasCloudAccount struct {
	ID       uuid.UUID `json:"id"`
	NativeID string    `json:"nativeId"`
	Name     string    `json:"name"`
}

// baasRegion is the wrapped region object in a CompleteBaasOnboarding request
// (CloudRegionInput).
type baasRegion struct {
	Region struct {
		AwsRegion aws.RegionEnum `json:"awsRegion"`
	} `json:"region"`
}

// CompleteBaasOnboarding completes the BaaS onboarding flow for the specified
// RSC cloud account, features and regions. It is the final step of the
// RSC-managed AWS onboarding flow.
func (a API) CompleteBaasOnboarding(ctx context.Context, accountID uuid.UUID, awsNativeID, name string, features []core.Feature, regions []aws.Region) error {
	a.log.Print(log.Trace)

	baasRegions := make([]baasRegion, 0, len(regions))
	for _, region := range regions {
		var r baasRegion
		r.Region.AwsRegion = region.ToRegionEnum()
		baasRegions = append(baasRegions, r)
	}

	query := completeBaasOnboardingQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccounts []baasCloudAccount `json:"cloudAccounts"`
		CloudProvider string             `json:"cloudProvider"`
		Features      []string           `json:"features"`
		Regions       []baasRegion       `json:"regions"`
	}{
		CloudAccounts: []baasCloudAccount{{ID: accountID, NativeID: awsNativeID, Name: name}},
		CloudProvider: "AWS",
		Features:      core.FeatureNames(features),
		Regions:       baasRegions,
	})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result any `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}
