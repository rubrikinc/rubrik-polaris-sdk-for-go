package polaris

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Between the account has been added and it has been removed we never fail
// fatally to allow the account to be removed in case of an error.
func TestAwsAccountAddAndRemove(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

	// Polaris client.
	polConfig, err := DefaultConfig("default")
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(polConfig, &log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	// AWS config.
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Add and verify AWS account.
	if err := client.AwsAccountAdd(ctx, awsConfig, []string{"us-east-2"}); err != nil {
		t.Fatal(err)
	}
	account, err := client.AwsAccountFromConfig(ctx, awsConfig)
	if err != nil {
		t.Error(err)
	}
	if name := account.Name; name != "Trinity-TPM-DevOps" {
		t.Errorf("invalid name, name=%v", name)
	}
	if id := account.NativeID; id != "627297623784" {
		t.Errorf("invalid native id, id=%v", id)
	}
	if n := len(account.Features); n != 1 {
		t.Errorf("invalid number of features, n=%v", n)
	}
	if name := account.Features[0].Feature; name != "CLOUD_NATIVE_PROTECTION" {
		t.Errorf("invalid feature name, name=%v", name)
	}
	if regions := account.Features[0].AwsRegions; reflect.DeepEqual(regions, []string{"us-east-2"}) {
		t.Errorf("invalid feature regions, regions=%v", regions)
	}
	if status := account.Features[0].Status; status != "CONNECTED" {
		t.Errorf("invalid feature status, status=%v", status)
	}

	// Set and verify regions for AWS account.
	if err := client.AwsAccountSetRegions(ctx, account.NativeID, []string{"us-west-2"}); err != nil {
		t.Error(err)
	}
	account, err = client.AwsAccountFromID(ctx, account.NativeID)
	if err != nil {
		t.Error(err)
	}
	if n := len(account.Features); n != 1 {
		t.Errorf("invalid number of features, n=%v", n)
	}
	if regions := account.Features[0].AwsRegions; reflect.DeepEqual(regions, []string{"us-west-2"}) {
		t.Errorf("invalid feature regions, regions=%v", regions)
	}

	// Remove AWS account and verify that it's gone.
	if err := client.AwsAccountRemove(ctx, awsConfig, ""); err != nil {
		t.Fatal(err)
	}
	account, err = client.AwsAccountFromConfig(ctx, awsConfig)
	if err != ErrAccountNotFound {
		t.Error(err)
	}
}
