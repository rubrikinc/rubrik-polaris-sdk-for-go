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
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CreateAzureKeyManagerParams holds the parameters for creating a cloud account
// provisioned Azure Key Vault key manager. RSC resolves the service principal
// credentials from the cloud account at call time, so no client secret is
// stored.
//
// Note that onboarding the subscription with the CLOUD_NATIVE_UEM_KEY_MANAGEMENT
// feature only grants RSC permission to read the vault. The key operations are
// granted separately by assigning RSC the Key Vault Crypto User role on the
// vault.
type CreateAzureKeyManagerParams struct {
	// CloudAccountID is the RSC cloud account ID of the Azure subscription
	// holding the key vault.
	CloudAccountID string `json:"cloudAccountId"`

	// Name is the name of the key manager, unique across all key managers.
	Name string `json:"instanceName"`

	// Region is the Azure region of the key vault.
	Region azure.Region `json:"region"`

	// VaultName is the name of the key vault, not its URL. RSC derives the URL
	// from the name and the cloud account's Azure environment.
	VaultName string `json:"vaultName"`

	Description string `json:"description,omitempty"`
}

// CreateAzureKeyManager creates a cloud account provisioned Azure Key Vault key
// manager and returns its ID.
func CreateAzureKeyManager(ctx context.Context, gql *graphql.Client, params CreateAzureKeyManagerParams) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	query := addAzureKeyVaultWithCloudAccountQuery
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

// UpdateAzureKeyManagerParams holds the parameters for updating a cloud account
// provisioned Azure Key Vault key manager. An empty field leaves the current
// value unchanged. The cloud account, the region and the vault name cannot be
// changed.
type UpdateAzureKeyManagerParams struct {
	Name        string `json:"instanceName,omitempty"`
	Description string `json:"description,omitempty"`
}

// UpdateAzureKeyManager updates the cloud account provisioned Azure Key Vault
// key manager with the specified ID.
func UpdateAzureKeyManager(ctx context.Context, gql *graphql.Client, keyManagerID uuid.UUID, params UpdateAzureKeyManagerParams) error {
	gql.Log().Print(log.Trace)

	query := updateAzureKeyVaultWithCloudAccountQuery
	if _, err := gql.Request(ctx, query, struct {
		KeyManagerID uuid.UUID `json:"kmsKeyVaultId"`
		UpdateAzureKeyManagerParams
	}{KeyManagerID: keyManagerID, UpdateAzureKeyManagerParams: params}); err != nil {
		return graphql.RequestError(query, err)
	}

	return nil
}

// CreateAzureSecureAuthKeyManagerParams holds the parameters for creating an
// Azure Key Vault key manager authenticating with a managed identity rather
// than with credentials. Use AzureAppID to obtain the application the identity
// belongs to.
type CreateAzureSecureAuthKeyManagerParams struct {
	ClientID    string `json:"clientId"`
	TenantID    string `json:"tenantId"`
	VaultURL    string `json:"vaultUrl"`
	Name        string `json:"instanceName"`
	Description string `json:"description,omitempty"`
}

// CreateAzureSecureAuthKeyManager creates an Azure Key Vault key manager
// authenticating with a managed identity and returns its ID.
func CreateAzureSecureAuthKeyManager(ctx context.Context, gql *graphql.Client, params CreateAzureSecureAuthKeyManagerParams) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	query := addSecureAuthAzureKeyVaultQuery
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

// AzureAppID returns the ID of the Azure AD application RSC uses for managed
// identity authentication against Azure Key Vault, creating it if it does not
// already exist.
func AzureAppID(ctx context.Context, gql *graphql.Client) (string, error) {
	gql.Log().Print(log.Trace)

	query := createOrGetKmsAzureAppIdQuery
	buf, err := gql.Request(ctx, query, struct{}{})
	if err != nil {
		return "", graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				AppID string `json:"appId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.AppID, nil
}

// AzureKeyNamesInKeyVault returns the names of the keys in the key vault the
// key manager connects to.
func AzureKeyNamesInKeyVault(ctx context.Context, gql *graphql.Client, keyManagerID uuid.UUID) ([]string, error) {
	gql.Log().Print(log.Trace)

	query := keyNamesInAzureKeyVaultQuery
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

// AzureKeyVersion represents a single version of a key in an Azure key vault.
type AzureKeyVersion struct {
	Version      string     `json:"keyVersion"`
	State        string     `json:"keyState"`
	CreatedAt    *time.Time `json:"createdAt"`
	ExpirationAt *time.Time `json:"expirationAt"`
}

// AzureKeyVersionsInKeyVault returns the versions of the named key in the key
// vault the key manager connects to. Unlike AWS KMS keys, an Azure key is
// referred to by name and version, so the version is part of what identifies
// the key to an archival location.
func AzureKeyVersionsInKeyVault(ctx context.Context, gql *graphql.Client, keyManagerID uuid.UUID, keyName string) ([]AzureKeyVersion, error) {
	gql.Log().Print(log.Trace)

	query := keyVersionsForKeyNameInAzureKeyVaultQuery
	buf, err := gql.Request(ctx, query, struct {
		KeyManagerID uuid.UUID `json:"kmsKeyVaultId"`
		KeyName      string    `json:"keyName"`
	}{KeyManagerID: keyManagerID, KeyName: keyName})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []AzureKeyVersion `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// ValidateAzureKeyVaultConnection checks whether RSC can reach the named key
// vault using the credentials of the cloud account, before registering it as a
// key manager. It returns the error reported by RSC, or an empty string when
// the connection is healthy.
func ValidateAzureKeyVaultConnection(ctx context.Context, gql *graphql.Client, cloudAccountID, vaultName string) (string, error) {
	gql.Log().Print(log.Trace)

	query := validateCloudAccountAzureKeyVaultConnectionQuery
	buf, err := gql.Request(ctx, query, struct {
		CloudAccountID string `json:"cloudAccountId"`
		VaultName      string `json:"vaultName"`
	}{CloudAccountID: cloudAccountID, VaultName: vaultName})
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

// ValidateAzureSecureAuthKeyVaultConnection checks whether RSC can reach the
// key vault at the URL using a managed identity, before registering it as a key
// manager. It returns the error reported by RSC, or an empty string when the
// connection is healthy.
func ValidateAzureSecureAuthKeyVaultConnection(ctx context.Context, gql *graphql.Client, tenantID, vaultURL string) (string, error) {
	gql.Log().Print(log.Trace)

	query := validateSecureAuthAzureKeyVaultConnectionQuery
	buf, err := gql.Request(ctx, query, struct {
		TenantID string `json:"tenantId"`
		VaultURL string `json:"vaultUrl"`
	}{TenantID: tenantID, VaultURL: vaultURL})
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
