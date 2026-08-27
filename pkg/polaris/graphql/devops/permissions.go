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
	"encoding/json"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Permissions holds the latest permission definitions available for a set of
// DevOps features and permission groups.
type Permissions struct {
	FeaturePermissions []FeaturePermission `json:"featurePermissions"`
	GroupPermissions   []GroupPermissions  `json:"groupPermissions"`
}

// ToFeatures flattens the permissions into a slice of core.Feature values, one
// per feature permission, each carrying the permission groups enabled for that
// feature.
func (p Permissions) ToFeatures() []core.Feature {
	features := make([]core.Feature, 0, len(p.FeaturePermissions))
	for _, fp := range p.FeaturePermissions {
		groups := make([]core.PermissionGroup, 0, len(fp.PermissionGroupVersions))
		for _, group := range fp.PermissionGroupVersions {
			groups = append(groups, group.PermissionGroup)
		}
		features = append(features, core.Feature{
			Name:             fp.Feature,
			PermissionGroups: groups,
		})
	}

	return features
}

// FeaturePermission holds the latest permission definition for a single
// feature. PermissionJSON is the permission document returned verbatim by RSC;
// decoding it is left to the caller.
type FeaturePermission struct {
	Feature                 string                   `json:"feature"`
	Permissions             string                   `json:"permissionJson"`
	PermissionGroupVersions []PermissionGroupVersion `json:"permissionsGroupVersions"`
}

// GroupPermissions holds the latest permissions for a single permission group.
type GroupPermissions struct {
	Group       core.PermissionGroup `json:"group"`
	Permissions []string             `json:"permissions"`
	Version     int                  `json:"version"`
}

// PermissionGroupVersion holds the current version of a single permission
// group.
type PermissionGroupVersion struct {
	PermissionGroup core.PermissionGroup `json:"permissionsGroup"`
	Version         int                  `json:"version"`
}

// ListLatestPermissions returns the most recent permission definitions
// available for the specified DevOps features and permission groups. When no
// permission groups are specified, RSC returns the permissions for all
// permission groups.
func ListLatestPermissions(ctx context.Context, gql *graphql.Client, features []core.Feature) (Permissions, error) {
	gql.Log().Print(log.Trace)

	query := devopsCloudAccountListLatestPermissionsQuery
	buf, err := gql.Request(ctx, query, struct {
		FeaturesWithPermissionsGroups []core.Feature `json:"featuresWithPermissionsGroups"`
	}{FeaturesWithPermissionsGroups: features})
	if err != nil {
		return Permissions{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result Permissions `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return Permissions{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// ListCurrentPermissions returns the permissions currently configured for the
// specified DevOps organization, for the given features and permission groups.
// When no permission groups are specified, RSC returns the permissions for the
// organization's current permission groups.
func ListCurrentPermissions(ctx context.Context, gql *graphql.Client, organizationID uuid.UUID, features []core.Feature) (Permissions, error) {
	gql.Log().Print(log.Trace)

	query := devopsCloudAccountListCurrentPermissionsQuery
	buf, err := gql.Request(ctx, query, struct {
		OrganizationID                uuid.UUID      `json:"organizationId"`
		FeaturesWithPermissionsGroups []core.Feature `json:"featuresWithPermissionsGroups"`
	}{OrganizationID: organizationID, FeaturesWithPermissionsGroups: features})
	if err != nil {
		return Permissions{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result Permissions `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return Permissions{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}
