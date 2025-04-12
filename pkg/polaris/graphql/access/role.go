// Copyright 2023 Rubrik, Inc.
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
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Role represents a user role in RSC.
type Role struct {
	ID                  uuid.UUID    `json:"id"`
	Name                string       `json:"name"`
	Description         string       `json:"description"`
	AssignedPermissions []Permission `json:"explicitlyAssignedPermissions"`
	IsOrgAdmin          bool         `json:"isOrgAdmin"`
}

// Permission represents the ability of a user to perform an operation on one
// or more objects under a snappable hierarchy.
type Permission struct {
	Operation                string                    `json:"operation"`
	ObjectsForHierarchyTypes []ObjectsForHierarchyType `json:"objectsForHierarchyTypes"`
}

// ObjectsForHierarchyType represents a set of objects under a specific
// snappable hierarchy.
type ObjectsForHierarchyType struct {
	SnappableType string   `json:"snappableType"`
	ObjectIDs     []string `json:"objectIds"`
}

// RolesByIDs returns the roles with the specified role IDs.
func RolesByIDs(ctx context.Context, gql *graphql.Client, roleIDs []uuid.UUID) ([]Role, error) {
	gql.Log().Print(log.Trace)

	query := getRolesByIdsQuery
	buf, err := gql.Request(ctx, query, struct {
		IDs []uuid.UUID `json:"roleIds"`
	}{IDs: roleIDs})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result []Role `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// ListRoles returns all roles matching the specified role name filter.
func ListRoles(ctx context.Context, gql *graphql.Client, nameFilter string) ([]Role, error) {
	gql.Log().Print(log.Trace)

	var cursor string
	var nodes []Role
	for {
		query := getAllRolesInOrgConnectionQuery
		buf, err := gql.Request(ctx, query, struct {
			After      string `json:"after,omitempty"`
			NameFilter string `json:"nameFilter,omitempty"`
		}{After: cursor, NameFilter: nameFilter})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}
		graphql.LogResponse(gql.Log(), query, buf)

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []Role `json:"nodes"`
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

// CreateRoleParams holds the parameters for a role create operation.
type CreateRoleParams struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// CreateRole creates a new role. Returns the ID of the new role.
func CreateRole(ctx context.Context, gql *graphql.Client, params CreateRoleParams) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	query := mutateRoleQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result string `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	id, err := uuid.Parse(payload.Data.Result)
	if err != nil {
		return uuid.Nil, graphql.ResponseError(query, fmt.Errorf("failed to parse role ID: %s", err))
	}

	return id, nil
}

// UpdateRoleParams holds the parameters for a role update operation.
type UpdateRoleParams struct {
	ID uuid.UUID `json:"roleId"`
	CreateRoleParams
}

// UpdateRole updates the role with the specified ID.
func UpdateRole(ctx context.Context, gql *graphql.Client, params UpdateRoleParams) error {
	gql.Log().Print(log.Trace)

	query := mutateRoleQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result string `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if payload.Data.Result == "" {
		return graphql.ResponseError(query, fmt.Errorf("failed to update role %q", params.ID))
	}

	return nil
}

// DeleteRole deletes the role with the specified ID.
func DeleteRole(ctx context.Context, gql *graphql.Client, id uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deleteRoleQuery
	buf, err := gql.Request(ctx, deleteRoleQuery, struct {
		ID uuid.UUID `json:"roleId"`
	}{ID: id})
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result {
		return graphql.ResponseError(query, fmt.Errorf("failed to delete role %q", id))
	}

	return nil
}

// AssignRoleParams holds the parameters for a role assign operation.
type AssignRoleParams struct {
	RoleIDs  []uuid.UUID `json:"roleIds"`
	UserIDs  []string    `json:"userIds,omitempty"`
	GroupIDs []string    `json:"groupIds,omitempty"`
}

// AssignRoles assigns the roles to the users and groups with the specified IDs.
func AssignRoles(ctx context.Context, gql *graphql.Client, params AssignRoleParams) error {
	gql.Log().Print(log.Trace)

	query := addRoleAssignmentQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result {
		return graphql.ResponseError(query, errors.New("failed to assign roles"))
	}

	return nil
}

// ReplaceRoleParams holds the parameters for a role replace operation.
type ReplaceRoleParams AssignRoleParams

// ReplaceRoles replaces the role assignment for the users and groups with the
// specified IDs.
func ReplaceRoles(ctx context.Context, gql *graphql.Client, params ReplaceRoleParams) error {
	gql.Log().Print(log.Trace)

	query := updateRoleAssignmentsQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result {
		return graphql.ResponseError(query, errors.New("failed to replace roles"))
	}

	return nil
}
