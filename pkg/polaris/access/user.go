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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/access"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// User represents a user in RSC.
type User struct {
	ID             string
	Email          string
	Status         string
	IsAccountOwner bool
	Roles          []Role
}

// HasRole returns true if the user is assigned the specified role, otherwise
// false.
func (u User) HasRole(roleID uuid.UUID) bool {
	for _, role := range u.Roles {
		if role.ID == roleID {
			return true
		}
	}

	return false
}

// User returns the user with the specified email address.
func (a API) User(ctx context.Context, userEmail string) (User, error) {
	a.client.Log().Print(log.Trace)

	users, err := a.Users(ctx, userEmail)
	if err != nil {
		return User{}, fmt.Errorf("failed to get users: %w", err)
	}

	user, err := findUserByEmail(users, userEmail)
	if err != nil {
		return User{}, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// findUserByEmail returns the user with an email address exactly matching the
// specified email address.
func findUserByEmail(users []User, userEmail string) (User, error) {
	for _, user := range users {
		if user.Email == userEmail {
			return user, nil
		}
	}

	return User{}, fmt.Errorf("user with email address %q %w", userEmail, graphql.ErrNotFound)
}

// Users returns the users matching the specified email address filter.
func (a API) Users(ctx context.Context, emailFilter string) ([]User, error) {
	a.client.Log().Print(log.Trace)

	users, err := access.Wrap(a.client).UsersInCurrentAndDescendantOrganization(ctx, emailFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup users by email: %w", err)
	}

	return toUsers(users), nil
}

// AddUser adds a new user with the specified email address and roles. Note that
// a user needs at least one role assigned at all times.
func (a API) AddUser(ctx context.Context, userEmail string, roleIDs []uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	if _, err := access.Wrap(a.client).CreateUser(ctx, userEmail, roleIDs); err != nil {
		return fmt.Errorf("failed to add user: %w", err)
	}

	return nil
}

// RemoveUser removes the user with the specified email address.
func (a API) RemoveUser(ctx context.Context, userEmail string) error {
	a.client.Log().Print(log.Trace)

	user, err := a.User(ctx, userEmail)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := access.Wrap(a.client).DeleteUserFromAccount(ctx, []string{user.ID}); err != nil {
		return fmt.Errorf("failed to remove user: %w", err)
	}

	return nil
}

// AssignRole assigns the role to the user with the specified user email
// address.
func (a API) AssignRole(ctx context.Context, userEmail string, roleID uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	accessClient := access.Wrap(a.client)

	user, err := a.User(ctx, userEmail)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := accessClient.AddRoleAssignment(ctx, []uuid.UUID{roleID}, []string{user.ID}, nil); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// UnassignRole unassigns the role from the user with the specified user email
// address.
func (a API) UnassignRole(ctx context.Context, userEmail string, roleID uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	accessClient := access.Wrap(a.client)

	user, err := a.User(ctx, userEmail)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	var roleIDs []uuid.UUID
	for _, role := range user.Roles {
		if role.ID != roleID {
			roleIDs = append(roleIDs, role.ID)
		}
	}
	if err := accessClient.UpdateRoleAssignment(ctx, []string{user.ID}, nil, roleIDs); err != nil {
		return fmt.Errorf("failed to unassign role: %w", err)
	}

	return nil
}

// ReplaceRoles replaces all the roles for the specified user.
func (a API) ReplaceRoles(ctx context.Context, userEmail string, newRoleIDs []uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	accessClient := access.Wrap(a.client)

	user, err := a.User(ctx, userEmail)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := accessClient.UpdateRoleAssignment(ctx, []string{user.ID}, nil, newRoleIDs); err != nil {
		return fmt.Errorf("failed to replace roles: %w", err)
	}

	return nil
}

func toUsers(accessUsers []access.User) []User {
	users := make([]User, 0, len(accessUsers))
	for _, user := range accessUsers {
		users = append(users, toUser(user))
	}

	return users
}

func toUser(accessUser access.User) User {
	return User{
		ID:             accessUser.ID,
		Email:          accessUser.Email,
		Status:         accessUser.Status,
		IsAccountOwner: accessUser.IsAccountOwner,
		Roles:          toRoles(accessUser.Roles),
	}
}
