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
	"slices"
	"testing"
)

func TestParseFeatureNoValidation(t *testing.T) {
	if feature := ParseFeatureNoValidation("CLOUD_NATIVE_PROTECTION"); !feature.Equal(FeatureCloudNativeProtection) {
		t.Errorf("invalid feature: %s", feature)
	}

	if feature := ParseFeatureNoValidation("cloud_native_protection"); !feature.Equal(FeatureCloudNativeProtection) {
		t.Errorf("invalid feature: %s", feature)
	}

	if feature := ParseFeatureNoValidation("cloud-native-protection"); !feature.Equal(FeatureCloudNativeProtection) {
		t.Errorf("invalid feature: %s", feature)
	}
}

func TestCloudNativeConfigProtection(t *testing.T) {
	if name := FeatureCloudNativeConfigProtection.Name; name != "CLOUD_NATIVE_CONFIG_PROTECTION" {
		t.Errorf("invalid feature name: %s", name)
	}

	if !FeatureCloudNativeConfigProtection.IsProtectionFeature() {
		t.Error("CLOUD_NATIVE_CONFIG_PROTECTION should be a protection feature")
	}

	if !slices.ContainsFunc(AllProtectionFeatures(CloudVendorAWS), FeatureCloudNativeConfigProtection.Equal) {
		t.Error("CLOUD_NATIVE_CONFIG_PROTECTION should be an AWS protection feature")
	}

	feature, err := ParseFeature("CLOUD_NATIVE_CONFIG_PROTECTION")
	if err != nil {
		t.Fatalf("failed to parse feature: %v", err)
	}
	if !feature.Equal(FeatureCloudNativeConfigProtection) {
		t.Errorf("invalid feature: %s", feature)
	}
}
