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

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// PermissionGroupInfo holds information about a GCP permission group.
type PermissionGroupInfo struct {
	PermissionGroup core.PermissionGroup `json:"permissionGroupType"`
	Version         int                  `json:"policyVersion"`
}

// FeaturePermissionGroups holds the permission groups for a GCP feature.
type FeaturePermissionGroups struct {
	Feature          string                `json:"feature"`
	PermissionGroups []PermissionGroupInfo `json:"permissionGroups"`
}

// AllPermissionsGroupsByFeature returns the permission groups for the specified
// GCP features.
func (a API) AllPermissionsGroupsByFeature(ctx context.Context, features []core.Feature) ([]FeaturePermissionGroups, error) {
	a.log.Print(log.Trace)

	query := allLatestPermissionsByPermissionsGroupGcpQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Features []string `json:"features"`
	}{Features: core.FeatureNames(features)})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []FeaturePermissionGroups `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}
