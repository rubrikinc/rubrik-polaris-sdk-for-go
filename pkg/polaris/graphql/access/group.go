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

// SSOGroup represents an SSO group in RSC.
type SSOGroup struct {
	ID         string    `json:"groupId"`
	Name       string    `json:"groupName"`
	DomainName string    `json:"domainName"`
	Roles      []RoleRef `json:"roles"`
	Users      []UserRef `json:"users"`
}

// HasRole returns true if the SSO group has the specified role, false
// otherwise.
func (group SSOGroup) HasRole(roleID uuid.UUID) bool {
	for _, role := range group.Roles {
		if role.ID == roleID {
			return true
		}
	}

	return false
}

// SSOGroupFilter holds the filter parameters for an SSO group list operation.
type SSOGroupFilter struct {
	AuthDomainIDs []string `json:"authDomainIdsFilter,omitempty"`
	Name          string   `json:"nameFilter,omitempty"`
	OrgIDs        []string `json:"orgIdsFilter,omitempty"`
}

// ListSSOGroups returns all SSO groups matching the specified SSO group filter.
func ListSSOGroups(ctx context.Context, gql *graphql.Client, filter SSOGroupFilter) ([]SSOGroup, error) {
	gql.Log().Print(log.Trace)

	var cursor string
	var groups []SSOGroup
	for {
		query := groupsInCurrentAndDescendantOrganizationQuery
		buf, err := gql.Request(ctx, query, struct {
			After  string         `json:"after,omitempty"`
			Filter SSOGroupFilter `json:"filter,omitempty"`
		}{After: cursor, Filter: filter})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []SSOGroup `json:"nodes"`
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
		groups = append(groups, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return groups, nil
}
