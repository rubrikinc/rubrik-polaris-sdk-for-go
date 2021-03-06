// Code generated by queries_gen.go DO NOT EDIT

// MIT License
//
// Copyright (c) 2021 Rubrik
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package aws

// allAwsCloudAccountsWithFeatures GraphQL query
var allAwsCloudAccountsWithFeaturesQuery = `query SdkGolangAllAwsCloudAccountsWithFeatures($feature: CloudAccountFeature!, $columnSearchFilter: String!) {
    result: allAwsCloudAccountsWithFeatures(awsCloudAccountsArg: {columnSearchFilter: $columnSearchFilter, statusFilters: [], feature: $feature}) {
        awsCloudAccount {
            cloudType
            id
            nativeId
            message
            accountName
        }
        featureDetails {
            feature
            roleArn
            stackArn
            status
            awsRegions
        }
    }
}`

// allAwsCloudAccountsWithFeaturesV0 GraphQL query
var allAwsCloudAccountsWithFeaturesV0Query = `query SdkGolangAllAwsCloudAccountsWithFeaturesV0($feature: CloudAccountFeatureEnum!, $columnSearchFilter: String!) {
    result: allAwsCloudAccountsWithFeatures(awsCloudAccountsArg: {columnSearchFilter: $columnSearchFilter, statusFilters: [], feature: $feature}) {
        awsCloudAccount {
            cloudType
            id
            nativeId
            message
            accountName
        }
        featureDetails {
            feature
            roleArn
            stackArn
            status
            awsRegions
        }
    }
}`

// allAwsExocomputeConfigs GraphQL query
var allAwsExocomputeConfigsQuery = `query SdkGolangAllAwsExocomputeConfigs($awsNativeAccountIdOrNamePrefix: String!) {
    result: allAwsExocomputeConfigs(awsNativeAccountIdOrNamePrefix: $awsNativeAccountIdOrNamePrefix) {
        awsCloudAccount {
            cloudType
            id
            nativeId
            message
            accountName
        }
        configs {
            areSecurityGroupsPolarisManaged
            clusterSecurityGroupId
            configUuid
            message
            nodeSecurityGroupId
            region
            subnet1 {
                availabilityZone
                subnetId
            }
            subnet2 {
                availabilityZone
                subnetId
             }
            vpcId
        }
        exocomputeEligibleRegions
        featureDetail {
            feature
            roleArn
            stackArn
            status
            awsRegions
        }
    }
}`

// allVpcsByRegionFromAws GraphQL query
var allVpcsByRegionFromAwsQuery = `query SdkGolangAllVpcsByRegionFromAws($awsAccountRubrikId: UUID!, $region: AwsNativeRegion!) {
    allVpcsByRegionFromAws(awsAccountRubrikId: $awsAccountRubrikId, region: $region) {
        id
        name
        subnets {
            id
            name
            availabilityZone
        }
        securityGroups {
            id
            name
        }
    }
}`

// allVpcsByRegionFromAwsV0 GraphQL query
var allVpcsByRegionFromAwsV0Query = `query SdkGolangAllVpcsByRegionFromAwsV0($awsAccountRubrikId: UUID!, $region: AwsNativeRegionEnum!) {
    allVpcsByRegionFromAws(awsAccountRubrikId: $awsAccountRubrikId, region: $region) {
        id
        name
        subnets {
            id
            name
            availabilityZone
        }
        securityGroups {
            id
            name
        }
    }
}`

// awsCloudAccountWithFeatures GraphQL query
var awsCloudAccountWithFeaturesQuery = `query SdkGolangAwsCloudAccountWithFeatures($cloudAccountId: UUID!, $features: [CloudAccountFeature!]!) {
    result: awsCloudAccountWithFeatures(cloudAccountId: $cloudAccountId, awsCloudAccountArg: {features: $features}) {
        awsCloudAccount {
            cloudType
            id
            nativeId
            message
            accountName
        }
        featureDetails {
            feature
            roleArn
            stackArn
            status
            awsRegions
        }
    }
}`

// awsCloudAccountWithFeaturesV0 GraphQL query
var awsCloudAccountWithFeaturesV0Query = `query SdkGolangAwsCloudAccountWithFeaturesV0($cloudAccountId: UUID!, $features: [CloudAccountFeatureEnum!]!) {
    result: awsCloudAccountWithFeatures(cloudAccountId: $cloudAccountId, awsCloudAccountArg: {features: $features}) {
        awsCloudAccount {
            cloudType
            id
            nativeId
            message
            accountName
        }
        featureDetails {
            feature
            roleArn
            stackArn
            status
            awsRegions
        }
    }
}`

// awsNativeAccount GraphQL query
var awsNativeAccountQuery = `query SdkGolangAwsNativeAccount($awsNativeAccountRubrikId: UUID!, $awsNativeProtectionFeature: AwsNativeProtectionFeature!) {
	awsNativeAccount(awsNativeAccountRubrikId: $awsNativeAccountRubrikId, awsNativeProtectionFeature: $awsNativeProtectionFeature) {
		id
		regionSpecs {
			region
			isExocomputeConfigured
		}
		status
		name
		slaAssignment
		configuredSlaDomain {
			id
			name
		}
		effectiveSlaDomain {
			id
			name
		}
	}
}`

// awsNativeAccountV0 GraphQL query
var awsNativeAccountV0Query = `query SdkGolangAwsNativeAccountV0($awsNativeAccountRubrikId: UUID!, $awsNativeProtectionFeature: AwsNativeProtectionFeatureEnum!) {
	awsNativeAccount(awsNativeAccountRubrikId: $awsNativeAccountRubrikId, awsNativeProtectionFeature: $awsNativeProtectionFeature) {
		id
		regionSpecs {
			region
			isExocomputeConfigured
		}
		status
		name
		slaAssignment
		configuredSlaDomain {
			id
			name
		}
		effectiveSlaDomain {
			id
			name
		}
	}
}`

// awsNativeAccounts GraphQL query
var awsNativeAccountsQuery = `query SdkGolangAwsNativeAccounts($after: String, $awsNativeProtectionFeature: AwsNativeProtectionFeature!, $filter: String!) {
	awsNativeAccounts(after: $after, awsNativeProtectionFeature: $awsNativeProtectionFeature, accountFilters: {nameSubstringFilter: {nameSubstring: $filter}}) {
		count
		edges {
			node {
				id
				regionSpecs {
					region
					isExocomputeConfigured
				}
				status
				name
				slaAssignment
				configuredSlaDomain {
					id
					name
				}
				effectiveSlaDomain {
					id
					name
				}
			}
		}
		pageInfo {
			endCursor
			hasNextPage
		}
	}
}`

// awsNativeAccountsV0 GraphQL query
var awsNativeAccountsV0Query = `query SdkGolangAwsNativeAccountsV0($after: String, $awsNativeProtectionFeature: AwsNativeProtectionFeatureEnum!, $filter: String!) {
	awsNativeAccounts(after: $after, awsNativeProtectionFeature: $awsNativeProtectionFeature, accountFilters: {nameSubstringFilter: {nameSubstring: $filter}}) {
		count
		edges {
			node {
				id
				regionSpecs {
					region
					isExocomputeConfigured
				}
				status
				name
				slaAssignment
				configuredSlaDomain {
					id
					name
				}
				effectiveSlaDomain {
					id
					name
				}
			}
		}
		pageInfo {
			endCursor
			hasNextPage
		}
	}
}`

// createAwsExocomputeConfigs GraphQL query
var createAwsExocomputeConfigsQuery = `mutation SdkGolangCreateAwsExocomputeConfigs($cloudAccountId: UUID!, $configs: [AwsExocomputeConfigInput!]!) {
    createAwsExocomputeConfigs(input: {cloudAccountId: $cloudAccountId, configs: $configs}) {
        configs {
            areSecurityGroupsPolarisManaged
            clusterSecurityGroupId
            configUuid
            message
            nodeSecurityGroupId
            region
            subnet1 {
                availabilityZone
                subnetId
            }
            subnet2 {
                availabilityZone
                subnetId
            }
            vpcId
        }
    }
}`

// deleteAwsExocomputeConfigs GraphQL query
var deleteAwsExocomputeConfigsQuery = `mutation SdkGolangDeleteAwsExocomputeConfigs($configIdsToBeDeleted: [UUID!]!) {
    deleteAwsExocomputeConfigs(input: {configIdsToBeDeleted: $configIdsToBeDeleted}) {
        deletionStatus {
            exocomputeConfigId
            success
        }
    }
}`

// finalizeAwsCloudAccountDeletion GraphQL query
var finalizeAwsCloudAccountDeletionQuery = `mutation SdkGolangFinalizeAwsCloudAccountDeletion($cloudAccountId: UUID!, $feature: CloudAccountFeature!) {
    finalizeAwsCloudAccountDeletion(input: {cloudAccountId: $cloudAccountId, feature: $feature}) {
        message
    }
}`

// finalizeAwsCloudAccountDeletionV0 GraphQL query
var finalizeAwsCloudAccountDeletionV0Query = `mutation SdkGolangFinalizeAwsCloudAccountDeletionV0($cloudAccountId: UUID!, $feature: CloudAccountFeatureEnum!) {
    finalizeAwsCloudAccountDeletion(input: {cloudAccountId: $cloudAccountId, feature: $feature}) {
        message
    }
}`

// finalizeAwsCloudAccountProtection GraphQL query
var finalizeAwsCloudAccountProtectionQuery = `mutation SdkGolangFinalizeAwsCloudAccountProtection($nativeId: String!, $accountName: String!, $awsRegions: [AwsCloudAccountRegion!], $externalId: String!, $featureVersion: [AwsCloudAccountFeatureVersionInput!]!, $feature: CloudAccountFeature!, $stackName: String!) {
    finalizeAwsCloudAccountProtection(input: {
        action: CREATE,
        awsChildAccounts: [{
            accountName: $accountName,
            nativeId: $nativeId,
        }],
        awsRegions: $awsRegions,
        externalId: $externalId,
        featureVersion: $featureVersion,
        features: [$feature],
        stackName: $stackName,
    }) {
       awsChildAccounts {
           accountName
           nativeId
           message
       }
       message
    }
}`

// finalizeAwsCloudAccountProtectionV0 GraphQL query
var finalizeAwsCloudAccountProtectionV0Query = `mutation SdkGolangFinalizeAwsCloudAccountProtectionV0($nativeId: String!, $accountName: String!, $awsRegions: [AwsCloudAccountRegionEnum!], $externalId: String!, $featureVersion: [AwsCloudAccountFeatureVersionInput!]!, $feature: CloudAccountFeatureEnum!, $stackName: String!) {
    finalizeAwsCloudAccountProtection(input: {
        action: CREATE,
        awsChildAccounts: [{
            accountName: $accountName,
            nativeId: $nativeId,
        }],
        awsRegions: $awsRegions,
        externalId: $externalId,
        featureVersion: $featureVersion,
        features: [$feature],
        stackName: $stackName,
    }) {
       awsChildAccounts {
           accountName
           nativeId
           message
       }
       message
    }
}`

// finalizeAwsCloudAccountProtectionV1 GraphQL query
var finalizeAwsCloudAccountProtectionV1Query = `mutation SdkGolangFinalizeAwsCloudAccountProtectionV1($nativeId: String!, $accountName: String!, $awsRegions: [AwsCloudAccountRegionEnum!], $externalId: String!, $featureVersion: [AwsCloudAccountFeatureVersionInput!]!, $feature: CloudAccountFeature!, $stackName: String!) {
    finalizeAwsCloudAccountProtection(input: {
        action: CREATE,
        awsChildAccounts: [{
            accountName: $accountName,
            nativeId: $nativeId,
        }],
        awsRegions: $awsRegions,
        externalId: $externalId,
        featureVersion: $featureVersion,
        features: [$feature],
        stackName: $stackName,
    }) {
       awsChildAccounts {
           accountName
           nativeId
           message
       }
       message
    }
}`

// prepareAwsCloudAccountDeletion GraphQL query
var prepareAwsCloudAccountDeletionQuery = `mutation SdkGolangPrepareAwsCloudAccountDeletion($cloudAccountId: UUID!, $feature: CloudAccountFeature!) {
    prepareAwsCloudAccountDeletion(input: {cloudAccountId: $cloudAccountId, feature: $feature}) {
        cloudFormationUrl
    }
}`

// prepareAwsCloudAccountDeletionV0 GraphQL query
var prepareAwsCloudAccountDeletionV0Query = `mutation SdkGolangPrepareAwsCloudAccountDeletionV0($cloudAccountId: UUID!, $feature: CloudAccountFeatureEnum!) {
    prepareAwsCloudAccountDeletion(input: {cloudAccountId: $cloudAccountId, feature: $feature}) {
        cloudFormationUrl
    }
}`

// prepareFeatureUpdateForAwsCloudAccount GraphQL query
var prepareFeatureUpdateForAwsCloudAccountQuery = `mutation SdkGolangPrepareFeatureUpdateForAwsCloudAccount($cloudAccountId: UUID!, $features: [CloudAccountFeature!]!) {
    result: prepareFeatureUpdateForAwsCloudAccount(input: {cloudAccountId: $cloudAccountId, features: $features}) {
        cloudFormationUrl
        templateUrl
    }
}`

// prepareFeatureUpdateForAwsCloudAccountV0 GraphQL query
var prepareFeatureUpdateForAwsCloudAccountV0Query = `mutation SdkGolangPrepareFeatureUpdateForAwsCloudAccountV0($cloudAccountId: UUID!, $features: [CloudAccountFeatureEnum!]!) {
    result: prepareFeatureUpdateForAwsCloudAccount(input: {cloudAccountId: $cloudAccountId, features: $features}) {
        cloudFormationUrl
        templateUrl
    }
}`

// startAwsExocomputeDisableJob GraphQL query
var startAwsExocomputeDisableJobQuery = `mutation SdkGolangStartAwsExocomputeDisableJob($cloudAccountId: UUID!) {
    result: startAwsExocomputeDisableJob(input: {cloudAccountId: $cloudAccountId}) {
        error
        jobId
    }
}`

// startAwsNativeAccountDisableJob GraphQL query
var startAwsNativeAccountDisableJobQuery = `mutation SdkGolangStartAwsNativeAccountDisableJob($awsAccountRubrikId: UUID!, $awsNativeProtectionFeature: AwsNativeProtectionFeature!, $shouldDeleteNativeSnapshots: Boolean!) {
    startAwsNativeAccountDisableJob(input: {
        awsAccountRubrikId:          $awsAccountRubrikId,
        shouldDeleteNativeSnapshots: $shouldDeleteNativeSnapshots,
        awsNativeProtectionFeature:  $awsNativeProtectionFeature
    }) {
        error
        jobId
    }
}`

// startAwsNativeAccountDisableJobV0 GraphQL query
var startAwsNativeAccountDisableJobV0Query = `mutation SdkGolangStartAwsNativeAccountDisableJobV0($awsAccountRubrikId: UUID!, $awsNativeProtectionFeature: AwsNativeProtectionFeatureEnum!, $shouldDeleteNativeSnapshots: Boolean!) {
    startAwsNativeAccountDisableJob(input: {
        awsAccountRubrikId:          $awsAccountRubrikId,
        shouldDeleteNativeSnapshots: $shouldDeleteNativeSnapshots,
        awsNativeProtectionFeature:  $awsNativeProtectionFeature
    }) {
        error
        jobId
    }
}`

// updateAwsCloudAccount GraphQL query
var updateAwsCloudAccountQuery = `mutation SdkGolangUpdateAwsCloudAccount($action: CloudAccountAction!, $cloudAccountId: UUID!, $awsRegions: [AwsCloudAccountRegion!]!, $feature: CloudAccountFeature!) {
    result: updateAwsCloudAccountFeature(input: {action: $action, cloudAccountId: $cloudAccountId, awsRegions: $awsRegions, feature: $feature}) {
        message
    }
}`

// updateAwsCloudAccountV0 GraphQL query
var updateAwsCloudAccountV0Query = `mutation SdkGolangUpdateAwsCloudAccountV0($action: CloudAccountActionEnum!, $cloudAccountId: UUID!, $awsRegions: [AwsCloudAccountRegionEnum!]!, $feature: CloudAccountFeatureEnum!) {
    result: updateAwsCloudAccount(input: {action: $action, cloudAccountId: $cloudAccountId, awsRegions: $awsRegions, feature: $feature}) {
        message
    }
}`

// updateAwsCloudAccountV1 GraphQL query
var updateAwsCloudAccountV1Query = `mutation SdkGolangUpdateAwsCloudAccountV1($action: CloudAccountActionEnum!, $cloudAccountId: UUID!, $awsRegions: [AwsCloudAccountRegionEnum!]!, $feature: CloudAccountFeature!) {
    result: updateAwsCloudAccount(input: {action: $action, cloudAccountId: $cloudAccountId, awsRegions: $awsRegions, feature: $feature}) {
        message
    }
}`

// updateAwsCloudAccountV2 GraphQL query
var updateAwsCloudAccountV2Query = `mutation SdkGolangUpdateAwsCloudAccountV2($action: CloudAccountAction!, $cloudAccountId: UUID!, $awsRegions: [AwsCloudAccountRegion!]!, $feature: CloudAccountFeature!) {
    result: updateAwsCloudAccount(input: {action: $action, cloudAccountId: $cloudAccountId, awsRegions: $awsRegions, feature: $feature}) {
        message
    }
}`

// validateAndCreateAwsCloudAccount GraphQL query
var validateAndCreateAwsCloudAccountQuery = `mutation SdkGolangValidateAndCreateAwsCloudAccount($nativeId: String!, $accountName: String!, $features: [CloudAccountFeature!]!) {
    result: validateAndCreateAwsCloudAccount(input: {
        action: CREATE,
        awsChildAccounts: [{
            accountName: $accountName,
            nativeId: $nativeId,
        }],
        features: $features
    }) {
        initiateResponse {
            cloudFormationUrl
            externalId
            featureVersions {
                feature
                version
            }
            stackName
            templateUrl
        }
        validateResponse {
            invalidAwsAccounts {
                accountName
                nativeId
                message
            }
            invalidAwsAdminAccount {
                accountName
                nativeId
                message
            }
        }
    }
}`

// validateAndCreateAwsCloudAccountV0 GraphQL query
var validateAndCreateAwsCloudAccountV0Query = `mutation SdkGolangValidateAndCreateAwsCloudAccountV0($nativeId: String!, $accountName: String!, $features: [CloudAccountFeatureEnum!]!) {
    result: validateAndCreateAwsCloudAccount(input: {
        action: CREATE,
        awsChildAccounts: [{
            accountName: $accountName,
            nativeId: $nativeId,
        }],
        features: $features
    }) {
        initiateResponse {
            cloudFormationUrl
            externalId
            featureVersions {
                feature
                version
            }
            stackName
            templateUrl
        }
        validateResponse {
            invalidAwsAccounts {
                accountName
                nativeId
                message
            }
            invalidAwsAdminAccount {
                accountName
                nativeId
                message
            }
        }
    }
}`
