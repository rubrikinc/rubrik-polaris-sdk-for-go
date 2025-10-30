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
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Deprecated: use FeaturePermissions instead.
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
// permissions to use the project with the given RSC features
func (a API) gcpCheckPermissions(ctx context.Context, creds *google.Credentials, projectID string, features []core.Feature) error {
	a.log.Print(log.Trace)

	perms, err := a.Permissions(ctx, features)
	if err != nil {
		return fmt.Errorf("failed to get permissions: %s", err)
	}

	client, err := cloudresourcemanager.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return fmt.Errorf("failed to create GCP Cloud Resource Manager client: %s", err)
	}

	// TestIamPermissions can only test 100 permissions at a time.
	for i, j := 0, 0; j < len(perms); i += 100 {
		j = min(i+100, len(perms))

		res, err := client.Projects.TestIamPermissions(projectID,
			&cloudresourcemanager.TestIamPermissionsRequest{Permissions: perms[i:j]}).Do()
		if err != nil {
			return fmt.Errorf("failed to test GCP IAM permissions: %s", err)
		}

		if missing := stringsDiff(perms[i:j], res.Permissions); len(missing) > 0 {
			return fmt.Errorf("missing permissions: %s", strings.Join(missing, ","))
		}
	}

	return nil
}

// Deprecated: use FeaturePermissions instead.
func (a API) Permissions(ctx context.Context, features []core.Feature) (Permissions, error) {
	a.log.Print(log.Trace)

	featurePerms, err := a.FeaturePermissions(ctx, features)
	if err != nil {
		return Permissions{}, err
	}

	var perms Permissions
	for _, feature := range featurePerms {
		perms = append(perms, feature.WithConditions...)
		perms = append(perms, feature.WithoutConditions...)
	}
	slices.Sort(perms)
	perms = slices.Compact(perms)

	return perms, nil
}

// FeaturePermissions holds the GCP permissions needed for a set of RSC
// features.
type FeaturePermissions struct {
	Feature core.Feature
	permissionsJson
}

type permissionsJson struct {
	Conditions        []string `json:"Conditions"`
	Services          []string `json:"Services"`
	WithConditions    []string `json:"PermissionsWithConditions"`
	WithoutConditions []string `json:"PermissionsWithoutConditions"`
}

func (a API) FeaturePermissions(ctx context.Context, features []core.Feature) ([]FeaturePermissions, error) {
	a.log.Print(log.Trace)

	permissions, err := gcp.Wrap(a.client).AllLatestFeaturePermissions(ctx, features)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %s", err)
	}

	featurePerms := make([]FeaturePermissions, 0, len(permissions))
	for _, feature := range permissions {
		var perms permissionsJson
		if err := json.Unmarshal([]byte(feature.PermissionJson), &perms); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions %s: %s", feature.PermissionJson, err)
		}
		slices.Sort(perms.Conditions)
		slices.Sort(perms.WithConditions)
		slices.Sort(perms.WithoutConditions)
		slices.Sort(perms.Services)

		var pgs []core.PermissionGroup
		for _, pg := range feature.PermissionGroupVersions {
			pgs = append(pgs, core.PermissionGroup(pg.PermissionGroup))
		}
		slices.Sort(pgs)

		featurePerms = append(featurePerms, FeaturePermissions{
			Feature:         core.Feature{Name: feature.Feature, PermissionGroups: pgs},
			permissionsJson: perms,
		})
	}

	return featurePerms, nil
}

// PermissionsUpdated notifies RSC that the permissions for the GCP service
// account for the RSC cloud account with the specified id has been updated.
// The permissions should be updated when a feature has the status
// StatusMissingPermissions. Updating the permissions is done outside this SDK.
// The features parameter is allowed to be nil. When features is nil all
// features are updated. Note that RSC is only notified about features with
// status StatusMissingPermissions.
func (a API) PermissionsUpdated(ctx context.Context, cloudAccountID uuid.UUID, features []core.Feature) error {
	a.log.Print(log.Trace)

	featureSet := make(map[string]struct{})
	for _, feature := range features {
		featureSet[feature.Name] = struct{}{}
	}

	account, err := a.ProjectByID(ctx, cloudAccountID)
	if err != nil {
		return fmt.Errorf("failed to get project: %s", err)
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

		err := gcp.Wrap(a.client).UpgradeCloudAccountPermissionsWithoutOAuth(ctx, account.ID, feature.Feature)
		if err != nil {
			return fmt.Errorf("failed to update permissions: %s", err)
		}
	}

	return nil
}

// PermissionsUpdatedForDefault notifies RSC that the permissions for the
// default GCP service account has been updated. The permissions should be
// updated when a feature has the status StatusMissingPermissions. Updating the
// permissions is done outside the SDK. The features parameter is allowed to be
// nil. When features is nil all features are updated. Note that RSC is only
// notified about features with status StatusMissingPermissions.
func (a API) PermissionsUpdatedForDefault(ctx context.Context, features []core.Feature) error {
	a.log.Print(log.Trace)

	featureSet := make(map[string]struct{})
	for _, feature := range features {
		featureSet[feature.Name] = struct{}{}
	}

	accounts, err := a.Projects(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get project: %s", err)
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

			err := gcp.Wrap(a.client).UpgradeCloudAccountPermissionsWithoutOAuth(ctx, account.ID, feature.Feature)
			if err != nil {
				return fmt.Errorf("failed to update permissions: %s", err)
			}
		}
	}

	return nil
}
