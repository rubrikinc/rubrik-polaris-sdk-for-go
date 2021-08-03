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
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

// GCP permissions.
type Permissions []string

// stringsDiff returns the difference between lhs and rhs, i.e. rhs subtracted
// from lhs.
func stringsDiff(lhs, rhs []string) []string {
	set := make(map[string]struct{})
	for _, s := range lhs {
		set[s] = struct{}{}
	}

	for _, s := range rhs {
		delete(set, s)
	}

	diff := make([]string, 0, len(set))
	for s := range set {
		diff = append(diff, s)
	}

	return diff
}

// checkPermissions checks that the specified credentials have the correct GCP
// permissions to use the project with the given Polaris features
func (a API) gcpCheckPermissions(ctx context.Context, creds *google.Credentials, projectID string, features []core.Feature) error {
	a.gql.Log().Print(log.Trace, "polaris/gcp.gcpCheckPermissions")

	perms, err := a.Permissions(ctx, features)
	if err != nil {
		return err
	}

	client, err := cloudresourcemanager.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return err
	}

	res, err := client.Projects.TestIamPermissions(projectID,
		&cloudresourcemanager.TestIamPermissionsRequest{Permissions: perms}).Do()
	if err != nil {
		return err
	}

	if missing := stringsDiff(perms, res.Permissions); len(missing) > 0 {
		return fmt.Errorf("polaris: missing permissions: %v", strings.Join(missing, ","))
	}

	return nil
}

// Permissions returns all GCP permissions requried to use the specified
// Polaris features.
func (a API) Permissions(ctx context.Context, features []core.Feature) (Permissions, error) {
	a.gql.Log().Print(log.Trace, "polaris/gcp.Permissions")

	permSet := make(map[string]struct{})
	for _, feature := range features {
		perms, err := gcp.Wrap(a.gql).CloudAccountListPermissions(ctx, feature)
		if err != nil {
			return Permissions{}, err
		}

		for _, perm := range perms {
			permSet[perm] = struct{}{}
		}
	}

	var perms []string
	for perm := range permSet {
		perms = append(perms, perm)
	}

	return perms, nil
}

// PermissionsUpdated should be called after the GCP permissions have been
// updated as a response to an account having the status
// StatusMissingPermissions. This will notify Polaris that the permissions have
// been updated. If features is nil the actual permissions after the update
// will not be verified.
func (a API) PermissionsUpdated(ctx context.Context, project ProjectFunc, features []core.Feature) error {
	a.gql.Log().Print(log.Trace, "polaris/gcp.NotifyPermissions")

	if project == nil {
		return errors.New("polaris: project is not allowed to be nil")
	}
	config, err := project(ctx)
	if err != nil {
		return err
	}

	// We need both project id and features to check the permissions of the
	// service account. This is not available when using a
	if config.id != "" && len(features) > 0 {
		err = a.gcpCheckPermissions(ctx, config.creds, config.id, features)
		if err != nil {
			return err
		}
	}

	// TODO: Invoke Polaris endpoint.

	return nil
}
