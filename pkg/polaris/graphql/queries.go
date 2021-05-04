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

package graphql

// awsCloudAccountDeleteInitiate GraphQL query
var awsCloudAccountDeleteInitiateQuery = `mutation SdkGolangAwsCloudAccountDeleteInitiate($polarisAccountId: UUID!) {
    awsCloudAccountDeleteInitiate(cloudAccountUuid: $polarisAccountId, awsCloudAccountDeleteInitiateArg: {feature: CLOUD_NATIVE_PROTECTION}) {
        cloudFormationUrl
    }
}`

// awsCloudAccountDeleteProcess GraphQL query
var awsCloudAccountDeleteProcessQuery = `mutation SdkGolangAwsCloudAccountDeleteProcess($polarisAccountId: UUID!) {
    awsCloudAccountDeleteProcess(cloudAccountUuid: $polarisAccountId, awsCloudAccountDeleteProcessArg: {feature: CLOUD_NATIVE_PROTECTION}) {
        message
    }
}`

// awsCloudAccountSave GraphQL query
var awsCloudAccountSaveQuery = `mutation SdkGolangAwsCloudAccountSave($polarisAccountId: UUID!, $awsRegions: [AwsCloudAccountRegionEnum!]!) {
  awsCloudAccountSave(cloudAccountUuid: $polarisAccountId, awsCloudAccountSaveArg: {action: UPDATE_REGIONS, feature: CLOUD_NATIVE_PROTECTION, awsRegions: $awsRegions}) {
    message
  }
}`

// awsCloudAccountUpdateFeatureInitiate GraphQL query
var awsCloudAccountUpdateFeatureInitiateQuery = `mutation SdkGolangAwsCloudAccountUpdateFeatureInitiate($polarisAccountId: UUID!) {
    awsCloudAccountUpdateFeatureInitiate(cloudAccountUuid: $polarisAccountId, features: [CLOUD_NATIVE_PROTECTION]) {
        cloudFormationUrl
        templateUrl
    }
}`

// awsCloudAccounts GraphQL query
var awsCloudAccountsQuery = `query SdkGolangAwsCloudAccounts($columnFilter: String = "") {
    awsCloudAccounts(awsCloudAccountsArg: {columnSearchFilter: $columnFilter, statusFilters: [], feature: CLOUD_NATIVE_PROTECTION}) {
        awsCloudAccounts {
            awsCloudAccount {
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
    }
}`

// awsNativeAccounts GraphQL query
var awsNativeAccountsQuery = `query SdkGolangAwsNativeAccounts($awsNativeProtectionFeature: AwsNativeProtectionFeatureEnum = EC2, $nameFilter: String = "") {
	awsNativeAccounts(awsNativeProtectionFeature: $awsNativeProtectionFeature, accountFilters: {nameSubstringFilter: {nameSubstring: $nameFilter}}) {
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

// awsNativeProtectionAccountAdd GraphQL query
var awsNativeProtectionAccountAddQuery = `mutation SdkGolangAwsNativeProtectionAccountAdd($accountId: String!, $accountName: String!, $regions: [String!]!) {
    awsNativeProtectionAccountAdd(awsNativeProtectionAccountAddArg: {accountId: $accountId, name: $accountName, regions: $regions}) {
       cloudFormationName
       cloudFormationUrl
       cloudFormationTemplateUrl
       errorMessage
    }
}`

// coreTaskchainStatus GraphQL query
var coreTaskchainStatusQuery = `query SdkGolangCoreTaskchainStatus($taskchainId: String!){
    getKorgTaskchainStatus(taskchainId: $taskchainId){
        taskchain {
            id
            state
            taskchainUuid
            ... on Taskchain{
                progressedAt
            }
        }
    }
}`

// gcpCloudAccountAddManualAuthProject GraphQL query
var gcpCloudAccountAddManualAuthProjectQuery = `mutation SdkGolangGcpCloudAccountAddManualAuthProject($gcp_native_project_id: String!, $gcp_native_project_name: String!, $gcp_native_project_number: Long!, $organization_name: String, $service_account_auth_key: String)
{
    gcpCloudAccountAddManualAuthProject(
        gcpNativeProjectId: $gcp_native_project_id,
        gcpProjectName: $gcp_native_project_name,
        gcpProjectNumber: $gcp_native_project_number,
        organizationName: $organization_name,
        serviceAccountJwtConfigOptional: $service_account_auth_key,
        features: [CLOUD_NATIVE_PROTECTION]
    )
}`

// gcpCloudAccountDeleteProjects GraphQL query
var gcpCloudAccountDeleteProjectsQuery = `mutation SdkGolangGcpCloudAccountDeleteProjects($native_protection_ids: [UUID!]!, $shared_vpc_host_project_ids: [UUID!]!, $cloud_account_project_ids: [UUID!]!) {
  gcpCloudAccountDeleteProjects(nativeProtectionProjectUuids: $native_protection_ids, sharedVpcHostProjectUuids: $shared_vpc_host_project_ids, cloudAccountsProjectUuids: $cloud_account_project_ids, skipResourceDeletion: true) {
    projectUuid
    success
    error
  }
}`

// gcpCloudAccountListPermissions GraphQL query
var gcpCloudAccountListPermissionsQuery = `query SdkGolangGcpCloudAccountListPermissions($feature: CloudAccountFeatureEnum = CLOUD_NATIVE_PROTECTION) {
    gcpCloudAccountListPermissions(feature: $feature){
        permission
    }
}`

// gcpCloudAccountListProjects GraphQL query
var gcpCloudAccountListProjectsQuery = `query SdkGolangGcpCloudAccountListProjects($feature: CloudAccountFeatureEnum = CLOUD_NATIVE_PROTECTION, $search_text: String!, $status_filters: [CloudAccountStatusEnum!]!) {
    gcpCloudAccountListProjects(feature: $feature, projectStatusFilters: $status_filters, projectSearchText: $search_text){
        project{
            projectId,
            projectNumber,
            name,
            id
        }
        featureDetail{
            feature
            status
        }
    }
}`

// gcpNativeDisableProject GraphQL query
var gcpNativeDisableProjectQuery = `mutation SdkGolangGcpNativeDisableProject($rubrik_project_id: UUID!, $delete_snapshots: Boolean!) {
  gcpNativeDisableProject(projectId: $rubrik_project_id, shouldDeleteNativeSnapshots: $delete_snapshots) {
    taskchainUuid
  }
}`

// gcpNativeProjectConnection GraphQL query
var gcpNativeProjectConnectionQuery = `query SdkGolangGcpNativeProjectConnection($filter: String = "") {
    gcpNativeProjectConnection(projectFilters: {nameOrNumberSubstringFilter: {nameOrNumberSubstring: $filter}}){
        count
        edges {
            node {
                id
                name
                nativeName
                nativeId
                projectNumber
                organizationName
                slaAssignment
                configuredSlaDomain{
                    id
                    name
                }
                effectiveSlaDomain{
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

// gcpSetDefaultServiceAccount GraphQL query
var gcpSetDefaultServiceAccountQuery = `mutation SdkGolangGcpSetDefaultServiceAccount($jwt_config: String!, $account_name: String!)
{
    gcpSetDefaultServiceAccountJwtConfig(serviceAccountJWTConfig: $jwt_config, serviceAccountName: $account_name)
}`

// startAwsNativeAccountDisableJob GraphQL query
var startAwsNativeAccountDisableJobQuery = `mutation SdkGolangStartAwsNativeAccountDisableJob($polarisAccountId: UUID!, $deleteNativeSnapshots: Boolean = false, $awsNativeProtectionFeature: AwsNativeProtectionFeatureEnum = EC2) {
    startAwsNativeAccountDisableJob(input: {awsNativeAccountId: $polarisAccountId, shouldDeleteNativeSnapshots: $deleteNativeSnapshots, awsNativeProtectionFeature: $awsNativeProtectionFeature}) {
        error
        jobId
    }
}`
