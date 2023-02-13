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
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/access"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NoProtectableClusters indicates that no protectable cluster is specified.
var NoProtectableClusters []string

// Role represents a user role in RSC.
type Role struct {
	ID                  uuid.UUID
	Name                string
	Description         string
	AssignedPermissions []Permission
	IsOrgAdmin          bool
	ProtectableClusters []string
}

// Permission represents the ability of a user to perform an operation on one
// or more objects under a snappable hierarchy.
type Permission struct {
	Operation   string
	Hierarchies []SnappableHierarchy
}

// SnappableHierarchy represents a set of objects under a specific snappable
// hierarchy.
type SnappableHierarchy struct {
	SnappableType string
	ObjectIDs     []string
}

// Role returns the role with the specified id.
func (a API) Role(ctx context.Context, id uuid.UUID) (Role, error) {
	a.client.Log.Print(log.Trace)

	roles, err := access.Wrap(a.client.GQL).RolesByIDs(ctx, []uuid.UUID{id})
	if err != nil {
		var gqlErr graphql.GQLError
		if errors.As(err, &gqlErr) {
			if gqlErr.Errors[0].Extensions.Code == 404 {
				err = fmt.Errorf("%s: %w", gqlErr.Error(), graphql.ErrNotFound)
			}
		}
		return Role{}, fmt.Errorf("failed to get role: %w", err)
	}
	if len(roles) > 1 {
		return Role{}, errors.New("multiple roles returned for one id")
	}

	return toRole(roles[0]), nil
}

// Roles returns the roles matching the specified role name filter. The name
// filter matches all roles that has the specified name as a prefix of their
// name.
func (a API) Roles(ctx context.Context, nameFilter string) ([]Role, error) {
	a.client.Log.Print(log.Trace)

	roles, err := access.Wrap(a.client.GQL).AllRolesInOrg(ctx, nameFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	return toRoles(roles), nil
}

// AddRole adds the specified role to RSC returning the id of the new role. Use
// the NoProtectableCluster value to indicate that no protectable clusters are
// specified.
func (a API) AddRole(ctx context.Context, name, description string, permissions []Permission, protectableClusters []string) (uuid.UUID, error) {
	a.client.Log.Print(log.Trace)

	id, err := access.Wrap(a.client.GQL).MutateRole(ctx, "", name, description, fromPermissions(permissions), protectableClusters)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add role: %w", err)
	}

	return id, nil
}

// UpdateRole updates the role with the specified id. Use the
// NoProtectableCluster value to indicate that no protectable clusters are
// specified.
func (a API) UpdateRole(ctx context.Context, id uuid.UUID, name, description string, permissions []Permission, protectableClusters []string) error {
	a.client.Log.Print(log.Trace)

	_, err := access.Wrap(a.client.GQL).MutateRole(ctx, id.String(), name, description, fromPermissions(permissions), protectableClusters)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

// RemoveRole removes the role with the specified id.
func (a API) RemoveRole(ctx context.Context, id uuid.UUID) error {
	a.client.Log.Print(log.Trace)

	// DeleteRole doesn't return an error if the role doesn't exist, so we start
	// by checking if the role exist.
	if _, err := a.Role(ctx, id); errors.Is(err, graphql.ErrNotFound) {
		return fmt.Errorf("failed to remove role %v: %w", id, graphql.ErrNotFound)
	}

	err := access.Wrap(a.client.GQL).DeleteRole(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

// RoleTemplate represents a named role template in RSC. A number of role
// templates comes bundled with RSC. They can be used to create roles with a
// predefined set of permissions.
type RoleTemplate struct {
	ID                  uuid.UUID
	Name                string
	Description         string
	AssignedPermissions []Permission
}

// RoleTemplates returns the role templates matching the specified role template
// name filter. The name filter matches all role templates that has the
// specified name as a prefix of their name.
func (a API) RoleTemplates(ctx context.Context, nameFilter string) ([]RoleTemplate, error) {
	a.client.Log.Print(log.Trace)

	roleTemplates, err := access.Wrap(a.client.GQL).RoleTemplates(ctx, nameFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get role templates: %w", err)
	}

	return toRoleTemplates(roleTemplates), nil
}

func toRoles(accessRole []access.Role) []Role {
	roles := make([]Role, 0, len(accessRole))
	for _, role := range accessRole {
		roles = append(roles, toRole(role))
	}

	return roles
}

func toRole(accessRole access.Role) Role {
	return Role{
		ID:                  accessRole.ID,
		Name:                accessRole.Name,
		Description:         accessRole.Description,
		AssignedPermissions: toPermissions(accessRole.ExplicitlyAssignedPermissions),
		IsOrgAdmin:          accessRole.IsOrgAdmin,
	}
}

func fromPermissions(permissions []Permission) []access.Permission {
	accessPermissions := make([]access.Permission, 0, len(permissions))
	for _, permission := range permissions {
		accessPermissions = append(accessPermissions, access.Permission{
			Operation:                permission.Operation,
			ObjectsForHierarchyTypes: fromHierarchies(permission.Hierarchies),
		})
	}

	return accessPermissions
}

func toPermissions(accessPermissions []access.Permission) []Permission {
	permissions := make([]Permission, 0, len(accessPermissions))
	for _, permission := range accessPermissions {
		permissions = append(permissions, Permission{
			Operation:   permission.Operation,
			Hierarchies: toHierarchies(permission.ObjectsForHierarchyTypes),
		})
	}

	return permissions
}

func fromHierarchies(hierarchies []SnappableHierarchy) []access.ObjectsForHierarchyType {
	accessHierarchies := make([]access.ObjectsForHierarchyType, 0, len(hierarchies))
	for _, hierarchy := range hierarchies {
		accessHierarchies = append(accessHierarchies, access.ObjectsForHierarchyType(hierarchy))
	}

	return accessHierarchies
}

func toHierarchies(accessHierarchies []access.ObjectsForHierarchyType) []SnappableHierarchy {
	hierarchies := make([]SnappableHierarchy, 0, len(accessHierarchies))
	for _, hierarchy := range accessHierarchies {
		hierarchies = append(hierarchies, SnappableHierarchy(hierarchy))
	}

	return hierarchies
}

func toRoleTemplates(accessTemplates []access.RoleTemplate) []RoleTemplate {
	templates := make([]RoleTemplate, 0, len(accessTemplates))
	for _, template := range accessTemplates {
		templates = append(templates, RoleTemplate{
			ID:                  template.ID,
			Name:                template.Name,
			Description:         template.Description,
			AssignedPermissions: toPermissions(template.AssignedPermissions),
		})
	}

	return templates
}
