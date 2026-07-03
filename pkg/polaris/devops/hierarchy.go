// Copyright 2024 Rubrik, Inc.
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
	"context"
	"fmt"

	"github.com/google/uuid"

	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	gqldevops "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/devops"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AzureOrganization is an onboarded Azure DevOps organization enriched with the Azure
// cloud type. The organization read itself does not return the cloud type, so it
// is resolved separately by the organization's Azure AD tenant domain.
type AzureOrganization struct {
	gqldevops.AzureOrganization
	Cloud gqlazure.Cloud
}

// AzureOrganizations returns all Azure DevOps organizations under the specified
// ancestor, each enriched with its Azure cloud type. Pass
// hierarchy.AzureDevOpsRoot as the ancestor ID to enumerate every organization
// in the account.
func (a API) AzureOrganizations(ctx context.Context, queryType gqldevops.QueryType, ancestorID string) ([]AzureOrganization, error) {
	a.log.Print(log.Trace)

	orgs, err := gqldevops.AzureOrganizations(ctx, a.client, queryType, ancestorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure DevOps organizations: %w", err)
	}

	clouds, err := a.azureCloudsByDomain(ctx)
	if err != nil {
		return nil, err
	}

	augmented := make([]AzureOrganization, 0, len(orgs))
	for _, org := range orgs {
		augmented = append(augmented, AzureOrganization{AzureOrganization: org, Cloud: clouds[org.TenantDomain]})
	}

	return augmented, nil
}

// AzureOrganizationByID returns the Azure DevOps organization with the specified
// workload ID, enriched with its Azure cloud type.
func (a API) AzureOrganizationByID(ctx context.Context, workloadID uuid.UUID) (AzureOrganization, error) {
	a.log.Print(log.Trace)

	org, err := gqldevops.AzureOrganizationByID(ctx, a.client, workloadID)
	if err != nil {
		return AzureOrganization{}, fmt.Errorf("failed to get Azure DevOps organization: %w", err)
	}

	clouds, err := a.azureCloudsByDomain(ctx)
	if err != nil {
		return AzureOrganization{}, err
	}

	return AzureOrganization{AzureOrganization: org, Cloud: clouds[org.TenantDomain]}, nil
}

// azureCloudsByDomain maps each Azure DevOps tenant's Azure AD domain to its Azure
// cloud type via a single allAzureCloudAccountTenants query. Only the cloud type
// and domain are reliable for DevOps tenants, so only those are used.
func (a API) azureCloudsByDomain(ctx context.Context) (map[string]gqlazure.Cloud, error) {
	tenants, err := gqlazure.Wrap(a.client).CloudAccountTenants(ctx, core.FeatureAzureDevOpsProtection, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure DevOps tenant cloud types: %w", err)
	}

	clouds := make(map[string]gqlazure.Cloud, len(tenants))
	for _, tenant := range tenants {
		clouds[tenant.DomainName] = tenant.Cloud
	}

	return clouds, nil
}

// AzureProjects returns all Azure DevOps projects under the specified ancestor
// (typically an organization ID).
func (a API) AzureProjects(ctx context.Context, queryType gqldevops.QueryType, ancestorID string) ([]gqldevops.AzureProject, error) {
	a.log.Print(log.Trace)

	projects, err := gqldevops.AzureProjects(ctx, a.client, queryType, ancestorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure DevOps projects: %w", err)
	}

	return projects, nil
}

// AzureProjectByID returns the Azure DevOps project with the specified workload ID.
func (a API) AzureProjectByID(ctx context.Context, workloadID uuid.UUID) (gqldevops.AzureProject, error) {
	a.log.Print(log.Trace)

	project, err := gqldevops.AzureProjectByID(ctx, a.client, workloadID)
	if err != nil {
		return gqldevops.AzureProject{}, fmt.Errorf("failed to get Azure DevOps project: %w", err)
	}

	return project, nil
}

// AzureRepositories returns all Azure DevOps repositories under the specified
// ancestor (typically an organization or project ID).
func (a API) AzureRepositories(ctx context.Context, queryType gqldevops.QueryType, ancestorID string) ([]gqldevops.AzureRepository, error) {
	a.log.Print(log.Trace)

	repos, err := gqldevops.AzureRepositories(ctx, a.client, queryType, ancestorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure DevOps repositories: %w", err)
	}

	return repos, nil
}

// AzureRepositoryByID returns the Azure DevOps repository with the specified workload
// ID.
func (a API) AzureRepositoryByID(ctx context.Context, workloadID uuid.UUID) (gqldevops.AzureRepository, error) {
	a.log.Print(log.Trace)

	repo, err := gqldevops.AzureRepositoryByID(ctx, a.client, workloadID)
	if err != nil {
		return gqldevops.AzureRepository{}, fmt.Errorf("failed to get Azure DevOps repository: %w", err)
	}

	return repo, nil
}
