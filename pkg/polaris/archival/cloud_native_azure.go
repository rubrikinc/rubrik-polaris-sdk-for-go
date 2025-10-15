// Copyright 2024 Rubrik, Inc.
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

package archival

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/archival"
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AzureTargetMappingByID returns the Azure target mapping with the specified
// ID. If no target mapping with the specified ID is found, graphql.ErrNotFound
// is returned.
func (a API) AzureTargetMappingByID(ctx context.Context, id uuid.UUID) (archival.AzureTargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []archival.ListTargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_ID",
		Text:  id.String(),
	}}
	targets, err := archival.ListTargetMappings[archival.AzureTargetMapping](ctx, a.client, filter)
	if err != nil {
		return archival.AzureTargetMapping{}, fmt.Errorf("failed to get target mappings: %s", err)
	}

	for _, target := range targets {
		if target.ID == id {
			return target, nil
		}
	}

	return archival.AzureTargetMapping{}, fmt.Errorf("target mapping for %q %w", id, graphql.ErrNotFound)
}

// AzureTargetMappingByName returns the Azure target mapping with the specified
// name. If no target mapping with the specified name is found,
// graphql.ErrNotFound is returned.
func (a API) AzureTargetMappingByName(ctx context.Context, name string) (archival.AzureTargetMapping, error) {
	a.log.Print(log.Trace)

	targets, err := a.AzureTargetMappings(ctx, name)
	if err != nil {
		return archival.AzureTargetMapping{}, err
	}

	name = strings.ToLower(name)
	for _, target := range targets {
		if strings.ToLower(target.Name) == name {
			return target, nil
		}
	}

	return archival.AzureTargetMapping{}, fmt.Errorf("target mapping for %q %w", name, graphql.ErrNotFound)
}

// AzureTargetMappings returns all Azure target mappings matching the specified
// name filter. The name filter can be used to search for prefixes of a name.
// If the name filter is empty, it will match all names.
func (a API) AzureTargetMappings(ctx context.Context, nameFilter string) ([]archival.AzureTargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []archival.ListTargetMappingFilter{{
		Field: "NAME",
		Text:  nameFilter,
	}}
	targets, err := archival.ListTargetMappings[archival.AzureTargetMapping](ctx, a.client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get target mappings: %s", err)
	}

	return targets, nil
}

// CreateAzureStorageSetting creates a cloud native archival location.
func (a API) CreateAzureStorageSetting(ctx context.Context, createParams archival.CreateAzureStorageSettingParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if createParams.NativeID == uuid.Nil {
		var err error
		if createParams.NativeID, err = lookupAzureNativeID(ctx, a.client, createParams.CloudAccountID); err != nil {
			return uuid.Nil, fmt.Errorf("failed to lookup azure native id: %s", err)
		}
	}

	if createParams.LocTemplate == "" {
		createParams.LocTemplate = "SPECIFIC_REGION"
		if createParams.StorageAccountRegion.Region == azure.RegionUnknown {
			createParams.LocTemplate = "SOURCE_REGION"
		}
	}

	targetMappingID, err := archival.CreateCloudNativeStorageSetting[archival.CreateAzureStorageSettingResult](ctx, a.client, createParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create cloud native storage setting: %s", err)
	}

	return targetMappingID, nil
}

// UpdateAzureStorageSetting updates the cloud native archival location with the
// specified target mapping ID.
func (a API) UpdateAzureStorageSetting(ctx context.Context, targetMappingID uuid.UUID, updateParams archival.UpdateAzureStorageSettingParams) error {
	a.log.Print(log.Trace)

	targetMapping, err := a.AzureTargetMappingByID(ctx, targetMappingID)
	if err != nil {
		return fmt.Errorf("failed to get azure target mapping: %s", err)
	}

	// All fields, except customer managed keys, are required.
	cloudNativeCompanion := targetMapping.TargetTemplate.CloudNativeCompanion
	if updateParams.Name == "" {
		updateParams.Name = targetMapping.Name
	}
	if updateParams.StorageTier == "" {
		updateParams.StorageTier = cloudNativeCompanion.StorageTier
	}
	if updateParams.StorageAccountTags.TagList == nil {
		updateParams.StorageAccountTags.TagList = cloudNativeCompanion.StorageAccountTags
	}

	if err := archival.UpdateCloudNativeStorageSetting[archival.UpdateAzureStorageSettingResult](ctx, a.client, targetMappingID, updateParams); err != nil {
		return fmt.Errorf("failed to update cloud native storage setting: %s", err)
	}

	return nil
}

// lookupAzureNativeID returns the native ID of the Azure cloud account with the
// specified cloud account ID.
func lookupAzureNativeID(ctx context.Context, client *graphql.Client, cloudAccountID uuid.UUID) (uuid.UUID, error) {
	tenants, err := gqlazure.Wrap(client).CloudAccountTenants(ctx, core.FeatureAll, true)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get azure cloud account tenants: %s", err)
	}

	for _, tenant := range tenants {
		for _, account := range tenant.Accounts {
			if account.ID == cloudAccountID {
				return account.NativeID, nil
			}
		}
	}
	return uuid.Nil, fmt.Errorf("azure cloud account %q: %w", cloudAccountID, graphql.ErrNotFound)
}
