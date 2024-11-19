// Code generated by queries_gen.go DO NOT EDIT.

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

package azure

// addAzureCloudAccountWithoutOauth GraphQL query
var addAzureCloudAccountWithoutOauthQuery = `mutation SdkGolangAddAzureCloudAccountWithoutOauth($tenantDomainName: String!, $azureCloudType: AzureCloudType!, $regions: [AzureCloudAccountRegion!]!, $feature: AddAzureCloudAccountFeatureInputWithoutOauth!, $subscriptionName: String!, $subscriptionId: String!) {
    result: addAzureCloudAccountWithoutOauth(input: {
        tenantDomainName: $tenantDomainName,
        azureCloudType:   $azureCloudType,
        subscriptions: {
            subscription: {
                name:     $subscriptionName,
                nativeId: $subscriptionId
            }
            features: [$feature]
        },
        regions: $regions,
    }) {
        tenantId
        status {
            azureSubscriptionRubrikId
            azureSubscriptionNativeId
            error
        }
    }
}`

// allAzureCloudAccountTenants GraphQL query
var allAzureCloudAccountTenantsQuery = `query SdkGolangAllAzureCloudAccountTenants($feature: CloudAccountFeature!, $includeSubscriptionDetails: Boolean!) {
    result: allAzureCloudAccountTenants(feature: $feature, includeSubscriptionDetails: $includeSubscriptionDetails) {
        cloudType
        azureCloudAccountTenantRubrikId
        clientId
        appName
        domainName
        subscriptionCount
        subscriptions {
            id
            name
            nativeId
            featureDetail {
                feature
                permissionsGroups
                status
                regions
                resourceGroup {
                    name
                    nativeId
                    region
                    tags {
                        key
                        value
                    }
                }
                userAssignedManagedIdentity {
                    name
                    nativeId
                    principalId
                }
            }
        }
    }
}`

// azureCloudAccountPermissionConfig GraphQL query
var azureCloudAccountPermissionConfigQuery = `query SdkGolangAzureCloudAccountPermissionConfig($feature: CloudAccountFeature!, $permissionsGroups: [PermissionsGroup!]!) {
    result: azureCloudAccountPermissionConfig(feature: $feature, permissionsGroups: $permissionsGroups) {
        permissionVersion
        permissionsGroupVersions {
            permissionsGroup
            version
        }
        resourceGroupRolePermissions {
            excludedActions
            excludedDataActions
            includedActions
            includedDataActions
        }
        rolePermissions {
            excludedActions
            excludedDataActions
            includedActions
            includedDataActions
        }
    }
}`

// azureNativeSubscriptions GraphQL query
var azureNativeSubscriptionsQuery = `query SdkGolangAzureNativeSubscriptions($after: String, $filter: String!) {
    result: azureNativeSubscriptions(after: $after, subscriptionFilters: {
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

// deleteAzureCloudAccountWithoutOauth GraphQL query
var deleteAzureCloudAccountWithoutOauthQuery = `mutation SdkGolangDeleteAzureCloudAccountWithoutOauth($subscriptionIds: [UUID!]!, $features: [CloudAccountFeature!]!) {
    result: deleteAzureCloudAccountWithoutOauth(input: {
        azureSubscriptionRubrikIds: $subscriptionIds
        features:                   $features,
    }) {
        status {
            azureSubscriptionNativeId
            isSuccess
            error
        }
    }
}`

// setAzureCloudAccountCustomerAppCredentials GraphQL query
var setAzureCloudAccountCustomerAppCredentialsQuery = `mutation SdkGolangSetAzureCloudAccountCustomerAppCredentials($azureCloudType: AzureCloudType!, $appId: String!, $appName: String, $appSecretKey: String!, $appTenantId: String, $tenantDomainName: String, $shouldReplace: Boolean!) {
    result: setAzureCloudAccountCustomerAppCredentials(input: {
        appId:            $appId,
        appSecretKey:     $appSecretKey,
        appTenantId:      $appTenantId,
        appName:          $appName,
        tenantDomainName: $tenantDomainName,
        shouldReplace:    $shouldReplace,
        azureCloudType:   $azureCloudType
    })
}`

// startDisableAzureCloudAccountJob GraphQL query
var startDisableAzureCloudAccountJobQuery = `mutation SdkGolangStartDisableAzureCloudAccountJob($cloudAccountId: UUID!, $feature: CloudAccountFeature!) {
  result: startDisableAzureCloudAccountJob(input: {
    feature:         $feature,
    cloudAccountIds: [$cloudAccountId],
  }) {
    jobIds {
      jobId
    }
    errors {
      error
    }
  }
}`

// startDisableAzureNativeSubscriptionProtectionJob GraphQL query
var startDisableAzureNativeSubscriptionProtectionJobQuery = `mutation SdkGolangStartDisableAzureNativeSubscriptionProtectionJob($azureSubscriptionRubrikId: UUID!, $shouldDeleteNativeSnapshots: Boolean!, $azureNativeProtectionFeature: AzureNativeProtectionFeature!) {
    result: startDisableAzureNativeSubscriptionProtectionJob(input: {
        azureSubscriptionRubrikId:    $azureSubscriptionRubrikId,
        shouldDeleteNativeSnapshots:  $shouldDeleteNativeSnapshots,
        azureNativeProtectionFeature: $azureNativeProtectionFeature,
    }) {
         jobId
     }
 }`

// updateAzureCloudAccount GraphQL query
var updateAzureCloudAccountQuery = `mutation SdkGolangUpdateAzureCloudAccount($features: [CloudAccountFeature!]!, $regionsToAdd: [AzureCloudAccountRegion!], $regionsToRemove: [AzureCloudAccountRegion!], $subscriptions: [AzureCloudAccountSubscriptionInput!]!) {
    result: updateAzureCloudAccount(input: {
        features:        $features,
        regionsToAdd:    $regionsToAdd,
        regionsToRemove: $regionsToRemove,
        subscriptions:   $subscriptions
    }) {
        status {
            azureSubscriptionNativeId
            isSuccess
        }
    }
}`

// upgradeAzureCloudAccountPermissionsWithoutOauth GraphQL query
var upgradeAzureCloudAccountPermissionsWithoutOauthQuery = `mutation SdkGolangUpgradeAzureCloudAccountPermissionsWithoutOauth($cloudAccountId: UUID!, $feature: CloudAccountFeature!) {
    result: upgradeAzureCloudAccountPermissionsWithoutOauth(input: {
        cloudAccountId: $cloudAccountId,
        feature:        $feature,
    }) {
        status
    }
}`

// upgradeAzureCloudAccountPermissionsWithoutOauthWithPermissionGroups GraphQL query
var upgradeAzureCloudAccountPermissionsWithoutOauthWithPermissionGroupsQuery = `mutation SdkGolangUpgradeAzureCloudAccountPermissionsWithoutOauthWithPermissionGroups($cloudAccountId: UUID!, $feature: UpgradeAzureCloudAccountFeatureInput!) {
    result: upgradeAzureCloudAccountPermissionsWithoutOauth(input: {
        cloudAccountId:   $cloudAccountId,
        featureToUpgrade: [$feature],
    }) {
        status
    }
}`
