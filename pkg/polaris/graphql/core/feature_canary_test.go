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

package core

import (
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

// TestFeatureEnumCanary acts as a canary test to catch changes to the
// CloudAccountFeature enum in RSC early. It compares the features in the
// featureInfoMap against the enum values in RSC. If RSC adds or removes enum
// values, this test will fail, alerting developers to update the
// featureInfoMap.
//
// To run this test against an RSC instance the following environment variables
// need to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
func TestFeatureEnumCanary(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	rscValues, err := graphql.EnumValuesAsSet(t.Context(), client.GQL, "CloudAccountFeature")
	if err != nil {
		t.Fatal(err)
	}

	for name := range featureInfoMap {
		if _, ok := rscValues[name]; !ok {
			t.Errorf("RSC enum CloudAccountFeature: %q exist in the SDK but not in RSC", name)
		}
	}

	for name := range rscValues {
		if name == "FEATURE_UNSPECIFIED" {
			// Skip the unspecified value, it's not a real feature.
			continue
		}
		if _, ok := featureInfoMap[name]; !ok {
			t.Errorf("RSC enum CloudAccountFeature: %q exist in RSC but not in the SDK", name)
		}
	}
}
