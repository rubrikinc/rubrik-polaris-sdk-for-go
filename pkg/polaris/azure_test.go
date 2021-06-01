// Copyright 2021 Rubrik, Inc.
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

package polaris

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Between the account has been added and it has been removed we never fail
// fatally to allow the account to be removed in case of an error.
func TestAzureSubscriptionAddAndRemove(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

	// Load configuration and create client. Usually resolved using the
	// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
	polAccount, err := DefaultServiceAccount()
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClientFromServiceAccount(polAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure service principal to Polaris. Usually resolved using
	// the environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	principal, err := AzureDefaultServicePrincipal()
	if err != nil {
		t.Fatal(err)
	}

	err = client.AzureServicePrincipalSet(ctx, principal)
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure subscription to Polaris. Usually resolved using the
	// environment variable AZURE_SUBSCRIPTION_LOCATION
	subscriptionIn, err := AzureDefaultSubscription()
	if err != nil {
		t.Fatal(err)
	}

	err = client.AzureSubscriptionAdd(ctx, subscriptionIn)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the subscription was successfully added.
	subscription, err := client.AzureSubscription(ctx, WithAzureSubscriptionID("8fa81a5e-a236-4a73-8e28-e1dcf863c56d"))
	if err != nil {
		t.Error(err)
	}
	if subscription.Name != "TrinityFDSE" {
		t.Errorf("invalid name: %v", subscription.Name)
	}
	if subscription.NativeID != uuid.MustParse("8fa81a5e-a236-4a73-8e28-e1dcf863c56d") {
		t.Errorf("invalid native id: %v", subscription.NativeID)
	}
	if subscription.Feature.Name != "CLOUD_NATIVE_PROTECTION" {
		t.Errorf("invalid feature name: %v", subscription.Feature.Name)
	}
	if regions := subscription.Feature.Regions; !reflect.DeepEqual(regions, []graphql.AzureRegion{graphql.AzureRegionEastUS2}) {
		t.Errorf("invalid feature regions: %v", regions)
	}
	if subscription.Feature.Status != "CONNECTED" {
		t.Errorf("invalid feature status: %v", subscription.Feature.Status)
	}

	// Set and verify regions for AWS account.
	err = client.AzureSubscriptionSetRegions(ctx, WithAzureSubscriptionID("8fa81a5e-a236-4a73-8e28-e1dcf863c56d"), graphql.AzureRegionWestUS2)
	if err != nil {
		t.Error(err)
	}
	subscription, err = client.AzureSubscription(ctx, WithAzureSubscriptionID("8fa81a5e-a236-4a73-8e28-e1dcf863c56d"))
	if err != nil {
		t.Error(err)
	}
	if regions := subscription.Feature.Regions; !reflect.DeepEqual(regions, []graphql.AzureRegion{graphql.AzureRegionWestUS2}) {
		t.Errorf("invalid feature regions: %v", regions)
	}

	//
	err = client.AzureSubscriptionRemove(ctx, WithAzureSubscriptionID("8fa81a5e-a236-4a73-8e28-e1dcf863c56d"), false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = client.AzureSubscription(ctx, WithAzureSubscriptionID("8fa81a5e-a236-4a73-8e28-e1dcf863c56d"))
	if !errors.Is(err, ErrNotFound) {
		t.Error(err)
	}
}
