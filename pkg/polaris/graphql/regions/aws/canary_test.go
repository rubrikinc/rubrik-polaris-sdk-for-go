// Copyright 2026 Rubrik, Inc.
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

package aws

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the RSC client shared across all integration tests in this package.
// Reusing a single client avoids repeated token creation, reducing the risk of
// hitting rate limits.
var client *polaris.Client

func TestMain(m *testing.M) {
	if testsetup.BoolEnvSet("TEST_INTEGRATION") {
		// Load configuration and create client. Usually resolved using the
		// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
		polAccount, err := polaris.DefaultServiceAccount(true)
		if err != nil {
			fmt.Printf("failed to get default service account: %v\n", err)
			os.Exit(1)
		}

		// The integration tests defaults the log level to INFO. Note that
		// RUBRIK_POLARIS_LOGLEVEL can be used to override this.
		logger := log.NewStandardLogger()
		logger.SetLogLevel(log.Info)
		if err := polaris.SetLogLevelFromEnv(logger); err != nil {
			fmt.Printf("failed to get log level from env: %v\n", err)
			os.Exit(1)
		}

		client, err = polaris.NewClientWithLogger(polAccount, logger)
		if err != nil {
			fmt.Printf("failed to create polaris client: %v\n", err)
			os.Exit(1)
		}

		version, err := client.GQL.DeploymentVersion(context.Background())
		if err != nil {
			fmt.Printf("failed to get deployment version: %v\n", err)
			os.Exit(1)
		}
		logger.Printf(log.Info, "Polaris version: %s", version)
	}

	os.Exit(m.Run())
}

// TestRegionEnumCanary acts as a canary test to catch changes to the AwsRegion
// enum in RSC early.
func TestRegionEnumCanary(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	sdkValues := make(map[string]struct{}, len(regionInfoMap))
	for _, info := range regionInfoMap {
		if info.regionEnum != "" {
			sdkValues[info.regionEnum] = struct{}{}
		}
	}

	rscValues, err := graphql.EnumValuesAsSet(t.Context(), client.GQL, "AwsRegion")
	if err != nil {
		t.Fatal(err)
	}

	assertEnumValues(t, "AwsRegion", sdkValues, rscValues)
}

// TestNativeRegionEnumCanary acts as a canary test to catch changes to the
// AwsNativeRegion enum in RSC early.
func TestNativeRegionEnumCanary(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	sdkValues := make(map[string]struct{}, len(regionInfoMap))
	for _, info := range regionInfoMap {
		if info.nativeRegionEnum != "" {
			sdkValues[info.nativeRegionEnum] = struct{}{}
		}
	}

	rscValues, err := graphql.EnumValuesAsSet(t.Context(), client.GQL, "AwsNativeRegion")
	if err != nil {
		t.Fatal(err)
	}

	assertEnumValues(t, "AwsNativeRegion", sdkValues, rscValues)
}

// TestRegionForReplicationEnumCanary acts as a canary test to catch changes to
// the AwsNativeRegionForReplication enum in RSC early.
func TestRegionForReplicationEnumCanary(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	sdkValues := make(map[string]struct{}, len(regionInfoMap))
	for _, info := range regionInfoMap {
		if info.regionForReplicationEnum != "" {
			sdkValues[info.regionForReplicationEnum] = struct{}{}
		}
	}

	rscValues, err := graphql.EnumValuesAsSet(t.Context(), client.GQL, "AwsNativeRegionForReplication")
	if err != nil {
		t.Fatal(err)
	}

	assertEnumValues(t, "AwsNativeRegionForReplication", sdkValues, rscValues)
}

// assertEnumValues checks that sdkValues and rscValues contain the same set of
// enum value names, reporting each discrepancy as a separate test error.
func assertEnumValues(t *testing.T, enumName string, sdkValues map[string]struct{}, rscValues map[string]graphql.EnumValue) {
	for value := range sdkValues {
		if _, ok := rscValues[value]; !ok {
			t.Errorf("RSC enum %s: %q exist in the SDK but not in RSC", enumName, value)
		}
	}

	for value := range rscValues {
		if _, ok := sdkValues[value]; !ok {
			t.Errorf("RSC enum %s: %q exist in RSC but not in the SDK", enumName, value)
		}
	}
}
