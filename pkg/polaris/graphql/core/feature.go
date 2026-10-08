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

var (
	FeatureInvalid                                     = Feature{Name: ""}
	FeatureAll                                         = Feature{Name: All}
	FeatureAlloyDBProtection                           = Feature{Name: AlloyDBProtection}
	FeatureAppFlows                                    = Feature{Name: AppFlows}
	FeatureArchival                                    = Feature{Name: Archival}
	FeatureAWSKMSKeySharing                            = Feature{Name: AWSKMSKeySharing}
	FeatureAzureCosmosNoSQLProtection                  = Feature{Name: AzureCosmosNoSQLProtection}
	FeatureAzureDevOpsArtifactsProtection              = Feature{Name: AzureDevOpsArtifactsProtection}
	FeatureAzureDevOpsDeveloperCollaborationProtection = Feature{Name: AzureDevOpsDeveloperCollaborationProtection}
	FeatureAzureDevOpsProtection                       = Feature{Name: AzureDevOpsProtection}
	FeatureAzureDevOpsRepositoryProtection             = Feature{Name: AzureDevOpsRepositoryProtection}
	FeatureAzureLocalCloudAccount                      = Feature{Name: AzureLocalCloudAccount}
	FeatureAzurePostgresFlexibleServerProtection       = Feature{Name: AzurePostgresFlexibleServerProtection}
	FeatureAzureSQLDBProtection                        = Feature{Name: AzureSQLDBProtection}
	FeatureAzureSQLMIProtection                        = Feature{Name: AzureSQLMIProtection}
	FeatureCCESBaaS                                    = Feature{Name: CCESBaaS}
	FeatureCloudAccounts                               = Feature{Name: CloudAccounts}
	FeatureCloudCostReport                             = Feature{Name: CloudCostReport}
	FeatureCloudDiscovery                              = Feature{Name: CloudDiscovery}
	FeatureCloudNativeArchival                         = Feature{Name: CloudNativeArchival}
	FeatureCloudNativeArchivalEncryption               = Feature{Name: CloudNativeArchivalEncryption}
	FeatureCloudNativeBlobProtection                   = Feature{Name: CloudNativeBlobProtection}
	FeatureCloudNativeConfigProtection                 = Feature{Name: CloudNativeConfigProtection}
	FeatureCloudNativeDynamoDBProtection               = Feature{Name: CloudNativeDynamoDBProtection}
	FeatureCloudNativeProtection                       = Feature{Name: CloudNativeProtection}
	FeatureCloudNativeS3Protection                     = Feature{Name: CloudNativeS3Protection}
	FeatureCloudNativeUEMKeyManagement                 = Feature{Name: CloudNativeUEMKeyManagement}
	FeatureCloudSQLProtection                          = Feature{Name: CloudSQLProtection}
	FeatureCriticalResourceProtection                  = Feature{Name: CriticalResourceProtection}
	FeatureCyberRecoveryDataClassificationData         = Feature{Name: CyberRecoveryDataClassificationData}
	FeatureCyberRecoveryDataClassificationMetadata     = Feature{Name: CyberRecoveryDataClassificationMetadata}
	FeatureDataCenterRoleBasedArchival                 = Feature{Name: DataCenterRoleBasedArchival}
	FeatureDSPMData                                    = Feature{Name: DSPMData}
	FeatureDSPMMetadata                                = Feature{Name: DSPMMetadata}
	FeatureExocompute                                  = Feature{Name: Exocompute}
	FeatureGCPBigQueryProtection                       = Feature{Name: GCPBigQueryProtection}
	FeatureGCPBigQueryReservation                      = Feature{Name: GCPBigQueryReservation}
	FeatureGCPSharedVPCHost                            = Feature{Name: GCPSharedVPCHost}
	FeatureGitHubDeveloperCollaborationProtection      = Feature{Name: GitHubDeveloperCollaborationProtection}
	FeatureGitHubPackagesProtection                    = Feature{Name: GitHubPackagesProtection}
	FeatureGitHubRepositoryProtection                  = Feature{Name: GitHubRepositoryProtection}
	FeatureGlueIcebergProtection                       = Feature{Name: GlueIcebergProtection}
	FeatureKubernetesProtection                        = Feature{Name: KubernetesProtection}
	FeatureLaminarCrossAccount                         = Feature{Name: LaminarCrossAccount}
	FeatureLaminarInternal                             = Feature{Name: LaminarInternal}
	FeatureLaminarOutpostApplication                   = Feature{Name: LaminarOutpostApplication}
	FeatureLaminarOutpostManagedIdentity               = Feature{Name: LaminarOutpostManagedIdentity}
	FeatureLaminarTargetApplication                    = Feature{Name: LaminarTargetApplication}
	FeatureLaminarTargetManagedIdentity                = Feature{Name: LaminarTargetManagedIdentity}
	FeatureOutpost                                     = Feature{Name: Outpost}
	FeatureRDSProtection                               = Feature{Name: RDSProtection}
	FeatureRoleChaining                                = Feature{Name: RoleChaining}
	FeatureS3TablesIcebergProtection                   = Feature{Name: S3TablesIcebergProtection}
	FeatureServerAndApps                               = Feature{Name: ServerAndApps}
)

// Feature represents an RSC cloud account feature with a set of permission
// groups.
type Feature struct {
	Name             string            `json:"featureType"`
	PermissionGroups []PermissionGroup `json:"permissionsGroups"`
}

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

// IsProtectionFeature returns true if the feature is a protection feature.
// Protection features, public or not, are marked as such in the feature info
// map. Unknown features are not protection features.
func (f Feature) IsProtectionFeature() bool {
	return featureInfoMap[f.Name].protection
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
//
// Deprecated: use CloudVendorProtectionFeatureNames instead.
func AllProtectionFeatures(cloud CloudVendor) []Feature {
	var features []Feature
	for _, name := range CloudVendorProtectionFeatureNames(cloud) {
		features = append(features, Feature{Name: name})
	}

	return features
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

// LookupFeatureName returns the feature with the specified name if it exists
// in the feature slice.
func LookupFeatureName(features []Feature, name string) (Feature, bool) {
	return LookupFeature(features, Feature{Name: name})
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

// Deprecated: use Feature.Name instead.
func FormatFeature(feature Feature) string {
	return strings.ReplaceAll(strings.ToLower(feature.Name), "_", "-")
}

// ParseFeature returns the Feature matching the given feature name. Returns an
// error if the feature name is not a known RSC feature, public or not.
//
// Deprecated: use the Parse<AccountType>Feature function of the cloud account
// type, e.g. ParseAWSCloudFormationFeature, or Feature{Name: <feature>} if no
// validation is needed.
func ParseFeature(feature string) (Feature, error) {
	f := ParseFeatureNoValidation(feature)
	if _, ok := featureInfoMap[f.Name]; ok {
		return f, nil
	}

	return FeatureInvalid, fmt.Errorf("invalid feature: %s", feature)
}

// ParseFeatureNoValidation returns the Feature matching the given feature name.
// No validation is performed.
//
// Deprecated: use the Parse<AccountType>Feature function of the cloud account
// type, e.g. ParseAWSCloudFormationFeature, or Feature{Name: <feature>} if no
// validation is needed.
func ParseFeatureNoValidation(feature string) Feature {
	return Feature{Name: strings.ToUpper(strings.ReplaceAll(feature, "-", "_"))}
}
