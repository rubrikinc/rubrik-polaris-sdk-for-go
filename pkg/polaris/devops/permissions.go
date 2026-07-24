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
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	gqldevops "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/devops"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ListPermissions returns the most recent permission definitions available for
// the specified Azure DevOps features and permission groups. When no permission
// groups are specified, RSC returns the permissions for all permission groups.
func (a API) ListPermissions(ctx context.Context, features []core.Feature) (gqldevops.Permissions, error) {
	a.log.Print(log.Trace)

	permissions, err := gqldevops.ListLatestPermissions(ctx, a.client, features)
	if err != nil {
		return gqldevops.Permissions{}, fmt.Errorf("failed to list latest Azure DevOps permissions: %s", err)
	}
	if err := sortPermissions(permissions); err != nil {
		return gqldevops.Permissions{}, fmt.Errorf("failed to list latest Azure DevOps permissions: %s", err)
	}

	return permissions, nil
}

// ListOrgPermissions returns the features, permission groups and permissions
// currently configured for the specified Azure DevOps organization.
func (a API) ListOrgPermissions(ctx context.Context, organizationID uuid.UUID) (gqldevops.Permissions, error) {
	a.log.Print(log.Trace)

	// Workaround: the endpoint currently requires at least the feature names to
	// be specified and returns nothing when they are omitted. Until this is
	// fixed the function defaults to the AZURE_DEVOPS_REPOSITORY_PROTECTION
	// feature.
	features := []core.Feature{core.FeatureAzureDevOpsRepositoryProtection}

	permissions, err := gqldevops.ListCurrentPermissions(ctx, a.client, organizationID, features)
	if err != nil {
		return gqldevops.Permissions{}, fmt.Errorf("failed to list current Azure DevOps permissions: %s", err)
	}
	if err := sortPermissions(permissions); err != nil {
		return gqldevops.Permissions{}, fmt.Errorf("failed to list current Azure DevOps permissions: %s", err)
	}

	return permissions, nil
}

// sortPermissions orders the permissions so the result is deterministic
// regardless of the order in which RSC returns the elements. It sorts the
// feature, permission-group and version slices, and canonicalizes each
// feature's permission JSON document with sortPermissionJSON so that callers
// which hash or compare the raw document get a stable result.
func sortPermissions(permissions gqldevops.Permissions) error {
	// Sort feature permissions.
	slices.SortFunc(permissions.FeaturePermissions, func(a, b gqldevops.FeaturePermission) int {
		return cmp.Compare(a.Feature, b.Feature)
	})
	for i := range permissions.FeaturePermissions {
		sortedPermissions, err := sortPermissionJSON(permissions.FeaturePermissions[i].Permissions)
		if err != nil {
			return fmt.Errorf("failed to canonicalize permission JSON: %s", err)
		}
		permissions.FeaturePermissions[i].Permissions = sortedPermissions

		slices.SortFunc(permissions.FeaturePermissions[i].PermissionGroupVersions, func(a, b gqldevops.PermissionGroupVersion) int {
			return cmp.Compare(a.PermissionGroup, b.PermissionGroup)
		})
	}

	// Sort group permissions.
	slices.SortFunc(permissions.GroupPermissions, func(a, b gqldevops.GroupPermissions) int {
		return cmp.Compare(a.Group, b.Group)
	})
	for i := range permissions.GroupPermissions {
		slices.Sort(permissions.GroupPermissions[i].Permissions)
	}

	return nil
}

// sortPermissionJSON returns the permission JSON document in a canonical form.
// The document is a JSON array of objects: the array elements are ordered by
// their marshaled bytes and, because encoding/json marshals map keys in sorted
// order, every object's keys are ordered too. An identical set of permissions
// therefore always produces an identical document, regardless of the order in
// which RSC returns it.
func sortPermissionJSON(permissionJSON string) (string, error) {
	if permissionJSON == "" {
		return permissionJSON, nil
	}

	var array []map[string]any
	if err := json.Unmarshal([]byte(permissionJSON), &array); err != nil {
		return "", err
	}

	slices.SortFunc(array, func(lhs, rhs map[string]any) int {
		buf1, _ := json.Marshal(lhs)
		buf2, _ := json.Marshal(rhs)
		return bytes.Compare(buf1, buf2)
	})

	buf, err := json.Marshal(array)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
