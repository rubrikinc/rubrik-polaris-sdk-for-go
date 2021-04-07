package polaris

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func TestAwsAccountAddAndRemove(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

	polConfig, err := DefaultConfig("default")
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(polConfig, &log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.AwsAccountAdd(ctx, awsConfig, []string{"us-east-2", "us-west-2"}); err != nil {
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
	if regions := account.Features[0].AwsRegions; reflect.DeepEqual(regions, []string{"us-east-2", "us-west-2"}) {
		t.Errorf("invalid feature regions, regions=%v", regions)
	}
	if status := account.Features[0].Status; status != "CONNECTED" {
		t.Errorf("invalid feature status, status=%v", status)
	}

	if err := client.AwsAccountRemove(ctx, awsConfig, ""); err != nil {
		t.Fatal(err)
	}
}
