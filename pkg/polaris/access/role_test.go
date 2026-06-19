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
	"cmp"
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlaccess "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/access"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
)

func TestRoleManagement(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	accessClient := Wrap(client)

	// Create role.
	roleID, err := accessClient.CreateRole(ctx, "Integration Test Role", "Test Role Description", []gqlaccess.Permission{{
		Operation: string(gqlaccess.OperationViewCluster),
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{hierarchy.ClusterRoot},
		}},
	}, {
		// RSC grants VIEW_CLUSTER_REFERENCE automatically with VIEW_CLUSTER, so
		// grant it explicitly to keep the role free of drift.
		Operation: string(gqlaccess.OperationViewClusterReference),
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{hierarchy.ClusterRoot},
		}},
	}})
	if err != nil {
		t.Fatal(err)
	}

	// Get role by id.
	role, err := accessClient.RoleByID(ctx, roleID)
	if err != nil {
		t.Error(err)
	}
	if role.ID != roleID {
		t.Errorf("invalid role id: %s", role.ID)
	}
	if role.Name != "Integration Test Role" {
		t.Errorf("invalid role name: %s", role.Name)
	}
	if role.Description != "Test Role Description" {
		t.Errorf("invalid role description: %s", role.Description)
	}
	if role.IsOrgAdmin {
		t.Error("is org admin is true")
	}
	if !reflect.DeepEqual(sortPermissions(role.AssignedPermissions), []gqlaccess.Permission{{
		Operation: string(gqlaccess.OperationViewCluster),
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{hierarchy.ClusterRoot},
		}},
	}, {
		Operation: string(gqlaccess.OperationViewClusterReference),
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{hierarchy.ClusterRoot},
		}},
	}}) {
		t.Errorf("invalid role permissions: %s", role.AssignedPermissions)
	}

	// Verify that Role returns ErrNotFound for non-existent UUIDs.
	_, err = accessClient.RoleByID(ctx, uuid.MustParse("c4c53ec0-aa02-4582-9443-bc4be0045653"))
	if err == nil || !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("expected graphql.ErrNotFound: %s", err)
	}

	// Get role by name.
	role, err = accessClient.RoleByName(ctx, "Integration Test Role")
	if err != nil {
		t.Error(err)
	}
	if role.ID != roleID {
		t.Errorf("invalid role id: %s", role.ID)
	}
	if role.Name != "Integration Test Role" {
		t.Errorf("invalid role name: %s", role.Name)
	}
	if role.Description != "Test Role Description" {
		t.Errorf("invalid role description: %s", role.Description)
	}
	if role.IsOrgAdmin {
		t.Error("is org admin is true")
	}
	if !reflect.DeepEqual(sortPermissions(role.AssignedPermissions), []gqlaccess.Permission{{
		Operation: string(gqlaccess.OperationViewCluster),
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{hierarchy.ClusterRoot},
		}},
	}, {
		Operation: string(gqlaccess.OperationViewClusterReference),
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{hierarchy.ClusterRoot},
		}},
	}}) {
		t.Errorf("invalid role permissions: %s", role.AssignedPermissions)
	}

	// Verify that RoleByName returns ErrNotFound for non-exact matches.
	_, err = accessClient.RoleByName(ctx, "Integration Test")
	if err == nil || !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("expected graphql.ErrNotFound: %s", err)
	}

	// Update role.
	err = accessClient.UpdateRole(ctx, roleID, "Integration Test Role Updated", "Test Role Description Updated", []gqlaccess.Permission{{
		Operation: string(gqlaccess.OperationViewCluster),
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{hierarchy.ClusterRoot},
		}},
	}, {
		Operation: string(gqlaccess.OperationViewClusterReference),
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{hierarchy.ClusterRoot},
		}},
	}, {
		Operation: "REMOVE_CLUSTER",
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{hierarchy.ClusterRoot},
		}},
	}})
	if err != nil {
		t.Error(err)
	}

	// Get role by name filter.
	roles, err := accessClient.Roles(ctx, "Integration Test")
	if err != nil {
		t.Error(err)
	}
	if n := len(roles); n != 1 {
		t.Errorf("invalid number of roles: %d", n)
	}
	if roles[0].ID != roleID {
		t.Errorf("invalid role id: %s", roles[0].ID)
	}
	if roles[0].Name != "Integration Test Role Updated" {
		t.Errorf("invalid role name: %s", roles[0].Name)
	}
	if roles[0].Description != "Test Role Description Updated" {
		t.Errorf("invalid role description: %s", roles[0].Description)
	}
	if roles[0].IsOrgAdmin {
		t.Error("is org admin is true")
	}

	// Sort permissions in ascending order before asserting.
	if !reflect.DeepEqual(sortPermissions(roles[0].AssignedPermissions), []gqlaccess.Permission{{
		Operation: "REMOVE_CLUSTER",
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{hierarchy.ClusterRoot},
		}},
	}, {
		Operation: string(gqlaccess.OperationViewCluster),
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{hierarchy.ClusterRoot},
		}},
	}, {
		Operation: string(gqlaccess.OperationViewClusterReference),
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{hierarchy.ClusterRoot},
		}},
	}}) {
		t.Errorf("invalid role permissions: %#v", roles[0].AssignedPermissions)
	}

	// Remove role.
	if err := accessClient.DeleteRole(ctx, roleID); err != nil {
		t.Fatal(err)
	}

	// Check that the role has been removed.
	if _, err = accessClient.RoleByID(ctx, roleID); err == nil || !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal("role should have been removed")
	}
}

// sortPermissions sorts permissions by operation in ascending order.
func sortPermissions(permissions []gqlaccess.Permission) []gqlaccess.Permission {
	slices.SortFunc(permissions, func(a, b gqlaccess.Permission) int {
		return cmp.Compare(a.Operation, b.Operation)
	})

	return permissions
}
