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
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqldevops "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/devops"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// GitHub organizations cannot be onboarded through the SDK due to GitHub API
// limitations. The methods below therefore only read the GitHub hierarchy; use
// the RSC UI to onboard GitHub organizations.

// GitHubOrganizationByID returns the GitHub organization with the specified
// workload ID. If no organization matches the ID, graphql.ErrNotFound is
// returned.
func (a API) GitHubOrganizationByID(ctx context.Context, workloadID uuid.UUID) (gqldevops.GitHubOrganization, error) {
	a.log.Print(log.Trace)

	orgs, err := a.GitHubOrganizations(ctx, gqldevops.QueryTypeChildren, hierarchy.GitHubRoot)
	if err != nil {
		return gqldevops.GitHubOrganization{}, err
	}
	for _, org := range orgs {
		if org.ID == workloadID {
			return org, nil
		}
	}

	return gqldevops.GitHubOrganization{}, fmt.Errorf("github organization %s %w", workloadID, graphql.ErrNotFound)
}

// GitHubOrganizations returns all GitHub organizations under the specified
// ancestor. Pass hierarchy.GitHubRoot as the ancestor ID to enumerate every
// organization in the account. Pass zero or more filters to narrow the results
// server-side.
func (a API) GitHubOrganizations(ctx context.Context, queryType gqldevops.QueryType, ancestorID string, filters ...hierarchy.Filter) ([]gqldevops.GitHubOrganization, error) {
	a.log.Print(log.Trace)

	orgs, err := gqldevops.GitHubOrganizations(ctx, a.client, queryType, ancestorID, filters...)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub organizations: %s", err)
	}

	return orgs, nil
}

// GitHubOrganizationsByName returns all GitHub organizations whose name starts
// with the specified prefix. Pass zero or more additional filters to narrow the
// search server-side.
func (a API) GitHubOrganizationsByName(ctx context.Context, namePrefix string, filters ...hierarchy.Filter) ([]gqldevops.GitHubOrganization, error) {
	a.log.Print(log.Trace)

	filters = append([]hierarchy.Filter{{Field: "NAME_PREFIX", Texts: []string{namePrefix}}}, filters...)
	orgs, err := a.GitHubOrganizations(ctx, gqldevops.QueryTypeChildren, hierarchy.GitHubRoot, filters...)
	if err != nil {
		return nil, err
	}

	return orgs, nil
}

// GitHubRepositoryByID returns the GitHub repository with the specified workload
// ID. If no repository matches the ID, graphql.ErrNotFound is returned.
func (a API) GitHubRepositoryByID(ctx context.Context, workloadID uuid.UUID) (gqldevops.GitHubRepository, error) {
	a.log.Print(log.Trace)

	repos, err := a.GitHubRepositories(ctx, gqldevops.QueryTypeDescendants, hierarchy.GitHubRoot)
	if err != nil {
		return gqldevops.GitHubRepository{}, err
	}
	for _, repo := range repos {
		if repo.ID == workloadID {
			return repo, nil
		}
	}

	return gqldevops.GitHubRepository{}, fmt.Errorf("github repository %s %w", workloadID, graphql.ErrNotFound)
}

// GitHubRepositories returns all GitHub repositories under the specified
// ancestor (typically an organization ID). Pass zero or more filters to narrow
// the results server-side.
func (a API) GitHubRepositories(ctx context.Context, queryType gqldevops.QueryType, ancestorID string, filters ...hierarchy.Filter) ([]gqldevops.GitHubRepository, error) {
	a.log.Print(log.Trace)

	repos, err := gqldevops.GitHubRepositories(ctx, a.client, queryType, ancestorID, filters...)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub repositories: %s", err)
	}

	return repos, nil
}

// GitHubRepositoriesByName returns all GitHub repositories whose name starts
// with the specified prefix. Repository names are only unique within an
// organization, so the results may span organizations; pass an
// organization-scoping filter to restrict them.
func (a API) GitHubRepositoriesByName(ctx context.Context, namePrefix string, filters ...hierarchy.Filter) ([]gqldevops.GitHubRepository, error) {
	a.log.Print(log.Trace)

	filters = append([]hierarchy.Filter{{Field: "NAME_PREFIX", Texts: []string{namePrefix}}}, filters...)
	repos, err := a.GitHubRepositories(ctx, gqldevops.QueryTypeDescendants, hierarchy.GitHubRoot, filters...)
	if err != nil {
		return nil, err
	}

	return repos, nil
}
