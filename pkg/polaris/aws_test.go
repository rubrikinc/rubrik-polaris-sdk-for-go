package polaris

import (
	"context"
	"reflect"
	"testing"

	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Between the account has been added and it has been removed we never fail
// fatally to allow the account to be removed in case of an error.
func TestAwsAccountAddAndRemove(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

	// Load configuration and create client.
	config, err := DefaultConfig("default")
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(config, &log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	// Add and verify AWS account using the default AWS account. Note that for
	// the Trinity lab we must use the name specified name since accounts cannot
	// be renamed.
	err = client.AwsAccountAdd(ctx, FromAwsDefault(), WithName("Trinity-AWS-FDSE"),
		WithRegion("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}
	account, err := client.AwsAccount(ctx, FromAwsDefault())
	if err != nil {
		t.Error(err)
	}
	if account.Name != "Trinity-AWS-FDSE" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.NativeID != "311033699123" {
		t.Errorf("invalid native id: %v", account.NativeID)
	}
	if n := len(account.Features); n != 1 {
		t.Errorf("invalid number of features: %v", n)
	}
	if account.Features[0].Feature != "CLOUD_NATIVE_PROTECTION" {
		t.Errorf("invalid feature name: %v", account.Features[0].Feature)
	}
	if regions := account.Features[0].AwsRegions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
		t.Errorf("invalid feature regions: %v", regions)
	}
	if account.Features[0].Status != "CONNECTED" {
		t.Errorf("invalid feature status: %v", account.Features[0].Status)
	}

	// Set and verify regions for AWS account.
	if err := client.AwsAccountSetRegions(ctx, WithUUID(account.ID), "us-west-2"); err != nil {
		t.Error(err)
	}
	account, err = client.AwsAccount(ctx, WithAwsID(account.NativeID))
	if err != nil {
		t.Error(err)
	}
	if n := len(account.Features); n != 1 {
		t.Errorf("invalid number of features: %v", n)
	}
	if regions := account.Features[0].AwsRegions; !reflect.DeepEqual(regions, []string{"us-west-2"}) {
		t.Errorf("invalid feature regions: %v", regions)
	}

	// Remove AWS account and verify that it's gone.
	if err := client.AwsAccountRemove(ctx, FromAwsDefault(), false); err != nil {
		t.Fatal(err)
	}
	account, err = client.AwsAccount(ctx, FromAwsDefault())
	if err != ErrAccountNotFound {
		t.Error(err)
	}
}
