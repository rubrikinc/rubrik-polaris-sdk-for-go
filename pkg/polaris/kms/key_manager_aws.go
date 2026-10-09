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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AWSKeyManagers returns the AWS KMS key managers matching the filter. Any
// types in the filter are ignored.
func (a API) AWSKeyManagers(ctx context.Context, filter kms.ListKeyManagerFilter) ([]kms.KeyManager, error) {
	a.log.Print(log.Trace)

	filter.Types = []kms.Type{kms.TypeAWSKMS}
	keyManagers, err := kms.ListKeyManagers(ctx, a.client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS KMS key managers: %s", err)
	}

	return keyManagers, nil
}

// AWSKeyManagerByID returns the AWS KMS key manager with the specified
// ID. If no such key manager exists, graphql.ErrNotFound is returned. This
// includes the ID of a key manager of another KMS type.
func (a API) AWSKeyManagerByID(ctx context.Context, keyManagerID uuid.UUID) (kms.KeyManager, error) {
	a.log.Print(log.Trace)

	// RSC has no query for a single key manager, so the AWS KMS key managers
	// are filtered client side.
	keyManagers, err := kms.ListKeyManagers(ctx, a.client, kms.ListKeyManagerFilter{Types: []kms.Type{kms.TypeAWSKMS}})
	if err != nil {
		return kms.KeyManager{}, fmt.Errorf("failed to get AWS KMS key manager: %s", err)
	}

	for _, keyManager := range keyManagers {
		if keyManager.KeyVaultInfo.ID == keyManagerID {
			return keyManager, nil
		}
	}

	return kms.KeyManager{}, fmt.Errorf("key manager %q of type AWS KMS %w", keyManagerID, graphql.ErrNotFound)
}

// AWSKeyManagerByName returns the AWS KMS key manager with the specified
// name. If no such key manager exists, graphql.ErrNotFound is returned. This
// includes the name of a key manager of another KMS type.
func (a API) AWSKeyManagerByName(ctx context.Context, name string) (kms.KeyManager, error) {
	a.log.Print(log.Trace)

	// The name filter matches on prefix, so the result still needs to be
	// narrowed to an exact match.
	keyManagers, err := kms.ListKeyManagers(ctx, a.client, kms.ListKeyManagerFilter{
		NamePrefix: name,
		Types:      []kms.Type{kms.TypeAWSKMS},
	})
	if err != nil {
		return kms.KeyManager{}, fmt.Errorf("failed to get AWS KMS key manager: %s", err)
	}

	for _, keyManager := range keyManagers {
		if keyManager.Name == name {
			return keyManager, nil
		}
	}

	return kms.KeyManager{}, fmt.Errorf("key manager %q of type AWS KMS %w", name, graphql.ErrNotFound)
}

// CreateAWSKeyManager creates a cloud account provisioned AWS KMS key manager
// and returns its ID. The AWS account must already be onboarded with the
// CLOUD_NATIVE_UEM_KEY_MANAGEMENT feature.
func (a API) CreateAWSKeyManager(ctx context.Context, params kms.CreateAWSKeyManagerParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	switch params.Region {
	case aws.RegionUnknown:
		return uuid.Nil, errors.New("region is required")
	case aws.RegionSource:
		return uuid.Nil, errors.New("source region is not supported")
	}

	keyManagerID, err := kms.CreateAWSKeyManager(ctx, a.client, params)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create AWS key manager: %s", err)
	}

	return keyManagerID, nil
}

// UpdateAWSKeyManager updates the cloud account provisioned AWS KMS key manager
// with the specified ID.
func (a API) UpdateAWSKeyManager(ctx context.Context, keyManagerID uuid.UUID, params kms.UpdateAWSKeyManagerParams) error {
	a.log.Print(log.Trace)

	if err := kms.UpdateAWSKeyManager(ctx, a.client, keyManagerID, params); err != nil {
		return fmt.Errorf("failed to update AWS key manager: %s", err)
	}

	return nil
}

// AWSKeyIDs returns the IDs of the AWS KMS keys reachable through the key
// manager. These are AWS KMS key IDs, and are the value to pass together with
// the key manager ID when selecting the encryption key of an archival location.
func (a API) AWSKeyIDs(ctx context.Context, keyManagerID uuid.UUID) ([]string, error) {
	a.log.Print(log.Trace)

	keyIDs, err := kms.AWSKeyIDsInKMS(ctx, a.client, keyManagerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS KMS key IDs: %s", err)
	}

	return keyIDs, nil
}

// AWSRegions returns the AWS regions which can host an AWS KMS key manager.
func (a API) AWSRegions(ctx context.Context) ([]string, error) {
	a.log.Print(log.Trace)

	regions, err := kms.AWSKMSRegions(ctx, a.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS KMS regions: %s", err)
	}

	return regions, nil
}
