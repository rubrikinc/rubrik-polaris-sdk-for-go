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

package devops

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/archival"
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	gqldevops "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/devops"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// GenerateAzureOnboardingScript generates the Azure DevOps onboarding scripts.
// Cloud defaults to the Azure public cloud when empty. The returned scripts are
// base64-encoded exactly as returned by RSC; decoding is left to the caller.
func (a API) GenerateAzureOnboardingScript(ctx context.Context, params gqldevops.GenerateAzureOnboardingScriptParams) (gqldevops.AzureOnboardingScript, error) {
	a.log.Print(log.Trace)

	if params.Cloud == "" {
		params.Cloud = gqlazure.PublicCloud
	}

	script, err := gqldevops.GenerateAzureOnboardingScript(ctx, a.client, params)
	if err != nil {
		return gqldevops.AzureOnboardingScript{}, fmt.Errorf("failed to generate Azure DevOps onboarding script: %s", err)
	}

	return script, nil
}

// AddAzureCloudAccount onboards one or more Azure DevOps organizations using a
// customer-supplied application (non-OAuth). Cloud defaults to the Azure public
// cloud when empty. The customer application must already be registered via the
// azure package using the azure.AppUseCaseDevOps use case.
//
// The onboarding mutation returns no result, so each organization is resolved
// by its native ID and returned in the same order as params.OrganizationNativeIDs.
// The RSC hierarchy is eventually consistent, so this blocks until the
// organizations appear or the context is cancelled.
func (a API) AddAzureCloudAccount(ctx context.Context, params gqldevops.AddAzureCloudAccountParams) ([]gqldevops.AzureOrganization, error) {
	a.log.Print(log.Trace)

	if params.Cloud == "" {
		params.Cloud = gqlazure.PublicCloud
	}

	if err := validateHostStorage(params.HostType, params.StorageType, params.BackupLocationID, params.ExocomputeCloudAccountID, params.ExocomputeRegion); err != nil {
		return nil, err
	}

	if err := gqldevops.AddAzureCloudAccountWithoutOauth(ctx, a.client, params); err != nil {
		return nil, fmt.Errorf("failed to add Azure DevOps cloud account: %s", err)
	}

	orgs := make([]gqldevops.AzureOrganization, 0, len(params.OrganizationNativeIDs))
	for _, nativeID := range params.OrganizationNativeIDs {
		org, err := a.azureOrganizationByNativeID(ctx, nativeID)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

// azureOrganizationByNativeID resolves an onboarded organization by its native
// ID, looking it up in the generic hierarchy and returning the full
// organization. Because the hierarchy is eventually consistent, it polls every
// 5 seconds until the organization appears, the caller's context is cancelled,
// or an internal 60-second deadline elapses. Exceeding the deadline returns a
// not-found error.
func (a API) azureOrganizationByNativeID(ctx context.Context, nativeID string) (gqldevops.AzureOrganization, error) {
	innerCtx, cancel := context.WithTimeoutCause(ctx, 60*time.Second, fmt.Errorf("organization %q not found in hierarchy", nativeID))
	defer cancel()

	for {
		objects, err := hierarchy.ObjectsByName[hierarchy.AzureDevOpsOrganization](innerCtx, hierarchy.Wrap(a.client), nativeID, hierarchy.WorkloadAllSubHierarchyType)
		if err != nil {
			return gqldevops.AzureOrganization{}, err
		}
		for _, obj := range objects {
			if obj.Object.Name == nativeID {
				return a.AzureOrganizationByID(ctx, obj.Object.ID)
			}
		}

		select {
		case <-innerCtx.Done():
			return gqldevops.AzureOrganization{}, context.Cause(innerCtx)
		case <-time.After(5 * time.Second):
		}
	}
}

// UpdateAzureCloudAccount updates the backup location/region, exocompute
// account/region and host/storage type of an onboarded Azure DevOps
// organization.
func (a API) UpdateAzureCloudAccount(ctx context.Context, params gqldevops.UpdateAzureCloudAccountParams) error {
	a.log.Print(log.Trace)

	if err := validateHostStorage(params.HostType, params.StorageType, params.BackupLocationID, params.ExocomputeCloudAccountID, params.ExocomputeRegion); err != nil {
		return err
	}

	// The RSC backend requires the backup region to be included in the update
	// request. Callers do not always supply it, so when absent default it to
	// the archival (backup) location's region.
	if targetID := params.BackupLocationID; params.BackupRegion == nil && targetID != nil {
		targetMapping, err := archival.WrapGQL(a.client).AzureTargetMappingByID(ctx, *targetID)
		if err != nil {
			return fmt.Errorf("failed to get archival location %q: %s", *targetID, err)
		}

		params.BackupRegion = &targetMapping.TargetTemplate.CloudNativeCompanion.StorageAccountRegion
	}

	if err := gqldevops.UpdateAzureCloudAccount(ctx, a.client, params); err != nil {
		return fmt.Errorf("failed to update Azure DevOps cloud account: %s", err)
	}

	return nil
}

// UpgradeAzureCloudAccount informs RSC that the permissions for the given
// features have been updated and applied in the Azure DevOps organization,
// advancing RSC to the latest permission version. Run the updated onboarding
// script in the organization to grant the new permissions before calling this.
func (a API) UpgradeAzureCloudAccount(ctx context.Context, params gqldevops.UpgradeAzureCloudAccountParams) error {
	a.log.Print(log.Trace)

	if err := gqldevops.UpgradeAzureCloudAccountWithoutOauth(ctx, a.client, params); err != nil {
		return fmt.Errorf("failed to upgrade Azure DevOps cloud account: %s", err)
	}

	return nil
}

// DeleteAzureCloudAccount removes an onboarded Azure DevOps organization, keyed
// by organizationID. If deleteSnapshots is true, the organization's snapshots
// are also deleted.
func (a API) DeleteAzureCloudAccount(ctx context.Context, organizationID uuid.UUID, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	if err := gqldevops.DeleteAzureCloudAccountWithoutOauth(ctx, a.client, organizationID, deleteSnapshots); err != nil {
		return fmt.Errorf("failed to delete Azure DevOps cloud account: %s", err)
	}

	return nil
}

// validateHostStorage enforces the conditional rules that couple host and
// storage type to their dependent identifiers: BYOS storage requires a backup
// location, customer-hosted exocompute requires an exocompute cloud account,
// and Rubrik-hosted exocompute requires an exocompute region.
func validateHostStorage(hostType gqldevops.HostType, storageType gqldevops.StorageType, backupLocationID, exocomputeCloudAccountID *uuid.UUID, exocomputeRegion *azure.Region) error {
	if storageType == gqldevops.StorageTypeBYOS && backupLocationID == nil {
		return fmt.Errorf("storage type %s requires a backup location id", storageType)
	}
	if hostType == gqldevops.HostTypeCustomer && exocomputeCloudAccountID == nil {
		return fmt.Errorf("host type %s requires an exocompute cloud account id", hostType)
	}
	if hostType == gqldevops.HostTypeRubrik && (exocomputeRegion == nil || *exocomputeRegion == azure.RegionUnknown) {
		return fmt.Errorf("host type %s requires an exocompute region", hostType)
	}

	return nil
}
