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

package aws

import (
	"context"
	"encoding/json"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AWSActionWithUseCase represents an AWS IAM action with the use cases that
// require it.
type AWSActionWithUseCase struct {
	Action   string   `json:"action"`
	UseCases []string `json:"usecase"`
}

// PermissionStatement represents an AWS IAM permission statement.
type PermissionStatement struct {
	Actions []AWSActionWithUseCase `json:"actions"`
}

// PermissionsGroupPermissions represents the permissions belonging to a
// permissions group, including the IAM action statements that grant them.
type PermissionsGroupPermissions struct {
	PermissionsGroup     core.PermissionGroup  `json:"permissionsGroup"`
	Version              int                   `json:"version"`
	PermissionStatements []PermissionStatement `json:"permissionStatements"`
}

// FeaturePermissions represents the permissions for an AWS cloud account
// feature, grouped by permissions group.
type FeaturePermissions struct {
	Feature                     string                        `json:"feature"`
	PermissionsGroupPermissions []PermissionsGroupPermissions `json:"permissionsGroupPermissions"`
}

// AllFeaturePermissions returns the latest AWS permissions grouped by
// permissions group, including the IAM action statements, for the specified
// features.
func (a API) AllFeaturePermissions(ctx context.Context, features []core.Feature) ([]FeaturePermissions, error) {
	a.log.Print(log.Trace)

	query := awsLatestPermissionsByPermissionsGroupQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Features []string `json:"features"`
	}{Features: core.FeatureNames(features)})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []FeaturePermissions `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}
