package access

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
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
	err = accessClient.AddUser(ctx, testConfig.UserEmail, []uuid.UUID{adminRole.ID})
	if err != nil {
		t.Fatal(err)
	}
	assertUserHasRoles(t, testConfig.UserEmail, adminRole.ID)

	// Get user by email address.
	user, err := accessClient.User(ctx, testConfig.UserEmail)
	if err != nil {
		t.Error(err)
	}
	if user.ID == "" {
		t.Errorf("invalid user id: %v", user.ID)
	}
	if user.Email != testConfig.UserEmail {
		t.Errorf("invalid user email: %v", user.Email)
	}
	if user.Status != "ACTIVE" {
		t.Errorf("invalid user status: %v", user.Status)
	}
	if user.IsAccountOwner {
		t.Errorf("invalid user account owner: true")
	}

	// Verify that RoleByName returns ErrNotFound for non-exact matches.
	if _, err := accessClient.User(ctx, "name@example.com"); err == nil || !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("expected graphql.ErrNotFound: %v", err)
	}
	email := testConfig.UserEmail[:len(testConfig.UserEmail)-1]
	if _, err := accessClient.User(ctx, email); err == nil || !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("expected graphql.ErrNotFound: %v", err)
	}

	// List users.
	users, err := accessClient.Users(ctx, testConfig.UserEmail)
	if err != nil {
		t.Error(err)
	}
	if n := len(users); n != 1 {
		t.Errorf("invalid number of users: %v", n)
	}
	assertUserHasRoles(t, testConfig.UserEmail, adminRole.ID)

	// Add new role.
	roleID, err := accessClient.AddRole(ctx, "Integration Test Role", "Test Role Description", []Permission{{
		Operation: "VIEW_CLUSTER",
		Hierarchies: []SnappableHierarchy{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"CLUSTER_ROOT"},
		}},
	}}, NoProtectableClusters)
	if err != nil {
		t.Error(err)
	}

	// Assign role to user.
	if err := accessClient.AssignRole(ctx, testConfig.UserEmail, roleID); err != nil {
		t.Error(err)
	}
	assertUserHasRoles(t, testConfig.UserEmail, adminRole.ID, roleID)

	// Unassign role from user.
	if err := accessClient.UnassignRole(ctx, testConfig.UserEmail, roleID); err != nil {
		t.Error(err)
	}
	assertUserHasRoles(t, testConfig.UserEmail, adminRole.ID)

	// Replace roles for user.
	if err := accessClient.ReplaceRoles(ctx, testConfig.UserEmail, []uuid.UUID{adminRole.ID, roleID}); err != nil {
		t.Error(err)
	}
	assertUserHasRoles(t, testConfig.UserEmail, adminRole.ID, roleID)

	// Unassign role from user.
	if err := accessClient.UnassignRole(ctx, testConfig.UserEmail, roleID); err != nil {
		t.Error(err)
	}
	assertUserHasRoles(t, testConfig.UserEmail, adminRole.ID)

	// Remove role.
	if err := accessClient.RemoveRole(ctx, roleID); err != nil {
		t.Error(err)
	}

	// Remove user.
	if err := accessClient.RemoveUser(ctx, testConfig.UserEmail); err != nil {
		t.Fatal(err)
	}

	// Check that the user has been removed.
	if _, err = accessClient.User(ctx, testConfig.UserEmail); err == nil || !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal("user should have been removed")
	}
}

func assertUserHasRoles(t *testing.T, userEmail string, roles ...uuid.UUID) {
	user, err := Wrap(client).User(context.Background(), userEmail)
	if err != nil {
		t.Errorf("failed to get user: %v", err)
	}

	if n := len(user.Roles); n != len(roles) {
		t.Errorf("invalid number of roles: %v", n)
	}

	for _, role := range roles {
		if !user.HasRole(role) {
			t.Errorf("user should be assigned role: %v", role)
		}
	}
}
