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
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/access"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// UserByID returns the user with the specified user ID.
func (a API) UserByID(ctx context.Context, userID string) (access.User, error) {
	a.client.Log().Print(log.Trace)

	users, err := a.Users(ctx, "")
	if err != nil {
		return access.User{}, err
	}

	for _, user := range users {
		if user.ID == userID {
			return user, nil
		}
	}

	return access.User{}, fmt.Errorf("user %q %w", userID, graphql.ErrNotFound)
}

// UserByEmail returns the user with the specified email address.
func (a API) UserByEmail(ctx context.Context, userEmail string) (access.User, error) {
	a.client.Log().Print(log.Trace)

	users, err := a.Users(ctx, userEmail)
	if err != nil {
		return access.User{}, err
	}

	for _, user := range users {
		if user.Email == userEmail {
			return user, nil
		}
	}

	return access.User{}, fmt.Errorf("user %q %w", userEmail, graphql.ErrNotFound)
}

// Users returns the users matching the specified email address filter.
func (a API) Users(ctx context.Context, emailFilter string) ([]access.User, error) {
	a.client.Log().Print(log.Trace)

	users, err := access.ListUsers(ctx, a.client, access.UserFilter{
		EmailFilter: emailFilter,
		UserDomains: []access.UserDomain{access.DomainLocal, access.DomainSSO},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get users with filter %q: %s", emailFilter, err)
	}

	return users, nil
}

// CreateUser creates a new user with the specified email address and roles.
// Returns the ID of the new role. Note, a user needs at least one role assigned
// at all times.
func (a API) CreateUser(ctx context.Context, userEmail string, roleIDs []uuid.UUID) (string, error) {
	a.client.Log().Print(log.Trace)

	userID, err := access.CreateUser(ctx, a.client, access.CreateUserParams{
		Email:   userEmail,
		RoleIDs: roleIDs,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create user %q: %s", userEmail, err)
	}

	return userID, nil
}

// DeleteUser deletes the user with the specified user ID.
func (a API) DeleteUser(ctx context.Context, userID string) error {
	a.client.Log().Print(log.Trace)

	if err := access.DeleteUser(ctx, a.client, userID); err != nil {
		return fmt.Errorf("failed to delete user %q: %s", userID, err)
	}

	return nil
}

// AssignUserRole assigns the role to the user with the specified user ID.
func (a API) AssignUserRole(ctx context.Context, userID string, roleID uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	if err := access.AssignRoles(ctx, a.client, access.AssignRoleParams{
		RoleIDs: []uuid.UUID{roleID},
		UserIDs: []string{userID},
	}); err != nil {
		return fmt.Errorf("failed to assign role %q to user %q: %s", roleID, userID, err)
	}

	return nil
}

// AssignUserRoles assigns the roles to the user with the specified user ID.
func (a API) AssignUserRoles(ctx context.Context, userID string, roleIDs []uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	if err := access.AssignRoles(ctx, a.client, access.AssignRoleParams{
		RoleIDs: roleIDs,
		UserIDs: []string{userID},
	}); err != nil {
		return fmt.Errorf("failed to assign roles %s to user %q: %s", joinUUIDs(roleIDs), userID, err)
	}

	return nil
}

// UnassignUserRole unassigns the role from the user with the specified user ID.
// Returns graphql.ErrNotFound if the user does not exist.
func (a API) UnassignUserRole(ctx context.Context, userID string, roleID uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	user, err := a.UserByID(ctx, userID)
	if err != nil {
		return err
	}

	var roleIDs []uuid.UUID
	for _, role := range user.Roles {
		if role.ID != roleID {
			roleIDs = append(roleIDs, role.ID)
		}
	}
	if err := access.ReplaceRoles(ctx, a.client, access.ReplaceRoleParams{
		RoleIDs: roleIDs,
		UserIDs: []string{userID},
	}); err != nil {
		return fmt.Errorf("failed to unassign role %q from user %q: %s", roleID, userID, err)
	}

	return nil
}

// UnassignUserRoles unassigns the roles from the user with the specified user
// ID. Returns graphql.ErrNotFound if the user does not exist.
func (a API) UnassignUserRoles(ctx context.Context, userID string, roleIDs []uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	user, err := a.UserByID(ctx, userID)
	if err != nil {
		return err
	}

	var keepRoleIDs []uuid.UUID
	for _, role := range user.Roles {
		if !slices.Contains(roleIDs, role.ID) {
			keepRoleIDs = append(keepRoleIDs, role.ID)
		}
	}
	if err := access.ReplaceRoles(ctx, a.client, access.ReplaceRoleParams{
		RoleIDs: keepRoleIDs,
		UserIDs: []string{userID},
	}); err != nil {
		return fmt.Errorf("failed to unassign role %q from user %q: %s", joinUUIDs(roleIDs), userID, err)
	}

	return nil
}

// ReplaceUserRoles replaces all the roles for the user with the specified user
// ID.
func (a API) ReplaceUserRoles(ctx context.Context, userID string, newRoleIDs []uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	if err := access.ReplaceRoles(ctx, a.client, access.ReplaceRoleParams{
		RoleIDs: newRoleIDs,
		UserIDs: []string{userID},
	}); err != nil {
		return fmt.Errorf("failed to replace roles for user %q: %s", userID, err)
	}

	return nil
}
