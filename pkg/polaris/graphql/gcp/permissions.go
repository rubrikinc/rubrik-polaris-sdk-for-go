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

package gcp

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Deprecated: use AllLatestFeaturePermissionsForCloudAccounts instead.
func (a API) FeaturePermissionsForCloudAccount(ctx context.Context, feature core.Feature) (permissions []string, err error) {
	a.log.Print(log.Trace)

	query := allFeaturePermissionsForGcpCloudAccountQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Feature string `json:"feature"`
	}{Feature: feature.Name})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Permissions []struct {
				Permission string `json:"permission"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	permissions = make([]string, 0, len(payload.Data.Permissions))
	for _, p := range payload.Data.Permissions {
		permissions = append(permissions, p.Permission)
	}

	return permissions, nil
}

// CloudAccountPermissions holds the features, permission groups and GCP
// permissions for an RSC cloud account.
type CloudAccountPermissions struct {
	CloudAccountID     uuid.UUID            `json:"cloudAccountId"`
	FeaturePermissions []FeaturePermissions `json:"featurePermissions"`
}

// AllLatestFeaturePermissionsForCloudAccounts list the features, permission
// groups and GCP permissions for the specified RSC cloud accounts.
func (a API) AllLatestFeaturePermissionsForCloudAccounts(ctx context.Context, cloudAccountIDs []uuid.UUID) ([]CloudAccountPermissions, error) {
	a.log.Print(log.Trace)

	// When no cloud account is specified, the query returns a list with a
	// single element having an empty string as the ID. Skip it.
	if len(cloudAccountIDs) == 0 {
		return nil, nil
	}

	// The query used is cloud-agnostic, but the JSON document returned by the
	// query is cloud specific.
	query := allLatestFeaturePermissionsForCloudAccountsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountIDs []uuid.UUID `json:"cloudAccountIds"`
	}{CloudAccountIDs: cloudAccountIDs})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []CloudAccountPermissions `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// FeaturePermissions holds the GCP permissions needed for a set of RSC
// features.
type FeaturePermissions struct {
	Feature                 string `json:"feature"`
	PermissionJson          string `json:"permissionJson"`
	PermissionGroupVersions []struct {
		PermissionGroup string `json:"permissionsGroup"`
		Version         int    `json:"version"`
	} `json:"permissionsGroupVersions"`
	Version int `json:"version"`
}

// AllLatestFeaturePermissions list the GCP permissions needed by RSC for the
// specified features.
func (a API) AllLatestFeaturePermissions(ctx context.Context, features []core.Feature) ([]FeaturePermissions, error) {
	a.log.Print(log.Trace)

	// The query used is cloud-agnostic, but the JSON document returned by the
	// query is cloud specific. Providing both features and featuresWithPG makes
	// the query work both with and without permission groups.
	query := allLatestFeaturePermissionsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Features       []string       `json:"features"`
		FeaturesWithPG []core.Feature `json:"featuresWithPG"`
	}{Features: core.FeatureNames(features), FeaturesWithPG: features})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []struct {
				FeaturePermissions []FeaturePermissions `json:"featurePermissions"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}
	if len(payload.Data.Result) != 1 {
		return nil, graphql.ResponseError(query, errors.New("expected a single result"))
	}

	return payload.Data.Result[0].FeaturePermissions, nil
}
