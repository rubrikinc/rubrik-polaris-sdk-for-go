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
	"fmt"
	"slices"
	"strings"
)

const (
	All                                         = "ALL"
	AlloyDBProtection                           = "ALLOY_DB_PROTECTION"
	AppFlows                                    = "APP_FLOWS"
	Archival                                    = "ARCHIVAL"
	AWSKMSKeySharing                            = "AWS_KMS_KEY_SHARING"
	AzureCosmosNoSQLProtection                  = "AZURE_COSMOS_NOSQL_PROTECTION"
	AzureDevOpsArtifactsProtection              = "AZURE_DEVOPS_ARTIFACTS_PROTECTION"
	AzureDevOpsDeveloperCollaborationProtection = "AZURE_DEVOPS_DEVELOPER_COLLABORATION_PROTECTION"
	AzureDevOpsProtection                       = "AZURE_DEVOPS_PROTECTION"
	AzureDevOpsRepositoryProtection             = "AZURE_DEVOPS_REPOSITORY_PROTECTION"
	AzureLocalCloudAccount                      = "AZURE_LOCAL_CLOUD_ACCOUNT"
	AzurePostgresFlexibleServerProtection       = "AZURE_POSTGRES_FLEXIBLE_SERVER_PROTECTION"
	AzureSQLDBProtection                        = "AZURE_SQL_DB_PROTECTION"
	AzureSQLMIProtection                        = "AZURE_SQL_MI_PROTECTION"
	CCESBaaS                                    = "CCES_BAAS"
	CloudAccounts                               = "CLOUDACCOUNTS"
	CloudCostReport                             = "CLOUD_COST_REPORT"
	CloudDiscovery                              = "CLOUD_DISCOVERY"
	CloudNativeArchival                         = "CLOUD_NATIVE_ARCHIVAL"
	CloudNativeArchivalEncryption               = "CLOUD_NATIVE_ARCHIVAL_ENCRYPTION"
	CloudNativeBlobProtection                   = "CLOUD_NATIVE_BLOB_PROTECTION"
	CloudNativeConfigProtection                 = "CLOUD_NATIVE_CONFIG_PROTECTION"
	CloudNativeDynamoDBProtection               = "CLOUD_NATIVE_DYNAMODB_PROTECTION"
	CloudNativeProtection                       = "CLOUD_NATIVE_PROTECTION"
	CloudNativeS3Protection                     = "CLOUD_NATIVE_S3_PROTECTION"
	CloudNativeUEMKeyManagement                 = "CLOUD_NATIVE_UEM_KEY_MANAGEMENT"
	CloudSQLProtection                          = "CLOUD_SQL_PROTECTION"
	CriticalResourceProtection                  = "CRITICAL_RESOURCE_PROTECTION"
	CyberRecoveryDataClassificationData         = "CYBERRECOVERY_DATA_CLASSIFICATION_DATA"
	CyberRecoveryDataClassificationMetadata     = "CYBERRECOVERY_DATA_CLASSIFICATION_METADATA"
	DataCenterRoleBasedArchival                 = "DATA_CENTER_ROLE_BASED_ARCHIVAL"
	DSPMData                                    = "DSPM_DATA"
	DSPMMetadata                                = "DSPM_METADATA"
	Exocompute                                  = "EXOCOMPUTE"
	GCPBigQueryProtection                       = "GCP_BIGQUERY_PROTECTION"
	GCPBigQueryReservation                      = "GCP_BIGQUERY_RESERVATION"
	GCPSharedVPCHost                            = "GCP_SHARED_VPC_HOST"
	GitHubDeveloperCollaborationProtection      = "GITHUB_DEVELOPER_COLLABORATION_PROTECTION"
	GitHubRepositoryProtection                  = "GITHUB_REPOSITORY_PROTECTION"
	GlueIcebergProtection                       = "GLUE_ICEBERG_PROTECTION"
	KubernetesProtection                        = "KUBERNETES_PROTECTION"
	LaminarCrossAccount                         = "LAMINAR_CROSS_ACCOUNT"
	LaminarInternal                             = "LAMINAR_INTERNAL"
	LaminarOutpostApplication                   = "LAMINAR_OUTPOST_APPLICATION"
	LaminarOutpostManagedIdentity               = "LAMINAR_OUTPOST_MANAGED_IDENTITY"
	LaminarTargetApplication                    = "LAMINAR_TARGET_APPLICATION"
	LaminarTargetManagedIdentity                = "LAMINAR_TARGET_MANAGED_IDENTITY"
	Outpost                                     = "OUTPOST"
	RDSProtection                               = "RDS_PROTECTION"
	RoleChaining                                = "ROLE_CHAINING"
	S3TablesIcebergProtection                   = "S3_TABLES_ICEBERG_PROTECTION"
	ServerAndApps                               = "SERVERS_AND_APPS"
)

// AWSCloudFormationFeatures returns the features, with permission groups, which
// can be onboarded for an AWS account using the CloudFormation workflow, sorted
// by name.
func AWSCloudFormationFeatures() []Feature {
	return featuresFor(awsAccountCFT)
}

// AWSCloudFormationFeatureNames returns the names of the features returned by
// AWSCloudFormationFeatures.
func AWSCloudFormationFeatureNames() []string {
	return FeatureNames(AWSCloudFormationFeatures())
}

// ParseAWSCloudFormationFeature returns the feature, with permission groups,
// matching the specified feature name. Returns an error if the feature can't
// be onboarded for an AWS account using the CloudFormation workflow.
func ParseAWSCloudFormationFeature(featureName string) (Feature, error) {
	if feature, ok := LookupFeatureName(AWSCloudFormationFeatures(), featureName); ok {
		return feature, nil
	}

	return FeatureInvalid, fmt.Errorf("invalid AWS CloudFormation feature: %s", featureName)
}

// AWSIAMRolesFeatures returns the features, with permission groups, which can
// be onboarded for an AWS account using the IAM roles workflow, sorted by name.
func AWSIAMRolesFeatures() []Feature {
	return featuresFor(awsAccountIAM)
}

// AWSIAMRolesFeatureNames returns the names of the features returned by
// AWSIAMRolesFeatures.
func AWSIAMRolesFeatureNames() []string {
	return FeatureNames(AWSIAMRolesFeatures())
}

// ParseAWSIAMRolesFeature returns the feature, with permission groups, matching
// the specified feature name. Returns an error if the feature can't be
// onboarded for an AWS account using the IAM roles workflow.
func ParseAWSIAMRolesFeature(featureName string) (Feature, error) {
	if feature, ok := LookupFeatureName(AWSIAMRolesFeatures(), featureName); ok {
		return feature, nil
	}

	return FeatureInvalid, fmt.Errorf("invalid AWS IAM roles feature: %s", featureName)
}

// AWSManagedFeatures returns the features, with permission groups, which can be
// onboarded for an RSC-managed (BaaS) AWS account, sorted by name.
func AWSManagedFeatures() []Feature {
	return featuresFor(awsManaged)
}

// AWSManagedFeatureNames returns the names of the features returned by
// AWSManagedFeatures.
func AWSManagedFeatureNames() []string {
	return FeatureNames(AWSManagedFeatures())
}

// ParseAWSManagedFeature returns the feature, with permission groups, matching
// the specified feature name. Returns an error if the feature can't be
// onboarded for an RSC-managed (BaaS) AWS account.
func ParseAWSManagedFeature(featureName string) (Feature, error) {
	if feature, ok := LookupFeatureName(AWSManagedFeatures(), featureName); ok {
		return feature, nil
	}

	return FeatureInvalid, fmt.Errorf("invalid AWS managed feature: %s", featureName)
}

// AzureSubscriptionFeatures returns the features, with permission groups, which
// can be onboarded for an Azure subscription, sorted by name.
func AzureSubscriptionFeatures() []Feature {
	return featuresFor(azureSubscription)
}

// AzureSubscriptionFeatureNames returns the names of the features returned by
// AzureSubscriptionFeatures.
func AzureSubscriptionFeatureNames() []string {
	return FeatureNames(AzureSubscriptionFeatures())
}

// ParseAzureSubscriptionFeature returns the feature, with permission groups,
// matching the specified feature name. Returns an error if the feature can't
// be onboarded for an Azure subscription.
func ParseAzureSubscriptionFeature(featureName string) (Feature, error) {
	if feature, ok := LookupFeatureName(AzureSubscriptionFeatures(), featureName); ok {
		return feature, nil
	}

	return FeatureInvalid, fmt.Errorf("invalid Azure subscription feature: %s", featureName)
}

// AzureDevOpsOrganizationFeatures returns the features, with permission groups,
// which can be onboarded for an Azure DevOps organization, sorted by name.
func AzureDevOpsOrganizationFeatures() []Feature {
	return featuresFor(azureDevOpsOrg)
}

// AzureDevOpsOrganizationFeatureNames returns the names of the features
// returned by AzureDevOpsOrganizationFeatures.
func AzureDevOpsOrganizationFeatureNames() []string {
	return FeatureNames(AzureDevOpsOrganizationFeatures())
}

// ParseAzureDevOpsOrganizationFeature returns the feature, with permission
// groups, matching the specified feature name. Returns an error if the feature
// can't be onboarded for an Azure DevOps organization.
func ParseAzureDevOpsOrganizationFeature(featureName string) (Feature, error) {
	if feature, ok := LookupFeatureName(AzureDevOpsOrganizationFeatures(), featureName); ok {
		return feature, nil
	}

	return FeatureInvalid, fmt.Errorf("invalid Azure DevOps organization feature: %s", featureName)
}

// GCPProjectFeatures returns the features, with permission groups, which can be
// onboarded for a GCP project, sorted by name.
func GCPProjectFeatures() []Feature {
	return featuresFor(gcpProject)
}

// GCPProjectFeatureNames returns the names of the features returned by
// GCPProjectFeatures.
func GCPProjectFeatureNames() []string {
	return FeatureNames(GCPProjectFeatures())
}

// ParseGCPProjectFeature returns the feature, with permission groups, matching
// the specified feature name. Returns an error if the feature can't be
// onboarded for a GCP project.
func ParseGCPProjectFeature(featureName string) (Feature, error) {
	if feature, ok := LookupFeatureName(GCPProjectFeatures(), featureName); ok {
		return feature, nil
	}

	return FeatureInvalid, fmt.Errorf("invalid GCP project feature: %s", featureName)
}

// GitHubOrganizationFeatures returns the features, with permission groups,
// which can be onboarded for a GitHub organization, sorted by name.
func GitHubOrganizationFeatures() []Feature {
	return featuresFor(githubOrg)
}

// GitHubOrganizationFeatureNames returns the names of the features returned by
// GitHubOrganizationFeatures.
func GitHubOrganizationFeatureNames() []string {
	return FeatureNames(GitHubOrganizationFeatures())
}

// ParseGitHubOrganizationFeature returns the feature, with permission groups,
// matching the specified feature name. Returns an error if the feature can't
// be onboarded for a GitHub organization.
func ParseGitHubOrganizationFeature(featureName string) (Feature, error) {
	if feature, ok := LookupFeatureName(GitHubOrganizationFeatures(), featureName); ok {
		return feature, nil
	}

	return FeatureInvalid, fmt.Errorf("invalid GitHub organization feature: %s", featureName)
}

// AWSAccountAllowlistNames returns the names of the features which are kept
// when reading AWS accounts from RSC, sorted by name. This includes features
// which are not public but allowed, e.g. CLOUD_COST_REPORT, which RSC enables
// on its own.
func AWSAccountAllowlistNames() []string {
	return allowlistFor(awsAccountCFT, awsAccountIAM, awsManaged)
}

// AzureSubscriptionAllowlistNames returns the names of the features which are
// kept when reading Azure subscriptions from RSC, sorted by name.
func AzureSubscriptionAllowlistNames() []string {
	return allowlistFor(azureSubscription)
}

// AzureDevOpsOrganizationAllowlistNames returns the names of the features which
// are kept when reading Azure DevOps organizations from RSC, sorted by name.
func AzureDevOpsOrganizationAllowlistNames() []string {
	return allowlistFor(azureDevOpsOrg)
}

// GCPProjectAllowlistNames returns the names of the features which are kept
// when reading GCP projects from RSC, sorted by name.
func GCPProjectAllowlistNames() []string {
	return allowlistFor(gcpProject)
}

// GitHubOrganizationAllowlistNames returns the names of the features which are
// kept when reading GitHub organizations from RSC, sorted by name.
func GitHubOrganizationAllowlistNames() []string {
	return allowlistFor(githubOrg)
}

// CloudVendorFeatureNames returns the names of the features which can be
// onboarded for any account type of the specified cloud vendor, sorted by name.
func CloudVendorFeatureNames(vendor CloudVendor) []string {
	var names []string
	switch vendor {
	case CloudVendorAWS:
		names = slices.Concat(AWSCloudFormationFeatureNames(), AWSIAMRolesFeatureNames(), AWSManagedFeatureNames())
	case CloudVendorAzure:
		names = slices.Concat(AzureSubscriptionFeatureNames(), AzureDevOpsOrganizationFeatureNames())
	case CloudVendorGCP:
		names = GCPProjectFeatureNames()
	default:
		return nil
	}

	slices.Sort(names)
	return slices.Compact(names)
}

// CloudVendorProtectionFeatureNames returns the names of the public protection
// features which can be onboarded for any account type of the specified cloud
// vendor, sorted by name.
func CloudVendorProtectionFeatureNames(vendor CloudVendor) []string {
	var accountTypes []cloudAccountType
	switch vendor {
	case CloudVendorAWS:
		accountTypes = []cloudAccountType{awsAccountCFT, awsAccountIAM, awsManaged}
	case CloudVendorAzure:
		accountTypes = []cloudAccountType{azureSubscription, azureDevOpsOrg}
	case CloudVendorGCP:
		accountTypes = []cloudAccountType{gcpProject}
	default:
		return nil
	}

	var names []string
	for name, info := range featureInfoMap {
		if !info.public || !info.protection {
			continue
		}
		if slices.ContainsFunc(accountTypes, func(accountType cloudAccountType) bool {
			_, ok := info.permissionGroups[accountType]
			return ok
		}) {
			names = append(names, name)
		}
	}
	slices.Sort(names)

	return names
}

// ValidateCloudVendorFeatureName returns an error if the specified feature name
// can't be onboarded for any account type of the specified cloud vendor.
func ValidateCloudVendorFeatureName(vendor CloudVendor, featureName string) error {
	if slices.Contains(CloudVendorFeatureNames(vendor), featureName) {
		return nil
	}

	return fmt.Errorf("invalid %s feature: %s", vendor, featureName)
}

// cloudAccountType identifies an RSC cloud account type, i.e. the kind of cloud
// account and the workflow used to onboard it. Cloud account types of the same
// cloud vendor can support different features and permission groups.
type cloudAccountType int

const (
	awsAccountCFT cloudAccountType = iota
	awsAccountIAM
	awsManaged
	azureSubscription
	azureDevOpsOrg
	gcpProject
	githubOrg
)

// cloudAccountPermissionGroups maps cloud account types to the permission
// groups a feature supports for each cloud account type.
type cloudAccountPermissionGroups map[cloudAccountType][]PermissionGroup

// featureInfoMap holds information about each RSC feature known to the SDK,
// keyed by feature name.
//
// permissionGroups holds the permission groups supported by the feature for
// each cloud account type. An empty list of permission groups means the cloud
// account type supports the feature, but the feature has no permission groups.
// A missing cloud account type means the cloud account type doesn't support
// the feature.
//
// allowed is true if a feature which is not public is still kept when reading
// cloud accounts of the supported cloud account types from RSC. Public
// features are always allowed.
//
// protection is true if the feature is a protection feature.
//
// public is true if the feature is exposed by the SDK. Features which are not
// public are known to the SDK but can't be used yet.
var featureInfoMap = map[string]struct {
	permissionGroups map[cloudAccountType][]PermissionGroup

	// Feature properties.
	allowed    bool
	protection bool
	public     bool
}{
	// Not an RSC feature, refers to all features of a cloud account.
	All: {},
	AlloyDBProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			gcpProject: {
				PermissionGroupBasic,
				PermissionGroupExportAndRestore,
			},
		},
		protection: true,
	},
	// Internal. Permission groups are behind feature flags. Not onboarded
	// through any of the cloud account types.
	AppFlows: {},
	// Data Center (CDM) archival feature, not onboarded through any of the
	// cloud account types but kept when reading AWS accounts.
	Archival: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {},
			awsAccountIAM: {},
		},
		allowed: true,
	},
	AWSKMSKeySharing: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {},
			awsAccountIAM: {},
		},
	},
	AzureCosmosNoSQLProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			azureSubscription: {
				PermissionGroupBasic,
			},
		},
		protection: true,
	},
	AzureDevOpsArtifactsProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			azureDevOpsOrg: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
		},
		protection: true,
	},
	AzureDevOpsDeveloperCollaborationProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			azureDevOpsOrg: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
		},
		protection: true,
	},
	// Parent feature of the other Azure DevOps features, not onboarded through
	// any of the cloud account types.
	AzureDevOpsProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			azureDevOpsOrg: {
				PermissionGroupBasic,
			},
		},
		protection: true,
	},
	AzureDevOpsRepositoryProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			azureDevOpsOrg: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
		},
		protection: true,
		public:     true,
	},
	AzureLocalCloudAccount: {
		permissionGroups: cloudAccountPermissionGroups{
			azureSubscription: {
				PermissionGroupBasic,
			},
		},
	},
	AzurePostgresFlexibleServerProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			azureSubscription: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
		},
		protection: true,
		public:     true,
	},
	AzureSQLDBProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			azureSubscription: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
				PermissionGroupBackupV2,
			},
		},
		protection: true,
		public:     true,
	},
	AzureSQLMIProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			azureSubscription: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
				PermissionGroupBackupV2,
			},
		},
		protection: true,
		public:     true,
	},
	// Not onboarded through any of the cloud account types.
	CCESBaaS: {},
	// Internal. Not onboarded through any of the cloud account types.
	CloudAccounts: {},
	CloudCostReport: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
			},
		},
		allowed: true,
	},
	CloudDiscovery: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
			},
			awsManaged: {},
			azureSubscription: {
				PermissionGroupBasic,
			},
		},
		protection: true,
		public:     true,
	},
	CloudNativeArchival: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
			},
			azureSubscription: {
				PermissionGroupBasic,
				PermissionGroupEncryption,
				PermissionGroupSQLArchival,
			},
			gcpProject: {
				PermissionGroupBasic,
				PermissionGroupEncryption,
			},
		},
		public: true,
	},
	CloudNativeArchivalEncryption: {
		permissionGroups: cloudAccountPermissionGroups{
			azureSubscription: {
				PermissionGroupBasic,
				PermissionGroupEncryption,
			},
		},
		public: true,
	},
	CloudNativeBlobProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			azureSubscription: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
		},
		protection: true,
		public:     true,
	},
	CloudNativeConfigProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
				PermissionGroupBasic2,
				PermissionGroupRecovery,
				PermissionGroupRecovery2,
				PermissionGroupRecovery3,
				PermissionGroupRecovery4,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
				PermissionGroupBasic2,
				PermissionGroupRecovery,
				PermissionGroupRecovery2,
				PermissionGroupRecovery3,
				PermissionGroupRecovery4,
			},
		},
		protection: true,
		public:     true,
	},
	CloudNativeDynamoDBProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
		},
		protection: true,
		public:     true,
	},
	CloudNativeProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
				PermissionGroupDownloadFile,
				PermissionGroupExportPowerOff,
				PermissionGroupExportPowerOn,
				PermissionGroupRestore,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
				PermissionGroupDownloadFile,
				PermissionGroupExportPowerOff,
				PermissionGroupExportPowerOn,
				PermissionGroupRestore,
			},
			awsManaged: {},
			azureSubscription: {
				PermissionGroupBasic,
				PermissionGroupCCES,
				PermissionGroupExportAndRestore,
				PermissionGroupExportAndRestorePowerOffVM,
				PermissionGroupFileLevelRecovery,
				PermissionGroupSnapshotPrivateAccess,
			},
			gcpProject: {
				PermissionGroupBasic,
				PermissionGroupExportAndRestore,
				PermissionGroupFileLevelRecovery,
			},
		},
		protection: true,
		public:     true,
	},
	CloudNativeS3Protection: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
				PermissionGroupExport,
				PermissionGroupRecovery,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
				PermissionGroupExport,
				PermissionGroupRecovery,
			},
			awsManaged: {},
		},
		protection: true,
		public:     true,
	},
	CloudNativeUEMKeyManagement: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
			},
			azureSubscription: {
				PermissionGroupBasic,
			},
			gcpProject: {
				PermissionGroupBasic,
			},
		},
	},
	CloudSQLProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			gcpProject: {
				PermissionGroupBasic,
				PermissionGroupExportAndRestore,
			},
		},
		protection: true,
		public:     true,
	},
	CriticalResourceProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
			},
		},
		protection: true,
	},
	CyberRecoveryDataClassificationData: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
		},
		public: true,
	},
	CyberRecoveryDataClassificationMetadata: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
		},
		public: true,
	},
	// Data Center (CDM) archival feature, not onboarded through any of the
	// cloud account types.
	DataCenterRoleBasedArchival: {},
	DSPMData: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
		},
		public: true,
	},
	DSPMMetadata: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
		},
		public: true,
	},
	Exocompute: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
				PermissionGroupRSCManagedCluster,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
				PermissionGroupRSCManagedCluster,
			},
			azureSubscription: {
				PermissionGroupBasic,
				PermissionGroupAKSCustomPrivateDNSZone,
				PermissionGroupAutomatedNetworkingSetup,
				PermissionGroupCustomerManagedCluster,
				PermissionGroupPrivateEndpoints,
				PermissionGroupServiceEndpointAutomation,
			},
			gcpProject: {
				PermissionGroupBasic,
				PermissionGroupAutomatedNetworkingSetup,
				PermissionGroupCloudSQL,
			},
		},
		public: true,
	},
	GCPBigQueryProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			gcpProject: {
				PermissionGroupBasic,
				PermissionGroupExportAndRestore,
			},
		},
		protection: true,
		public:     true,
	},
	GCPBigQueryReservation: {
		permissionGroups: cloudAccountPermissionGroups{
			gcpProject: {
				PermissionGroupBasic,
			},
		},
		public: true,
	},
	GCPSharedVPCHost: {
		permissionGroups: cloudAccountPermissionGroups{
			gcpProject: {
				PermissionGroupBasic,
				PermissionGroupCloudSQL,
			},
		},
		public: true,
	},
	GitHubDeveloperCollaborationProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			githubOrg: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
		},
		protection: true,
	},
	GitHubRepositoryProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			githubOrg: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
		},
		protection: true,
		public:     true,
	},
	GlueIcebergProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
		},
		protection: true,
	},
	KubernetesProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
			},
		},
		protection: true,
		public:     true,
	},
	LaminarCrossAccount: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
		},
		public: true,
	},
	LaminarInternal: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
		},
		public: true,
	},
	LaminarOutpostApplication: {
		permissionGroups: cloudAccountPermissionGroups{
			azureSubscription: {
				PermissionGroupBasic,
				PermissionGroupAutomatedNetworkingSetup,
			},
		},
	},
	LaminarOutpostManagedIdentity: {
		permissionGroups: cloudAccountPermissionGroups{
			azureSubscription: {
				PermissionGroupBasic,
			},
		},
	},
	LaminarTargetApplication: {
		permissionGroups: cloudAccountPermissionGroups{
			azureSubscription: {
				PermissionGroupBasic,
			},
		},
	},
	LaminarTargetManagedIdentity: {
		permissionGroups: cloudAccountPermissionGroups{
			azureSubscription: {
				PermissionGroupBasic,
			},
		},
	},
	Outpost: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
		},
		public: true,
	},
	RDSProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
			awsManaged: {},
		},
		protection: true,
		public:     true,
	},
	RoleChaining: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
			},
		},
		public: true,
	},
	S3TablesIcebergProtection: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
			awsAccountIAM: {
				PermissionGroupBasic,
				PermissionGroupRecovery,
			},
		},
		protection: true,
	},
	ServerAndApps: {
		permissionGroups: cloudAccountPermissionGroups{
			awsAccountCFT: {
				PermissionGroupCCES,
			},
			awsAccountIAM: {
				PermissionGroupCCES,
			},
			azureSubscription: {
				PermissionGroupCCES,
				PermissionGroupSAPHanaSSBasic,
				PermissionGroupSAPHanaSSRecovery,
			},
			gcpProject: {
				PermissionGroupCCES,
			},
		},
		public: true,
	},
}

// allowlistFor returns the names of the public or allowed features supported by
// any of the specified cloud account types, sorted by name.
func allowlistFor(accountTypes ...cloudAccountType) []string {
	var names []string
	for name, info := range featureInfoMap {
		if !info.public && !info.allowed {
			continue
		}
		if slices.ContainsFunc(accountTypes, func(accountType cloudAccountType) bool {
			_, ok := info.permissionGroups[accountType]
			return ok
		}) {
			names = append(names, name)
		}
	}
	slices.Sort(names)

	return names
}

// featuresFor returns the public features, with permission groups, supported by
// the specified cloud account type, sorted by name.
func featuresFor(accountType cloudAccountType) []Feature {
	var features []Feature
	for name, info := range featureInfoMap {
		if !info.public {
			continue
		}
		if groups, ok := info.permissionGroups[accountType]; ok {
			features = append(features, Feature{
				Name:             name,
				PermissionGroups: slices.Sorted(slices.Values(groups)),
			})
		}
	}
	slices.SortFunc(features, func(lhs, rhs Feature) int {
		return strings.Compare(lhs.Name, rhs.Name)
	})

	return features
}
