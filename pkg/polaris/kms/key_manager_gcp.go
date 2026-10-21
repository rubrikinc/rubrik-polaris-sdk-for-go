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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// GCPKeyManagers returns the GCP Cloud KMS key managers matching the filter.
// Any types in the filter are ignored.
func (a API) GCPKeyManagers(ctx context.Context, filter kms.ListKeyManagerFilter) ([]kms.KeyManager, error) {
	a.log.Print(log.Trace)

	filter.Types = []kms.Type{kms.TypeGCPKMS}
	keyManagers, err := kms.ListKeyManagers(ctx, a.client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCP KMS key managers: %s", err)
	}

	return keyManagers, nil
}

// GCPKeyManagerByID returns the GCP Cloud KMS key manager with the specified
// ID. If no such key manager exists, graphql.ErrNotFound is returned. This
// includes the ID of a key manager of another KMS type.
func (a API) GCPKeyManagerByID(ctx context.Context, keyManagerID uuid.UUID) (kms.KeyManager, error) {
	a.log.Print(log.Trace)

	// RSC has no query for a single key manager, so the GCP KMS key managers
	// are filtered client side.
	keyManagers, err := kms.ListKeyManagers(ctx, a.client, kms.ListKeyManagerFilter{Types: []kms.Type{kms.TypeGCPKMS}})
	if err != nil {
		return kms.KeyManager{}, fmt.Errorf("failed to get GCP KMS key manager: %s", err)
	}

	for _, keyManager := range keyManagers {
		if keyManager.KeyVaultInfo.ID == keyManagerID {
			return keyManager, nil
		}
	}

	return kms.KeyManager{}, fmt.Errorf("key manager %q of type GCP KMS %w", keyManagerID, graphql.ErrNotFound)
}

// GCPKeyManagerByName returns the GCP Cloud KMS key manager with the specified
// name. If no such key manager exists, graphql.ErrNotFound is returned. This
// includes the name of a key manager of another KMS type.
func (a API) GCPKeyManagerByName(ctx context.Context, name string) (kms.KeyManager, error) {
	a.log.Print(log.Trace)

	// The name filter matches on prefix, so the result still needs to be
	// narrowed to an exact match.
	keyManagers, err := kms.ListKeyManagers(ctx, a.client, kms.ListKeyManagerFilter{
		NamePrefix: name,
		Types:      []kms.Type{kms.TypeGCPKMS},
	})
	if err != nil {
		return kms.KeyManager{}, fmt.Errorf("failed to get GCP KMS key manager: %s", err)
	}

	for _, keyManager := range keyManagers {
		if keyManager.Name == name {
			return keyManager, nil
		}
	}

	return kms.KeyManager{}, fmt.Errorf("key manager %q of type GCP KMS %w", name, graphql.ErrNotFound)
}

// CreateGCPKeyManager creates a cloud account provisioned GCP Cloud KMS key
// manager and returns its ID. The GCP project must already be onboarded with
// the CLOUD_NATIVE_UEM_KEY_MANAGEMENT feature.
func (a API) CreateGCPKeyManager(ctx context.Context, params kms.CreateGCPKeyManagerParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if params.Location.Region == gcp.RegionUnknown {
		return uuid.Nil, errors.New("location is required")
	}
	if params.Location.ToKMSLocationEnumPtr() == nil {
		return uuid.Nil, fmt.Errorf("location %q is not a GCP KMS location", params.Location.Name())
	}

	keyManagerID, err := kms.CreateGCPKeyManager(ctx, a.client, params)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create GCP key manager: %s", err)
	}

	return keyManagerID, nil
}

// UpdateGCPKeyManager updates the cloud account provisioned GCP Cloud KMS key
// manager with the specified ID.
func (a API) UpdateGCPKeyManager(ctx context.Context, keyManagerID uuid.UUID, params kms.UpdateGCPKeyManagerParams) error {
	a.log.Print(log.Trace)

	if err := kms.UpdateGCPKeyManager(ctx, a.client, keyManagerID, params); err != nil {
		return fmt.Errorf("failed to update GCP key manager: %s", err)
	}

	return nil
}

// GCPCryptoKeys returns the crypto keys in the key ring of the key manager.
// Only keys with the ENCRYPT_DECRYPT purpose can be used for encryption.
func (a API) GCPCryptoKeys(ctx context.Context, keyManagerID uuid.UUID) ([]kms.GCPCryptoKey, error) {
	a.log.Print(log.Trace)

	keys, err := kms.GCPCryptoKeysInKMS(ctx, a.client, keyManagerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCP KMS crypto keys: %s", err)
	}

	return keys, nil
}

// GCPCryptoKeyVersions returns the versions of the named crypto key in the key
// ring of the key manager. A GCP crypto key is identified by name and version,
// so both are needed to select the encryption key of an archival location.
func (a API) GCPCryptoKeyVersions(ctx context.Context, keyManagerID uuid.UUID, cryptoKeyName string) ([]kms.GCPCryptoKeyVersion, error) {
	a.log.Print(log.Trace)

	keyVersions, err := kms.GCPCryptoKeyVersionsInKMS(ctx, a.client, keyManagerID, cryptoKeyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCP KMS crypto key versions: %s", err)
	}

	return keyVersions, nil
}

// GCPKeyRings returns the names of the key rings in the Cloud KMS location of
// the GCP project with the specified RSC cloud account ID.
func (a API) GCPKeyRings(ctx context.Context, cloudAccountID uuid.UUID, location gcp.Region) ([]string, error) {
	a.log.Print(log.Trace)

	if location == gcp.RegionUnknown {
		return nil, errors.New("location is required")
	}
	if location.ToKMSLocationEnumPtr() == nil {
		return nil, fmt.Errorf("location %q is not a GCP KMS location", location.Name())
	}

	keyRings, err := kms.GCPKeyRingsInProject(ctx, a.client, cloudAccountID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCP KMS key rings: %s", err)
	}

	return keyRings, nil
}
