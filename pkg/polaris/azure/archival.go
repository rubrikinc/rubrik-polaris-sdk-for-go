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

package azure

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// TargetMapping represents an Azure cloud archival location.
type TargetMapping struct {
	ID                   uuid.UUID
	Name                 string
	ArchivalGroup        string
	ArchivalTarget       string
	ConnectionStatus     string
	ContainerName        string
	StorageAccountName   string
	StorageAccountRegion string
	StorageAccountTags   map[string]string
	LocTemplate          string
	Redundancy           string
	StorageTier          string
	NativeID             uuid.UUID
	CustomerKeys         []CustomerKey
}

// CustomerKey represents the customer managed key information required for
// encryption of Azure storage.
type CustomerKey struct {
	Name      string
	Region    string
	VaultName string
}

// TargetMappingByID returns the Azure target mapping with the specified ID.
// If no target mapping with the specified ID is found, graphql.ErrNotFound is
// returned.
func (a API) TargetMappingByID(ctx context.Context, id uuid.UUID) (TargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []azure.TargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_ID",
		Text:  id.String(),
	}}
	targets, err := archival.ListTargetMappings[azure.TargetMapping](ctx, a.client, filter)
	if err != nil {
		return TargetMapping{}, fmt.Errorf("failed to get target mappings: %s", err)
	}

	for _, target := range targets {
		if target.ID == id {
			return toTargetMapping(target), nil
		}
	}

	return TargetMapping{}, fmt.Errorf("target mapping for %q %w", id, graphql.ErrNotFound)
}

// TargetMappingByName returns the Azure target mapping with the specified name.
// If no target mapping with the specified ID is found, graphql.ErrNotFound is
// returned.
func (a API) TargetMappingByName(ctx context.Context, name string) (TargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []azure.TargetMappingFilter{{
		Field: "NAME",
		Text:  name,
	}}
	targets, err := archival.ListTargetMappings[azure.TargetMapping](ctx, a.client, filter)
	if err != nil {
		return TargetMapping{}, fmt.Errorf("failed to get target mappings: %s", err)
	}

	for _, target := range targets {
		if target.Name == name {
			return toTargetMapping(target), nil
		}
	}

	return TargetMapping{}, fmt.Errorf("target mapping for %q %w", name, graphql.ErrNotFound)
}

// TargetMappings returns all Azure target mappings that match the specified
// archival group and name filter. The name filter can be used to search for
// prefixes of a name. If the name filter is empty, is will match all names.
// In RSC cloud, archival locations are also referred to as target mappings.
func (a API) TargetMappings(ctx context.Context, nameFilter string) ([]TargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []azure.TargetMappingFilter{{
		Field: "NAME",
		Text:  nameFilter,
	}}
	targets, err := archival.ListTargetMappings[azure.TargetMapping](ctx, a.client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get target mappings: %s", err)
	}

	targetMappings := make([]TargetMapping, 0, len(targets))
	for _, target := range targets {
		targetMappings = append(targetMappings, toTargetMapping(target))
	}

	return targetMappings, nil
}

// DeleteTargetMapping deletes the target mapping with the specified ID. In RSC
// cloud archival locations are also referred to as target mappings.
func (a API) DeleteTargetMapping(ctx context.Context, id uuid.UUID) error {
	a.log.Print(log.Trace)

	return archival.DeleteTargetMapping(ctx, a.client, id)
}

// CreateStorageSetting creates a cloud native archival location. The storage
// account region, the storage account tags, and the customer managed keys are
// optional.
func (a API) CreateStorageSetting(ctx context.Context, id IdentityFunc, name, redundancy, storageTier, storageAccountName, storageAccountRegion string, storageAccountTags map[string]string, customerKeys []CustomerKey) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	cloudAccountID, err := a.toCloudAccountID(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}

	saRegion := azure.RegionUnknown
	if storageAccountRegion != "" && storageAccountRegion != "UNKNOWN_AZURE_REGION" {
		saRegion = azure.ParseRegionNoValidation(storageAccountRegion)
	}

	var tags *struct {
		TagList []azure.Tag `json:"tagList"`
	}
	if len(storageAccountTags) > 0 {
		tags = &struct {
			TagList []azure.Tag `json:"tagList"`
		}{TagList: make([]azure.Tag, 0, len(storageAccountTags))}
		for key, value := range storageAccountTags {
			tags.TagList = append(tags.TagList, azure.Tag{Key: key, Value: value})
		}
	}

	locTemplate := "SPECIFIC_REGION"
	if saRegion == azure.RegionUnknown {
		locTemplate = "SOURCE_REGION"
	}

	keys := make([]azure.CustomerKey, 0, len(customerKeys))
	for _, key := range customerKeys {
		keys = append(keys, azure.CustomerKey{
			KeyName:      key.Name,
			KeyVaultName: key.VaultName,
			Region:       azure.ParseRegionNoValidation(key.Region),
		})
	}

	targetMappingID, err := archival.CreateCloudNativeStorageSetting[azure.StorageSettingCreateResult](ctx, a.client,
		cloudAccountID, azure.StorageSettingCreateParams{
			LocTemplate:          locTemplate,
			Name:                 name,
			Redundancy:           redundancy,
			StorageTier:          storageTier,
			NativeID:             cloudAccountID,
			StorageAccountName:   storageAccountName,
			StorageAccountRegion: saRegion,
			StorageAccountTags:   tags,
			CMKInfo:              keys,
		})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create cloud native storage setting: %s", err)
	}

	return targetMappingID, nil
}

// UpdateStorageSetting updates the cloud native archival location with the
// specified ID. Note that not all properties can be updated, only the name,
// storage tier, storage account tags and customer managed keys.
func (a API) UpdateStorageSetting(ctx context.Context, targetMappingID uuid.UUID, name, storageTier string, storageAccountTags map[string]string, customerKeys []CustomerKey) error {
	a.log.Print(log.Trace)

	tags := make([]azure.Tag, 0, len(storageAccountTags))
	for key, value := range storageAccountTags {
		tags = append(tags, azure.Tag{Key: key, Value: value})
	}

	keys := make([]azure.CustomerKey, 0, len(customerKeys))
	for _, key := range customerKeys {
		keys = append(keys, azure.CustomerKey{
			KeyName:      key.Name,
			KeyVaultName: key.VaultName,
			Region:       azure.ParseRegionNoValidation(key.Region),
		})
	}

	err := archival.UpdateCloudNativeStorageSetting[azure.StorageSettingUpdateResult](ctx, a.client, targetMappingID,
		azure.StorageSettingUpdateParams{
			Name:        name,
			StorageTier: storageTier,
			StorageAccountTags: struct {
				TagList []azure.Tag `json:"tagList"`
			}{TagList: tags},
			CMKInfo: keys,
		})
	if err != nil {
		return fmt.Errorf("failed to update cloud native storage setting: %s", err)
	}

	return nil
}

// toTargetMapping converts an aws.TargetMapping to a TargetMapping.
func toTargetMapping(target azure.TargetMapping) TargetMapping {
	tags := make(map[string]string, len(target.TargetTemplate.CloudNativeCompanion.StorageAccountTags))
	for _, tag := range target.TargetTemplate.CloudNativeCompanion.StorageAccountTags {
		tags[tag.Key] = tag.Value
	}

	region := ""
	if target.TargetTemplate.CloudNativeCompanion.StorageAccountRegion != azure.RegionUnknown {
		region = azure.FormatRegion(target.TargetTemplate.CloudNativeCompanion.StorageAccountRegion)
	}

	keys := make([]CustomerKey, 0, len(target.TargetTemplate.CloudNativeCompanion.CMKInfo))
	for _, key := range target.TargetTemplate.CloudNativeCompanion.CMKInfo {
		keys = append(keys, CustomerKey{
			Name:      key.KeyName,
			Region:    azure.FormatRegion(key.Region),
			VaultName: key.KeyVaultName,
		})
	}

	return TargetMapping{
		ID:                   target.ID,
		Name:                 target.Name,
		ArchivalGroup:        target.GroupType,
		ArchivalTarget:       target.TargetType,
		ConnectionStatus:     target.ConnectionStatus.Status,
		ContainerName:        target.TargetTemplate.ContainerNamePrefix,
		StorageAccountName:   target.TargetTemplate.StorageAccountName,
		StorageAccountRegion: region,
		StorageAccountTags:   tags,
		LocTemplate:          target.TargetTemplate.CloudNativeCompanion.LocTemplate,
		Redundancy:           target.TargetTemplate.CloudNativeCompanion.Redundancy,
		StorageTier:          target.TargetTemplate.CloudNativeCompanion.StorageTier,
		NativeID:             target.TargetTemplate.CloudNativeCompanion.NativeID,
		CustomerKeys:         keys,
	}
}
