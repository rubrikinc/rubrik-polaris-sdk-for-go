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

package polaris

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

// contains returns true if all values are present in slice.
func contains(slice, values []string) bool {
	for _, v := range values {
		found := false
		for _, s := range slice {
			if v == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// GcpProjectFeature GCP project feature.
type GcpProjectFeature struct {
	Feature string
	Status  string
}

// GcpProject GCP project.
type GcpProject struct {
	ID       string
	Name     string
	Features []GcpProjectFeature

	// GCP
	ProjectName      string
	ProjectID        string
	ProjectNumber    int64
	OrganizationName string
}

// GcpProject returns a single GCP project matching the query option or an
// error. At the moment only projects with Cloud Native Protection are
// returned.
func (c *Client) GcpProject(ctx context.Context, opt QueryOption) (GcpProject, error) {
	c.log.Print(log.Trace, "polaris.Client.GcpProject")

	projects, err := c.GcpProjects(ctx, opt)
	switch {
	case err != nil:
		return GcpProject{}, err
	case len(projects) < 1:
		return GcpProject{}, fmt.Errorf("polaris: project %w", ErrNotFound)
	case len(projects) > 1:
		return GcpProject{}, fmt.Errorf("polaris: project %w", ErrNotUnique)
	}

	return projects[0], nil
}

// GcpProjects returns all GCP projects matching the given query option. At the
// moment only projects with Cloud Native Protection are returned.
func (c *Client) GcpProjects(ctx context.Context, opt QueryOption) ([]GcpProject, error) {
	c.log.Print(log.Trace, "polaris.Client.GcpProjects")

	opts := options{}
	if opt == nil {
		return nil, errors.New("polaris: option not allowed to be nil")
	}
	if err := opt.query(ctx, &opts); err != nil {
		return nil, err
	}

	var filter string
	switch {
	case opts.gcpID != "":
		filter = opts.gcpID
	case opts.gcpNumber != 0:
		filter = strconv.FormatInt(opts.gcpNumber, 10)
	case opts.gcpName != "":
		filter = opts.gcpName
	case opts.name != "":
		filter = opts.name
	}

	// Allow us to search for project name, id and number.
	projs, err := c.gql.GcpCloudAccountListProjects(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	// Merge features on project number and lookup additional information from
	// the native project structure.
	projSet := make(map[int64]*GcpProject)
	for _, p := range projs {
		pn := p.Project.ProjectNumber

		project, ok := projSet[pn]
		if !ok {
			natives, err := c.gql.GcpNativeProjectConnection(ctx, strconv.FormatInt(pn, 10))
			if err != nil {
				return nil, err
			}
			if len(natives) != 1 {
				return nil, fmt.Errorf("polaris: native project %w", ErrNotUnique)
			}

			project = &GcpProject{
				ID:               p.Project.ID,
				Name:             natives[0].Name,
				ProjectName:      natives[0].NativeName,
				ProjectID:        natives[0].NativeID,
				ProjectNumber:    pn,
				OrganizationName: natives[0].OrganizationName,
			}
			projSet[pn] = project
		}

		project.Features = append(project.Features, GcpProjectFeature{
			Feature: p.FeatureDetail.Feature,
			Status:  p.FeatureDetail.Status,
		})
	}

	projects := make([]GcpProject, 0, len(projSet))
	for _, project := range projSet {
		projects = append(projects, *project)
	}

	return projects, nil
}

// GcpProjectAdd adds the GCP project identified by the GcpConfigOption to
// Polaris. Note that passing a FromGcpProject as the GcpConfigOption requires
// that a GCP service account has been set.
func (c *Client) GcpProjectAdd(ctx context.Context, opt GcpConfigOption) error {
	c.log.Print(log.Trace, "polaris.Client.GcpProjectAdd")

	opts := options{}
	if opt == nil {
		return errors.New("polaris: option not allowed to be nil")
	}
	if err := opt.gcpConfig(ctx, &opts); err != nil {
		return err
	}

	// If we got a service account we check that it has all the permissions
	// required by Polaris.
	var jwtConfig string
	if opts.gcpCreds != nil {
		perms, err := c.gql.GcpCloudAccountListPermissions(ctx)
		if err != nil {
			return err
		}

		client, err := cloudresourcemanager.NewService(ctx, option.WithCredentials(opts.gcpCreds))
		if err != nil {
			return err
		}

		req := client.Projects.TestIamPermissions(opts.gcpID, &cloudresourcemanager.TestIamPermissionsRequest{
			Permissions: perms,
		})
		res, err := req.Do()
		if err != nil {
			return err
		}
		if !contains(res.Permissions, perms) {
			return errors.New("polaris: service account missing permissions")
		}

		jwtConfig = string(opts.gcpCreds.JSON)
	}

	return c.gql.GcpCloudAccountAddManualAuthProject(ctx, opts.gcpName, opts.gcpID, opts.gcpNumber,
		opts.gcpOrgName, jwtConfig)
}

// GcpProjectRemove removes the GCP project identified by the GcpConfigOption
// from Polaris. If deleteSnapshots are true the snapshots are deleted otherwise
// they are kept.
func (c *Client) GcpProjectRemove(ctx context.Context, opt QueryOption, deleteSnapshots bool) error {
	c.log.Print(log.Trace, "polaris.Client.GcpProjectRemove")

	opts := options{}
	if opt == nil {
		return errors.New("polaris: option not allowed to be nil")
	}
	if err := opt.query(ctx, &opts); err != nil {
		return err
	}

	var queryOpt QueryOption
	switch {
	case opts.gcpNumber != 0:
		queryOpt = WithGcpProjectNumber(opts.gcpNumber)
	default:
		queryOpt = WithGcpProjectID(opts.gcpID)
	}

	project, err := c.GcpProject(ctx, queryOpt)
	if err != nil {
		return err
	}

	// Disable project. At the moment GCP only support the Cloud Native
	// Protection feature.
	if n := len(project.Features); n != 1 {
		return fmt.Errorf("polaris: invalid number of features: %d", n)
	}
	if project.Features[0].Status != "DISCONNECTED" {
		natives, err := c.gql.GcpNativeProjectConnection(ctx, strconv.FormatInt(project.ProjectNumber, 10))
		if err != nil {
			return err
		}
		if len(natives) != 1 {
			return fmt.Errorf("polaris: %w", ErrNotUnique)
		}

		taskChainID, err := c.gql.GcpNativeDisableProject(ctx, natives[0].ID, deleteSnapshots)
		if err != nil {
			return err
		}

		taskChainState, err := c.gql.WaitForTaskChain(ctx, taskChainID, 10*time.Second)
		if err != nil {
			return err
		}
		if taskChainState != graphql.TaskChainSucceeded {
			return fmt.Errorf("polaris: taskchain failed: taskChainID=%v, state=%v", taskChainID, taskChainState)
		}
	}

	// Remove project.
	return c.gql.GcpCloudAccountDeleteProjects(ctx, []string{project.ID}, nil, nil)
}

// GcpServiceAccount gets the default GCP service account name. If no default
// GCP service account has been set an empty string is returned.
func (c *Client) GcpServiceAccount(ctx context.Context) (string, error) {
	c.log.Print(log.Trace, "polaris.Client.GcpServiceAccount")

	return c.gql.GcpGetDefaultCredentialsServiceAccount(ctx)
}

// GcpSetServiceAccount sets the default GCP service account. The set service
// account will be used for GCP projects added without a service account key
// file. The optional AddOption can be used to specify a name for the service
// account, otherwise the service account's project name will be used. Note
// that it's not possible to remove a service account once it has been set.
func (c *Client) GcpServiceAccountSet(ctx context.Context, gcpOpt GcpConfigOption, addOpts ...AddOption) error {
	c.log.Print(log.Trace, "polaris.Client.GcpServiceAccountSet")

	opts := options{}
	if gcpOpt == nil {
		return errors.New("polaris: option not allowed to be nil")
	}
	if err := gcpOpt.gcpConfig(ctx, &opts); err != nil {
		return err
	}
	if opts.gcpCreds == nil {
		return errors.New("polaris: missing gcp credentials")
	}
	for _, addOpt := range addOpts {
		if err := addOpt.add(ctx, &opts); err != nil {
			return err
		}
	}
	if opts.name == "" {
		opts.name = opts.gcpName
	}

	// Check that the service account has all permissions required by Polaris.
	perms, err := c.gql.GcpCloudAccountListPermissions(ctx)
	if err != nil {
		return err
	}

	client, err := cloudresourcemanager.NewService(ctx, option.WithCredentials(opts.gcpCreds))
	if err != nil {
		return err
	}

	req := client.Projects.TestIamPermissions(opts.gcpID, &cloudresourcemanager.TestIamPermissionsRequest{
		Permissions: perms,
	})
	res, err := req.Do()
	if err != nil {
		return err
	}
	if !contains(res.Permissions, perms) {
		return errors.New("polaris: service account missing permissions")
	}

	return c.gql.GcpSetDefaultServiceAccount(ctx, string(opts.gcpCreds.JSON), opts.name)
}
