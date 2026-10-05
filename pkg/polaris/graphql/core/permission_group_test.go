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

func TestAccountTypePermissionGroupNames(t *testing.T) {
	for name, tc := range map[string]struct {
		names    []string
		features []Feature
	}{
		"AWSCloudFormationPermissionGroupNames": {
			names:    AWSCloudFormationPermissionGroupNames(),
			features: AWSCloudFormationFeatures(),
		},
		"AWSIAMRolesPermissionGroupNames": {
			names:    AWSIAMRolesPermissionGroupNames(),
			features: AWSIAMRolesFeatures(),
		},
		"AWSManagedPermissionGroupNames": {
			names:    AWSManagedPermissionGroupNames(),
			features: AWSManagedFeatures(),
		},
		"AzureSubscriptionPermissionGroupNames": {
			names:    AzureSubscriptionPermissionGroupNames(),
			features: AzureSubscriptionFeatures(),
		},
		"AzureDevOpsOrganizationPermissionGroupNames": {
			names:    AzureDevOpsOrganizationPermissionGroupNames(),
			features: AzureDevOpsOrganizationFeatures(),
		},
		"GCPProjectPermissionGroupNames": {
			names:    GCPProjectPermissionGroupNames(),
			features: GCPProjectFeatures(),
		},
		"GitHubOrganizationPermissionGroupNames": {
			names:    GitHubOrganizationPermissionGroupNames(),
			features: GitHubOrganizationFeatures(),
		},
	} {
		t.Run(name, func(t *testing.T) {
			if !slices.IsSorted(tc.names) {
				t.Errorf("permission groups are not sorted: %v", tc.names)
			}
			if len(slices.Compact(slices.Clone(tc.names))) != len(tc.names) {
				t.Errorf("permission groups contain duplicates: %v", tc.names)
			}
			for _, feature := range tc.features {
				for _, group := range feature.PermissionGroups {
					if !slices.Contains(tc.names, string(group)) {
						t.Errorf("permission groups are missing %s of %s", group, feature.Name)
					}
				}
			}
		})
	}
}
