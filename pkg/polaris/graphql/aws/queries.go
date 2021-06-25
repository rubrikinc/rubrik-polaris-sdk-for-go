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

// allAwsCloudAccounts GraphQL query
var allAwsCloudAccountsQuery = `query SdkGolangAllAwsCloudAccounts($feature: CloudAccountFeatureEnum!, $columnSearchFilter: String!) {
    allAwsCloudAccounts(awsCloudAccountsArg: {columnSearchFilter: $columnSearchFilter, statusFilters: [], feature: $feature}) {
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
    allAwsExocomputeConfigs(awsNativeAccountIdOrNamePrefix: $awsNativeAccountIdOrNamePrefix) {
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
        featureDetails {
            feature
            roleArn
            stackArn
            status
            awsRegions
        }
    }
}`

// awsCloudAccountSelector GraphQL query
var awsCloudAccountSelectorQuery = `query SdkGolangAwsCloudAccountSelector($cloudAccountId: UUID!, $feature: CloudAccountFeatureEnum!) {
    awsCloudAccountSelector(cloudAccountId: $cloudAccountId, awsCloudAccountArg: {features: [$feature]}) {
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
var awsNativeAccountQuery = `query SdkGolangAwsNativeAccount($fid: UUID!, $awsNativeProtectionFeature: AwsNativeProtectionFeatureEnum!) {
	awsNativeAccount(awsNativeProtectionFeature: $awsNativeProtectionFeature, fid: $fid) {
		id
		regions
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
var awsNativeAccountsQuery = `query SdkGolangAwsNativeAccounts($after: String, $awsNativeProtectionFeature: AwsNativeProtectionFeatureEnum!, $filter: String) {
	awsNativeAccounts(after: $after, awsNativeProtectionFeature: $awsNativeProtectionFeature, accountFilters: {nameSubstringFilter: {nameSubstring: $filter}}) {
		count
		edges {
			node {
				id
				regions
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
var finalizeAwsCloudAccountDeletionQuery = `mutation SdkGolangFinalizeAwsCloudAccountDeletion($cloudAccountId: UUID!, $feature: CloudAccountFeatureEnum!) {
    finalizeAwsCloudAccountDeletion(input: {cloudAccountId: $cloudAccountId, feature: $feature}) {
        message
    }
}`

// finalizeAwsCloudAccountProtection GraphQL query
var finalizeAwsCloudAccountProtectionQuery = `mutation SdkGolangFinalizeAwsCloudAccountProtection($nativeId: String!, $accountName: String!, $awsRegions: [AwsCloudAccountRegionEnum!], $externalId: String!, $featureVersion: [AwsCloudAccountFeatureVersionInput!]!, $feature: CloudAccountFeatureEnum!, $stackName: String!) {
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
var prepareAwsCloudAccountDeletionQuery = `mutation SdkGolangPrepareAwsCloudAccountDeletion($cloudAccountId: UUID!, $feature: CloudAccountFeatureEnum!) {
    prepareAwsCloudAccountDeletion(input: {cloudAccountId: $cloudAccountId, feature: $feature}) {
        cloudFormationUrl
    }
}`

// startAwsNativeAccountDisableJob GraphQL query
var startAwsNativeAccountDisableJobQuery = `mutation SdkGolangStartAwsNativeAccountDisableJob($awsAccountRubrikId: UUID!, $awsNativeProtectionFeature: AwsNativeProtectionFeatureEnum!, $shouldDeleteNativeSnapshots: Boolean!) {
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
var updateAwsCloudAccountQuery = `mutation SdkGolangUpdateAwsCloudAccount($action: CloudAccountActionEnum!, $cloudAccountId: UUID!, $awsRegions: [AwsCloudAccountRegionEnum!]!, $feature: CloudAccountFeatureEnum!) {
    updateAwsCloudAccount(input: {action: $action, cloudAccountId: $cloudAccountId, awsRegions: $awsRegions, feature: $feature}) {
        message
    }
}`

// validateAndCreateAwsCloudAccount GraphQL query
var validateAndCreateAwsCloudAccountQuery = `mutation SdkGolangValidateAndCreateAwsCloudAccount($nativeId: String!, $accountName: String!, $feature: CloudAccountFeatureEnum!) {
    validateAndCreateAwsCloudAccount(input: {
        action: CREATE, 
        awsChildAccounts: [{
            accountName: $accountName,
            nativeId: $nativeId,
        }], 
        features: [$feature]
    }) {
        initiateResponse {
            cloudFormationUrl
            externalId
            featureVersionList {
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
