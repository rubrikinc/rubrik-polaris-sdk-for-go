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

// RoleByID returns the role with the specified role ID.
func (a API) RoleByID(ctx context.Context, roleID uuid.UUID) (access.Role, error) {
	a.log.Print(log.Trace)

	roles, err := access.RolesByIDs(ctx, a.client, []uuid.UUID{roleID})
	if err != nil {
		var gqlErr graphql.GQLError
		if errors.As(err, &gqlErr) && len(gqlErr.Errors) > 0 {
			if gqlErr.Errors[0].Extensions.Code == 404 {
				return access.Role{}, fmt.Errorf("failed to get role %q: %s: %w", roleID, gqlErr.Error(), graphql.ErrNotFound)
			}
		}
		return access.Role{}, fmt.Errorf("failed to get role %q: %s", roleID, err)
	}
	if len(roles) > 1 {
		return access.Role{}, fmt.Errorf("expected a single role, got: %d", len(roles))
	}

	return roles[0], nil
}

// RoleByName returns the role with the specified name.
func (a API) RoleByName(ctx context.Context, name string) (access.Role, error) {
	a.log.Print(log.Trace)

	roles, err := a.Roles(ctx, name)
	if err != nil {
		return access.Role{}, err
	}

	for _, role := range roles {
		if role.Name == name {
			return role, nil
		}
	}

	return access.Role{}, fmt.Errorf("role %q %w", name, graphql.ErrNotFound)
}

// Roles returns the roles matching the specified role name filter. The name
// filter matches all roles that has the specified name filter as part of their
// name.
func (a API) Roles(ctx context.Context, nameFilter string) ([]access.Role, error) {
	a.client.Log().Print(log.Trace)

	roles, err := access.ListRoles(ctx, a.client, nameFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles with filter %q: %s", nameFilter, err)
	}

	return roles, nil
}

// CreateRole creates a new role. Returns the ID of the new role.
func (a API) CreateRole(ctx context.Context, name, description string, permissions []access.Permission) (uuid.UUID, error) {
	a.client.Log().Print(log.Trace)

	id, err := access.CreateRole(ctx, a.client, access.CreateRoleParams{
		Name:        name,
		Description: description,
		Permissions: permissions,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create role %q: %s", name, err)
	}

	return id, nil
}

// UpdateRole updates the role with the specified role ID.
func (a API) UpdateRole(ctx context.Context, roleID uuid.UUID, name, description string, permissions []access.Permission) error {
	a.client.Log().Print(log.Trace)

	err := access.UpdateRole(ctx, a.client, access.UpdateRoleParams{
		ID: roleID,
		CreateRoleParams: access.CreateRoleParams{
			Name:        name,
			Description: description,
			Permissions: permissions,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update role %q: %s", roleID, err)
	}

	return nil
}

// DeleteRole deletes the role with the specified role ID.
func (a API) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	// DeleteRole doesn't return an error if the role doesn't exist, so we start
	// by checking if the role exist.
	if _, err := a.RoleByID(ctx, roleID); errors.Is(err, graphql.ErrNotFound) {
		return fmt.Errorf("failed to remove role %q: %w", roleID, graphql.ErrNotFound)
	}

	if err := access.DeleteRole(ctx, a.client, roleID); err != nil {
		return fmt.Errorf("failed to remove role %q: %s", roleID, err)
	}

	return nil
}
