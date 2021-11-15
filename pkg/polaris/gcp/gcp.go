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

// Package gcp provides a high level interface to the GCP part of the Polaris
// platform.
package gcp

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for Google Cloud Platform.
type API struct {
	Version string
	gql     *graphql.Client
}

// NewAPI returns a new API instance. Note that this is a very cheap call to
// make.
func NewAPI(gql *graphql.Client) API {
	return API{Version: gql.Version, gql: gql}
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
		if f.Name == feature {
			return f, true
		}
	}

	return Feature{}, false
}

// Feature for Google Cloud Platform projects.
type Feature struct {
	Name   core.Feature
	Status core.Status
}

// toNativeID returns the GCP project id for the specified identity. If the
// identity is a GCP project id no remote endpoint is called.
func (a API) toNativeID(ctx context.Context, id IdentityFunc) (string, error) {
	a.gql.Log().Print(log.Trace)

	if id == nil {
		return "", errors.New("polaris: id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return "", err
	}

	if !identity.internal {
		return identity.id, nil
	}

	uid, err := uuid.Parse(identity.id)
	if err != nil {
		return "", nil
	}

	selectors, err := gcp.Wrap(a.gql).CloudAccountListProjects(ctx, core.FeatureCloudNativeProtection, "")
	if err != nil {
		return "", err
	}

	for _, selector := range selectors {
		if selector.Account.ID == uid {
			return selector.Account.ProjectID, nil
		}
	}

	return "", fmt.Errorf("polaris: account %w", graphql.ErrNotFound)
}

// Polaris does not support the AllFeatures for GCP cloud accounts. We work
// around this by translating FeatureAll to the following list of features.
var allFeatures = []core.Feature{
	core.FeatureCloudAccounts,
	core.FeatureCloudNativeProtection,
	core.FeatureGCPSharedVPCHost,
}

// projects return all projects for the given feature and filter. Note that the
// organization name of the cloud account is not set.
func (a API) projects(ctx context.Context, feature core.Feature, filter string) ([]CloudAccount, error) {
	a.gql.Log().Print(log.Trace)

	accountsWithFeature, err := gcp.Wrap(a.gql).CloudAccountListProjects(ctx, feature, filter)
	if err != nil {
		return nil, err
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
				Name:   accountWithFeature.Feature.Name,
				Status: accountWithFeature.Feature.Status,
			}},
		})
	}

	return accounts, nil
}

// projectsAllFeatures return all projects with all features for the given
// filter. Note that the organization name of the cloud account is not set.
func (a API) projectsAllFeatures(ctx context.Context, filter string) ([]CloudAccount, error) {
	a.gql.Log().Print(log.Trace)

	accountMap := make(map[uuid.UUID]*CloudAccount)
	for _, feature := range allFeatures {
		accounts, err := a.projects(ctx, feature, filter)
		if err != nil {
			return nil, err
		}

		for _, account := range accounts {
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
	a.gql.Log().Print(log.Trace)

	if id == nil {
		return CloudAccount{}, errors.New("polaris: id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return CloudAccount{}, err
	}

	if identity.internal {
		id, err := uuid.Parse(identity.id)
		if err != nil {
			return CloudAccount{}, err
		}

		accounts, err := a.Projects(ctx, feature, "")
		if err != nil {
			return CloudAccount{}, err
		}

		for _, account := range accounts {
			if account.ID == id {
				return account, nil
			}
		}
	} else {
		accounts, err := a.Projects(ctx, feature, identity.id)
		if err != nil {
			return CloudAccount{}, err
		}
		if len(accounts) > 1 {
			return CloudAccount{}, fmt.Errorf("polaris: account %w", graphql.ErrNotUnique)
		}
		if len(accounts) == 1 {
			return accounts[0], nil
		}
	}

	return CloudAccount{}, fmt.Errorf("polaris: account %w", graphql.ErrNotFound)
}

// Projects return all projects with the specified feature matching the filter.
// The filter can be used to search for project id, project name and project
// number.
func (a API) Projects(ctx context.Context, feature core.Feature, filter string) ([]CloudAccount, error) {
	a.gql.Log().Print(log.Trace)

	var accounts []CloudAccount
	var err error
	if feature == core.FeatureAll {
		accounts, err = a.projectsAllFeatures(ctx, filter)
	} else {
		accounts, err = a.projects(ctx, feature, filter)
	}
	if err != nil {
		return nil, err
	}

	// Look up organization name for cloud accounts. Note that if the native
	// project has been disabled we cannot read the organization name and leave
	// it blank. This can happen when RemoveProject times out and gets re-run
	// with the native project already disabled.
	for i := range accounts {
		natives, err := gcp.Wrap(a.gql).NativeProjects(ctx, strconv.FormatInt(accounts[i].ProjectNumber, 10))
		if err != nil {
			return nil, err
		}
		if len(natives) < 1 {
			accounts[i].OrganizationName = "<Native Disabled>"
			continue
		}
		if len(natives) > 1 {
			return nil, fmt.Errorf("polaris: native project %w", graphql.ErrNotUnique)
		}

		accounts[i].OrganizationName = natives[0].OrganizationName
	}

	return accounts, nil
}

// AddProject adds the specified project to Polaris for the given feature.
// If name or organization aren't given as a options they are derived from
// information in the cloud. The result can vary slightly depending on
// permissions. Returns the Polaris cloud account id of the added project.
func (a API) AddProject(ctx context.Context, project ProjectFunc, feature core.Feature, opts ...OptionFunc) (uuid.UUID, error) {
	a.gql.Log().Print(log.Trace)

	if feature != core.FeatureCloudNativeProtection {
		return uuid.Nil, fmt.Errorf("polaris: feature not supported on gcp: %v", core.FormatFeature(feature))
	}

	if project == nil {
		return uuid.Nil, errors.New("polaris: project is not allowed to be nil")
	}
	config, err := project(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return uuid.Nil, err
		}
	}
	if options.name != "" {
		config.name = options.name
	}
	if options.orgName != "" {
		config.orgName = options.orgName
	}

	// If we got a service account we check that it has all the permissions
	// required by Polaris.
	var jwtConfig string
	if config.creds != nil {
		err = a.gcpCheckPermissions(ctx, config.creds, config.id, []core.Feature{feature})
		if err != nil {
			return uuid.Nil, err
		}

		jwtConfig = string(config.creds.JSON)
	}

	err = gcp.Wrap(a.gql).CloudAccountAddManualAuthProject(ctx, config.id, config.name, config.number,
		config.orgName, jwtConfig, feature)
	if err != nil {
		return uuid.Nil, err
	}

	// Get the cloud account id of the newly added project
	account, err := a.Project(ctx, ProjectID(config.id), feature)
	if err != nil {
		return uuid.Nil, err
	}

	return account.ID, nil
}

// RemoveProject removes the project with the specified id from Polaris for the
// given feature. If deleteSnapshots is true the snapshots are deleted otherwise
// they are kept. Note that snapshots are only considered to be deleted when
// removing the cloud native protection feature.
func (a API) RemoveProject(ctx context.Context, id IdentityFunc, feature core.Feature, deleteSnapshots bool) error {
	a.gql.Log().Print(log.Trace)

	account, err := a.Project(ctx, id, feature)
	if err != nil {
		return err
	}
	if n := len(account.Features); n != 1 {
		return fmt.Errorf("polaris: feature %w", graphql.ErrNotFound)
	}

	if account.Features[0].Name == core.FeatureCloudNativeProtection && account.Features[0].Status != core.StatusDisabled {
		// Lookup the Polaris Native ID from the GCP project number. The Polaris
		// Native Account ID is needed to delete the Polaris Native Project.
		natives, err := gcp.Wrap(a.gql).NativeProjects(ctx, strconv.FormatInt(account.ProjectNumber, 10))
		if err != nil {
			return err
		}
		if len(natives) < 1 {
			return fmt.Errorf("polaris: native account: %w", graphql.ErrNotFound)
		}
		if len(natives) > 1 {
			return fmt.Errorf("polaris: native account: %w", graphql.ErrNotUnique)
		}

		jobID, err := gcp.Wrap(a.gql).NativeDisableProject(ctx, natives[0].ID, deleteSnapshots)
		if err != nil {
			return err
		}

		state, err := core.Wrap(a.gql).WaitForTaskChain(ctx, jobID, 10*time.Second)
		if err != nil {
			return err
		}
		if state != core.TaskChainSucceeded {
			return fmt.Errorf("polaris: taskchain failed: jobID=%v, state=%v", jobID, state)
		}
	}

	err = gcp.Wrap(a.gql).CloudAccountDeleteProject(ctx, account.ID)
	if err != nil {
		return err
	}

	return nil
}

// ServiceAccount returns the default service account name. If no default
// service account has been set an empty string is returned.
func (a API) ServiceAccount(ctx context.Context) (string, error) {
	a.gql.Log().Print(log.Trace)

	return gcp.Wrap(a.gql).DefaultCredentialsServiceAccount(ctx)
}

// SetServiceAccount sets the default service account. The service account set
// will be used for projects added without a service account key file. If name
// isn't given as an option it will be derived from information in the cloud.
// The result can vary slightly depending on permissions. The organization
// option does nothing. Note that it's not possible to remove a service account
// once it has been set.
func (a API) SetServiceAccount(ctx context.Context, project ProjectFunc, opts ...OptionFunc) error {
	a.gql.Log().Print(log.Trace)

	if project == nil {
		return errors.New("polaris: project is not allowed to be nil")
	}
	config, err := project(ctx)
	if err != nil {
		return err
	}
	if config.creds == nil {
		return errors.New("polaris: project is missing google credentials")
	}

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return err
		}
	}
	if options.name != "" {
		config.name = options.name
	}

	err = gcp.Wrap(a.gql).SetDefaultServiceAccount(ctx, config.name, string(config.creds.JSON))
	if err != nil {
		return err
	}

	return nil
}
