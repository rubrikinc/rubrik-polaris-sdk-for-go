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

package azure

// addAzureCloudAccountExocomputeConfigurations GraphQL query
var addAzureCloudAccountExocomputeConfigurationsQuery = `mutation SdkGolangAddAzureCloudAccountExocomputeConfigurations($cloudAccountId: UUID!, $azureExocomputeRegionConfigs: [AzureExocomputeAddConfigInputType!]!) {
    result: addAzureCloudAccountExocomputeConfigurations(input: {
        cloudAccountId: $cloudAccountId, azureExocomputeRegionConfigs: $azureExocomputeRegionConfigs
    }) {
        configs {
            configUuid
            isPolarisManaged
            message
            region
            subnetNativeId
        }
    }
}`

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
        regions:          $regions,
    }) {
        tenantId
        status {
            azureSubscriptionRubrikId
            azureSubscriptionNativeId
            error
        }
    }
}`

// addAzureCloudAccountWithoutOauthV0 GraphQL query
var addAzureCloudAccountWithoutOauthV0Query = `mutation SdkGolangAddAzureCloudAccountWithoutOauthV0($tenantDomainName: String!, $azureCloudType: AzureCloudTypeEnum!, $regions: [AzureCloudAccountRegionEnum!]!, $feature: CloudAccountFeatureEnum!, $subscriptionName: String!, $subscriptionId: String!, $policyVersion: Int!) {
    result: addAzureCloudAccountWithoutOAuth(input: {
        tenantDomainName: $tenantDomainName,
        azureCloudType:   $azureCloudType,
        features:         [$feature],
        subscriptions: {
            name:     $subscriptionName,
            nativeId: $subscriptionId
        },
        regions:          $regions,
        policyVersion:    $policyVersion
    }) {
        tenantId
        status {
            azureSubscriptionRubrikId
            azureSubscriptionNativeId
            error
        }
    }
}`

// addAzureCloudAccountWithoutOauthV1 GraphQL query
var addAzureCloudAccountWithoutOauthV1Query = `mutation SdkGolangAddAzureCloudAccountWithoutOauthV1($tenantDomainName: String!, $azureCloudType: AzureCloudTypeEnum!, $regions: [AzureCloudAccountRegionEnum!]!, $feature: CloudAccountFeatureEnum!, $subscriptionName: String!, $subscriptionId: String!, $policyVersion: Int!) {
    result: addAzureCloudAccountWithoutOAuth(input: {
        tenantDomainName: $tenantDomainName,
        azureCloudType:   $azureCloudType,
        subscriptions: {
            subscription: {
                name:     $subscriptionName,
                nativeId: $subscriptionId
            }
            features: [{
                featureType: $feature,
            }]
        },
        regions:          $regions,
        policyVersion:    $policyVersion
    }) {
        tenantId
        status {
            azureSubscriptionRubrikId
            azureSubscriptionNativeId
            error
        }
    }
}`

// addAzureCloudAccountWithoutOauthV2 GraphQL query
var addAzureCloudAccountWithoutOauthV2Query = `mutation SdkGolangAddAzureCloudAccountWithoutOauthV2($tenantDomainName: String!, $azureCloudType: AzureCloudTypeEnum!, $regions: [AzureCloudAccountRegionEnum!]!, $feature: AddAzureCloudAccountFeatureInputWithoutOauth!, $subscriptionName: String!, $subscriptionId: String!) {
    result: addAzureCloudAccountWithoutOAuth(input: {
        tenantDomainName: $tenantDomainName,
        azureCloudType:   $azureCloudType,
        subscriptions: {
            subscription: {
                name:     $subscriptionName,
                nativeId: $subscriptionId
            }
            features: [$feature]
        },
        regions:          $regions,
    }) {
        tenantId
        status {
            azureSubscriptionRubrikId
            azureSubscriptionNativeId
            error
        }
    }
}`

// addAzureCloudAccountWithoutOauthV3 GraphQL query
var addAzureCloudAccountWithoutOauthV3Query = `mutation SdkGolangAddAzureCloudAccountWithoutOauthV3($tenantDomainName: String!, $azureCloudType: AzureCloudTypeEnum!, $regions: [AzureCloudAccountRegionEnum!]!, $feature: AddAzureCloudAccountFeatureInputWithoutOauth!, $subscriptionName: String!, $subscriptionId: String!) {
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
        regions:          $regions,
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
        domainName
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

// allAzureCloudAccountTenantsV0 GraphQL query
var allAzureCloudAccountTenantsV0Query = `query SdkGolangAllAzureCloudAccountTenantsV0($feature: CloudAccountFeatureEnum!, $includeSubscriptionDetails: Boolean!) {
    result: allAzureCloudAccountTenants(feature: $feature, includeSubscriptionDetails: $includeSubscriptionDetails) {
        cloudType
        azureCloudAccountTenantRubrikId
        domainName
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

// allAzureExocomputeConfigsInAccount GraphQL query
var allAzureExocomputeConfigsInAccountQuery = `query SdkGolangAllAzureExocomputeConfigsInAccount($cloudAccountIDs: [UUID!], $azureExocomputeSearchQuery: String!) {
    result: allAzureExocomputeConfigsInAccount(cloudAccountIDs: $cloudAccountIDs, azureExocomputeSearchQuery: $azureExocomputeSearchQuery) {
        azureCloudAccount {
            id
            name
            nativeId
            featureDetail {
                feature
                regions
                status
            }
        }
        configs {
            configUuid
            isPolarisManaged
            message
            region
            subnetNativeId
        }
        exocomputeEligibleRegions
        featureDetails {
            feature
            regions
            status
        }
    }
}`

// azureCloudAccountPermissionConfig GraphQL query
var azureCloudAccountPermissionConfigQuery = `query SdkGolangAzureCloudAccountPermissionConfig($feature: CloudAccountFeature!) {
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

// azureCloudAccountPermissionConfigV0 GraphQL query
var azureCloudAccountPermissionConfigV0Query = `query SdkGolangAzureCloudAccountPermissionConfigV0($feature: CloudAccountFeatureEnum!) {
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

// azureCloudAccountTenant GraphQL query
var azureCloudAccountTenantQuery = `query SdkGolangAzureCloudAccountTenant($tenantId: UUID!, $feature: CloudAccountFeature!, $subscriptionSearchText: String!) {
    result: azureCloudAccountTenant(tenantId: $tenantId, feature: $feature, subscriptionSearchText: $subscriptionSearchText, subscriptionStatusFilters: []) {
        cloudType
        azureCloudAccountTenantRubrikId
        clientId
        domainName
        subscriptions {
            id
            name
            nativeId
            featureDetail {
                feature
                regions
                status
            }
        }
    }
}`

// azureCloudAccountTenantV0 GraphQL query
var azureCloudAccountTenantV0Query = `query SdkGolangAzureCloudAccountTenantV0($tenantId: UUID!, $feature: CloudAccountFeatureEnum!, $subscriptionSearchText: String!) {
    result: azureCloudAccountTenant(tenantId: $tenantId, feature: $feature, subscriptionSearchText: $subscriptionSearchText, subscriptionStatusFilters: []) {
        cloudType
        azureCloudAccountTenantRubrikId
        clientId
        domainName
        subscriptions {
            id
            name
            nativeId
            featureDetail {
                feature
                regions
                status
            }
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

// deleteAzureCloudAccountExocomputeConfigurations GraphQL query
var deleteAzureCloudAccountExocomputeConfigurationsQuery = `mutation SdkGolangDeleteAzureCloudAccountExocomputeConfigurations($cloudAccountIds: [UUID!]!) {
    result: deleteAzureCloudAccountExocomputeConfigurations(input: {
        cloudAccountIds: $cloudAccountIds
    }) {
        deletionFailedIds
        deletionSuccessIds
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

// deleteAzureCloudAccountWithoutOauthV0 GraphQL query
var deleteAzureCloudAccountWithoutOauthV0Query = `mutation SdkGolangDeleteAzureCloudAccountWithoutOauthV0($subscriptionIds: [UUID!]!, $features: [CloudAccountFeatureEnum!]!) {
    result: deleteAzureCloudAccountWithoutOAuth(input: {
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

// deleteAzureCloudAccountWithoutOauthV1 GraphQL query
var deleteAzureCloudAccountWithoutOauthV1Query = `mutation SdkGolangDeleteAzureCloudAccountWithoutOauthV1($subscriptionIds: [UUID!]!, $features: [CloudAccountFeatureEnum!]!) {
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

// setAzureCloudAccountCustomerAppCredentialsV0 GraphQL query
var setAzureCloudAccountCustomerAppCredentialsV0Query = `mutation SdkGolangSetAzureCloudAccountCustomerAppCredentialsV0($azureCloudType: AzureCloudTypeEnum!, $appId: String!, $appName: String, $appSecretKey: String!, $appTenantId: String, $tenantDomainName: String) {
    result: setAzureCloudAccountCustomerAppCredentials(input: {
        appId:            $appId,
        appSecretKey:     $appSecretKey,
        appTenantId:      $appTenantId,
        appName:          $appName,
        tenantDomainName: $tenantDomainName,
        azureCloudType:   $azureCloudType
    })
}`

// setAzureCloudAccountCustomerAppCredentialsV1 GraphQL query
var setAzureCloudAccountCustomerAppCredentialsV1Query = `mutation SdkGolangSetAzureCloudAccountCustomerAppCredentialsV1($azureCloudType: AzureCloudType!, $appId: String!, $appName: String, $appSecretKey: String!, $appTenantId: String, $tenantDomainName: String) {
    result: setAzureCloudAccountCustomerAppCredentials(input: {
        appId:            $appId,
        appSecretKey:     $appSecretKey,
        appTenantId:      $appTenantId,
        appName:          $appName,
        tenantDomainName: $tenantDomainName,
        azureCloudType:   $azureCloudType
    })
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

// updateAzureCloudAccountV0 GraphQL query
var updateAzureCloudAccountV0Query = `mutation SdkGolangUpdateAzureCloudAccountV0($features: [CloudAccountFeatureEnum!]!, $regionsToAdd: [AzureCloudAccountRegionEnum!], $regionsToRemove: [AzureCloudAccountRegionEnum!], $subscriptions: [AzureCloudAccountSubscriptionInput!]!) {
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

// updateAzureCloudAccountV1 GraphQL query
var updateAzureCloudAccountV1Query = `mutation SdkGolangUpdateAzureCloudAccountV1($features: [CloudAccountFeatureEnum!]!, $regionsToAdd: [AzureCloudAccountRegion!], $regionsToRemove: [AzureCloudAccountRegion!], $subscriptions: [AzureCloudAccountSubscriptionInput!]!) {
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
        cloudAccountId: $cloudAccountId
        feature:        $feature,
    }) {
        status
    }
}`

// upgradeAzureCloudAccountPermissionsWithoutOauthV0 GraphQL query
var upgradeAzureCloudAccountPermissionsWithoutOauthV0Query = `mutation SdkGolangUpgradeAzureCloudAccountPermissionsWithoutOauthV0($cloudAccountId: UUID!, $feature: CloudAccountFeatureEnum!) {
    result: upgradeAzureCloudAccountPermissionsWithoutOauth(input: {
        cloudAccountId: $cloudAccountId
        feature:        $feature,
    }) {
        status
    }
}`
