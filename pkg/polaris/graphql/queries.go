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

// awsAllCloudAccounts GraphQL query
var awsAllCloudAccountsQuery = `query SdkGolangAwsAllCloudAccounts($column_filter: String = "") {
    allAwsCloudAccounts(awsCloudAccountsArg: {columnSearchFilter: $column_filter, statusFilters: [], feature: CLOUD_NATIVE_PROTECTION}) {
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
}`

// awsCloudAccountSelector GraphQL query
var awsCloudAccountSelectorQuery = `query SdkGolangAwsCloudAccountSelector($cloud_account_id: String!) {
    awsCloudAccountSelector(awsCloudAccountsArg: {features: [CLOUD_NATIVE_PROTECTION], cloudAccountId: $cloud_account_id}) {
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
}`

// awsFinalizeCloudAccountDeletion GraphQL query
var awsFinalizeCloudAccountDeletionQuery = `mutation SdkGolangAwsFinalizeCloudAccountDeletion($cloud_account_uuid: UUID!) {
    finalizeAwsCloudAccountDeletion(input: {cloudAccountId: $cloud_account_uuid, feature: CLOUD_NATIVE_PROTECTION}) {
        message
    }
}`

// awsFinalizeCloudAccountProtection GraphQL query
var awsFinalizeCloudAccountProtectionQuery = `mutation SdkGolangAwsFinalizeCloudAccountProtection($account_name: String!, $aws_account_id: String!, $aws_regions: [AwsCloudAccountRegionEnum!], $external_id: String!, $feature_versions: [AwsCloudAccountFeatureVersionInput!]!, $stack_name: String!) {
    finalizeAwsCloudAccountProtection(input: {
        action: CREATE,
        awsChildAccounts: [{
            accountName: $account_name,
            nativeId: $aws_account_id,
        }],
        awsRegions: $aws_regions,
        externalId: $external_id,
        featureVersion: $feature_versions,
        features: [CLOUD_NATIVE_PROTECTION],
        stackName: $stack_name,
    }) {
       awsChildAccounts {
           accountName
           nativeId
           message
       }
       message
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

// awsPrepareCloudAccountDeletion GraphQL query
var awsPrepareCloudAccountDeletionQuery = `mutation SdkGolangAwsPrepareCloudAccountDeletion($cloud_account_uuid: UUID!) {
    prepareAwsCloudAccountDeletion(input: {cloudAccountId: $cloud_account_uuid, feature: CLOUD_NATIVE_PROTECTION}) {
        cloudFormationUrl
    }
}`

// awsStartNativeAccountDisableJob GraphQL query
var awsStartNativeAccountDisableJobQuery = `mutation SdkGolangAwsStartNativeAccountDisableJob($aws_account_rubrik_id: UUID!, $delete_native_snapshots: Boolean = false, $aws_native_protection_feature: AwsNativeProtectionFeatureEnum = EC2) {
    startAwsNativeAccountDisableJob(input: {
        awsAccountRubrikId:          $aws_account_rubrik_id,
        shouldDeleteNativeSnapshots: $delete_native_snapshots,
        awsNativeProtectionFeature:  $aws_native_protection_feature
    }) {
        error
        jobId
    }
}`

// awsUpdateCloudAccount GraphQL query
var awsUpdateCloudAccountQuery = `mutation SdkGolangAwsUpdateCloudAccount($cloud_account_uuid: UUID!, $aws_regions: [AwsCloudAccountRegionEnum!]!) {
  updateAwsCloudAccount(input: {cloudAccountId: $cloud_account_uuid, action: UPDATE_REGIONS, awsRegions: $aws_regions, feature: CLOUD_NATIVE_PROTECTION}) {
    message
  }
}`

// awsValidateAndCreateCloudAccount GraphQL query
var awsValidateAndCreateCloudAccountQuery = `mutation SdkGolangAwsValidateAndCreateCloudAccount($account_name: String!, $aws_account_id: String!) {
    validateAndCreateAwsCloudAccount(input: {
        action: CREATE, 
        awsChildAccounts: [{
            accountName: $account_name,
            nativeId: $aws_account_id,
        }], 
        features: [CLOUD_NATIVE_PROTECTION]
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

// azureAddCloudAccountWithoutOauth GraphQL query
var azureAddCloudAccountWithoutOauthQuery = `mutation SdkGolangAzureAddCloudAccountWithoutOauth($azure_tenant_domain_name: String!, $azure_cloud_type: AzureCloudTypeEnum!, $azure_regions: [AzureCloudAccountRegionEnum!]!, $feature: CloudAccountFeatureEnum!, $azure_subscriptions: [AzureSubscriptionInput!]!, $azure_policy_version: Int!) {
    result: addAzureCloudAccountWithoutOAuth(input: {
        tenantDomainName: $azure_tenant_domain_name,
        azureCloudType:   $azure_cloud_type,
        feature:          $feature,
        subscriptions:    $azure_subscriptions,
        regions:          $azure_regions,
        policyVersion:    $azure_policy_version
    }) {
        tenantId
        status {
            azureSubscriptionRubrikId
            azureSubscriptionNativeId
            error
        }
    }
}`

// azureAddCloudAccountWithoutOauthV0 GraphQL query
var azureAddCloudAccountWithoutOauthV0Query = `mutation SdkGolangAzureAddCloudAccountWithoutOauthV0($azure_tenant_domain_name: String!, $azure_cloud_type: AzureCloudTypeEnum!, $azure_regions: [AzureCloudAccountRegionEnum!]!, $feature: CloudAccountFeatureEnum!, $azure_subscriptions: [AzureSubscriptionInput!]!, $azure_policy_version: Int!) {
    result: azureCloudAccountAddWithoutOAuth(tenantDomainName: $azure_tenant_domain_name, azureCloudType: $azure_cloud_type, feature: $feature, subscriptions: $azure_subscriptions, regions: $azure_regions, policyVersion: $azure_policy_version) {
        tenantId
        status {
            azureSubscriptionRubrikId: subscriptionId
            azureSubscriptionNativeId: subscriptionNativeId
            error
        }
    }
}`

// azureCloudAccountPermissionConfig GraphQL query
var azureCloudAccountPermissionConfigQuery = `query SdkGolangAzureCloudAccountPermissionConfig($feature: CloudAccountFeatureEnum = CLOUD_NATIVE_PROTECTION) {
    azureCloudAccountPermissionConfig(feature: $feature) {
        permissionVersion
        rolePermissions {
            excludedActions
            excludedDataActions
            includedActions
            includedDataActions
        }
    }
}`

// azureCloudAccountTenants GraphQL query
var azureCloudAccountTenantsQuery = `query SdkGolangAzureCloudAccountTenants($feature: CloudAccountFeatureEnum!, $include_subscriptions: Boolean!) {
    result: azureCloudAccountTenants(feature: $feature, includeSubscriptionDetails: $include_subscriptions) {
        cloudType
        azureCloudAccountTenantRubrikId
        domainName
        subscriptionCount
        subscriptions {
            id
            name
            nativeId
            featureDetail {
                feature
                status
                regions
            }
        }
    }
}`

// azureCloudAccountTenantsV0 GraphQL query
var azureCloudAccountTenantsV0Query = `query SdkGolangAzureCloudAccountTenantsV0($feature: CloudAccountFeatureEnum!, $include_subscriptions: Boolean!) {
    result: azureCloudAccountTenants(feature: $feature, includeSubscriptionDetails: $include_subscriptions) {
        cloudType
        azureCloudAccountTenantRubrikId: id
        domainName
        subscriptionCount
        subscriptions {
            id
            name
            nativeId
            featureDetail {
                feature
                status
                regions
            }
        }
    }
}`

// azureDeleteCloudAccountWithoutOauth GraphQL query
var azureDeleteCloudAccountWithoutOauthQuery = `mutation SdkGolangAzureDeleteCloudAccountWithoutOauth($feature: CloudAccountFeatureEnum!, $azure_subscription_ids: [UUID!]!) {
    result: deleteAzureCloudAccountWithoutOAuth(input: {
        feature:                    $feature,
        azureSubscriptionRubrikIds: $azure_subscription_ids
    }) {
        status {
            azureSubscriptionNativeId
            isSuccess
            error
        }
    }
}`

// azureDeleteCloudAccountWithoutOauthV0 GraphQL query
var azureDeleteCloudAccountWithoutOauthV0Query = `mutation SdkGolangAzureDeleteCloudAccountWithoutOauthV0($feature: CloudAccountFeatureEnum!, $azure_subscription_ids: [UUID!]!) {
    result: azureCloudAccountDeleteWithoutOAuth(feature: $feature, subscriptionIds: $azure_subscription_ids) {
        status {
            azureSubscriptionNativeId: subscriptionId
            isSuccess: success
            error
        }
    }
}`

// azureNativeSubscriptions GraphQL query
var azureNativeSubscriptionsQuery = `query SdkGolangAzureNativeSubscriptions($filter: String = "") {
    result: azureNativeSubscriptions(subscriptionFilters: {
        nameSubstringFilter: {
            nameSubstring: $filter
        }
    }) {
        count
        edges {
            node {
                id
                azureSubscriptionNativeId
                name
                azureSubscriptionStatus
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

// azureNativeSubscriptionsV0 GraphQL query
var azureNativeSubscriptionsV0Query = `query SdkGolangAzureNativeSubscriptionsV0($filter: String = "") {
    result: azureNativeSubscriptionConnection(subscriptionFilters: {
        nameSubstringFilter: {
            nameSubstring: $filter
        }
    }) {
        count
        edges {
            node {
                id
                azureSubscriptionNativeId: nativeId
                name
                azureSubscriptionStatus: status
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

// azureSetCloudAccountCustomerAppCredentials GraphQL query
var azureSetCloudAccountCustomerAppCredentialsQuery = `mutation SdkGolangAzureSetCloudAccountCustomerAppCredentials($azure_app_id: String!, $azure_app_secret_key: String!, $azure_app_tenant_id: String, $azure_app_name: String, $azure_tenant_domain_name: String, $azure_cloud_type: AzureCloudTypeEnum!) {
    result: setAzureCloudAccountCustomerAppCredentials(input: {
        appId:            $azure_app_id,
        appSecretKey:     $azure_app_secret_key,
        appTenantId:      $azure_app_tenant_id,
        appName:          $azure_app_name,
        tenantDomainName: $azure_tenant_domain_name,
        azureCloudType:   $azure_cloud_type
    })
}`

// azureSetCloudAccountCustomerAppCredentialsV0 GraphQL query
var azureSetCloudAccountCustomerAppCredentialsV0Query = `mutation SdkGolangAzureSetCloudAccountCustomerAppCredentialsV0($azure_app_id: String!, $azure_app_secret_key: String!, $azure_app_tenant_id: String, $azure_app_name: String, $azure_tenant_domain_name: String, $azure_cloud_type: AzureCloudTypeEnum!) {
    result: azureSetCustomerAppCredentials(appId: $azure_app_id, appSecretKey: $azure_app_secret_key, appTenantId: $azure_app_tenant_id, appName: $azure_app_name, tenantDomainName: $azure_tenant_domain_name, azureCloudType: $azure_cloud_type)
}`

// azureStartDisableNativeSubscriptionProtectionJob GraphQL query
var azureStartDisableNativeSubscriptionProtectionJobQuery = `mutation SdkGolangAzureStartDisableNativeSubscriptionProtectionJob($subscription_id: UUID!, $delete_snapshots: Boolean!) {
    result: startDisableAzureNativeSubscriptionProtectionJob(input: {
        azureSubscriptionRubrikId:   $subscription_id,
        shouldDeleteNativeSnapshots: $delete_snapshots
    }) {
        jobId
    }
}`

// azureStartDisableNativeSubscriptionProtectionJobV0 GraphQL query
var azureStartDisableNativeSubscriptionProtectionJobV0Query = `mutation SdkGolangAzureStartDisableNativeSubscriptionProtectionJobV0($subscription_id: UUID!, $delete_snapshots: Boolean!) {
    result: deleteAzureNativeSubscription(subscriptionId: $subscription_id, shouldDeleteNativeSnapshots: $delete_snapshots) {
        jobId: taskchainUuid
    }
}`

// azureUpdateCloudAccount GraphQL query
var azureUpdateCloudAccountQuery = `mutation SdkGolangAzureUpdateCloudAccount($feature: CloudAccountFeatureEnum!, $regions_to_add: [AzureCloudAccountRegionEnum!], $regions_to_remove: [AzureCloudAccountRegionEnum!], $subscriptions: [AzureCloudAccountSubscriptionInput!]!) {
    result: updateAzureCloudAccount(input: {
        feature:         $feature,
        regionsToAdd:    $regions_to_add,
        regionsToRemove: $regions_to_remove,
        subscriptions:   $subscriptions
    }) {
        status {
            azureSubscriptionNativeId
            isSuccess
        }
    }
}`

// azureUpdateCloudAccountV0 GraphQL query
var azureUpdateCloudAccountV0Query = `mutation SdkGolangAzureUpdateCloudAccountV0($feature: CloudAccountFeatureEnum!, $regions_to_add: [AzureCloudAccountRegionEnum!], $regions_to_remove: [AzureCloudAccountRegionEnum!], $subscriptions: [AzureCloudAccountSubscriptionInput!]!) {
    result: azureCloudAccountUpdate(feature: $feature, regionsToAdd: $regions_to_add, regionsToRemove: $regions_to_remove, subscriptions: $subscriptions) {
        status {
            azureSubscriptionNativeId: subscriptionId
            isSuccess: success
        }
    }
}`

// coreDeploymentVersion GraphQL query
var coreDeploymentVersionQuery = `query SdkGolangCoreDeploymentVersion {
    deploymentVersion
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
    gcpCloudAccountListProjects(feature: $feature, projectStatusFilters: $status_filters, projectSearchText: $search_text) {
        project {
            projectId,
            projectNumber,
            name,
            id
        }
        featureDetail {
            feature
            status
        }
    }
}`

// gcpGetDefaultCredentialsServiceAccount GraphQL query
var gcpGetDefaultCredentialsServiceAccountQuery = `query SdkGolangGcpGetDefaultCredentialsServiceAccount {
    gcpGetDefaultCredentialsServiceAccount
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
