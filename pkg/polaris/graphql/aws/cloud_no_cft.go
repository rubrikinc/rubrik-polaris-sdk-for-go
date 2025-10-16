// Copyright 2023 Rubrik, Inc.
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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// PermissionPolicyArtifact holds the permission policies for a specific cloud
// and feature set.
type PermissionPolicyArtifact struct {
	ArtifactKey             string   `json:"externalArtifactKey"`
	ManagedPolicies         []string `json:"awsManagedPolicies"`
	CustomerManagedPolicies []struct {
		Feature        string `json:"feature"`
		PolicyName     string `json:"policyName"`
		PolicyDocument string `json:"policyDocumentJson"`
	} `json:"customerManagedPolicies"`
}

// AllPermissionPolicies returns all permission policies for the specified cloud
// and feature set.
func (a API) AllPermissionPolicies(ctx context.Context, cloud Cloud, features []core.Feature, ec2RecoveryRolePath string) ([]PermissionPolicyArtifact, error) {
	a.log.Print(log.Trace)

	featuresWithoutPG, featuresWithPG, err := core.FilterFeaturesOnPermissionGroups(features)
	if err != nil {
		return nil, err
	}

	query := allAwsPermissionPoliciesQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Cloud          Cloud          `json:"cloudType"`
		Features       []string       `json:"features,omitempty"`
		FeaturesWithPG []core.Feature `json:"featuresWithPG,omitempty"`
		RolePath       string         `json:"ec2RecoveryRolePath,omitempty"`
	}{Cloud: cloud, Features: featuresWithoutPG, FeaturesWithPG: featuresWithPG, RolePath: ec2RecoveryRolePath})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []PermissionPolicyArtifact `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// TrustPolicyAccount holds the native ID and external ID.
type TrustPolicyAccount struct {
	ID         string `json:"id"`
	ExternalID string `json:"externalId,omitempty"`
}

// TrustPolicy holds the native ID and the artifacts.
type TrustPolicy struct {
	NativeID  string                `json:"awsNativeId"`
	Artifacts []TrustPolicyArtifact `json:"artifacts"`
}

// TrustPolicyArtifact holds the artifact key and the corresponding trust policy
// document. If an error occurs ErrorMessage will be set.
type TrustPolicyArtifact struct {
	ExternalArtifactKey string `json:"externalArtifactKey"`
	TrustPolicyDoc      string `json:"trustPolicyDoc"`
	ErrorMessage        string `json:"errorMessage"`
}

// TrustPolicy returns the trust policy for the specified account and external
// id.
func (a API) TrustPolicy(ctx context.Context, cloud Cloud, features []core.Feature, trustPolicyAccounts []TrustPolicyAccount) ([]TrustPolicy, error) {
	a.log.Print(log.Trace)

	query := awsTrustPolicyQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Cloud          Cloud                `json:"cloudType"`
		Features       []string             `json:"features"`
		NativeAccounts []TrustPolicyAccount `json:"awsNativeAccounts"`
	}{Cloud: cloud, Features: core.FeatureNames(features), NativeAccounts: trustPolicyAccounts})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				TrustPolicies []TrustPolicy `json:"result"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.TrustPolicies, nil
}

// AccountFeatureArtifact holds the artifacts for a cloud account, identified by
// the native ID.
type AccountFeatureArtifact struct {
	NativeID  string             `json:"awsNativeId"`
	Features  []string           `json:"features"`
	Artifacts []ExternalArtifact `json:"externalArtifacts"`
}

// ExternalArtifact holds the key and value for an artifact.
type ExternalArtifact struct {
	ExternalArtifactKey   string `json:"externalArtifactKey"`
	ExternalArtifactValue string `json:"externalArtifactValue"`
}

// NativeIDToRSCIDMapping holds a mapping between cloud account ID and native
// ID. If an error occurs Message will be set.
type NativeIDToRSCIDMapping struct {
	CloudAccountID string `json:"awsCloudAccountId"`
	NativeID       string `json:"awsNativeId"`
	Message        string `json:"Message"`
}

// RegisterFeatureArtifacts registers the specified artifacts with the cloud
// account identified by the native ID.
func (a API) RegisterFeatureArtifacts(ctx context.Context, cloud Cloud, artifacts []AccountFeatureArtifact) ([]NativeIDToRSCIDMapping, error) {
	a.log.Print(log.Trace)

	query := registerAwsFeatureArtifactsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Cloud     Cloud                    `json:"cloudType"`
		Artifacts []AccountFeatureArtifact `json:"awsArtifacts"`
	}{Cloud: cloud, Artifacts: artifacts})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Mappings []NativeIDToRSCIDMapping `json:"allAwsNativeIdtoRscIdMappings"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Mappings, nil
}

// FeatureResult gives the result of the delete operation for a feature.
type FeatureResult struct {
	Feature string
	Success bool
}

// DeleteCloudAccountWithoutCft deletes the cloud account identified by the
// native ID. Note that certain features needs to be disabled before being
// deleted.
func (a API) DeleteCloudAccountWithoutCft(ctx context.Context, nativeID string, features []core.Feature) ([]FeatureResult, error) {
	a.log.Print(log.Trace)

	query := bulkDeleteAwsCloudAccountWithoutCftQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		NativeID string   `json:"awsNativeId"`
		Features []string `json:"features"`
	}{NativeID: nativeID, Features: core.FeatureNames(features)})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				FeatureResult []FeatureResult `json:"deleteAwsCloudAccountWithoutCftResp"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.FeatureResult, nil
}

// ArtifactsToDelete holds the feature and the artifacts to delete.
type ArtifactsToDelete struct {
	Feature           string             `json:"feature"`
	ArtifactsToDelete []ExternalArtifact `json:"artifactsToDelete"`
}

// ArtifactsToDelete returns all feature artifacts registered with the cloud
// account.
func (a API) ArtifactsToDelete(ctx context.Context, nativeID string) ([]ArtifactsToDelete, error) {
	a.log.Print(log.Trace)

	query := awsArtifactsToDeleteQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		NativeID string   `json:"awsNativeId"`
		Features []string `json:"features"`
	}{NativeID: nativeID, Features: []string{}})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				ArtifactsToDelete []ArtifactsToDelete `json:"artifactsToDelete"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.ArtifactsToDelete, nil
}

// UpgradeCloudAccountFeaturesWithoutCft notifies RSC that permissions have been
// updated for the specified cloud account and features.
func (a API) UpgradeCloudAccountFeaturesWithoutCft(ctx context.Context, cloudAccountID uuid.UUID, features []core.Feature) error {
	a.log.Print(log.Trace)

	query := upgradeAwsCloudAccountFeaturesWithoutCftQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID `json:"awsCloudAccountId"`
		Features       []string  `json:"features"`
	}{CloudAccountID: cloudAccountID, Features: core.FeatureNames(features)})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result {
		return graphql.ResponseError(query, err)
	}

	return nil
}
