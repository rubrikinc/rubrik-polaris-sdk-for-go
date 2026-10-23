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

// Package devops provides a high-level interface to the DevOps (Azure DevOps
// and GitHub) part of the RSC platform. Azure DevOps onboarding assumes the
// customer application has already been registered via the azure package using
// the azure.AppUseCaseDevOps use case. GitHub cannot be onboarded through the
// SDK due to GitHub API limitations; its organizations and repositories are
// read-only.
package devops

import (
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for Azure DevOps organization management.
type API struct {
	client *graphql.Client
	log    log.Logger
}

// Wrap the RSC client in the devops API.
func Wrap(client *polaris.Client) API {
	return API{client: client.GQL, log: client.GQL.Log()}
}

// WrapGQL wraps the GQL client in the devops API.
func WrapGQL(client *graphql.Client) API {
	return API{client: client, log: client.Log()}
}

// AzureSupportedFeatures returns the features and permission groups supported
// by Azure DevOps organizations.
//
// Deprecated: use core.AzureDevOpsOrganizationFeatures instead.
func AzureSupportedFeatures() []core.Feature {
	return core.AzureDevOpsOrganizationFeatures()
}

// AzureSupportedFeatureNames returns the name of all features supported by
// Azure DevOps organizations.
//
// Deprecated: use core.AzureDevOpsOrganizationFeatureNames instead.
func AzureSupportedFeatureNames() []string {
	return core.AzureDevOpsOrganizationFeatureNames()
}

// AzureSupportedPermissionGroups returns the deduplicated set of permission
// groups across all features supported by Azure DevOps organizations.
//
// Deprecated: use core.AzureDevOpsOrganizationPermissionGroups instead
func AzureSupportedPermissionGroups() []core.PermissionGroup {
	return core.AzureDevOpsOrganizationPermissionGroups()
}

// AzureSupportedPermissionGroupNames returns the name of all permission groups
// supported by Azure DevOps organizations.
//
// Deprecated: use core.AzureDevOpsOrganizationPermissionGroupNames instead.
func AzureSupportedPermissionGroupNames() []string {
	return core.AzureDevOpsOrganizationPermissionGroupNames()
}

// AzureCheckFeature returns nil if the feature and all of its permission groups
// are supported by Azure DevOps organizations, otherwise it returns an error
// describing what is not supported.
//
// Deprecated: use core.CheckAzureDevOpsOrganizationFeature instead.
func AzureCheckFeature(feature core.Feature) error {
	return core.CheckAzureDevOpsOrganizationFeature(feature)
}

// GitHubSupportedFeatures returns the features and permission groups supported
// by GitHub organizations.
//
// Deprecated: use core.GitHubOrganizationFeatures instead.
func GitHubSupportedFeatures() []core.Feature {
	return core.GitHubOrganizationFeatures()
}

// GitHubSupportedFeatureNames returns the name of all features supported by
// GitHub organizations.
//
// Deprecated: use core.GitHubOrganizationFeatureNames instead.
func GitHubSupportedFeatureNames() []string {
	return core.GitHubOrganizationFeatureNames()
}

// GitHubSupportedPermissionGroups returns the deduplicated set of permission
// groups across all features supported by GitHub organizations.
//
// Deprecated: use core.GitHubOrganizationPermissionGroups instead
func GitHubSupportedPermissionGroups() []core.PermissionGroup {
	return core.GitHubOrganizationPermissionGroups()
}

// GitHubSupportedPermissionGroupNames returns the name of all permission groups
// supported by GitHub organizations.
//
// Deprecated: use core.GitHubOrganizationPermissionGroupNames instead.
func GitHubSupportedPermissionGroupNames() []string {
	return core.GitHubOrganizationPermissionGroupNames()
}

// GitHubCheckFeature returns nil if the feature and all of its permission groups
// are supported by GitHub organizations, otherwise it returns an error
// describing what is not supported.
//
// Deprecated: use core.CheckGitHubOrganizationFeature instead
func GitHubCheckFeature(feature core.Feature) error {
	return core.CheckGitHubOrganizationFeature(feature)
}
