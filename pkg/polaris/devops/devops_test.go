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

package devops

import (
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// TestGitHubCheckFeature covers the supported-feature validation for GitHub
// organizations, including unsupported features and unsupported permission
// groups on an otherwise supported feature.
func TestGitHubCheckFeature(t *testing.T) {
	tests := []struct {
		name    string
		feature core.Feature
		wantErr bool
	}{{
		name:    "supported feature without permission groups",
		feature: core.FeatureGitHubRepositoryProtection,
		wantErr: false,
	}, {
		name: "supported feature with supported permission groups",
		feature: core.FeatureGitHubRepositoryProtection.WithPermissionGroups(
			core.PermissionGroupBasic,
			core.PermissionGroupRecovery,
		),
		wantErr: false,
	}, {
		name: "supported feature with unsupported permission group",
		feature: core.FeatureGitHubRepositoryProtection.WithPermissionGroups(
			core.PermissionGroupRestore,
		),
		wantErr: true,
	}, {
		name:    "unsupported feature",
		feature: core.FeatureAzureDevOpsRepositoryProtection,
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GitHubCheckFeature(tt.feature)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitHubCheckFeature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAzureCheckFeature covers the supported-feature validation for Azure
// DevOps organizations, including unsupported features and unsupported
// permission groups on an otherwise supported feature.
func TestAzureCheckFeature(t *testing.T) {
	tests := []struct {
		name    string
		feature core.Feature
		wantErr bool
	}{{
		name:    "supported feature without permission groups",
		feature: core.FeatureAzureDevOpsRepositoryProtection,
		wantErr: false,
	}, {
		name: "supported feature with supported permission groups",
		feature: core.FeatureAzureDevOpsRepositoryProtection.WithPermissionGroups(
			core.PermissionGroupBasic,
			core.PermissionGroupRecovery,
		),
		wantErr: false,
	}, {
		name: "supported feature with unsupported permission group",
		feature: core.FeatureAzureDevOpsRepositoryProtection.WithPermissionGroups(
			core.PermissionGroupRestore,
		),
		wantErr: true,
	}, {
		name:    "unsupported feature",
		feature: core.FeatureGitHubRepositoryProtection,
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AzureCheckFeature(tt.feature)
			if (err != nil) != tt.wantErr {
				t.Errorf("AzureCheckFeature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
