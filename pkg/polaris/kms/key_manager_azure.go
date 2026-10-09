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
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/kms"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AzureKeyManagers returns the Azure Key Vault key managers matching the filter. Any
// types in the filter are ignored.
func (a API) AzureKeyManagers(ctx context.Context, filter kms.ListKeyManagerFilter) ([]kms.KeyManager, error) {
	a.log.Print(log.Trace)

	filter.Types = []kms.Type{kms.TypeAzureKeyVault}
	keyManagers, err := kms.ListKeyManagers(ctx, a.client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure Key Vault key managers: %s", err)
	}

	return keyManagers, nil
}

// AzureKeyManagerByID returns the Azure Key Vault key manager with the specified
// ID. If no such key manager exists, graphql.ErrNotFound is returned. This
// includes the ID of a key manager of another KMS type.
func (a API) AzureKeyManagerByID(ctx context.Context, keyManagerID uuid.UUID) (kms.KeyManager, error) {
	a.log.Print(log.Trace)

	// RSC has no query for a single key manager, so the Azure Key Vault key managers
	// are filtered client side.
	keyManagers, err := kms.ListKeyManagers(ctx, a.client, kms.ListKeyManagerFilter{Types: []kms.Type{kms.TypeAzureKeyVault}})
	if err != nil {
		return kms.KeyManager{}, fmt.Errorf("failed to get Azure Key Vault key manager: %s", err)
	}

	for _, keyManager := range keyManagers {
		if keyManager.KeyVaultInfo.ID == keyManagerID {
			return keyManager, nil
		}
	}

	return kms.KeyManager{}, fmt.Errorf("key manager %q of type Azure Key Vault %w", keyManagerID, graphql.ErrNotFound)
}

// AzureKeyManagerByName returns the Azure Key Vault key manager with the specified
// name. If no such key manager exists, graphql.ErrNotFound is returned. This
// includes the name of a key manager of another KMS type.
func (a API) AzureKeyManagerByName(ctx context.Context, name string) (kms.KeyManager, error) {
	a.log.Print(log.Trace)

	// The name filter matches on prefix, so the result still needs to be
	// narrowed to an exact match.
	keyManagers, err := kms.ListKeyManagers(ctx, a.client, kms.ListKeyManagerFilter{
		NamePrefix: name,
		Types:      []kms.Type{kms.TypeAzureKeyVault},
	})
	if err != nil {
		return kms.KeyManager{}, fmt.Errorf("failed to get Azure Key Vault key manager: %s", err)
	}

	for _, keyManager := range keyManagers {
		if keyManager.Name == name {
			return keyManager, nil
		}
	}

	return kms.KeyManager{}, fmt.Errorf("key manager %q of type Azure Key Vault %w", name, graphql.ErrNotFound)
}

// CreateAzureKeyManager creates a cloud account provisioned Azure Key Vault key
// manager and returns its ID. The Azure subscription must already be onboarded
// with the CLOUD_NATIVE_UEM_KEY_MANAGEMENT feature.
//
// Note that the feature only grants RSC permission to read the vault. RSC must
// additionally be assigned the Key Vault Crypto User role on the vault for the
// key operations to succeed.
func (a API) CreateAzureKeyManager(ctx context.Context, params kms.CreateAzureKeyManagerParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	switch params.Region {
	case azure.RegionUnknown:
		return uuid.Nil, errors.New("region is required")
	case azure.RegionSource:
		return uuid.Nil, errors.New("source region is not supported")
	}

	keyManagerID, err := kms.CreateAzureKeyManager(ctx, a.client, params)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create Azure key manager: %s", err)
	}

	return keyManagerID, nil
}

// UpdateAzureKeyManager updates the cloud account provisioned Azure Key Vault
// key manager with the specified ID.
func (a API) UpdateAzureKeyManager(ctx context.Context, keyManagerID uuid.UUID, params kms.UpdateAzureKeyManagerParams) error {
	a.log.Print(log.Trace)

	if err := kms.UpdateAzureKeyManager(ctx, a.client, keyManagerID, params); err != nil {
		return fmt.Errorf("failed to update Azure key manager: %s", err)
	}

	return nil
}

// CreateAzureSecureAuthKeyManager creates an Azure Key Vault key manager
// authenticating with a managed identity and returns its ID. Use AzureAppID to
// obtain the application the identity belongs to.
func (a API) CreateAzureSecureAuthKeyManager(ctx context.Context, createParams kms.CreateAzureSecureAuthKeyManagerParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	keyManagerID, err := kms.CreateAzureSecureAuthKeyManager(ctx, a.client, createParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create Azure secure auth key manager: %s", err)
	}

	return keyManagerID, nil
}

// AzureAppID returns the ID of the Azure AD application RSC uses for managed
// identity authentication against Azure Key Vault, creating it if it does not
// already exist.
func (a API) AzureAppID(ctx context.Context) (string, error) {
	a.log.Print(log.Trace)

	appID, err := kms.AzureAppID(ctx, a.client)
	if err != nil {
		return "", fmt.Errorf("failed to get Azure app ID: %s", err)
	}

	return appID, nil
}

// AzureKeyNames returns the names of the keys in the key vault the key manager
// connects to.
func (a API) AzureKeyNames(ctx context.Context, keyManagerID uuid.UUID) ([]string, error) {
	a.log.Print(log.Trace)

	keyNames, err := kms.AzureKeyNamesInKeyVault(ctx, a.client, keyManagerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure key vault key names: %s", err)
	}

	return keyNames, nil
}

// AzureKeyVersions returns the versions of the named key in the key vault the
// key manager connects to. An Azure key is identified by name and version, so
// both are needed to select the encryption key of an archival location.
func (a API) AzureKeyVersions(ctx context.Context, keyManagerID uuid.UUID, keyName string) ([]kms.AzureKeyVersion, error) {
	a.log.Print(log.Trace)

	keyVersions, err := kms.AzureKeyVersionsInKeyVault(ctx, a.client, keyManagerID, keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure key vault key versions: %s", err)
	}

	return keyVersions, nil
}

// ValidateAzureKeyVaultConnection checks whether RSC can reach the named key
// vault using the credentials of the cloud account, before registering it as a
// key manager. It returns the error reported by RSC, or an empty string when
// the connection is healthy.
func (a API) ValidateAzureKeyVaultConnection(ctx context.Context, cloudAccountID, vaultName string) (string, error) {
	a.log.Print(log.Trace)

	errorString, err := kms.ValidateAzureKeyVaultConnection(ctx, a.client, cloudAccountID, vaultName)
	if err != nil {
		return "", fmt.Errorf("failed to validate Azure key vault connection: %s", err)
	}

	return errorString, nil
}

// ValidateAzureSecureAuthKeyVaultConnection checks whether RSC can reach the key
// vault at the URL using a managed identity, before registering it as a key
// manager. It returns the error reported by RSC, or an empty string when the
// connection is healthy.
func (a API) ValidateAzureSecureAuthKeyVaultConnection(ctx context.Context, tenantID, vaultURL string) (string, error) {
	a.log.Print(log.Trace)

	errorString, err := kms.ValidateAzureSecureAuthKeyVaultConnection(ctx, a.client, tenantID, vaultURL)
	if err != nil {
		return "", fmt.Errorf("failed to validate Azure secure auth key vault connection: %s", err)
	}

	return errorString, nil
}
