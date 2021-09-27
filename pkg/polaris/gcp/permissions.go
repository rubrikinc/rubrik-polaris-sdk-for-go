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

// PermissionsUpdated notifies Polaris that the permissions for the GCP service
// account for the Polaris cloud account with the specified id has been
// updated. The permissions should be updated when a feature has the status
// StatusMissingPermissions. Updating the permissions is done outside of this
// SDK. The features parameter is alowed to be nil. When features is
// nil all features are updated. Note that Polaris is only notified about
// features with status StatusMissingPermissions.
func (a API) PermissionsUpdated(ctx context.Context, id IdentityFunc, features []core.Feature) error {
	a.gql.Log().Print(log.Trace, "polaris/gcp.PermissionsUpdated")

	featureSet := make(map[core.Feature]struct{})
	for _, feature := range features {
		featureSet[feature] = struct{}{}
	}

	account, err := a.Project(ctx, id, core.FeatureAll)
	if err != nil {
		return err
	}

	for _, feature := range account.Features {
		if feature.Status != core.StatusMissingPermissions {
			continue
		}

		// Check that the feature is in the feature set unless the set is
		// empty which is when all features should be updated.
		if _, ok := featureSet[feature.Name]; len(featureSet) > 0 && !ok {
			continue
		}

		err := gcp.Wrap(a.gql).UpgradeCloudAccountPermissionsWithoutOAuth(ctx, account.ID, feature.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

// PermissionsUpdatedForDefault notifies Polaris that the permissions for the
// default GCP service account has been updated. The permissions should be
// updated when a feature has the status StatusMissingPermissions. Updating the
// permissions is done outside of the SDK. The features parameter is alowed to
// be nil. When features is nil all features are updated. Note that Polaris is
// only notified about features with status StatusMissingPermissions.
func (a API) PermissionsUpdatedForDefault(ctx context.Context, features []core.Feature) error {
	a.gql.Log().Print(log.Trace, "polaris/gcp.PermissionsUpdatedForDefault")

	featureSet := make(map[core.Feature]struct{})
	for _, feature := range features {
		featureSet[feature] = struct{}{}
	}

	accounts, err := a.Projects(ctx, core.FeatureAll, "")
	if err != nil {
		return err
	}

	for _, account := range accounts {
		if !account.DefaultServiceAccount {
			continue
		}

		for _, feature := range account.Features {
			if feature.Status != core.StatusMissingPermissions {
				continue
			}

			// Check that the feature is in the feature set unless the set is
			// empty which is when all features should be updated.
			if _, ok := featureSet[feature.Name]; len(featureSet) > 0 && !ok {
				continue
			}

			err := gcp.Wrap(a.gql).UpgradeCloudAccountPermissionsWithoutOAuth(ctx, account.ID, feature.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
