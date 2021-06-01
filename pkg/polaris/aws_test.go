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
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	polaris_log "github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// testAwsAccount holds information about the AWS account used in the
// integration tests. Normally used to assert that the account information read
// from Polaris is correct.
type testAwsAccount struct {
	Name      string `json:"name"`
	AccountID string `json:"accountId"`
}

// TestAwsAccountAddAndRemove verifies that the SDK can perform the basic AWS
// account operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * SDK_INTEGRATION=1
//   * SDK_AWSACCOUNT_FILE=<path-to-test-aws-account-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * AWS_ACCESS_KEY_ID=<aws-access-key>
//   * AWS_SECRET_ACCESS_KEY=<aws-secret-key>
//   * AWS_DEFAULT_REGION=<aws-default-region>
//
// The file referred to by SDK_AWSACCOUNT_FILE should contain a single
// testAwsAccount JSON object.
//
// Note that between the project has been added and it has been removed we
// never fail fatally to allow the project to be removed in case of an error.
func TestAwsAccountAddAndRemove(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

	// Load test account information from the file pointed to by the
	// SDK_AWSACCOUNT_FILE environment variable.
	buf, err := os.ReadFile(os.Getenv("SDK_AWSACCOUNT_FILE"))
	if err != nil {
		t.Fatalf("failed to read file pointed to by SDK_AWSACCOUNT_FILE: %v", err)
	}
	testAccount := testAwsAccount{}
	if err := json.Unmarshal(buf, &testAccount); err != nil {
		t.Fatal(err)
	}

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

	// Add the default AWS account to Polaris. Usually resolved using the
	// environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and
	// AWS_DEFAULT_REGION. Note that for the Trinity lab we must use the name
	// specified name since accounts cannot be renamed.
	err = client.AwsAccountAdd(ctx, FromAwsDefault(), WithName("Trinity-AWS-FDSE"),
		WithRegion("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully added.
	account, err := client.AwsAccount(ctx, FromAwsDefault())
	if err != nil {
		t.Error(err)
	}
	if account.Name != testAccount.Name {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.NativeID != testAccount.AccountID {
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
	if n := len(account.Features); n == 1 {
		if regions := account.Features[0].AwsRegions; !reflect.DeepEqual(regions, []string{"us-west-2"}) {
			t.Errorf("invalid feature regions: %v", regions)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Remove AWS account from Polaris.
	if err := client.AwsAccountRemove(ctx, FromAwsDefault(), false); err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	account, err = client.AwsAccount(ctx, FromAwsDefault())
	if !errors.Is(err, ErrNotFound) {
		t.Error(err)
	}
}
