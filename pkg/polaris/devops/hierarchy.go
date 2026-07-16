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

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqldevops "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/devops"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AzureOrganizations returns all Azure DevOps organizations under the specified
// ancestor. Pass hierarchy.AzureDevOpsRoot as the ancestor ID to enumerate
// every organization in the account.
func (a API) AzureOrganizations(ctx context.Context, queryType gqldevops.QueryType, ancestorID string) ([]gqldevops.AzureOrganization, error) {
	a.log.Print(log.Trace)

	orgs, err := gqldevops.AzureOrganizations(ctx, a.client, queryType, ancestorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure DevOps organizations: %s", err)
	}

	return orgs, nil
}

// AzureOrganizationByID returns the Azure DevOps organization with the
// specified workload ID. If no organization matches the ID, graphql.ErrNotFound
// is returned.
func (a API) AzureOrganizationByID(ctx context.Context, workloadID uuid.UUID) (gqldevops.AzureOrganization, error) {
	a.log.Print(log.Trace)

	orgs, err := a.AzureOrganizations(ctx, gqldevops.QueryTypeDescendants, hierarchy.AzureDevOpsRoot)
	if err != nil {
		return gqldevops.AzureOrganization{}, err
	}
	for _, org := range orgs {
		if org.ID == workloadID {
			return org, nil
		}
	}

	return gqldevops.AzureOrganization{}, fmt.Errorf("azure devops organization %s %w", workloadID, graphql.ErrNotFound)
}

// AzureProjects returns all Azure DevOps projects under the specified ancestor,
// typically an organization ID.
func (a API) AzureProjects(ctx context.Context, queryType gqldevops.QueryType, ancestorID string) ([]gqldevops.AzureProject, error) {
	a.log.Print(log.Trace)

	projects, err := gqldevops.AzureProjects(ctx, a.client, queryType, ancestorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure DevOps projects: %s", err)
	}

	return projects, nil
}

// AzureProjectByID returns the Azure DevOps project with the specified workload
// ID. If no project matches the ID, graphql.ErrNotFound is returned.
//
// RSC does not surface a not-found signal for a single-project lookup, so the
// projects are enumerated and matched by ID on the client side.
func (a API) AzureProjectByID(ctx context.Context, workloadID uuid.UUID) (gqldevops.AzureProject, error) {
	a.log.Print(log.Trace)

	projects, err := a.AzureProjects(ctx, gqldevops.QueryTypeDescendants, hierarchy.AzureDevOpsRoot)
	if err != nil {
		return gqldevops.AzureProject{}, err
	}
	for _, project := range projects {
		if project.ID == workloadID {
			return project, nil
		}
	}

	return gqldevops.AzureProject{}, fmt.Errorf("azure devops project %s %w", workloadID, graphql.ErrNotFound)
}

// AzureRepositories returns all Azure DevOps repositories under the specified
// ancestor (typically an organization or project ID).
func (a API) AzureRepositories(ctx context.Context, queryType gqldevops.QueryType, ancestorID string) ([]gqldevops.AzureRepository, error) {
	a.log.Print(log.Trace)

	repos, err := gqldevops.AzureRepositories(ctx, a.client, queryType, ancestorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure DevOps repositories: %s", err)
	}

	return repos, nil
}

// AzureRepositoryByID returns the Azure DevOps repository with the specified
// workload ID. If no repository matches the ID, graphql.ErrNotFound is
// returned.
func (a API) AzureRepositoryByID(ctx context.Context, workloadID uuid.UUID) (gqldevops.AzureRepository, error) {
	a.log.Print(log.Trace)

	repos, err := a.AzureRepositories(ctx, gqldevops.QueryTypeDescendants, hierarchy.AzureDevOpsRoot)
	if err != nil {
		return gqldevops.AzureRepository{}, err
	}
	for _, repo := range repos {
		if repo.ID == workloadID {
			return repo, nil
		}
	}

	return gqldevops.AzureRepository{}, fmt.Errorf("azure devops repository %s %w", workloadID, graphql.ErrNotFound)
}
