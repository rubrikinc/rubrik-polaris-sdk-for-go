// Copyright 2022 Rubrik, Inc.
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

package appliance

import (
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// TestApplianceTokenFromServiceAccount verifies that the SDK can retrieve
// the API token for an appliance on behalf of a service account.
//
// To run this test against a Polaris instance the following environment
// variable needs to be set:
//  * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
// 	* TEST_INTEGRATION=1
//  * TEST_APPLIANCE_ID=<appliance-uuid/cluster-uuid>
//
// In addition to the above environment variables, an appliance must be added to
// the Polaris instance.
func TestApplianceTokenFromServiceAccount(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testApplianceID := os.Getenv("TEST_APPLIANCE_ID")
	if testApplianceID == "" {
		t.Skipf("skipping due to env TEST_APPLIANCE_ID not set")
	}

	// Load service account credentials. Usually resolved using the
	// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		t.Fatal(err)
	}
	applianceID, err := uuid.Parse(testApplianceID)
	if err != nil {
		t.Fatal(err)
	}

	token, err := TokenFromServiceAccount(polAccount, applianceID, log.NewStandardLogger())
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("TokenFromServiceAccount returned an empty token")
	}
}
