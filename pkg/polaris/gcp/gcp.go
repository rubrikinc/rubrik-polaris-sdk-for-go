// Copyright 2021 Rubrik, Inc.
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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for GCP project management.
type API struct {
	client *graphql.Client
	log    log.Logger
}

// Deprecated: use Wrap instead.
func NewAPI(gql *graphql.Client) API {
	return API{client: gql, log: gql.Log()}
}

// Wrap the RSC client in the azure API.
func Wrap(client *polaris.Client) API {
	return API{client: client.GQL, log: client.GQL.Log()}
}

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

// RSC does not support the AllFeatures for GCP cloud accounts. We work around
// this by translating FeatureAll to the following list of features.
var allFeatures = []core.Feature{
	core.FeatureCloudAccounts,
	core.FeatureCloudNativeProtection,
	core.FeatureGCPSharedVPCHost,
}

// projects return all projects for the given feature and filter. Note that the
// organization name of the cloud account is not set.
func (a API) projects(ctx context.Context, feature core.Feature, filter string) ([]CloudAccount, error) {
	a.log.Print(log.Trace)

	accountsWithFeature, err := gcp.Wrap(a.client).CloudAccountProjectsByFeature(ctx, feature, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %v", err)
	}

	accounts := make([]CloudAccount, 0, len(accountsWithFeature))
	for _, accountWithFeature := range accountsWithFeature {
		accounts = append(accounts, CloudAccount{
			ID:                    accountWithFeature.Account.ID,
			NativeID:              accountWithFeature.Account.ProjectID,
			Name:                  accountWithFeature.Account.Name,
			ProjectNumber:         accountWithFeature.Account.ProjectNumber,
			DefaultServiceAccount: accountWithFeature.Account.UsesGlobalConfig,
			Features: []Feature{{
				Feature: core.Feature{Name: accountWithFeature.Feature.Feature},
				Status:  accountWithFeature.Feature.Status,
			}},
		})
	}

	return accounts, nil
}

// projectsAllFeatures return all projects with all features for the given
// filter. Note that the organization name of the cloud account is not set.
func (a API) projectsAllFeatures(ctx context.Context, filter string) ([]CloudAccount, error) {
	a.log.Print(log.Trace)

	accountMap := make(map[uuid.UUID]*CloudAccount)
	for _, feature := range allFeatures {
		accounts, err := a.projects(ctx, feature, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to get projects: %v", err)
		}

		for i := range accounts {
			// We need to create a copy of the account here since we use it as a
			// pointer further down.
			account := accounts[i]

			if mapped, ok := accountMap[account.ID]; ok {
				mapped.Features = append(mapped.Features, account.Features...)
			} else {
				accountMap[account.ID] = &account
			}
		}
	}

	accounts := make([]CloudAccount, 0, len(accountMap))
	for _, account := range accountMap {
		accounts = append(accounts, *account)
	}

	return accounts, nil
}

// Project returns the project with specified id.
func (a API) Project(ctx context.Context, id IdentityFunc, feature core.Feature) (CloudAccount, error) {
	a.log.Print(log.Trace)

	if id == nil {
		return CloudAccount{}, errors.New("id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return CloudAccount{}, fmt.Errorf("failed to lookup identity: %v", err)
	}

	if identity.kind == internalID {
		id, err := uuid.Parse(identity.id)
		if err != nil {
			return CloudAccount{}, fmt.Errorf("failed to parse identity: %v", err)
		}

		accounts, err := a.Projects(ctx, feature, "")
		if err != nil {
			return CloudAccount{}, fmt.Errorf("failed to get projects: %v", err)
		}

		// Find the exact match.
		for _, account := range accounts {
			if account.ID == id {
				return account, nil
			}
		}
	} else {
		accounts, err := a.Projects(ctx, feature, identity.id)
		if err != nil {
			return CloudAccount{}, fmt.Errorf("failed to get projects: %v", err)
		}

		// Find the exact match.
		if identity.kind == externalID {
			for _, account := range accounts {
				if account.NativeID == identity.id {
					return account, nil
				}
			}
		} else {
			for _, account := range accounts {
				if strconv.FormatInt(account.ProjectNumber, 10) == identity.id {
					return account, nil
				}
			}
		}
	}

	return CloudAccount{}, fmt.Errorf("project %w", graphql.ErrNotFound)
}

// Projects return all projects with the specified feature matching the filter.
// The filter can be used to search for project id, project name and project
// number.
func (a API) Projects(ctx context.Context, feature core.Feature, filter string) ([]CloudAccount, error) {
	a.log.Print(log.Trace)

	var accounts []CloudAccount
	var err error
	if feature.Equal(core.FeatureAll) {
		accounts, err = a.projectsAllFeatures(ctx, filter)
	} else {
		accounts, err = a.projects(ctx, feature, filter)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %v", err)
	}

	// Look up organization name for cloud accounts. Note that if the native
	// project has been disabled we cannot read the organization name and leave
	// it blank. This can happen when RemoveProject times out and gets re-run
	// with the native project already disabled.
	for i := range accounts {
		natives, err := gcp.Wrap(a.client).NativeProjects(ctx, strconv.FormatInt(accounts[i].ProjectNumber, 10))
		if err != nil {
			return nil, fmt.Errorf("failed to get native projects: %v", err)
		}

		// Find the exact match.
		accounts[i].OrganizationName = "<Native Disabled>"
		for _, native := range natives {
			if native.NativeID == accounts[i].NativeID {
				accounts[i].OrganizationName = native.OrganizationName
				break
			}
		}
	}

	return accounts, nil
}

// AddProject adds the specified project to RSC for the given feature. If name
// or organization aren't given as an options they are derived from information
// in the cloud. The result can vary slightly depending on permissions. Returns
// the RSC cloud account id of the added project.
func (a API) AddProject(ctx context.Context, project ProjectFunc, feature core.Feature, opts ...OptionFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if !feature.Equal(core.FeatureCloudNativeProtection) {
		return uuid.Nil, fmt.Errorf("feature not supported on gcp: %v", feature)
	}

	if project == nil {
		return uuid.Nil, errors.New("project is not allowed to be nil")
	}
	config, err := project(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup project: %v", err)
	}

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return uuid.Nil, fmt.Errorf("failed to lookup option: %v", err)
		}
	}
	if options.name != "" {
		config.name = options.name
	}
	if options.orgName != "" {
		config.orgName = options.orgName
	}

	// If we got a service account we check that it has all the permissions
	// required by RSC.
	var jwtConfig string
	if config.creds != nil {
		err = a.gcpCheckPermissions(ctx, config.creds, config.id, []core.Feature{feature})
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to check permissions: %v", err)
		}

		jwtConfig = string(config.creds.JSON)
	}

	err = gcp.Wrap(a.client).CloudAccountAddManualAuthProject(ctx, config.id, config.name, config.number,
		config.orgName, jwtConfig, feature)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add project: %v", err)
	}

	// Get the cloud account id of the newly added project
	account, err := a.Project(ctx, ProjectID(config.id), feature)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get project: %v", err)
	}

	return account.ID, nil
}

// RemoveProject removes the project with the specified id from RSC for the
// given feature. If deleteSnapshots is true the snapshots are deleted otherwise
// they are kept. Note that snapshots are only considered to be deleted when
// removing the cloud native protection feature.
func (a API) RemoveProject(ctx context.Context, id IdentityFunc, feature core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	account, err := a.Project(ctx, id, feature)
	if err != nil {
		return fmt.Errorf("failed to lookup project: %v", err)
	}
	if n := len(account.Features); n != 1 {
		return fmt.Errorf("feature %w", graphql.ErrNotFound)
	}

	if account.Features[0].Equal(core.FeatureCloudNativeProtection) && account.Features[0].Status != core.StatusDisabled {
		// Lookup the RSC Native ID from the GCP project number. The RSC Native
		// Account ID is needed to delete the RSC Native Project.
		natives, err := gcp.Wrap(a.client).NativeProjects(ctx, strconv.FormatInt(account.ProjectNumber, 10))
		if err != nil {
			return fmt.Errorf("failed to get native projects: %v", err)
		}

		// Find the exact match.
		var nativeID uuid.UUID
		for _, native := range natives {
			if native.NativeID == account.NativeID {
				nativeID = native.ID
				break
			}
		}
		if nativeID == uuid.Nil {
			return fmt.Errorf("native project %w", graphql.ErrNotFound)
		}

		jobID, err := gcp.Wrap(a.client).NativeDisableProject(ctx, nativeID, deleteSnapshots)
		if err != nil {
			return fmt.Errorf("failed to disable native project: %v", err)
		}

		if err := core.Wrap(a.client).WaitForFeatureDisableTaskChain(ctx, jobID, func(ctx context.Context) (bool, error) {
			account, err := a.Project(ctx, id, feature)
			if err != nil {
				return false, fmt.Errorf("failed to retrieve status for feature %s: %s", feature, err)
			}

			feature, ok := account.Feature(feature)
			if !ok {
				return false, fmt.Errorf("failed to retrieve status for feature %s: not found", feature)
			}
			return feature.Status == core.StatusDisabled, nil
		}); err != nil {
			return fmt.Errorf("failed to wait for task chain %s: %s", jobID, err)
		}
	}

	err = gcp.Wrap(a.client).CloudAccountDeleteProject(ctx, account.ID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %v", err)
	}

	return nil
}

// ServiceAccount returns the default service account name. If no default
// service account has been set an empty string is returned.
func (a API) ServiceAccount(ctx context.Context) (string, error) {
	a.log.Print(log.Trace)

	account, err := gcp.Wrap(a.client).DefaultCredentialsServiceAccount(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get default service account: %v", err)
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
		return fmt.Errorf("failed to lookup project: %v", err)
	}
	if config.creds == nil {
		return errors.New("project is missing credentials")
	}

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return fmt.Errorf("failed to lookup option: %v", err)
		}
	}
	if options.name != "" {
		config.name = options.name
	}

	err = gcp.Wrap(a.client).SetDefaultServiceAccount(ctx, config.name, string(config.creds.JSON))
	if err != nil {
		return fmt.Errorf("failed to set default service account: %v", err)
	}

	return nil
}
