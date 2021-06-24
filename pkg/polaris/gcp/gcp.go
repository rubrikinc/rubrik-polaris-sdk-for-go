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

package gcp

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

// API for Google Cloud Platform.
type API struct {
	gql *graphql.Client
}

// NewAPI returns a new API instance. Note that this is a very cheap call to
// make.
func NewAPI(gql *graphql.Client) API {
	return API{gql: gql}
}

// CloudAccount for Google Cloud Platform projects.
type CloudAccount struct {
	ID               uuid.UUID
	NativeID         string
	Name             string
	ProjectNumber    int64
	OrganizationName string
	Features         []Feature
}

// Feature for Google Cloud Platform projects.
type Feature struct {
	Name   core.CloudAccountFeature
	Status core.CloudAccountStatus
}

// Project returns the project with specified id and feature.
func (a API) Project(ctx context.Context, id IdentityFunc, feature core.CloudAccountFeature) (CloudAccount, error) {
	a.gql.Log().Print(log.Trace, "polaris/gcp.Project")

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
func (a API) Projects(ctx context.Context, feature core.CloudAccountFeature, filter string) ([]CloudAccount, error) {
	a.gql.Log().Print(log.Trace, "polaris/gcp.Projects")

	selectors, err := gcp.Wrap(a.gql).CloudAccountListProjects(ctx, feature, filter)
	if err != nil {
		return nil, err
	}

	accountMap := make(map[uuid.UUID]*CloudAccount)
	for _, selector := range selectors {
		if account, ok := accountMap[selector.Account.ID]; ok {
			account.Features = append(account.Features, Feature{
				Name:   selector.Feature.Name,
				Status: selector.Feature.Status,
			})
		} else {
			// Look up organization name for cloud account.
			natives, err := gcp.Wrap(a.gql).NativeProjects(ctx, strconv.FormatInt(selector.Account.ProjectNumber, 10))
			if err != nil {
				return nil, err
			}
			if len(natives) != 1 {
				return nil, fmt.Errorf("polaris: native project %w", graphql.ErrNotUnique)
			}

			accountMap[selector.Account.ID] = &CloudAccount{
				ID:               selector.Account.ID,
				NativeID:         selector.Account.ProjectID,
				Name:             selector.Account.Name,
				ProjectNumber:    selector.Account.ProjectNumber,
				OrganizationName: natives[0].OrganizationName,
				Features: []Feature{{
					Name:   selector.Feature.Name,
					Status: selector.Feature.Status,
				}},
			}
		}
	}

	accounts := make([]CloudAccount, 0, len(accountMap))
	for _, account := range accountMap {
		accounts = append(accounts, *account)
	}

	return accounts, nil
}

// diffPermissions returns the required permissions missing in the available
// permissions.
func diffPermissions(required, available []string) []string {
	reqSet := make(map[string]struct{})
	for _, v := range required {
		reqSet[v] = struct{}{}
	}

	for _, perm := range available {
		delete(reqSet, perm)
	}

	missing := make([]string, 0, len(reqSet))
	for perm := range reqSet {
		missing = append(missing, perm)
	}

	return missing
}

// AddProject adds the specified project to Polaris. If name or organization
// aren't given as a options they are derived from information in the cloud.
// The result can vary slightly depending on permissions. Returns the Polaris
// cloud account id of the added project.
func (a API) AddProject(ctx context.Context, project ProjectFunc, opts ...OptionFunc) (uuid.UUID, error) {
	a.gql.Log().Print(log.Trace, "polaris/gcp.AddProject")

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

	// If we got a service account we check that it has all the permissions
	// required by Polaris.
	var jwtConfig string
	if config.creds != nil {
		perms, err := gcp.Wrap(a.gql).CloudAccountListPermissions(ctx, core.CloudNativeProtection)
		if err != nil {
			return uuid.Nil, err
		}

		client, err := cloudresourcemanager.NewService(ctx, option.WithCredentials(config.creds))
		if err != nil {
			return uuid.Nil, err
		}

		res, err := client.Projects.TestIamPermissions(config.id,
			&cloudresourcemanager.TestIamPermissionsRequest{Permissions: perms}).Do()
		if err != nil {
			return uuid.Nil, err
		}

		missing := diffPermissions(perms, res.Permissions)
		if len(missing) > 0 {
			return uuid.Nil, fmt.Errorf("polaris: service account missing permissions: %v", strings.Join(missing, ","))
		}

		jwtConfig = string(config.creds.JSON)
	}

	err = gcp.Wrap(a.gql).CloudAccountAddManualAuthProject(ctx, config.id, config.name, config.number, config.orgName,
		jwtConfig, core.CloudNativeProtection)
	if err != nil {
		return uuid.Nil, err
	}

	// Get the cloud account id of the newly added project
	account, err := a.Project(ctx, ProjectID(config.id), core.CloudNativeProtection)
	if err != nil {
		return uuid.Nil, err
	}

	return account.ID, nil
}

// RemoveProject removes the project with the specified id from Polaris. If
// deleteSnapshots is true the snapshots are deleted otherwise they are kept.
func (a API) RemoveProject(ctx context.Context, id IdentityFunc, deleteSnapshots bool) error {
	a.gql.Log().Print(log.Trace, "polaris/gcp.RemoveProject")

	account, err := a.Project(ctx, id, core.CloudNativeProtection)
	if err != nil {
		return err
	}
	if n := len(account.Features); n != 1 {
		return fmt.Errorf("polaris: invalid number of features: %v", n)
	}

	if account.Features[0].Status != core.Disabled {
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
	a.gql.Log().Print(log.Trace, "polaris/gcp.ServiceAccount")

	return gcp.Wrap(a.gql).DefaultCredentialsServiceAccount(ctx)
}

// SetServiceAccount sets the default service account. The service account set
// will be used for projects added without a service account key file. If name
// isn't given as an option it will be derived from information in the cloud.
// The result can vary slightly depending on permissions. The organization
// option does nothing. Note that it's not possible to remove a service account
// once it has been set.
func (a API) SetServiceAccount(ctx context.Context, project ProjectFunc, opts ...OptionFunc) error {
	a.gql.Log().Print(log.Trace, "polaris/gcp.GcpServiceAccountSet")

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

	// Check that the service account has all permissions required by Polaris.
	perms, err := gcp.Wrap(a.gql).CloudAccountListPermissions(ctx, core.CloudNativeProtection)
	if err != nil {
		return err
	}

	client, err := cloudresourcemanager.NewService(ctx, option.WithCredentials(config.creds))
	if err != nil {
		return err
	}

	res, err := client.Projects.TestIamPermissions(config.id,
		&cloudresourcemanager.TestIamPermissionsRequest{Permissions: perms}).Do()
	if err != nil {
		return err
	}

	missing := diffPermissions(perms, res.Permissions)
	if len(missing) > 0 {
		return fmt.Errorf("polaris: service account missing permissions: %v", strings.Join(missing, ","))
	}

	err = gcp.Wrap(a.gql).SetDefaultServiceAccount(ctx, config.name, string(config.creds.JSON))
	if err != nil {
		return err
	}

	return nil
}
