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

	if !slices.Contains(CloudVendorProtectionFeatureNames(CloudVendorAWS), FeatureCloudNativeConfigProtection.Name) {
		t.Error("CLOUD_NATIVE_CONFIG_PROTECTION should be an AWS protection feature")
	}
}

func TestAccountTypeFeatures(t *testing.T) {
	for name, features := range map[string][]Feature{
		"AWSCloudFormationFeatures":       AWSCloudFormationFeatures(),
		"AWSIAMRolesFeatures":             AWSIAMRolesFeatures(),
		"AWSManagedFeatures":              AWSManagedFeatures(),
		"AzureSubscriptionFeatures":       AzureSubscriptionFeatures(),
		"AzureDevOpsOrganizationFeatures": AzureDevOpsOrganizationFeatures(),
		"GCPProjectFeatures":              GCPProjectFeatures(),
		"GitHubOrganizationFeatures":      GitHubOrganizationFeatures(),
	} {
		t.Run(name, func(t *testing.T) {
			if len(features) == 0 {
				t.Fatal("no features")
			}

			names := FeatureNames(features)
			if !slices.IsSorted(names) {
				t.Errorf("features are not sorted by name: %v", names)
			}
			if len(slices.Compact(slices.Clone(names))) != len(names) {
				t.Errorf("features contain duplicates: %v", names)
			}
		})
	}
}

func TestCloudVendorFeatureNames(t *testing.T) {
	// Every account type feature must be part of the cloud vendor union.
	for cloud, features := range map[CloudVendor][]Feature{
		CloudVendorAWS:   slices.Concat(AWSCloudFormationFeatures(), AWSIAMRolesFeatures(), AWSManagedFeatures()),
		CloudVendorAzure: slices.Concat(AzureSubscriptionFeatures(), AzureDevOpsOrganizationFeatures()),
		CloudVendorGCP:   GCPProjectFeatures(),
	} {
		names := CloudVendorFeatureNames(cloud)
		if !slices.IsSorted(names) {
			t.Errorf("%s feature names are not sorted: %v", cloud, names)
		}
		if len(slices.Compact(slices.Clone(names))) != len(names) {
			t.Errorf("%s feature names contain duplicates: %v", cloud, names)
		}
		for _, feature := range features {
			if !slices.Contains(names, feature.Name) {
				t.Errorf("%s features are missing %s", cloud, feature.Name)
			}
		}
	}
}

func TestLookupFeatureName(t *testing.T) {
	features := []Feature{
		FeatureCloudNativeProtection.WithPermissionGroups(PermissionGroupBasic),
		FeatureExocompute,
	}

	feature, ok := LookupFeatureName(features, CloudNativeProtection)
	if !ok {
		t.Fatal("expected CLOUD_NATIVE_PROTECTION to be found")
	}
	if !feature.DeepEqual(features[0]) {
		t.Errorf("unexpected feature: %s", feature)
	}

	if _, ok := LookupFeatureName(features, CloudSQLProtection); ok {
		t.Error("expected CLOUD_SQL_PROTECTION not to be found")
	}
}

// TestAllowlists verifies that the allowlists keep the same features when
// reading cloud accounts from RSC as previous versions of the SDK.
func TestAllowlists(t *testing.T) {
	for name, tc := range map[string]struct {
		names []string
		want  []string
	}{
		"AWSAccountAllowlistNames": {
			names: AWSAccountAllowlistNames(),
			want: []string{
				Archival, CloudCostReport, CloudDiscovery, CloudNativeArchival, CloudNativeConfigProtection,
				CloudNativeDynamoDBProtection, CloudNativeProtection, CloudNativeS3Protection,
				CyberRecoveryDataClassificationData, CyberRecoveryDataClassificationMetadata, DSPMData, DSPMMetadata,
				Exocompute, KubernetesProtection, LaminarCrossAccount, LaminarInternal, Outpost, RDSProtection,
				RoleChaining, ServerAndApps,
			},
		},
		"AzureSubscriptionAllowlistNames": {
			names: AzureSubscriptionAllowlistNames(),
			want: []string{
				AzurePostgresFlexibleServerProtection, AzureSQLDBProtection, AzureSQLMIProtection, CloudDiscovery,
				CloudNativeArchival, CloudNativeArchivalEncryption, CloudNativeBlobProtection, CloudNativeProtection,
				Exocompute, ServerAndApps,
			},
		},
		"GCPProjectAllowlistNames": {
			names: GCPProjectAllowlistNames(),
			want: []string{
				CloudNativeArchival, CloudNativeProtection, CloudSQLProtection, Exocompute, GCPBigQueryProtection,
				GCPBigQueryReservation, GCPSharedVPCHost, ServerAndApps,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if got := tc.names; !slices.Equal(got, tc.want) {
				t.Errorf("unexpected features\n got: %v\nwant: %v", got, tc.want)
			}
		})
	}
}
