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
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/kms"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// KeyManagers returns the key managers matching the filter. Pass the zero value
// filter to return every key manager.
func (a API) KeyManagers(ctx context.Context, filter kms.ListKeyManagerFilter) ([]kms.KeyManager, error) {
	a.log.Print(log.Trace)

	keyManagers, err := kms.ListKeyManagers(ctx, a.client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get key managers: %s", err)
	}

	return keyManagers, nil
}

// KeyManagerByID returns the key manager with the specified ID. If no such key
// manager exists, graphql.ErrNotFound is returned.
func (a API) KeyManagerByID(ctx context.Context, keyManagerID uuid.UUID) (kms.KeyManager, error) {
	a.log.Print(log.Trace)

	// RSC has no query for a single key manager, so the full list is filtered
	// client side.
	keyManagers, err := kms.ListKeyManagers(ctx, a.client, kms.ListKeyManagerFilter{})
	if err != nil {
		return kms.KeyManager{}, fmt.Errorf("failed to get key manager: %s", err)
	}

	for _, keyManager := range keyManagers {
		if keyManager.KeyVaultInfo.ID == keyManagerID {
			return keyManager, nil
		}
	}

	return kms.KeyManager{}, fmt.Errorf("key manager %q %w", keyManagerID, graphql.ErrNotFound)
}

// KeyManagerByName returns the key manager with the specified name. Key manager
// names are unique across KMS types. If no such key manager exists,
// graphql.ErrNotFound is returned.
func (a API) KeyManagerByName(ctx context.Context, name string) (kms.KeyManager, error) {
	a.log.Print(log.Trace)

	// The name filter matches on prefix, so the result still needs to be
	// narrowed to an exact match.
	keyManagers, err := kms.ListKeyManagers(ctx, a.client, kms.ListKeyManagerFilter{NamePrefix: name})
	if err != nil {
		return kms.KeyManager{}, fmt.Errorf("failed to get key manager: %s", err)
	}

	for _, keyManager := range keyManagers {
		if keyManager.Name == name {
			return keyManager, nil
		}
	}

	return kms.KeyManager{}, fmt.Errorf("key manager %q %w", name, graphql.ErrNotFound)
}

// DeleteKeyManager deletes the key manager with the specified ID. RSC refuses
// to delete a key manager which still encrypts a data location.
func (a API) DeleteKeyManager(ctx context.Context, keyManagerID uuid.UUID) error {
	a.log.Print(log.Trace)

	if err := kms.DeleteKeyManager(ctx, a.client, keyManagerID); err != nil {
		return fmt.Errorf("failed to delete key manager: %s", err)
	}

	return nil
}

// ValidateKeyManagerConnection checks whether RSC can reach the key management
// service of the registered key manager. It returns the error reported by RSC,
// or an empty string when the connection is healthy.
//
// Note that RSC currently only implements connection validation for Azure Key
// Vault. For all other key manager types, an error is returned.
func (a API) ValidateKeyManagerConnection(ctx context.Context, keyManagerID uuid.UUID) (string, error) {
	a.log.Print(log.Trace)

	errorString, err := kms.ValidateKeyManagerConnection(ctx, a.client, keyManagerID)
	if err != nil {
		return "", fmt.Errorf("failed to validate key manager connection: %s", err)
	}

	return errorString, nil
}

// KeyManagerNameAvailable returns true if the name is not already used by
// another key manager.
func (a API) KeyManagerNameAvailable(ctx context.Context, name string) (bool, error) {
	a.log.Print(log.Trace)

	available, err := kms.VerifyKeyManagerName(ctx, a.client, name)
	if err != nil {
		return false, fmt.Errorf("failed to verify key manager name: %s", err)
	}

	return available, nil
}
