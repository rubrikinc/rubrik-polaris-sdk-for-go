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

package archival

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Rubrik Cloud Vault, RCV, is Rubrik hosted archival storage. The cloud in the
// function names below is the destination cloud, the cloud hosting the vault,
// and not the cloud the archived workloads originate from.

// RCVAWSTargetMappingByID returns the RCV AWS target mapping with the specified
// ID. If no target mapping with the specified ID is found, graphql.ErrNotFound
// is returned.
func (a API) RCVAWSTargetMappingByID(ctx context.Context, id uuid.UUID) (archival.RCVAWSTargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []archival.ListTargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_ID",
		Text:  id.String(),
	}}
	targets, err := archival.ListTargetMappings[archival.RCVAWSTargetMapping](ctx, a.client, filter)
	if err != nil {
		return archival.RCVAWSTargetMapping{}, fmt.Errorf("failed to get target mappings: %s", err)
	}

	for _, target := range targets {
		if target.ID == id {
			return target, nil
		}
	}

	return archival.RCVAWSTargetMapping{}, fmt.Errorf("target mapping for %q %w", id, graphql.ErrNotFound)
}

// RCVAWSTargetMappingByName returns the RCV AWS target mapping with the
// specified name. If no target mapping with the specified name is found,
// graphql.ErrNotFound is returned.
func (a API) RCVAWSTargetMappingByName(ctx context.Context, name string) (archival.RCVAWSTargetMapping, error) {
	a.log.Print(log.Trace)

	targets, err := a.RCVAWSTargetMappings(ctx, name)
	if err != nil {
		return archival.RCVAWSTargetMapping{}, err
	}

	name = strings.ToLower(name)
	for _, target := range targets {
		if strings.ToLower(target.Name) == name {
			return target, nil
		}
	}

	return archival.RCVAWSTargetMapping{}, fmt.Errorf("target mapping for %q %w", name, graphql.ErrNotFound)
}

// RCVAWSTargetMappings returns all RCV AWS target mappings matching the
// specified name filter. The name filter can be used to search for prefixes of
// a name. If the name filter is empty, it will match all names.
func (a API) RCVAWSTargetMappings(ctx context.Context, nameFilter string) ([]archival.RCVAWSTargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []archival.ListTargetMappingFilter{{
		Field: "NAME",
		Text:  nameFilter,
	}}
	targets, err := archival.ListTargetMappings[archival.RCVAWSTargetMapping](ctx, a.client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get target mappings: %s", err)
	}

	return targets, nil
}

// CreateRCVAWSStorageSetting creates an RCV AWS cloud native archival location.
func (a API) CreateRCVAWSStorageSetting(ctx context.Context, createParams archival.CreateRCVAWSStorageSettingParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if createParams.LocTemplate == "" {
		createParams.LocTemplate = archival.LocTemplateSpecificRegion
		if createParams.Region.Region == aws.RegionUnknown {
			createParams.LocTemplate = archival.LocTemplateSourceRegion
		}
	}

	targetMappingID, err := archival.CreateCloudNativeStorageSetting[archival.CreateRCVAWSStorageSettingResult](ctx, a.client, createParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create cloud native storage setting: %s", err)
	}

	return targetMappingID, nil
}

// UpdateRCVAWSStorageSetting updates the RCV AWS cloud native archival location
// with the specified target mapping ID. The RSC API only supports updating the
// name and the tier.
func (a API) UpdateRCVAWSStorageSetting(ctx context.Context, targetMappingID uuid.UUID, updateParams archival.UpdateRCVAWSStorageSettingParams) error {
	a.log.Print(log.Trace)

	// The name is required by the API, so read it back when the caller only
	// wants to change the tier.
	if updateParams.Name == "" {
		targetMapping, err := a.RCVAWSTargetMappingByID(ctx, targetMappingID)
		if err != nil {
			return fmt.Errorf("failed to get rcv aws target mapping: %s", err)
		}
		updateParams.Name = targetMapping.Name
	}

	if err := archival.UpdateCloudNativeStorageSetting[archival.UpdateRCVAWSStorageSettingResult](ctx, a.client, targetMappingID, updateParams); err != nil {
		return fmt.Errorf("failed to update cloud native storage setting: %s", err)
	}

	return nil
}

// RCVAzureTargetMappingByID returns the RCV Azure target mapping with the
// specified ID. If no target mapping with the specified ID is found,
// graphql.ErrNotFound is returned.
func (a API) RCVAzureTargetMappingByID(ctx context.Context, id uuid.UUID) (archival.RCVAzureTargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []archival.ListTargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_ID",
		Text:  id.String(),
	}}
	targets, err := archival.ListTargetMappings[archival.RCVAzureTargetMapping](ctx, a.client, filter)
	if err != nil {
		return archival.RCVAzureTargetMapping{}, fmt.Errorf("failed to get target mappings: %s", err)
	}

	for _, target := range targets {
		if target.ID == id {
			return target, nil
		}
	}

	return archival.RCVAzureTargetMapping{}, fmt.Errorf("target mapping for %q %w", id, graphql.ErrNotFound)
}

// RCVAzureTargetMappingByName returns the RCV Azure target mapping with the
// specified name. If no target mapping with the specified name is found,
// graphql.ErrNotFound is returned.
func (a API) RCVAzureTargetMappingByName(ctx context.Context, name string) (archival.RCVAzureTargetMapping, error) {
	a.log.Print(log.Trace)

	targets, err := a.RCVAzureTargetMappings(ctx, name)
	if err != nil {
		return archival.RCVAzureTargetMapping{}, err
	}

	name = strings.ToLower(name)
	for _, target := range targets {
		if strings.ToLower(target.Name) == name {
			return target, nil
		}
	}

	return archival.RCVAzureTargetMapping{}, fmt.Errorf("target mapping for %q %w", name, graphql.ErrNotFound)
}

// RCVAzureTargetMappings returns all RCV Azure target mappings matching the
// specified name filter. The name filter can be used to search for prefixes of
// a name. If the name filter is empty, it will match all names.
func (a API) RCVAzureTargetMappings(ctx context.Context, nameFilter string) ([]archival.RCVAzureTargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []archival.ListTargetMappingFilter{{
		Field: "NAME",
		Text:  nameFilter,
	}}
	targets, err := archival.ListTargetMappings[archival.RCVAzureTargetMapping](ctx, a.client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get target mappings: %s", err)
	}

	return targets, nil
}

// CreateRCVAzureStorageSetting creates an RCV Azure cloud native archival
// location.
//
// The RSC API takes both an Azure region and an RCV region. Both are derived
// from the Region field to keep the two from drifting apart. Any value in the
// RCVRegion field is overwritten.
func (a API) CreateRCVAzureStorageSetting(ctx context.Context, createParams archival.CreateRCVAzureStorageSettingParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if createParams.LocTemplate == "" {
		createParams.LocTemplate = archival.LocTemplateSpecificRegion
		if createParams.Region.Region == azure.RegionUnknown {
			createParams.LocTemplate = archival.LocTemplateSourceRegion
		}
	}
	createParams.RCVRegion = createParams.Region.Region.ToRCSRegionEnum()

	targetMappingID, err := archival.CreateCloudNativeStorageSetting[archival.CreateRCVAzureStorageSettingResult](ctx, a.client, createParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create cloud native storage setting: %s", err)
	}

	return targetMappingID, nil
}

// UpdateRCVAzureStorageSetting updates the RCV Azure cloud native archival
// location with the specified target mapping ID. The RSC API only supports
// updating the name, the tier and the redundancy.
func (a API) UpdateRCVAzureStorageSetting(ctx context.Context, targetMappingID uuid.UUID, updateParams archival.UpdateRCVAzureStorageSettingParams) error {
	a.log.Print(log.Trace)

	// The name is required by the API, so read it back when the caller only
	// wants to change the other fields.
	if updateParams.Name == "" {
		targetMapping, err := a.RCVAzureTargetMappingByID(ctx, targetMappingID)
		if err != nil {
			return fmt.Errorf("failed to get rcv azure target mapping: %s", err)
		}
		updateParams.Name = targetMapping.Name
	}

	if err := archival.UpdateCloudNativeStorageSetting[archival.UpdateRCVAzureStorageSettingResult](ctx, a.client, targetMappingID, updateParams); err != nil {
		return fmt.Errorf("failed to update cloud native storage setting: %s", err)
	}

	return nil
}

// RCVGCPTargetMappingByID returns the RCV GCP target mapping with the specified
// ID. If no target mapping with the specified ID is found, graphql.ErrNotFound
// is returned.
func (a API) RCVGCPTargetMappingByID(ctx context.Context, id uuid.UUID) (archival.RCVGCPTargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []archival.ListTargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_ID",
		Text:  id.String(),
	}}
	targets, err := archival.ListTargetMappings[archival.RCVGCPTargetMapping](ctx, a.client, filter)
	if err != nil {
		return archival.RCVGCPTargetMapping{}, fmt.Errorf("failed to get target mappings: %s", err)
	}

	for _, target := range targets {
		if target.ID == id {
			return target, nil
		}
	}

	return archival.RCVGCPTargetMapping{}, fmt.Errorf("target mapping for %q %w", id, graphql.ErrNotFound)
}

// rekeyPollInterval is the interval between checks of the rekey status when
// waiting for a rekey to complete.
const rekeyPollInterval = 10 * time.Second

// RekeyRCVGCPMasterKey rekeys the master key of the RCV GCP archival location
// with the specified target mapping ID, moving it to the specified GCP Cloud
// KMS key. The function returns once the rekey has completed. If the rekey
// fails for any of the group's archival locations, an error holding the error
// message reported by RSC is returned.
func (a API) RekeyRCVGCPMasterKey(ctx context.Context, targetMappingID uuid.UUID, params archival.RekeyArchivalGroupMasterKeyParams) error {
	a.log.Print(log.Trace)

	// The encryption details can only be filtered by name, so the name of the
	// target mapping is needed to find the rekey status.
	targetMapping, err := a.RCVGCPTargetMappingByID(ctx, targetMappingID)
	if err != nil {
		return fmt.Errorf("failed to get rcv gcp target mapping: %s", err)
	}

	if err := archival.RekeyArchivalGroupMasterKey(ctx, a.client, targetMappingID, params); err != nil {
		return fmt.Errorf("failed to rekey master key: %s", err)
	}

	if err := a.waitForRekey(ctx, targetMappingID, targetMapping.Name); err != nil {
		return fmt.Errorf("failed to wait for rekey: %s", err)
	}

	return nil
}

// RekeyRCVAWSMasterKey rekeys the master key of the RCV AWS archival location
// with the specified target mapping ID, moving it to the AWS KMS key given in
// the parameters. The function returns once the rekey has completed. If the
// rekey fails for any of the group's archival locations, an error holding the
// error message reported by RSC is returned.
func (a API) RekeyRCVAWSMasterKey(ctx context.Context, targetMappingID uuid.UUID, params archival.RekeyArchivalGroupMasterKeyParams) error {
	a.log.Print(log.Trace)

	// The encryption details can only be filtered by name, so the name of the
	// target mapping is needed to find the rekey status.
	targetMapping, err := a.RCVAWSTargetMappingByID(ctx, targetMappingID)
	if err != nil {
		return fmt.Errorf("failed to get rcv aws target mapping: %s", err)
	}

	if err := archival.RekeyArchivalGroupMasterKey(ctx, a.client, targetMappingID, params); err != nil {
		return fmt.Errorf("failed to rekey master key: %s", err)
	}

	if err := a.waitForRekey(ctx, targetMappingID, targetMapping.Name); err != nil {
		return fmt.Errorf("failed to wait for rekey: %s", err)
	}

	return nil
}

// RekeyRCVAzureMasterKey rekeys the master key of the RCV Azure archival
// location with the specified target mapping ID, moving it to the Azure Key
// Vault key given in the parameters. The function returns once the rekey has
// completed. If the rekey fails for any of the group's archival locations, an
// error holding the error message reported by RSC is returned.
func (a API) RekeyRCVAzureMasterKey(ctx context.Context, targetMappingID uuid.UUID, params archival.RekeyArchivalGroupMasterKeyParams) error {
	a.log.Print(log.Trace)

	// The encryption details can only be filtered by name, so the name of the
	// target mapping is needed to find the rekey status.
	targetMapping, err := a.RCVAzureTargetMappingByID(ctx, targetMappingID)
	if err != nil {
		return fmt.Errorf("failed to get rcv azure target mapping: %s", err)
	}

	if err := archival.RekeyArchivalGroupMasterKey(ctx, a.client, targetMappingID, params); err != nil {
		return fmt.Errorf("failed to rekey master key: %s", err)
	}

	if err := a.waitForRekey(ctx, targetMappingID, targetMapping.Name); err != nil {
		return fmt.Errorf("failed to wait for rekey: %s", err)
	}

	return nil
}

// waitForRekey waits for the master key rekey of the archival group with the
// specified ID to complete. A failed rekey of any of the group's archival
// locations is returned as an error.
func (a API) waitForRekey(ctx context.Context, groupID uuid.UUID, groupName string) error {
	a.log.Print(log.Trace)

	for {
		groups, err := archival.ListArchivalGroupEncryptionDetails(ctx, a.client, groupName)
		if err != nil {
			return fmt.Errorf("failed to get encryption details: %s", err)
		}

		var group *archival.ArchivalGroupEncryptionDetails
		for i := range groups {
			if groups[i].ID == groupID {
				group = &groups[i]
				break
			}
		}
		if group == nil {
			return fmt.Errorf("encryption details for %q %w", groupID, graphql.ErrNotFound)
		}

		// A failed location keeps the group in progress, so failures must be
		// checked before the group status.
		for _, location := range group.Locations {
			job := location.EncryptionJobs.LastMasterKeyRekey
			if job != nil && job.Status == archival.EncryptionJobFailed {
				return fmt.Errorf("rekey of archival location %q failed: %s", location.ID, job.ErrorMessage)
			}
		}
		if group.RekeyStatus == archival.RekeyStatusSuccess {
			return nil
		}

		select {
		case <-time.After(rekeyPollInterval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// RCVGCPTargetMappingByName returns the RCV GCP target mapping with the
// specified name. If no target mapping with the specified name is found,
// graphql.ErrNotFound is returned.
func (a API) RCVGCPTargetMappingByName(ctx context.Context, name string) (archival.RCVGCPTargetMapping, error) {
	a.log.Print(log.Trace)

	targets, err := a.RCVGCPTargetMappings(ctx, name)
	if err != nil {
		return archival.RCVGCPTargetMapping{}, err
	}

	name = strings.ToLower(name)
	for _, target := range targets {
		if strings.ToLower(target.Name) == name {
			return target, nil
		}
	}

	return archival.RCVGCPTargetMapping{}, fmt.Errorf("target mapping for %q %w", name, graphql.ErrNotFound)
}

// RCVGCPTargetMappings returns all RCV GCP target mappings matching the
// specified name filter. The name filter can be used to search for prefixes of
// a name. If the name filter is empty, it will match all names.
func (a API) RCVGCPTargetMappings(ctx context.Context, nameFilter string) ([]archival.RCVGCPTargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []archival.ListTargetMappingFilter{{
		Field: "NAME",
		Text:  nameFilter,
	}}
	targets, err := archival.ListTargetMappings[archival.RCVGCPTargetMapping](ctx, a.client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get target mappings: %s", err)
	}

	return targets, nil
}

// CreateRCVGCPStorageSetting creates an RCV GCP cloud native archival location.
func (a API) CreateRCVGCPStorageSetting(ctx context.Context, createParams archival.CreateRCVGCPStorageSettingParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	// GCP requires the tier to be explicitly specified. Default it to backup
	// which is the default used by AWS and Azure.
	if createParams.Tier == "" {
		createParams.Tier = archival.RCVTierBackup
	}

	if createParams.LocTemplate == "" {
		createParams.LocTemplate = archival.LocTemplateSpecificRegion
		if createParams.Region.Region == gcp.RegionUnknown {
			createParams.LocTemplate = archival.LocTemplateSourceRegion
		}
	}

	targetMappingID, err := archival.CreateCloudNativeStorageSetting[archival.CreateRCVGCPStorageSettingResult](ctx, a.client, createParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create cloud native storage setting: %s", err)
	}

	return targetMappingID, nil
}

// UpdateRCVGCPStorageSetting updates the RCV GCP cloud native archival location
// with the specified target mapping ID. The RSC API only supports updating the
// name.
func (a API) UpdateRCVGCPStorageSetting(ctx context.Context, targetMappingID uuid.UUID, updateParams archival.UpdateRCVGCPStorageSettingParams) error {
	a.log.Print(log.Trace)

	// The name is required by the API, and it is also the only field the API
	// accepts, so read it back to keep the call consistent with the AWS and
	// Azure update methods.
	if updateParams.Name == "" {
		targetMapping, err := a.RCVGCPTargetMappingByID(ctx, targetMappingID)
		if err != nil {
			return fmt.Errorf("failed to get rcv gcp target mapping: %s", err)
		}
		updateParams.Name = targetMapping.Name
	}

	if err := archival.UpdateCloudNativeStorageSetting[archival.UpdateRCVGCPStorageSettingResult](ctx, a.client, targetMappingID, updateParams); err != nil {
		return fmt.Errorf("failed to update cloud native storage setting: %s", err)
	}

	return nil
}
