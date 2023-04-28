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
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Role represents a user role in RSC.
type Role struct {
	ID                            uuid.UUID    `json:"id"`
	Name                          string       `json:"name"`
	Description                   string       `json:"description"`
	ExplicitlyAssignedPermissions []Permission `json:"explicitlyAssignedPermissions"`
	IsOrgAdmin                    bool         `json:"isOrgAdmin"`
	ProtectableClusters           []string     `json:"protectableClusters"`
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

// AllRolesInOrg returns all roles in the user's organization matching the
// specified name filter.
func (a API) AllRolesInOrg(ctx context.Context, nameFilter string) ([]Role, error) {
	a.log.Print(log.Trace)

	var roles []Role
	var cursor string
	for {
		buf, err := a.GQL.Request(ctx, getAllRolesInOrgConnectionQuery, struct {
			After      string `json:"after,omitempty"`
			NameFilter string `json:"nameFilter,omitempty"`
		}{After: cursor, NameFilter: nameFilter})
		if err != nil {
			return nil, fmt.Errorf("failed to request getAllRolesInOrgConnection: %w", err)
		}
		a.log.Printf(log.Debug, "getAllRolesInOrgConnection(%q): %s", nameFilter, string(buf))

		var payload struct {
			Data struct {
				Result struct {
					Edges []struct {
						Node Role `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal getAllRolesInOrgConnection response: %v", err)
		}
		for _, account := range payload.Data.Result.Edges {
			roles = append(roles, account.Node)
		}

		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return roles, nil
}

// RolesByIDs returns the roles with the specified ids.
func (a API) RolesByIDs(ctx context.Context, IDs []uuid.UUID) ([]Role, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, getRolesByIdsQuery, struct {
		IDs []uuid.UUID `json:"roleIds"`
	}{IDs: IDs})
	if err != nil {
		return nil, fmt.Errorf("failed to request getRolesByIds: %w", err)
	}
	a.log.Printf(log.Debug, "getRolesByIds(%v): %s", IDs, string(buf))

	var payload struct {
		Data struct {
			Result []Role `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getRolesByIds response: %v", err)
	}

	return payload.Data.Result, nil
}

// MutateRole creates or updates a role. To create a role pass in the empty
// string for id.
func (a API) MutateRole(ctx context.Context, id string, name, description string, permissions []Permission, protectableClusters []string) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if protectableClusters == nil {
		protectableClusters = []string{}
	}

	buf, err := a.GQL.Request(ctx, mutateRoleQuery, struct {
		ID                  string       `json:"roleId,omitempty"`
		Name                string       `json:"name"`
		Description         string       `json:"description"`
		Permissions         []Permission `json:"permissions"`
		ProtectableClusters []string     `json:"protectableClusters"`
	}{ID: id, Name: name, Description: description, Permissions: permissions, ProtectableClusters: protectableClusters})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to request mutateRole: %w", err)
	}
	a.log.Printf(log.Debug, "mutateRole(%q, %q, %q, %v, %v): %s", id, name, description, permissions, protectableClusters, string(buf))

	var payload struct {
		Data struct {
			Result uuid.UUID `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal mutateRole response: %v", err)
	}

	return payload.Data.Result, nil
}

// DeleteRole deletes the role with the specified id.
func (a API) DeleteRole(ctx context.Context, id uuid.UUID) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, deleteRoleQuery, struct {
		ID uuid.UUID `json:"roleId"`
	}{ID: id})
	if err != nil {
		return fmt.Errorf("failed to request deleteRole: %w", err)
	}
	a.log.Printf(log.Debug, "deleteRole(%q): %s", id, string(buf))

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal deleteRole response: %v", err)
	}
	if !payload.Data.Result {
		return fmt.Errorf("failed to delete role: %s", id)
	}

	return nil
}

// AddRoleAssignment assigns the roles to the users with the specified ids.
func (a API) AddRoleAssignment(ctx context.Context, roleIDs []uuid.UUID, userIDs, groupIDs []string) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, addRoleAssignmentQuery, struct {
		RoleIDs  []uuid.UUID `json:"roleIds"`
		UserIDs  []string    `json:"userIds"`
		GroupIDs []string    `json:"groupIds,omitempty"`
	}{RoleIDs: roleIDs, UserIDs: userIDs, GroupIDs: groupIDs})
	if err != nil {
		return fmt.Errorf("failed to request addRoleAssignment: %w", err)
	}
	a.log.Printf(log.Debug, "addRoleAssignment(%v, %v, %v): %s", roleIDs, userIDs, groupIDs, string(buf))

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal addRoleAssignment response: %v", err)
	}
	if !payload.Data.Result {
		return fmt.Errorf("failed to add assignments for roles: %v", roleIDs)
	}

	return nil
}

// UpdateRoleAssignment updates the role assignments for the users with the
// specified ids.
func (a API) UpdateRoleAssignment(ctx context.Context, userIDs, groupIDs []string, roleIDs []uuid.UUID) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, updateRoleAssignmentsQuery, struct {
		UserIDs  []string    `json:"userIds"`
		GroupIDs []string    `json:"groupIds"`
		RoleIDs  []uuid.UUID `json:"roleIds"`
	}{UserIDs: userIDs, GroupIDs: groupIDs, RoleIDs: roleIDs})
	if err != nil {
		return fmt.Errorf("failed to request updateRoleAssignments: %w", err)
	}
	a.log.Printf(log.Debug, "updateRoleAssignments(%v, %v, %v): %s", roleIDs, userIDs, groupIDs, string(buf))

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal updateRoleAssignments response: %v", err)
	}
	if !payload.Data.Result {
		return fmt.Errorf("failed to update assignments for roles: %v", roleIDs)
	}

	return nil
}

// RoleTemplate represents a named role template in RSC.
type RoleTemplate struct {
	ID                  uuid.UUID    `json:"id"`
	Name                string       `json:"name"`
	Description         string       `json:"description"`
	AssignedPermissions []Permission `json:"explicitlyAssignedPermissions"`
}

// RoleTemplates returns the role templates matching the specified name filter.
func (a API) RoleTemplates(ctx context.Context, nameFilter string) ([]RoleTemplate, error) {
	a.log.Print(log.Trace)

	var templates []RoleTemplate
	var cursor string
	for {
		buf, err := a.GQL.Request(ctx, roleTemplatesQuery, struct {
			After      string `json:"after,omitempty"`
			NameFilter string `json:"nameFilter,omitempty"`
		}{After: cursor, NameFilter: nameFilter})
		if err != nil {
			return nil, fmt.Errorf("failed to request roleTemplates: %w", err)
		}
		a.log.Printf(log.Debug, "roleTemplates(%q): %s", nameFilter, string(buf))

		var payload struct {
			Data struct {
				Result struct {
					Edges []struct {
						Node RoleTemplate `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal roleTemplates response: %v", err)
		}
		for _, edge := range payload.Data.Result.Edges {
			templates = append(templates, edge.Node)
		}

		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return templates, nil
}
