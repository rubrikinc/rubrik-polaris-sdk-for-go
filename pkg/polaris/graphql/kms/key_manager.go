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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// KeyManager represents a KMS key manager registered in RSC.
type KeyManager struct {
	Type               Type   `json:"kmsType"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	DataLocationsCount int    `json:"dataLocationsCount"`

	// NaturalID identifies whatever the key manager connects to, in the terms
	// of that system. It is not an RSC ID and its meaning depends on both the
	// KMS type and the provisioning method: the IAM role ARN for a direct AWS
	// key manager, the RSC cloud account ID for a cloud account provisioned
	// one, the vault URL for Azure and the key ring for GCP.
	NaturalID string `json:"naturalId"`

	// KeyVaultInfo holds the fields which depend on the KMS type. The fields
	// below the common ones only apply to a single KMS type and are empty for
	// the others. Which of them are set follows Type.
	KeyVaultInfo struct {
		// ID is the value archival locations refer to the key manager by, e.g.
		// the keyManagerId field of an AWS KMS key identifier. It is empty for
		// a KMIP server, which has no UUID.
		ID                 uuid.UUID          `json:"id"`
		ProvisioningMethod ProvisioningMethod `json:"provisioningMethod"`

		// CloudAccountID is the RSC cloud account backing the key manager.
		// Empty unless ProvisioningMethod is MethodCloudAccount.
		CloudAccountID string `json:"cloudAccountId"`

		// AWS.
		IAMRoleARN        string         `json:"iamRoleArn"`
		AWSRegion         aws.RegionEnum `json:"awsRegion"`
		STSRegionEndpoint string         `json:"stsRegionEndpoint"`
		KMSRegionEndpoint string         `json:"kmsRegionEndpoint"`

		// Azure.
		KeyVaultURL string                       `json:"keyVaultUrl"`
		TenantID    string                       `json:"tenantId"`
		ClientID    string                       `json:"clientId"`
		AzureRegion azure.CloudAccountRegionEnum `json:"azureRegion"`

		// GCP.
		KeyRing     string              `json:"keyRing"`
		GCPLocation gcp.KMSLocationEnum `json:"gcpLocation"`

		// KMIP server.
		Address string `json:"address"`
		Port    int    `json:"port"`
	} `json:"keyVaultInfo"`
}

// ListKeyManagerFilter narrows the key managers returned by ListKeyManagers.
// The zero value matches every key manager.
type ListKeyManagerFilter struct {
	// NamePrefix matches key managers whose name starts with it.
	NamePrefix string

	// Types matches key managers of the KMS types. All types when empty.
	Types []Type

	// ProvisioningMethods matches key managers provisioned by any of the
	// methods. All methods when empty.
	ProvisioningMethods []ProvisioningMethod
}

// ListKeyManagers returns the KMS key managers matching the filter.
func ListKeyManagers(ctx context.Context, gql *graphql.Client, filter ListKeyManagerFilter) ([]KeyManager, error) {
	gql.Log().Print(log.Trace)

	var cursor string
	var keyManagers []KeyManager
	for {
		query := kmsInstancesQuery
		buf, err := gql.Request(ctx, query, struct {
			First        int                  `json:"first"`
			After        string               `json:"after,omitempty"`
			Name         string               `json:"kmsName,omitempty"`
			TypeFilter   []Type               `json:"kmsTypeFilter,omitempty"`
			MethodFilter []ProvisioningMethod `json:"provisioningMethodFilter,omitempty"`
		}{
			First:        100,
			After:        cursor,
			Name:         filter.NamePrefix,
			TypeFilter:   filter.Types,
			MethodFilter: filter.ProvisioningMethods,
		})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []KeyManager `json:"nodes"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, graphql.UnmarshalError(query, err)
		}

		keyManagers = append(keyManagers, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return keyManagers, nil
}

// DeleteKeyManager deletes the key manager with the specified ID. The same
// mutation deletes key managers of every KMS type.
func DeleteKeyManager(ctx context.Context, gql *graphql.Client, keyManagerID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deleteKmsInstanceQuery
	if _, err := gql.Request(ctx, query, struct {
		KeyManagerID uuid.UUID `json:"kmsKeyVaultId"`
	}{KeyManagerID: keyManagerID}); err != nil {
		return graphql.RequestError(query, err)
	}

	return nil
}

// VerifyKeyManagerName returns true if the name is available, i.e. not already
// used by another key manager. Key manager names are unique across KMS types.
func VerifyKeyManagerName(ctx context.Context, gql *graphql.Client, name string) (bool, error) {
	gql.Log().Print(log.Trace)

	query := verifyKmsNameQuery
	buf, err := gql.Request(ctx, query, struct {
		Name string `json:"kmsName"`
	}{Name: name})
	if err != nil {
		return false, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				IsVerified bool `json:"isVerified"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return false, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.IsVerified, nil
}

// ValidateKeyManagerConnection checks whether RSC can reach the key management
// service of the registered key manager. It returns the error reported by RSC,
// or an empty string when the connection is healthy.
//
// The mutation accepts the ID of a key manager of any KMS type, but RSC
// currently only implements connection validation for Azure Key Vault. For all
// other key manager types, RSC returns an error.
func ValidateKeyManagerConnection(ctx context.Context, gql *graphql.Client, keyManagerID uuid.UUID) (string, error) {
	gql.Log().Print(log.Trace)

	query := validateKmsConnectionQuery
	buf, err := gql.Request(ctx, query, struct {
		KeyManagerID uuid.UUID `json:"kmsKeyVaultId"`
	}{KeyManagerID: keyManagerID})
	if err != nil {
		return "", graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				ErrorString string `json:"errorString"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.ErrorString, nil
}
