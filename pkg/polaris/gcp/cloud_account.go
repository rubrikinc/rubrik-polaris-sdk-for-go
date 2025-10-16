// Copyright 2025 Rubrik, Inc.
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

// Package gcp provides a high level interface to the GCP part of the RSC
// platform.
package gcp

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CloudAccount for Google Cloud Platform projects. If DefaultServiceAccount is
// true the cloud account depends on the default service account.
type CloudAccount struct {
	ID                    uuid.UUID
	NativeID              string
	Name                  string
	ProjectNumber         int64
	OrganizationName      string
	DefaultServiceAccount bool
	Features              []Feature
}

// Feature returns the specified feature from the CloudAccount's features.
func (c CloudAccount) Feature(feature core.Feature) (Feature, bool) {
	for _, f := range c.Features {
		if f.Feature.Equal(feature) {
			return f, true
		}
	}

	return Feature{}, false
}

// Feature for Google Cloud Platform projects.
type Feature struct {
	core.Feature
	Status core.Status
}

// ProjectByID returns the project with the specified RSC cloud account ID.
func (a API) ProjectByID(ctx context.Context, cloudAccountID uuid.UUID) (CloudAccount, error) {
	a.log.Print(log.Trace)

	projects, err := a.Projects(ctx, "")
	if err != nil {
		return CloudAccount{}, err
	}
	for _, project := range projects {
		if project.ID == cloudAccountID {
			return project, nil
		}
	}

	return CloudAccount{}, fmt.Errorf("project %q %w", cloudAccountID, graphql.ErrNotFound)
}

// ProjectByNativeID returns the project with the specified native ID.
// For GCP cloud accounts the native ID is the project name.
func (a API) ProjectByNativeID(ctx context.Context, nativeID string) (CloudAccount, error) {
	a.log.Print(log.Trace)

	projects, err := a.Projects(ctx, "")
	if err != nil {
		return CloudAccount{}, err
	}
	for _, project := range projects {
		if project.NativeID == nativeID {
			return project, nil
		}
	}

	return CloudAccount{}, fmt.Errorf("project %q %w", nativeID, graphql.ErrNotFound)
}

// ProjectByName returns the project with the specified name.
func (a API) ProjectByName(ctx context.Context, name string) (CloudAccount, error) {
	a.log.Print(log.Trace)

	projects, err := a.Projects(ctx, name)
	if err != nil {
		return CloudAccount{}, err
	}

	for _, project := range projects {
		if project.Name == name {
			return project, nil
		}
	}

	return CloudAccount{}, fmt.Errorf("project %q %w", name, graphql.ErrNotFound)

}

// Projects return all projects matching the specified filter. The filter can
// be used to search for project ID, project name and project number.
func (a API) Projects(ctx context.Context, filter string) ([]CloudAccount, error) {
	a.log.Print(log.Trace)

	rawProjects, err := gcp.Wrap(a.client).CloudAccountProjectsByFeature(ctx, core.FeatureAll, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %s", err)
	}

	// Look up organization name for cloud accounts. Note that if the native
	// project has been disabled we cannot read the organization name.
	// This can happen when RemoveProject times out and gets re-run with the
	// native project already disabled.
	accounts := toProjects(rawProjects)
	for i := range accounts {
		natives, err := gcp.Wrap(a.client).NativeProjects(ctx, strconv.FormatInt(accounts[i].ProjectNumber, 10))
		if err != nil {
			return nil, fmt.Errorf("failed to get native projects: %s", err)
		}
		accounts[i].OrganizationName = "<DISABLED>"
		for _, native := range natives {
			if native.NativeID == accounts[i].NativeID {
				accounts[i].OrganizationName = native.OrganizationName
				break
			}
		}
	}

	return accounts, nil
}

// AddProject adds the specified project and features to RSC. Returns the RSC
// cloud account ID of the project.
// If name or organization isn't given as an option they are derived from
// information in the cloud. The result can vary slightly depending on
// permissions.
func (a API) AddProject(ctx context.Context, project ProjectFunc, features []core.Feature, opts ...OptionFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if project == nil {
		return uuid.Nil, errors.New("project is not allowed to be nil")
	}
	config, err := project(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup project: %s", err)
	}

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return uuid.Nil, fmt.Errorf("failed to lookup option: %s", err)
		}
	}
	if options.name != "" {
		config.name = options.name
	}
	if options.orgName != "" {
		config.orgName = options.orgName
	}

	// If the user provided a service account, we check that it has all the
	// permissions required by RSC.
	var jwtConfig string
	if config.creds != nil {
		err = a.gcpCheckPermissions(ctx, config.creds, config.id, features)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to check permissions: %s", err)
		}
		jwtConfig = string(config.creds.JSON)
	}

	if err := gcp.Wrap(a.client).CloudAccountAddManualAuthProject(
		ctx, config.id, config.name, config.number, config.orgName, jwtConfig, features); err != nil {
		return uuid.Nil, fmt.Errorf("failed to add project: %s", err)
	}

	account, err := a.ProjectByNativeID(ctx, config.id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get project: %s", err)
	}

	return account.ID, nil
}

// RemoveProject removes the features from the project with the specified RSC
// cloud account ID. If deleteSnapshots is true and a cloud native protection
// feature is being removed, the snapshots are deleted, otherwise they are kept.
func (a API) RemoveProject(ctx context.Context, cloudAccountID uuid.UUID, features []core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	account, err := a.ProjectByID(ctx, cloudAccountID)
	if err != nil {
		return fmt.Errorf("failed to lookup project: %s", err)
	}
	for _, feature := range features {
		if _, ok := account.Feature(feature); !ok {
			return fmt.Errorf("cannot remove feature not added: %s", feature)
		}
	}

	jobs, err := gcp.Wrap(a.client).CloudAccountDeleteProjectV2(ctx, account.ID, features, deleteSnapshots)
	if err != nil {
		return fmt.Errorf("failed to delete project: %s", err)
	}

	for _, job := range jobs {
		feature := job.Feature
		if err := core.Wrap(a.client).WaitForFeatureDisableTaskChain(ctx, job.ID, func(ctx context.Context) (bool, error) {
			account, err := a.ProjectByID(ctx, account.ID)
			if errors.Is(err, graphql.ErrNotFound) {
				return true, nil
			}
			if err != nil {
				return false, err
			}

			if _, ok := account.Feature(feature); ok {
				return false, nil
			}
			return true, nil
		}); err != nil {
			return fmt.Errorf("failed to wait for task chain %s: %s", job.ID, err)
		}
	}

	return nil
}

// ServiceAccount returns the default service account name. If no default
// service account has been set an empty string is returned.
func (a API) ServiceAccount(ctx context.Context) (string, error) {
	a.log.Print(log.Trace)

	account, err := gcp.Wrap(a.client).DefaultCredentialsServiceAccount(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get default service account: %s", err)
	}

	return account, nil
}

// SetServiceAccount sets the default service account. The service account set
// will be used for projects added without a service account key file. If name
// isn't given as an option it will be derived from information in the cloud.
// The result can vary slightly depending on permissions. The organization
// option does nothing. Note that it's not possible to remove a service account
// once it has been set.
func (a API) SetServiceAccount(ctx context.Context, project ProjectFunc, opts ...OptionFunc) error {
	a.log.Print(log.Trace)

	if project == nil {
		return errors.New("project is not allowed to be nil")
	}
	config, err := project(ctx)
	if err != nil {
		return fmt.Errorf("failed to lookup project: %s", err)
	}
	if config.creds == nil {
		return errors.New("project is missing credentials")
	}

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return fmt.Errorf("failed to lookup option: %s", err)
		}
	}
	if options.name != "" {
		config.name = options.name
	}

	err = gcp.Wrap(a.client).SetDefaultServiceAccount(ctx, config.name, string(config.creds.JSON))
	if err != nil {
		return fmt.Errorf("failed to set default service account: %s", err)
	}

	return nil
}

func toProjects(projectsByFeature []gcp.CloudAccountWithFeature) []CloudAccount {
	accountMap := make(map[uuid.UUID]*CloudAccount)

	for _, projectByFeature := range projectsByFeature {
		account, ok := accountMap[projectByFeature.Account.ID]
		if !ok {
			account = &CloudAccount{
				ID:                    projectByFeature.Account.ID,
				NativeID:              projectByFeature.Account.ProjectID,
				Name:                  projectByFeature.Account.Name,
				ProjectNumber:         projectByFeature.Account.ProjectNumber,
				DefaultServiceAccount: projectByFeature.Account.UsesGlobalConfig,
			}
			accountMap[projectByFeature.Account.ID] = account
		}

		account.Features = append(account.Features, Feature{
			Feature: core.Feature{Name: projectByFeature.Feature.Feature},
			Status:  projectByFeature.Feature.Status,
		})
	}

	accounts := make([]CloudAccount, 0, len(accountMap))
	for _, account := range accountMap {
		accounts = append(accounts, *account)
	}

	return accounts
}
