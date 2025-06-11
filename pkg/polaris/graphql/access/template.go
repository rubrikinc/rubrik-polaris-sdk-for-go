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

package access

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// RoleTemplate represents a named role template in RSC. A number of role
// templates comes bundled with RSC. They can be used to create roles with a
// predefined set of permissions.
type RoleTemplate struct {
	ID                  uuid.UUID    `json:"id"`
	Name                string       `json:"name"`
	Description         string       `json:"description"`
	AssignedPermissions []Permission `json:"explicitlyAssignedPermissions"`
}

// ListRoleTemplates returns the role templates matching the specified name
// filter.
func ListRoleTemplates(ctx context.Context, gql *graphql.Client, nameFilter string) ([]RoleTemplate, error) {
	gql.Log().Print(log.Trace)

	var nodes []RoleTemplate
	var cursor string
	for {
		query := roleTemplatesQuery
		buf, err := gql.Request(ctx, query, struct {
			After      string `json:"after,omitempty"`
			NameFilter string `json:"nameFilter,omitempty"`
		}{After: cursor, NameFilter: nameFilter})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []RoleTemplate `json:"nodes"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, graphql.UnmarshalError(query, err)
		}
		nodes = append(nodes, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return nodes, nil
}
