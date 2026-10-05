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
	"slices"
)

// PermissionGroup represents a named set of permissions for a feature. Note,
// not all permission groups are applicable to all features.
type PermissionGroup string

// Note: When adding new PermissionGroup constants, also add them to the
// canaryPermissionGroups in permission_group_canary_test.go.
const (
	PermissionGroupAKSCustomPrivateDNSZone       PermissionGroup = "AKS_CUSTOM_PRIVATE_DNS_ZONE"
	PermissionGroupAlloyDB                       PermissionGroup = "ALLOYDB"
	PermissionGroupAdvancedDiagnostics           PermissionGroup = "ADVANCED_DIAGNOSTICS"
	PermissionGroupArcVMExport                   PermissionGroup = "ARC_VM_EXPORT"
	PermissionGroupAutomatedNetworkingSetup      PermissionGroup = "AUTOMATED_NETWORKING_SETUP"
	PermissionGroupBaaSBasic                     PermissionGroup = "BAAS_BASIC"
	PermissionGroupBackupV2                      PermissionGroup = "BACKUP_V2"
	PermissionGroupBasic                         PermissionGroup = "BASIC"
	PermissionGroupBasic2                        PermissionGroup = "BASIC_2"
	PermissionGroupCCES                          PermissionGroup = "CLOUD_CLUSTER_ES"
	PermissionGroupCloudSQL                      PermissionGroup = "CLOUDSQL"
	PermissionGroupCustomerHostedLogging         PermissionGroup = "CUSTOMER_HOSTED_LOGGING"
	PermissionGroupCustomerManagedCluster        PermissionGroup = "CUSTOMER_MANAGED_BASIC"
	PermissionGroupCustomerManagedStorageIndexng PermissionGroup = "CUSTOMER_MANAGED_STORAGE_INDEXING"
	PermissionGroupDataCenterConsolidation       PermissionGroup = "DATA_CENTER_CONSOLIDATION"
	PermissionGroupDataCenterImmutability        PermissionGroup = "DATA_CENTER_IMMUTABILITY"
	PermissionGroupDataCenterKMS                 PermissionGroup = "DATA_CENTER_KMS"
	PermissionGroupDownloadFile                  PermissionGroup = "DOWNLOAD_FILE"
	PermissionGroupEncryption                    PermissionGroup = "ENCRYPTION"
	PermissionGroupExport                        PermissionGroup = "EXPORT"
	PermissionGroupExportAndRestore              PermissionGroup = "EXPORT_AND_RESTORE"
	PermissionGroupExportAndRestorePowerOffVM    PermissionGroup = "EXPORT_AND_RESTORE_POWER_OFF_VM"
	PermissionGroupExportPowerOff                PermissionGroup = "EXPORT_POWER_OFF"
	PermissionGroupExportPowerOn                 PermissionGroup = "EXPORT_POWER_ON"
	PermissionGroupFileLevelRecovery             PermissionGroup = "FILE_LEVEL_RECOVERY"
	PermissionGroupGatewayKeyCreation            PermissionGroup = "GATEWAY_KEY_CREATION"
	PermissionGroupInvalid                       PermissionGroup = "GROUP_UNSPECIFIED"
	PermissionGroupInventoryGeneration           PermissionGroup = "INVENTORY_GENERATION"
	PermissionGroupKMSKeySharing                 PermissionGroup = "KMS_KEY_SHARING"
	PermissionGroupNATGateway                    PermissionGroup = "NAT_GATEWAY"
	PermissionGroupPrivateEndpoints              PermissionGroup = "PRIVATE_ENDPOINTS"
	PermissionGroupRecoverToS3                   PermissionGroup = "RECOVER_TO_S3"
	PermissionGroupRecovery                      PermissionGroup = "RECOVERY"
	PermissionGroupRecovery2                     PermissionGroup = "RECOVERY_2"
	PermissionGroupRecovery3                     PermissionGroup = "RECOVERY_3"
	PermissionGroupRecovery4                     PermissionGroup = "RECOVERY_4"
	PermissionGroupRecoveryNetworking            PermissionGroup = "RECOVERY_NETWORKING"
	PermissionGroupRecoveryRDSConnectivity       PermissionGroup = "RECOVERY_RDS_CONNECTIVITY"
	PermissionGroupRestore                       PermissionGroup = "RESTORE"
	PermissionGroupRSCManagedCluster             PermissionGroup = "RSC_MANAGED_CLUSTER"
	PermissionGroupSAPHanaSSBasic                PermissionGroup = "SAP_HANA_SS_BASIC"
	PermissionGroupSAPHanaSSRecovery             PermissionGroup = "SAP_HANA_SS_RECOVERY"
	PermissionGroupServiceEndpointAutomation     PermissionGroup = "SERVICE_ENDPOINT_AUTOMATION"
	PermissionGroupSnapshotPrivateAccess         PermissionGroup = "SNAPSHOT_PRIVATE_ACCESS"
	PermissionGroupSQLArchival                   PermissionGroup = "SQL_ARCHIVAL"
	PermissionGroupSurgicalRecovery              PermissionGroup = "SURGICAL_RECOVERY"
)

// AWSCloudFormationPermissionGroupNames returns the names of the permission
// groups of the features which can be onboarded for an AWS account using the
// CloudFormation workflow, sorted by name.
func AWSCloudFormationPermissionGroupNames() []string {
	return PermissionGroupNames(permissionGroupsFor(AWSCloudFormationFeatures()))
}

// AWSIAMRolesPermissionGroupNames returns the names of the permission groups of
// the features which can be onboarded for an AWS account using the IAM roles
// workflow, sorted by name.
func AWSIAMRolesPermissionGroupNames() []string {
	return PermissionGroupNames(permissionGroupsFor(AWSIAMRolesFeatures()))
}

// AWSManagedPermissionGroupNames returns the names of the permission groups of
// the features which can be onboarded for an RSC-managed (BaaS) AWS account,
// sorted by name.
func AWSManagedPermissionGroupNames() []string {
	return PermissionGroupNames(permissionGroupsFor(AWSManagedFeatures()))
}

// AzureSubscriptionPermissionGroupNames returns the names of the permission
// groups of the features which can be onboarded for an Azure subscription,
// sorted by name.
func AzureSubscriptionPermissionGroupNames() []string {
	return PermissionGroupNames(permissionGroupsFor(AzureSubscriptionFeatures()))
}

// AzureDevOpsOrganizationPermissionGroupNames returns the names of the
// permission groups of the features which can be onboarded for an Azure DevOps
// organization, sorted by name.
func AzureDevOpsOrganizationPermissionGroupNames() []string {
	return PermissionGroupNames(permissionGroupsFor(AzureDevOpsOrganizationFeatures()))
}

// GCPProjectPermissionGroupNames returns the names of the permission groups of
// the features which can be onboarded for a GCP project, sorted by name.
func GCPProjectPermissionGroupNames() []string {
	return PermissionGroupNames(permissionGroupsFor(GCPProjectFeatures()))
}

// GitHubOrganizationPermissionGroupNames returns the names of the permission
// groups of the features which can be onboarded for a GitHub organization,
// sorted by name.
func GitHubOrganizationPermissionGroupNames() []string {
	return PermissionGroupNames(permissionGroupsFor(GitHubOrganizationFeatures()))
}

// PermissionGroupNames returns the names of the permission groups.
func PermissionGroupNames(groups []PermissionGroup) []string {
	var names []string
	for _, group := range groups {
		names = append(names, string(group))
	}

	return names
}

// permissionGroupsFor returns the permission groups for the specified feature,
// sorted by name.
func permissionGroupsFor(features []Feature) []PermissionGroup {
	var groups []PermissionGroup
	for _, feature := range features {
		groups = append(groups, feature.PermissionGroups...)
	}
	slices.Sort(groups)

	return slices.Compact(groups)
}
