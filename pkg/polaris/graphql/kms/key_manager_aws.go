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

package kms

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CreateAWSKeyManagerParams holds the parameters for creating a cloud account
// provisioned AWS KMS key manager. RSC resolves the credentials from the cloud
// account at call time, so no access key is stored.
type CreateAWSKeyManagerParams struct {
	// CloudAccountID is the RSC cloud account ID of the AWS account holding the
	// KMS keys. The account must be onboarded with the
	// CLOUD_NATIVE_UEM_KEY_MANAGEMENT feature in a connected state.
	CloudAccountID string `json:"cloudAccountId"`

	// Name is the name of the key manager, unique across all key managers.
	Name string `json:"instanceName"`

	// Region is the AWS region of the KMS keys.
	Region aws.Region `json:"region"`

	Description       string `json:"description,omitempty"`
	STSRegionEndpoint string `json:"stsRegionEndpoint,omitempty"`
	KMSRegionEndpoint string `json:"kmsRegionEndpoint,omitempty"`
}

// CreateAWSKeyManager creates a cloud account provisioned AWS KMS key manager
// and returns its ID.
func CreateAWSKeyManager(ctx context.Context, gql *graphql.Client, params CreateAWSKeyManagerParams) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	query := addAwsKmsWithCloudAccountQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				KeyManagerID uuid.UUID `json:"kmsKeyVaultId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}

	if payload.Data.Result.KeyManagerID == uuid.Nil {
		return uuid.Nil, graphql.ResponseError(query, fmt.Errorf("no key manager ID returned"))
	}

	return payload.Data.Result.KeyManagerID, nil
}

// UpdateAWSKeyManagerParams holds the parameters for updating a cloud account
// provisioned AWS KMS key manager. An empty field leaves the current value
// unchanged. The cloud account and the region cannot be changed.
type UpdateAWSKeyManagerParams struct {
	Name              string `json:"instanceName,omitempty"`
	Description       string `json:"description,omitempty"`
	STSRegionEndpoint string `json:"stsRegionEndpoint,omitempty"`
	KMSRegionEndpoint string `json:"kmsRegionEndpoint,omitempty"`
}

// UpdateAWSKeyManager updates the cloud account provisioned AWS KMS key manager
// with the specified ID.
func UpdateAWSKeyManager(ctx context.Context, gql *graphql.Client, keyManagerID uuid.UUID, params UpdateAWSKeyManagerParams) error {
	gql.Log().Print(log.Trace)

	query := updateAwsKmsWithCloudAccountQuery
	if _, err := gql.Request(ctx, query, struct {
		KeyManagerID uuid.UUID `json:"kmsKeyVaultId"`
		UpdateAWSKeyManagerParams
	}{KeyManagerID: keyManagerID, UpdateAWSKeyManagerParams: params}); err != nil {
		return graphql.RequestError(query, err)
	}

	return nil
}

// AWSKeyIDsInKMS returns the IDs of the AWS KMS keys reachable through the key
// manager. The IDs are AWS KMS key IDs and are filtered by RSC to enabled
// customer managed symmetric encryption keys.
func AWSKeyIDsInKMS(ctx context.Context, gql *graphql.Client, keyManagerID uuid.UUID) ([]string, error) {
	gql.Log().Print(log.Trace)

	query := keyIdsInAwsKmsQuery
	buf, err := gql.Request(ctx, query, struct {
		KeyManagerID uuid.UUID `json:"kmsKeyVaultId"`
	}{KeyManagerID: keyManagerID})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []string `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// AWSKMSRegions returns the AWS regions which can host an AWS KMS key manager.
func AWSKMSRegions(ctx context.Context, gql *graphql.Client) ([]string, error) {
	gql.Log().Print(log.Trace)

	query := awsKmsRegionsQuery
	buf, err := gql.Request(ctx, query, struct{}{})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Regions []string `json:"regions"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Regions, nil
}
