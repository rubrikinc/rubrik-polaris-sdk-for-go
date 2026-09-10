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

package core

import (
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

// canaryPermissionGroups maps RSC version tags to the known PermissionGroup
// constants for that version. This is used by the canary test to detect changes
// to the PermissionsGroup enum in RSC.
//
// When running against a specific RSC version, the test will use the constants
// for the most recent version tag that is not newer than the deployment version.
// The version tags should be ordered from oldest to newest. Each entry can have
// multiple version tags to support different version formats (e.g., master-XXXXX
// and vYYYYMMDD-XX).
var canaryPermissionGroups = []struct {
	versionTags      []string
	permissionGroups map[PermissionGroup]struct{}
}{
	{
		// Base version, should be first in the list.
		versionTags: nil,
		permissionGroups: map[PermissionGroup]struct{}{
			PermissionGroupInvalid:                       {},
			PermissionGroupAKSCustomPrivateDNSZone:       {},
			PermissionGroupAutomatedNetworkingSetup:      {},
			PermissionGroupBackupV2:                      {},
			PermissionGroupBasic:                         {},
			PermissionGroupCCES:                          {},
			PermissionGroupCustomerHostedLogging:         {},
			PermissionGroupCustomerManagedCluster:        {},
			PermissionGroupCustomerManagedStorageIndexng: {},
			PermissionGroupDataCenterConsolidation:       {},
			PermissionGroupDataCenterImmutability:        {},
			PermissionGroupDataCenterKMS:                 {},
			PermissionGroupDownloadFile:                  {},
			PermissionGroupEncryption:                    {},
			PermissionGroupExportAndRestore:              {},
			PermissionGroupExportAndRestorePowerOffVM:    {},
			PermissionGroupExportPowerOff:                {},
			PermissionGroupExportPowerOn:                 {},
			PermissionGroupFileLevelRecovery:             {},
			PermissionGroupPrivateEndpoints:              {},
			PermissionGroupRecovery:                      {},
			PermissionGroupRestore:                       {},
			PermissionGroupRSCManagedCluster:             {},
			PermissionGroupSAPHanaSSBasic:                {},
			PermissionGroupSAPHanaSSRecovery:             {},
			PermissionGroupServiceEndpointAutomation:     {},
			PermissionGroupSnapshotPrivateAccess:         {},
			PermissionGroupSQLArchival:                   {},
		},
	},
	{
		versionTags: []string{"master-86924"},
		permissionGroups: map[PermissionGroup]struct{}{
			PermissionGroupInvalid:                       {},
			PermissionGroupAKSCustomPrivateDNSZone:       {},
			PermissionGroupAutomatedNetworkingSetup:      {},
			PermissionGroupBackupV2:                      {},
			PermissionGroupBasic:                         {},
			PermissionGroupCCES:                          {},
			PermissionGroupCustomerHostedLogging:         {},
			PermissionGroupCustomerManagedCluster:        {},
			PermissionGroupCustomerManagedStorageIndexng: {},
			PermissionGroupDataCenterConsolidation:       {},
			PermissionGroupDataCenterImmutability:        {},
			PermissionGroupDataCenterKMS:                 {},
			PermissionGroupDownloadFile:                  {},
			PermissionGroupEncryption:                    {},
			PermissionGroupExportAndRestore:              {},
			PermissionGroupExportAndRestorePowerOffVM:    {},
			PermissionGroupExportPowerOff:                {},
			PermissionGroupExportPowerOn:                 {},
			PermissionGroupFileLevelRecovery:             {},
			PermissionGroupNATGateway:                    {},
			PermissionGroupPrivateEndpoints:              {},
			PermissionGroupRecovery:                      {},
			PermissionGroupRestore:                       {},
			PermissionGroupRSCManagedCluster:             {},
			PermissionGroupSAPHanaSSBasic:                {},
			PermissionGroupSAPHanaSSRecovery:             {},
			PermissionGroupServiceEndpointAutomation:     {},
			PermissionGroupSnapshotPrivateAccess:         {},
			PermissionGroupSQLArchival:                   {},
		},
	},
	{
		versionTags: []string{"master-88479"},
		permissionGroups: map[PermissionGroup]struct{}{
			PermissionGroupInvalid:                       {},
			PermissionGroupAKSCustomPrivateDNSZone:       {},
			PermissionGroupAutomatedNetworkingSetup:      {},
			PermissionGroupBackupV2:                      {},
			PermissionGroupBasic:                         {},
			PermissionGroupCCES:                          {},
			PermissionGroupCloudSQL:                      {},
			PermissionGroupCustomerHostedLogging:         {},
			PermissionGroupCustomerManagedCluster:        {},
			PermissionGroupCustomerManagedStorageIndexng: {},
			PermissionGroupDataCenterConsolidation:       {},
			PermissionGroupDataCenterImmutability:        {},
			PermissionGroupDataCenterKMS:                 {},
			PermissionGroupDownloadFile:                  {},
			PermissionGroupEncryption:                    {},
			PermissionGroupExportAndRestore:              {},
			PermissionGroupExportAndRestorePowerOffVM:    {},
			PermissionGroupExportPowerOff:                {},
			PermissionGroupExportPowerOn:                 {},
			PermissionGroupFileLevelRecovery:             {},
			PermissionGroupNATGateway:                    {},
			PermissionGroupPrivateEndpoints:              {},
			PermissionGroupRecovery:                      {},
			PermissionGroupRestore:                       {},
			PermissionGroupRSCManagedCluster:             {},
			PermissionGroupSAPHanaSSBasic:                {},
			PermissionGroupSAPHanaSSRecovery:             {},
			PermissionGroupServiceEndpointAutomation:     {},
			PermissionGroupSnapshotPrivateAccess:         {},
			PermissionGroupSQLArchival:                   {},
		},
	},
	{
		versionTags: []string{"master-89707"},
		permissionGroups: map[PermissionGroup]struct{}{
			PermissionGroupInvalid:                       {},
			PermissionGroupAKSCustomPrivateDNSZone:       {},
			PermissionGroupAlloyDB:                       {},
			PermissionGroupAutomatedNetworkingSetup:      {},
			PermissionGroupBackupV2:                      {},
			PermissionGroupBasic:                         {},
			PermissionGroupCCES:                          {},
			PermissionGroupCloudSQL:                      {},
			PermissionGroupCustomerHostedLogging:         {},
			PermissionGroupCustomerManagedCluster:        {},
			PermissionGroupCustomerManagedStorageIndexng: {},
			PermissionGroupDataCenterConsolidation:       {},
			PermissionGroupDataCenterImmutability:        {},
			PermissionGroupDataCenterKMS:                 {},
			PermissionGroupDownloadFile:                  {},
			PermissionGroupEncryption:                    {},
			PermissionGroupExportAndRestore:              {},
			PermissionGroupExportAndRestorePowerOffVM:    {},
			PermissionGroupExportPowerOff:                {},
			PermissionGroupExportPowerOn:                 {},
			PermissionGroupFileLevelRecovery:             {},
			PermissionGroupNATGateway:                    {},
			PermissionGroupPrivateEndpoints:              {},
			PermissionGroupRecovery:                      {},
			PermissionGroupRecoveryNetworking:            {},
			PermissionGroupRestore:                       {},
			PermissionGroupRSCManagedCluster:             {},
			PermissionGroupSAPHanaSSBasic:                {},
			PermissionGroupSAPHanaSSRecovery:             {},
			PermissionGroupServiceEndpointAutomation:     {},
			PermissionGroupSnapshotPrivateAccess:         {},
			PermissionGroupSQLArchival:                   {},
		},
	},
	{
		versionTags: []string{"master-90207"},
		permissionGroups: map[PermissionGroup]struct{}{
			PermissionGroupInvalid:                       {},
			PermissionGroupAKSCustomPrivateDNSZone:       {},
			PermissionGroupAlloyDB:                       {},
			PermissionGroupAutomatedNetworkingSetup:      {},
			PermissionGroupBaaSBasic:                     {},
			PermissionGroupBackupV2:                      {},
			PermissionGroupBasic:                         {},
			PermissionGroupCCES:                          {},
			PermissionGroupCloudSQL:                      {},
			PermissionGroupCustomerHostedLogging:         {},
			PermissionGroupCustomerManagedCluster:        {},
			PermissionGroupCustomerManagedStorageIndexng: {},
			PermissionGroupDataCenterConsolidation:       {},
			PermissionGroupDataCenterImmutability:        {},
			PermissionGroupDataCenterKMS:                 {},
			PermissionGroupDownloadFile:                  {},
			PermissionGroupEncryption:                    {},
			PermissionGroupExportAndRestore:              {},
			PermissionGroupExportAndRestorePowerOffVM:    {},
			PermissionGroupExportPowerOff:                {},
			PermissionGroupExportPowerOn:                 {},
			PermissionGroupFileLevelRecovery:             {},
			PermissionGroupNATGateway:                    {},
			PermissionGroupPrivateEndpoints:              {},
			PermissionGroupRecovery:                      {},
			PermissionGroupRecoveryNetworking:            {},
			PermissionGroupRestore:                       {},
			PermissionGroupRSCManagedCluster:             {},
			PermissionGroupSAPHanaSSBasic:                {},
			PermissionGroupSAPHanaSSRecovery:             {},
			PermissionGroupServiceEndpointAutomation:     {},
			PermissionGroupSnapshotPrivateAccess:         {},
			PermissionGroupSQLArchival:                   {},
		},
	},
	{
		versionTags: []string{"master-94619"},
		permissionGroups: map[PermissionGroup]struct{}{
			PermissionGroupInvalid:                       {},
			PermissionGroupAKSCustomPrivateDNSZone:       {},
			PermissionGroupAlloyDB:                       {},
			PermissionGroupAutomatedNetworkingSetup:      {},
			PermissionGroupBaaSBasic:                     {},
			PermissionGroupBackupV2:                      {},
			PermissionGroupBasic:                         {},
			PermissionGroupCCES:                          {},
			PermissionGroupCloudSQL:                      {},
			PermissionGroupCustomerHostedLogging:         {},
			PermissionGroupCustomerManagedCluster:        {},
			PermissionGroupCustomerManagedStorageIndexng: {},
			PermissionGroupDataCenterConsolidation:       {},
			PermissionGroupDataCenterImmutability:        {},
			PermissionGroupDataCenterKMS:                 {},
			PermissionGroupDownloadFile:                  {},
			PermissionGroupEncryption:                    {},
			PermissionGroupExportAndRestore:              {},
			PermissionGroupExportAndRestorePowerOffVM:    {},
			PermissionGroupExportPowerOff:                {},
			PermissionGroupExportPowerOn:                 {},
			PermissionGroupFileLevelRecovery:             {},
			PermissionGroupKMSKeySharing:                 {},
			PermissionGroupNATGateway:                    {},
			PermissionGroupPrivateEndpoints:              {},
			PermissionGroupRecovery:                      {},
			PermissionGroupRecoveryNetworking:            {},
			PermissionGroupRestore:                       {},
			PermissionGroupRSCManagedCluster:             {},
			PermissionGroupSAPHanaSSBasic:                {},
			PermissionGroupSAPHanaSSRecovery:             {},
			PermissionGroupServiceEndpointAutomation:     {},
			PermissionGroupSnapshotPrivateAccess:         {},
			PermissionGroupSQLArchival:                   {},
		},
	},
	{
		versionTags: []string{"master-96408"},
		permissionGroups: map[PermissionGroup]struct{}{
			PermissionGroupInvalid:                       {},
			PermissionGroupAKSCustomPrivateDNSZone:       {},
			PermissionGroupAdvancedDiagnostics:           {},
			PermissionGroupAlloyDB:                       {},
			PermissionGroupAutomatedNetworkingSetup:      {},
			PermissionGroupBaaSBasic:                     {},
			PermissionGroupBackupV2:                      {},
			PermissionGroupBasic:                         {},
			PermissionGroupBasic2:                        {},
			PermissionGroupCCES:                          {},
			PermissionGroupCloudSQL:                      {},
			PermissionGroupCustomerHostedLogging:         {},
			PermissionGroupCustomerManagedCluster:        {},
			PermissionGroupCustomerManagedStorageIndexng: {},
			PermissionGroupDataCenterConsolidation:       {},
			PermissionGroupDataCenterImmutability:        {},
			PermissionGroupDataCenterKMS:                 {},
			PermissionGroupDownloadFile:                  {},
			PermissionGroupEncryption:                    {},
			PermissionGroupExportAndRestore:              {},
			PermissionGroupExportAndRestorePowerOffVM:    {},
			PermissionGroupExportPowerOff:                {},
			PermissionGroupExportPowerOn:                 {},
			PermissionGroupFileLevelRecovery:             {},
			PermissionGroupGatewayKeyCreation:            {},
			PermissionGroupInventoryGeneration:           {},
			PermissionGroupKMSKeySharing:                 {},
			PermissionGroupNATGateway:                    {},
			PermissionGroupPrivateEndpoints:              {},
			PermissionGroupRecoverToS3:                   {},
			PermissionGroupRecovery:                      {},
			PermissionGroupRecovery2:                     {},
			PermissionGroupRecovery3:                     {},
			PermissionGroupRecovery4:                     {},
			PermissionGroupRecoveryNetworking:            {},
			PermissionGroupRecoveryRDSConnectivity:       {},
			PermissionGroupRestore:                       {},
			PermissionGroupRSCManagedCluster:             {},
			PermissionGroupSAPHanaSSBasic:                {},
			PermissionGroupSAPHanaSSRecovery:             {},
			PermissionGroupServiceEndpointAutomation:     {},
			PermissionGroupSnapshotPrivateAccess:         {},
			PermissionGroupSQLArchival:                   {},
			PermissionGroupSurgicalRecovery:              {},
		},
	},
	{
		versionTags: []string{"master-97996"},
		permissionGroups: map[PermissionGroup]struct{}{
			PermissionGroupInvalid:                       {},
			PermissionGroupAKSCustomPrivateDNSZone:       {},
			PermissionGroupAdvancedDiagnostics:           {},
			PermissionGroupAlloyDB:                       {},
			PermissionGroupAutomatedNetworkingSetup:      {},
			PermissionGroupBaaSBasic:                     {},
			PermissionGroupBackupV2:                      {},
			PermissionGroupBasic:                         {},
			PermissionGroupBasic2:                        {},
			PermissionGroupCCES:                          {},
			PermissionGroupCloudSQL:                      {},
			PermissionGroupCustomerHostedLogging:         {},
			PermissionGroupCustomerManagedCluster:        {},
			PermissionGroupCustomerManagedStorageIndexng: {},
			PermissionGroupDataCenterConsolidation:       {},
			PermissionGroupDataCenterImmutability:        {},
			PermissionGroupDataCenterKMS:                 {},
			PermissionGroupDownloadFile:                  {},
			PermissionGroupEncryption:                    {},
			PermissionGroupExport:                        {},
			PermissionGroupExportAndRestore:              {},
			PermissionGroupExportAndRestorePowerOffVM:    {},
			PermissionGroupExportPowerOff:                {},
			PermissionGroupExportPowerOn:                 {},
			PermissionGroupFileLevelRecovery:             {},
			PermissionGroupGatewayKeyCreation:            {},
			PermissionGroupInventoryGeneration:           {},
			PermissionGroupKMSKeySharing:                 {},
			PermissionGroupNATGateway:                    {},
			PermissionGroupPrivateEndpoints:              {},
			PermissionGroupRecoverToS3:                   {},
			PermissionGroupRecovery:                      {},
			PermissionGroupRecovery2:                     {},
			PermissionGroupRecovery3:                     {},
			PermissionGroupRecovery4:                     {},
			PermissionGroupRecoveryNetworking:            {},
			PermissionGroupRecoveryRDSConnectivity:       {},
			PermissionGroupRestore:                       {},
			PermissionGroupRSCManagedCluster:             {},
			PermissionGroupSAPHanaSSBasic:                {},
			PermissionGroupSAPHanaSSRecovery:             {},
			PermissionGroupServiceEndpointAutomation:     {},
			PermissionGroupSnapshotPrivateAccess:         {},
			PermissionGroupSQLArchival:                   {},
			PermissionGroupSurgicalRecovery:              {},
		},
	},
}

// canaryPermissionGroupsForVersion returns the known PermissionGroup constants
// for the given RSC version. It finds the last version entry (from oldest to
// newest) where the deployment version is after or equal to the version tags.
// Entries without a matching version format are skipped.
func canaryPermissionGroupsForVersion(version graphql.Version) map[PermissionGroup]struct{} {
	// Start with the first entry (base version).
	result := canaryPermissionGroups[0].permissionGroups

	for _, entry := range canaryPermissionGroups {
		// Version is after or equal to the version tags if it's not before them.
		// Skip entries where no version tag matches the version format.
		before, err := version.Before(entry.versionTags...)
		if err == nil && !before {
			result = entry.permissionGroups
		}
	}

	return result
}

// TestPermissionGroupCanary acts as a canary test to catch changes to the
// PermissionsGroup enum in RSC early. It uses GraphQL introspection to read the
// current enum values from RSC and compares them against the known SDK
// constants. If RSC adds or removes enum values, this test will fail, alerting
// developers to update the SDK's PermissionGroup constants.
//
// To run this test against an RSC instance the following environment variables
// need to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
func TestPermissionGroupCanary(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	// Get the known permission groups for this version.
	knownPermissionGroups := canaryPermissionGroupsForVersion(version)

	rscEnumValues, err := graphql.EnumValuesAsSet(t.Context(), client.GQL, "PermissionsGroup")
	if err != nil {
		t.Fatal(err)
	}

	// Check that all known SDK constants are present in RSC.
	for pg := range knownPermissionGroups {
		if pg == PermissionGroupInvalid {
			// Skip the invalid constant, it's not a real enum value.
			continue
		}
		if _, ok := rscEnumValues[string(pg)]; !ok {
			t.Errorf("SDK PermissionGroup %q not found in RSC PermissionsGroup enum", pg)
		}
	}

	// Check if RSC has any enum values that are not in the SDK.
	for enumValue := range rscEnumValues {
		if _, ok := knownPermissionGroups[PermissionGroup(enumValue)]; !ok {
			t.Errorf("RSC PermissionsGroup enum value %q not found in SDK constants", enumValue)
		}
	}
}
