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
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Note: When adding new PermissionGroup constants, also add them to the
// canaryPermissionGroups in permission_group_canary_test.go.
const (
	PermissionGroupAKSCustomPrivateDNSZone       PermissionGroup = "AKS_CUSTOM_PRIVATE_DNS_ZONE"
	PermissionGroupAlloyDB                       PermissionGroup = "ALLOYDB"
	PermissionGroupAutomatedNetworkingSetup      PermissionGroup = "AUTOMATED_NETWORKING_SETUP"
	PermissionGroupBaaSBasic                     PermissionGroup = "BAAS_BASIC"
	PermissionGroupBackupV2                      PermissionGroup = "BACKUP_V2"
	PermissionGroupBasic                         PermissionGroup = "BASIC"
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
	PermissionGroupExportAndRestore              PermissionGroup = "EXPORT_AND_RESTORE"
	PermissionGroupExportAndRestorePowerOffVM    PermissionGroup = "EXPORT_AND_RESTORE_POWER_OFF_VM"
	PermissionGroupExportPowerOff                PermissionGroup = "EXPORT_POWER_OFF"
	PermissionGroupExportPowerOn                 PermissionGroup = "EXPORT_POWER_ON"
	PermissionGroupFileLevelRecovery             PermissionGroup = "FILE_LEVEL_RECOVERY"
	PermissionGroupInvalid                       PermissionGroup = "GROUP_UNSPECIFIED"
	PermissionGroupKMSKeySharing                 PermissionGroup = "KMS_KEY_SHARING"
	PermissionGroupNATGateway                    PermissionGroup = "NAT_GATEWAY"
	PermissionGroupPrivateEndpoints              PermissionGroup = "PRIVATE_ENDPOINTS"
	PermissionGroupRecovery                      PermissionGroup = "RECOVERY"
	PermissionGroupRecoveryNetworking            PermissionGroup = "RECOVERY_NETWORKING"
	PermissionGroupRestore                       PermissionGroup = "RESTORE"
	PermissionGroupRSCManagedCluster             PermissionGroup = "RSC_MANAGED_CLUSTER"
	PermissionGroupSAPHanaSSBasic                PermissionGroup = "SAP_HANA_SS_BASIC"
	PermissionGroupSAPHanaSSRecovery             PermissionGroup = "SAP_HANA_SS_RECOVERY"
	PermissionGroupServiceEndpointAutomation     PermissionGroup = "SERVICE_ENDPOINT_AUTOMATION"
	PermissionGroupSnapshotPrivateAccess         PermissionGroup = "SNAPSHOT_PRIVATE_ACCESS"
	PermissionGroupSQLArchival                   PermissionGroup = "SQL_ARCHIVAL"
)

var (
	FeatureInvalid                                     = Feature{Name: ""}
	FeatureAll                                         = Feature{Name: "ALL"}
	FeatureAppFlows                                    = Feature{Name: "APP_FLOWS"}
	FeatureArchival                                    = Feature{Name: "ARCHIVAL"}
	FeatureAzureDevOpsDeveloperCollaborationProtection = Feature{Name: "AZURE_DEVOPS_DEVELOPER_COLLABORATION_PROTECTION"}
	FeatureAzureDevOpsProtection                       = Feature{Name: "AZURE_DEVOPS_PROTECTION"} // Deprecated: use FeatureAzureDevOpsRepositoryProtection
	FeatureAzureDevOpsRepositoryProtection             = Feature{Name: "AZURE_DEVOPS_REPOSITORY_PROTECTION"}
	FeatureAzureSQLDBProtection                        = Feature{Name: "AZURE_SQL_DB_PROTECTION"}
	FeatureAzureSQLMIProtection                        = Feature{Name: "AZURE_SQL_MI_PROTECTION"}
	FeatureCloudAccounts                               = Feature{Name: "CLOUDACCOUNTS"} // Deprecated: no replacement.
	FeatureCloudDiscovery                              = Feature{Name: "CLOUD_DISCOVERY"}
	FeatureCloudNativeArchival                         = Feature{Name: "CLOUD_NATIVE_ARCHIVAL"}
	FeatureCloudNativeArchivalEncryption               = Feature{Name: "CLOUD_NATIVE_ARCHIVAL_ENCRYPTION"}
	FeatureCloudNativeBlobProtection                   = Feature{Name: "CLOUD_NATIVE_BLOB_PROTECTION"}
	FeatureCloudNativeDynamoDBProtection               = Feature{Name: "CLOUD_NATIVE_DYNAMODB_PROTECTION"}
	FeatureCloudNativeProtection                       = Feature{Name: "CLOUD_NATIVE_PROTECTION"}
	FeatureCloudNativeS3Protection                     = Feature{Name: "CLOUD_NATIVE_S3_PROTECTION"}
	FeatureCyberRecoveryDataClassificationData         = Feature{Name: "CYBERRECOVERY_DATA_CLASSIFICATION_DATA"}
	FeatureCyberRecoveryDataClassificationMetadata     = Feature{Name: "CYBERRECOVERY_DATA_CLASSIFICATION_METADATA"}
	FeatureDSPMData                                    = Feature{Name: "DSPM_DATA"}
	FeatureDSPMMetadata                                = Feature{Name: "DSPM_METADATA"}
	FeatureExocompute                                  = Feature{Name: "EXOCOMPUTE"}
	FeatureGCPSharedVPCHost                            = Feature{Name: "GCP_SHARED_VPC_HOST"}
	FeatureGitHubRepositoryProtection                  = Feature{Name: "GITHUB_REPOSITORY_PROTECTION"}
	FeatureKubernetesProtection                        = Feature{Name: "KUBERNETES_PROTECTION"}
	FeatureLaminarCrossAccount                         = Feature{Name: "LAMINAR_CROSS_ACCOUNT"}
	FeatureLaminarInternal                             = Feature{Name: "LAMINAR_INTERNAL"}
	FeatureLaminarOutpostApplication                   = Feature{Name: "LAMINAR_OUTPOST_APPLICATION"}
	FeatureLaminarOutpostManagedIdentity               = Feature{Name: "LAMINAR_OUTPOST_MANAGED_IDENTITY"}
	FeatureLaminarTargetApplication                    = Feature{Name: "LAMINAR_TARGET_APPLICATION"}
	FeatureLaminarTargetManagedIdentity                = Feature{Name: "LAMINAR_TARGET_MANAGED_IDENTITY"}
	FeatureOutpost                                     = Feature{Name: "OUTPOST"}
	FeatureRDSProtection                               = Feature{Name: "RDS_PROTECTION"}
	FeatureRoleChaining                                = Feature{Name: "ROLE_CHAINING"}
	FeatureServerAndApps                               = Feature{Name: "SERVERS_AND_APPS"}
)

// Feature represents an RSC cloud account feature with a set of permission
// groups.
type Feature struct {
	Name             string            `json:"featureType"`
	PermissionGroups []PermissionGroup `json:"permissionsGroups"`
}

// PermissionGroup represents a named set of permissions for a feature. Note,
// not all permission groups are applicable to all features.
type PermissionGroup string

// Equal returns true if the features have the same name. Note, this function
// does not compare the permission groups.
func (f Feature) Equal(other Feature) bool {
	return f.Name == other.Name
}

// DeepEqual returns true if the features are equal. The features are equal if
// they have the same name and the same permission groups.
func (f Feature) DeepEqual(feature Feature) bool {
	if !f.Equal(feature) {
		return false
	}

	set := make(map[PermissionGroup]struct{}, len(f.PermissionGroups))
	for _, permissionGroup := range f.PermissionGroups {
		set[permissionGroup] = struct{}{}
	}
	for _, permissionGroup := range feature.PermissionGroups {
		if _, ok := set[permissionGroup]; !ok {
			return false
		}
		delete(set, permissionGroup)
	}

	return len(set) == 0
}

// HasPermissionGroup returns true if the feature has the specified permission
// group.
func (f Feature) HasPermissionGroup(permissionGroup PermissionGroup) bool {
	return slices.Contains(f.PermissionGroups, permissionGroup)
}

// IsProtectionFeature
func (f Feature) IsProtectionFeature() bool {
	protectionFeatures := []Feature{
		FeatureAzureDevOpsDeveloperCollaborationProtection,
		FeatureAzureDevOpsProtection,
		FeatureAzureDevOpsRepositoryProtection,
		FeatureAzureSQLDBProtection,
		FeatureAzureSQLMIProtection,
		FeatureCloudDiscovery,
		FeatureCloudNativeBlobProtection,
		FeatureCloudNativeDynamoDBProtection,
		FeatureCloudNativeProtection,
		FeatureCloudNativeS3Protection,
		FeatureGitHubRepositoryProtection,
		FeatureKubernetesProtection,
		FeatureRDSProtection,
	}

	return slices.ContainsFunc(protectionFeatures, func(feature Feature) bool {
		return f.Equal(feature)
	})
}

// String returns a string representation of the feature.
func (f Feature) String() string {
	if len(f.PermissionGroups) == 0 {
		return f.Name
	}

	var buf strings.Builder
	permissionGroups := slices.Clone(f.PermissionGroups)
	slices.Sort(permissionGroups)
	for _, permissionGroup := range permissionGroups {
		buf.WriteString(string(permissionGroup))
		buf.WriteString(",")
	}

	return fmt.Sprintf("%s(%s)", f.Name, buf.String()[:buf.Len()-1])
}

// WithPermissionGroups returns a copy of the feature with the specified
// permission groups added.
func (f Feature) WithPermissionGroups(permissionGroups ...PermissionGroup) Feature {
	groups := append(f.PermissionGroups, permissionGroups...)
	return Feature{Name: f.Name, PermissionGroups: groups}
}

// AllProtectionFeatures returns the protection features for the specified cloud
// vendor.
func AllProtectionFeatures(cloud CloudVendor) []Feature {
	switch cloud {
	case CloudVendorAWS:
		return []Feature{
			FeatureCloudNativeDynamoDBProtection,
			FeatureCloudNativeProtection,
			FeatureCloudNativeS3Protection,
			FeatureKubernetesProtection,
			FeatureRDSProtection,
		}
	case CloudVendorAzure:
		return []Feature{
			FeatureAzureSQLDBProtection,
			FeatureAzureSQLMIProtection,
			FeatureCloudNativeBlobProtection,
			FeatureCloudNativeProtection,
			FeatureKubernetesProtection,
		}
	case CloudVendorGCP:
		return []Feature{
			FeatureCloudNativeProtection,
		}
	default:
		return nil
	}
}

// FeatureNames returns the names of the features.
func FeatureNames(features []Feature) []string {
	var names []string
	for _, feature := range features {
		names = append(names, feature.Name)
	}

	return names
}

// FilterFeaturesOnPermissionGroups verifies that all features either have no
// permission groups or all have permission groups. The features are returned
// in two different slices, depending on whether they have permission groups
// or not.
func FilterFeaturesOnPermissionGroups(features []Feature) ([]string, []Feature, error) {
	if len(features) == 0 {
		return nil, nil, errors.New("no features specified")
	}

	// Check that all features have the same use of permission groups.
	usePG := len(features[0].PermissionGroups) > 0
	for _, feature := range features[1:] {
		if pg := len(feature.PermissionGroups) > 0; pg != usePG {
			return nil, nil, errors.New("features with and without permission groups cannot be mixed")
		}
	}
	if usePG {
		return nil, features, nil
	}

	return FeatureNames(features), nil, nil
}

// LookupFeature returns the specified feature if it exists in the feature
// slice.
func LookupFeature(features []Feature, feature Feature) (Feature, bool) {
	for _, f := range features {
		if f.Equal(feature) {
			return f, true
		}
	}

	return Feature{}, false
}

// ValidateRoleChaining returns an error if ROLE_CHAINING is combined with
// other features. The ROLE_CHAINING feature is mutually exclusive with all
// other features.
func ValidateRoleChaining(features []Feature) error {
	if len(features) < 2 {
		return nil
	}
	if _, ok := LookupFeature(features, FeatureRoleChaining); ok {
		return errors.New("ROLE_CHAINING is mutually exclusive with all other features")
	}
	return nil
}

// Deprecated: no replacement.
var validFeatures = map[string]struct{}{
	FeatureAll.Name:      {},
	FeatureAppFlows.Name: {},
	FeatureArchival.Name: {},
	FeatureAzureDevOpsDeveloperCollaborationProtection.Name: {},
	FeatureAzureDevOpsProtection.Name:                       {}, // Deprecated: use FeatureAzureDevOpsRepositoryProtection
	FeatureAzureDevOpsRepositoryProtection.Name:             {},
	FeatureAzureSQLDBProtection.Name:                        {},
	FeatureAzureSQLMIProtection.Name:                        {},
	FeatureCloudAccounts.Name:                               {}, // Deprecated: no replacement.
	FeatureCloudDiscovery.Name:                              {},
	FeatureCloudNativeArchival.Name:                         {},
	FeatureCloudNativeArchivalEncryption.Name:               {},
	FeatureCloudNativeBlobProtection.Name:                   {},
	FeatureCloudNativeDynamoDBProtection.Name:               {},
	FeatureCloudNativeProtection.Name:                       {},
	FeatureCloudNativeS3Protection.Name:                     {},
	FeatureCyberRecoveryDataClassificationData.Name:         {},
	FeatureCyberRecoveryDataClassificationMetadata.Name:     {},
	FeatureDSPMData.Name:                                    {},
	FeatureDSPMMetadata.Name:                                {},
	FeatureExocompute.Name:                                  {},
	FeatureGCPSharedVPCHost.Name:                            {},
	FeatureGitHubRepositoryProtection.Name:                  {},
	FeatureKubernetesProtection.Name:                        {},
	FeatureLaminarCrossAccount.Name:                         {},
	FeatureLaminarInternal.Name:                             {},
	FeatureLaminarOutpostApplication.Name:                   {},
	FeatureLaminarOutpostManagedIdentity.Name:               {},
	FeatureLaminarTargetApplication.Name:                    {},
	FeatureLaminarTargetManagedIdentity.Name:                {},
	FeatureOutpost.Name:                                     {},
	FeatureRDSProtection.Name:                               {},
	FeatureRoleChaining.Name:                                {},
	FeatureServerAndApps.Name:                               {},
}

// Deprecated: use Feature.Name instead.
func FormatFeature(feature Feature) string {
	return strings.ReplaceAll(strings.ToLower(feature.Name), "_", "-")
}

// Deprecated: use Feature{Name: <feature>} instead or ParseFeatureNoValidation
// if you need to remain backwards compatible with previously accepted feature
// names.
func ParseFeature(feature string) (Feature, error) {
	f := ParseFeatureNoValidation(feature)
	if _, ok := validFeatures[f.Name]; ok {
		return f, nil
	}

	return FeatureInvalid, fmt.Errorf("invalid feature: %s", feature)
}

// ParseFeatureNoValidation returns the Feature matching the given feature name.
// No validation is performed.
func ParseFeatureNoValidation(feature string) Feature {
	return Feature{Name: strings.ToUpper(strings.ReplaceAll(feature, "-", "_"))}
}
