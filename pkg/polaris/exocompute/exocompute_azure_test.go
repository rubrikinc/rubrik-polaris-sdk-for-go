// Copyright 2023 Rubrik, Inc.
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

package exocompute

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// TestAzureExocompute verifies that the SDK can perform basic Exocompute
// operations on a real RSC instance.
//
// To run this test against an RSC instance, the following environment variables
// need to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
//   - TEST_AZURESUBSCRIPTION_FILE=<path-to-test-azure-subscription-file>
//   - AZURE_AUTH_LOCATION=<path-to-azure-sdk-auth-file>
//
// The file referred to by TEST_AZURESUBSCRIPTION_FILE should contain a single
// testAzureSubscription JSON object.
func TestAzureExocompute(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testSubscription, err := testsetup.AzureSubscription()
	if err != nil {
		t.Fatal(err)
	}

	azureClient := azure.Wrap(client)

	// Add default Azure service principal to RSC. Usually resolved using the
	// environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	_, err = azureClient.SetServicePrincipal(ctx, azure.Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	// Add Azure subscription to RSC.
	subscription := azure.Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	cnpRegions := azure.Regions(testSubscription.CloudNativeProtection.Regions...)
	cnpResourceGroup := azure.ResourceGroup(testSubscription.CloudNativeProtection.ResourceGroupName,
		testSubscription.CloudNativeProtection.ResourceGroupRegion, nil)
	accountID, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection,
		azure.Name(testSubscription.SubscriptionName), cnpRegions, cnpResourceGroup)
	if err != nil {
		t.Fatal(err)
	}

	// Enable the exocompute feature for the account.
	exoRegions := azure.Regions(testSubscription.Exocompute.Regions...)
	exoResourceGroup := azure.ResourceGroup(testSubscription.Exocompute.ResourceGroupName,
		testSubscription.Exocompute.ResourceGroupRegion, nil)
	exoAccountID, err := azureClient.AddSubscription(ctx, subscription, core.FeatureExocompute, exoRegions,
		exoResourceGroup)
	if err != nil {
		t.Fatal(err)
	}
	if accountID != exoAccountID {
		t.Fatal("cloud native protection and exocompute added to different cloud accounts")
	}

	account, err := azureClient.SubscriptionByID(ctx, accountID)
	if err != nil {
		t.Fatal(err)
	}
	if name := account.Name; name != testSubscription.SubscriptionName {
		t.Fatalf("invalid name: %v", name)
	}
	if id := account.NativeID; id != testSubscription.SubscriptionID {
		t.Fatalf("invalid native id: %v", id)
	}
	if domain := account.TenantDomain; domain != testSubscription.TenantDomain {
		t.Fatalf("invalid tenant domain: %v", domain)
	}
	if n := len(account.Features); n != 2 {
		t.Fatalf("invalid number of features: %v", n)
	}
	feature, ok := account.Feature(core.FeatureExocompute)
	if !ok {
		t.Fatalf("%s feature not found", core.FeatureExocompute)
	}
	if name := feature.Name; name != core.FeatureExocompute.Name {
		t.Fatalf("invalid feature name: %v", name)
	}
	slices.Sort(feature.Regions)
	slices.Sort(testSubscription.Exocompute.Regions)
	if regions := feature.Regions; !slices.Equal(regions, testSubscription.Exocompute.Regions) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if status := feature.Status; status != "CONNECTED" {
		t.Fatalf("invalid feature status: %v", status)
	}
	if name := feature.ResourceGroup.Name; name != testSubscription.Exocompute.ResourceGroupName {
		t.Fatalf("invalid feature resource group name: %v", name)
	}
	if region := feature.ResourceGroup.Region; region != testSubscription.Exocompute.ResourceGroupRegion {
		t.Fatalf("invalid feature resource group region: %v", region)
	}
	if tags := feature.ResourceGroup.Tags; !maps.Equal(tags, map[string]string{}) {
		t.Fatalf("invalid feature resource group tags: %v", tags)
	}

	exoClient := Wrap(client)

	// Add an exocompute configuration to the account.
	if len(testSubscription.Exocompute.Regions) == 0 {
		t.Fatalf("exocompute test data must contain at least one region")
	}
	exoConfigRegion := gqlazure.RegionFromName(testSubscription.Exocompute.Regions[0])
	exoID, err := exoClient.AddAzureConfiguration(ctx, accountID, AzureManaged(exoConfigRegion, testSubscription.Exocompute.SubnetID, false))
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve the exocompute configuration added.
	exoConfig, err := exoClient.AzureConfigurationByID(ctx, exoID)
	if err != nil {
		t.Fatal(err)
	}
	if id := exoConfig.ID; id != exoID {
		t.Fatalf("invalid exocompute config id: %v", id)
	}
	if region := exoConfig.Region; region.Region != exoConfigRegion {
		t.Fatalf("invalid exocompute config region: %v", region)
	}
	if id := exoConfig.SubnetID; id != testSubscription.Exocompute.SubnetID {
		t.Fatalf("invalid exocompute config subnet id: %v", id)
	}
	if !exoConfig.ManagedByRubrik {
		t.Fatalf("invalid exocompute config managed state: %t", exoConfig.ManagedByRubrik)
	}

	// Remove the exocompute configuration.
	if err := exoClient.RemoveAzureConfiguration(ctx, exoID); err != nil {
		t.Fatal(err)
	}

	// Verify that the exocompute configuration was successfully removed.
	exoConfig, err = exoClient.AzureConfigurationByID(ctx, exoID)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}

	// Remove exocompute feature.
	if err := azureClient.RemoveSubscription(ctx, accountID, core.FeatureExocompute, false); err != nil {
		t.Fatal(err)
	}

	// Verify that the feature was successfully removed.
	account, err = azureClient.SubscriptionByID(ctx, accountID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := account.Feature(core.FeatureExocompute); ok {
		t.Fatalf("Exocompute feature still available")
	}

	// Remove subscription.
	if err := azureClient.RemoveSubscription(ctx, accountID, core.FeatureCloudNativeProtection, false); err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	if _, err := azureClient.SubscriptionByID(ctx, accountID); !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAddAzureConfigurationValidation verifies that validation warnings do not
// block an Azure exocompute configuration from being created, while validation
// errors do.
func TestAddAzureConfigurationValidation(t *testing.T) {
	configID := uuid.MustParse("5f1b8c0e-4a1d-4c53-9d57-1d1f4a3f2b10")
	cloudAccountID := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")

	testCases := []struct {
		name           string
		validationInfo string
		expectCreate   bool
	}{{
		name:           "NoValidationResults",
		validationInfo: `{}`,
		expectCreate:   true,
	}, {
		name:           "PrivateDnsZoneInDifferentSubscriptionIsWarning",
		validationInfo: `{"isAksCustomPrivateDnsZoneInDifferentSubscription": true}`,
		expectCreate:   true,
	}, {
		name:           "BlockedSecurityRulesIsWarning",
		validationInfo: `{"hasBlockedSecurityRules": true}`,
		expectCreate:   true,
	}, {
		name:           "SubnetDelegatedIsError",
		validationInfo: `{"isSubnetDelegated": true}`,
		expectCreate:   false,
	}, {
		name:           "WarningDoesNotMaskError",
		validationInfo: `{"isAksCustomPrivateDnsZoneInDifferentSubscription": true, "isSubnetDelegated": true}`,
		expectCreate:   false,
	}}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var created bool
			srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Error(err)
					return
				}

				switch {
				case strings.Contains(string(body), "validateAzureCloudAccountExocomputeConfigurations"):
					_, err = io.WriteString(w, `{"data": {"result": {"validationInfo": [`+tc.validationInfo+`]}}}`)
				case strings.Contains(string(body), "addAzureCloudAccountExocomputeConfigurations"):
					created = true
					_, err = io.WriteString(w, `{"data": {"result": {"configs": [{"configUuid": "`+configID.String()+`", "message": ""}]}}}`)
				default:
					t.Errorf("unexpected request: %s", body)
				}
				if err != nil {
					t.Error(err)
				}
			}))
			defer srv.Close()

			id, err := Wrap(&polaris.Client{GQL: graphql.NewTestClient(srv)}).AddAzureConfiguration(
				t.Context(), cloudAccountID, AzureManaged(gqlazure.RegionEastUS2, "subnet-id", false))
			if tc.expectCreate {
				if err != nil {
					t.Fatal(err)
				}
				if id != configID {
					t.Errorf("invalid configuration id: %s, expected: %s", id, configID)
				}
			} else if err == nil {
				t.Error("expected add configuration call to fail")
			}
			if created != tc.expectCreate {
				t.Errorf("configuration created: %t, expected: %t", created, tc.expectCreate)
			}
		})
	}
}

// TestUpdateAzureHealthCheckVMNamePrefix verifies that the prefix of a single
// configuration is changed while the other configurations of the cloud account
// are sent back unchanged, and that the update is refused when RSC manages the
// configurations of the account itself.
func TestUpdateAzureHealthCheckVMNamePrefix(t *testing.T) {
	cloudAccountID := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	targetID := uuid.MustParse("5f1b8c0e-4a1d-4c53-9d57-1d1f4a3f2b10")
	otherID := uuid.MustParse("0c6d2f3a-8d1e-4b0f-9a55-2e1d0d6f7a11")

	configs := `[{"azureCloudAccount": {"id": "` + cloudAccountID.String() + `"}, "configs": [{
		"configUuid": "` + targetID.String() + `", "region": "EASTUS2", "isRscManaged": true,
		"subnetNativeId": "subnet-east",
		"optionalConfig": {"aksClusterAccessType": "AKS_CLUSTER_ACCESS_TYPE_PUBLIC", "aksClusterTier": "AKS_CLUSTER_TIER_FREE",
			"aksNodeCountBucket": "AKS_NODE_COUNT_BUCKET_MEDIUM", "healthCheckVmNamePrefix": "old", "diskEncryptionSetId": "des"}
	}, {
		"configUuid": "` + otherID.String() + `", "region": "WESTUS", "isRscManaged": true,
		"subnetNativeId": "subnet-west", "optionalConfig": null
	}]}]`

	testCases := []struct {
		name           string
		permissions    string
		updateReply    string
		prefix         string
		expectUpdate   bool
		expectError    bool
		wantTargetOpts string
	}{{
		name:           "ChangesOnlyTheTarget",
		permissions:    `["BASIC"]`,
		updateReply:    `{"deleteStatus": [], "exocomputeConfigs": []}`,
		prefix:         "new",
		expectUpdate:   true,
		wantTargetOpts: `"healthCheckVmNamePrefix":"new"`,
	}, {
		name:           "EmptyPrefixClearsIt",
		permissions:    `["BASIC"]`,
		updateReply:    `{"deleteStatus": [], "exocomputeConfigs": []}`,
		prefix:         "",
		expectUpdate:   true,
		wantTargetOpts: `"aksNodeCountBucket":"AKS_NODE_COUNT_BUCKET_MEDIUM"`,
	}, {
		name:         "RefusedWithAutomatedNetworkingSetup",
		permissions:  `["BASIC", "AUTOMATED_NETWORKING_SETUP"]`,
		prefix:       "new",
		expectUpdate: false,
		expectError:  true,
	}, {
		name:         "DeletedConfigurationIsAnError",
		permissions:  `["BASIC"]`,
		updateReply:  `{"deleteStatus": [{"exocomputeConfigId": "` + otherID.String() + `", "success": true}], "exocomputeConfigs": []}`,
		prefix:       "new",
		expectUpdate: true,
		expectError:  true,
	}, {
		name:         "CreatedConfigurationIsAnError",
		permissions:  `["BASIC"]`,
		updateReply:  `{"deleteStatus": [], "exocomputeConfigs": [{"configUuid": "` + targetID.String() + `", "message": ""}]}`,
		prefix:       "new",
		expectUpdate: true,
		expectError:  true,
	}, {
		name:         "ConfigurationMessageIsAnError",
		permissions:  `["BASIC"]`,
		updateReply:  `{"deleteStatus": [], "exocomputeConfigs": [{"configUuid": "", "message": "boom"}]}`,
		prefix:       "new",
		expectUpdate: true,
		expectError:  true,
	}}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var sent string
			srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Error(err)
					return
				}

				switch {
				case strings.Contains(string(body), "allAzureExocomputeConfigsInAccount"):
					_, err = io.WriteString(w, `{"data": {"result": `+configs+`}}`)
				case strings.Contains(string(body), "allAzureCloudAccountTenants"):
					_, err = io.WriteString(w, `{"data": {"result": [{"cloudType": "AZUREPUBLICCLOUD",
						"azureCloudAccountTenantRubrikId": "`+uuid.Nil.String()+`", "domainName": "example.onmicrosoft.com",
						"subscriptionCount": 1, "subscriptions": [{"id": "`+cloudAccountID.String()+`", "name": "sub",
						"nativeId": "`+uuid.Nil.String()+`", "featureDetail": {"feature": "EXOCOMPUTE",
						"permissionsGroups": `+tc.permissions+`, "status": "CONNECTED", "regions": ["EASTUS2"]}}]}]}}`)
				case strings.Contains(string(body), "updateAzureExocomputeConfigs"):
					sent = string(body)
					_, err = io.WriteString(w, `{"data": {"result": `+tc.updateReply+`}}`)
				default:
					t.Errorf("unexpected request: %s", body)
				}
				if err != nil {
					t.Error(err)
				}
			}))
			defer srv.Close()

			err := Wrap(&polaris.Client{GQL: graphql.NewTestClient(srv)}).UpdateAzureHealthCheckVMNamePrefix(t.Context(), targetID, tc.prefix)
			if tc.expectError && err == nil {
				t.Error("expected update to fail")
			}
			if !tc.expectError && err != nil {
				t.Fatal(err)
			}
			if updated := sent != ""; updated != tc.expectUpdate {
				t.Fatalf("update sent: %t, expected: %t", updated, tc.expectUpdate)
			}
			if !tc.expectUpdate || tc.wantTargetOpts == "" {
				return
			}

			// Every configuration of the account must be sent back, the target with
			// the new prefix and the other one unchanged.
			var request struct {
				Variables struct {
					CloudAccountID uuid.UUID        `json:"cloudAccountId"`
					Configs        []map[string]any `json:"azureExocomputeRegionConfigs"`
				} `json:"variables"`
			}
			if err := json.Unmarshal([]byte(sent), &request); err != nil {
				t.Fatal(err)
			}
			if request.Variables.CloudAccountID != cloudAccountID {
				t.Errorf("invalid cloud account id: %s", request.Variables.CloudAccountID)
			}
			if n := len(request.Variables.Configs); n != 2 {
				t.Fatalf("expected both configurations to be sent, got: %d", n)
			}
			east, west := request.Variables.Configs[0], request.Variables.Configs[1]
			if east["region"] != "EASTUS2" || east["subnetNativeId"] != "subnet-east" || east["isRscManaged"] != true {
				t.Errorf("target configuration changed: %v", east)
			}
			eastOpts, err := json.Marshal(east["optionalConfig"])
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(eastOpts), tc.wantTargetOpts) {
				t.Errorf("target optional config %s does not contain %s", eastOpts, tc.wantTargetOpts)
			}
			if tc.prefix == "" && strings.Contains(string(eastOpts), "healthCheckVmNamePrefix") {
				t.Errorf("prefix not cleared: %s", eastOpts)
			}
			if !strings.Contains(string(eastOpts), `"diskEncryptionSetId":"des"`) {
				t.Errorf("other optional config fields not preserved: %s", eastOpts)
			}
			if west["region"] != "WESTUS" || west["subnetNativeId"] != "subnet-west" {
				t.Errorf("other configuration changed: %v", west)
			}
			if _, ok := west["optionalConfig"]; ok {
				t.Errorf("other configuration gained an optional config: %v", west)
			}
		})
	}
}

// TestUpdateAzureHealthCheckVMNamePrefixWithoutOptionalConfig verifies that a
// configuration which RSC stores no optional configuration for is sent with
// valid enum values. RSC rejects an empty enum value, so sending the zero
// value of the optional config would fail the whole update.
func TestUpdateAzureHealthCheckVMNamePrefixWithoutOptionalConfig(t *testing.T) {
	cloudAccountID := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	targetID := uuid.MustParse("5f1b8c0e-4a1d-4c53-9d57-1d1f4a3f2b10")

	var sent string
	srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Error(err)
			return
		}

		switch {
		case strings.Contains(string(body), "allAzureExocomputeConfigsInAccount"):
			_, err = io.WriteString(w, `{"data": {"result": [{"azureCloudAccount": {"id": "`+cloudAccountID.String()+`"}, "configs": [{
				"configUuid": "`+targetID.String()+`", "region": "WESTUS2", "isRscManaged": true,
				"subnetNativeId": "subnet", "optionalConfig": null}]}]}}`)
		case strings.Contains(string(body), "allAzureCloudAccountTenants"):
			_, err = io.WriteString(w, `{"data": {"result": [{"cloudType": "AZUREPUBLICCLOUD",
				"azureCloudAccountTenantRubrikId": "`+uuid.Nil.String()+`", "domainName": "example.onmicrosoft.com",
				"subscriptionCount": 1, "subscriptions": [{"id": "`+cloudAccountID.String()+`", "name": "sub",
				"nativeId": "`+uuid.Nil.String()+`", "featureDetail": {"feature": "EXOCOMPUTE",
				"permissionsGroups": ["BASIC"], "status": "CONNECTED", "regions": ["WESTUS2"]}}]}]}}`)
		case strings.Contains(string(body), "updateAzureExocomputeConfigs"):
			sent = string(body)
			_, err = io.WriteString(w, `{"data": {"result": {"deleteStatus": [], "exocomputeConfigs": []}}}`)
		default:
			t.Errorf("unexpected request: %s", body)
		}
		if err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	err := Wrap(&polaris.Client{GQL: graphql.NewTestClient(srv)}).UpdateAzureHealthCheckVMNamePrefix(t.Context(), targetID, "prefix")
	if err != nil {
		t.Fatal(err)
	}

	var request struct {
		Variables struct {
			Configs []struct {
				OptionalConfig map[string]any `json:"optionalConfig"`
			} `json:"azureExocomputeRegionConfigs"`
		} `json:"variables"`
	}
	if err := json.Unmarshal([]byte(sent), &request); err != nil {
		t.Fatal(err)
	}
	if n := len(request.Variables.Configs); n != 1 {
		t.Fatalf("expected one configuration, got: %d", n)
	}
	for field, want := range map[string]string{
		"aksClusterAccessType":    "AKS_CLUSTER_ACCESS_TYPE_UNSPECIFIED",
		"aksClusterTier":          "AKS_CLUSTER_TIER_UNSPECIFIED",
		"aksNodeCountBucket":      "AKS_NODE_COUNT_BUCKET_UNSPECIFIED",
		"healthCheckVmNamePrefix": "prefix",
	} {
		if got := request.Variables.Configs[0].OptionalConfig[field]; got != want {
			t.Errorf("%s: %v, expected: %s", field, got, want)
		}
	}
}
