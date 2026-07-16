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

// Package devops provides a high-level interface to the DevOps (Azure DevOps)
// part of the RSC platform. Onboarding assumes the customer application has
// already been registered via the azure package using the
// azure.AppUseCaseDevOps use case.
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

// SupportedFeatures returns the features supported by Azure DevOps
// organizations.
func SupportedFeatures() []core.Feature {
	return []core.Feature{
		core.FeatureAzureDevOpsProtection,
		core.FeatureAzureDevOpsRepositoryProtection,
		core.FeatureAzureDevOpsDeveloperCollaborationProtection,
	}
}

// SupportedFeatureNames returns the features supported by Azure DevOps
// organizations.
func SupportedFeatureNames() []string {
	return []string{
		core.FeatureAzureDevOpsProtection.Name,
		core.FeatureAzureDevOpsRepositoryProtection.Name,
		core.FeatureAzureDevOpsDeveloperCollaborationProtection.Name,
	}
}
