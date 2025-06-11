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
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlaccess "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/access"
)

func TestUserManagement(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testConfig, err := testsetup.RSCConfig()
	if err != nil {
		t.Fatal(err)
	}

	accessClient := Wrap(client)

	// Look up the administrator role.
	adminRole, err := accessClient.RoleByName(ctx, "Administrator")
	if err != nil {
		t.Fatal(err)
	}

	// Add user with administrator role.
	userID, err := accessClient.CreateUser(ctx, testConfig.NewUserEmail, []uuid.UUID{adminRole.ID})
	if err != nil {
		t.Fatal(err)
	}
	assertUserHasRoles(t, userID, adminRole.ID)

	// Get user by ID.
	user1, err := accessClient.UserByID(ctx, userID)
	if err != nil {
		t.Error(err)
	}
	if user1.ID != userID {
		t.Errorf("invalid user id: %s", user1.ID)
	}
	if user1.Email != testConfig.NewUserEmail {
		t.Errorf("invalid user email: %s", user1.Email)
	}

	// Get user by email address.
	user2, err := accessClient.UserByEmail(ctx, testConfig.NewUserEmail, gqlaccess.DomainLocal)
	if err != nil {
		t.Error(err)
	}
	if user2.ID != userID {
		t.Errorf("invalid user id: %s", user2.ID)
	}
	if user2.Email != testConfig.NewUserEmail {
		t.Errorf("invalid user email: %s", user2.Email)
	}
	if user2.Status != "ACTIVE" {
		t.Errorf("invalid user status: %s", user2.Status)
	}
	if user2.IsAccountOwner {
		t.Errorf("invalid user account owner: true")
	}

	// Verify that UserByID returns ErrNotFound for non-existing UUIDs.
	if _, err := accessClient.UserByEmail(ctx, "c4c53ec0-aa02-4582-9443-bc4be0045653", gqlaccess.DomainLocal); err == nil || !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("expected graphql.ErrNotFound: %s", err)
	}

	// Verify that UserByEmail returns ErrNotFound for non-exact matches.
	if _, err := accessClient.UserByEmail(ctx, "name@example.com", gqlaccess.DomainLocal); err == nil || !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("expected graphql.ErrNotFound: %s", err)
	}
	email := testConfig.NewUserEmail[:len(testConfig.NewUserEmail)-1]
	if _, err := accessClient.UserByEmail(ctx, email, gqlaccess.DomainLocal); err == nil || !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("expected graphql.ErrNotFound: %s", err)
	}

	// List users.
	users, err := accessClient.Users(ctx, testConfig.NewUserEmail)
	if err != nil {
		t.Error(err)
	}
	if n := len(users); n != 1 {
		t.Errorf("invalid number of users: %d", n)
	}
	assertUserHasRoles(t, userID, adminRole.ID)

	// Add new role.
	roleID, err := accessClient.CreateRole(ctx, "Integration Test Role", "Test Role Description", []gqlaccess.Permission{{
		Operation: "VIEW_CLUSTER",
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"CLUSTER_ROOT"},
		}},
	}})
	if err != nil {
		t.Error(err)
	}

	// Assign role to user.
	if err := accessClient.AssignUserRole(ctx, userID, roleID); err != nil {
		t.Error(err)
	}
	assertUserHasRoles(t, userID, adminRole.ID, roleID)

	// Unassign role from user.
	if err := accessClient.UnassignUserRole(ctx, userID, roleID); err != nil {
		t.Error(err)
	}
	assertUserHasRoles(t, userID, adminRole.ID)

	// Replace roles for user.
	if err := accessClient.ReplaceUserRoles(ctx, userID, []uuid.UUID{adminRole.ID, roleID}); err != nil {
		t.Error(err)
	}
	assertUserHasRoles(t, userID, adminRole.ID, roleID)

	// Unassign role from user.
	if err := accessClient.UnassignUserRole(ctx, userID, roleID); err != nil {
		t.Error(err)
	}
	assertUserHasRoles(t, userID, adminRole.ID)

	// Delete role and user.
	if err := accessClient.DeleteRole(ctx, roleID); err != nil {
		t.Error(err)
	}
	if err := accessClient.DeleteUser(ctx, userID); err != nil {
		t.Fatal(err)
	}

	// Check that the user has been deleted.
	if _, err = accessClient.UserByID(ctx, userID); err == nil || !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal("user should have been deleted")
	}
}

func assertUserHasRoles(t *testing.T, userID string, roles ...uuid.UUID) {
	user, err := Wrap(client).UserByID(context.Background(), userID)
	if err != nil {
		t.Errorf("failed to get user: %s", err)
	}
	if n := len(user.Roles); n != len(roles) {
		t.Errorf("invalid number of roles: %d", n)
	}
	for _, role := range roles {
		if !user.HasRole(role) {
			t.Errorf("user should be assigned role: %s", role)
		}
	}
}
