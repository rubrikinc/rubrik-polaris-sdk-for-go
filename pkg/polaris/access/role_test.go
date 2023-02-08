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
	"os"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the common Polaris client used for tests. By reusing the same
// client we reduce the risk of hitting rate limits when access tokens are
// created.
var client *polaris.Client

func TestMain(m *testing.M) {
	if testsetup.BoolEnvSet("TEST_INTEGRATION") {
		// Load configuration and create client. Usually resolved using the
		// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
		polAccount, err := polaris.DefaultServiceAccount(true)
		if err != nil {
			fmt.Printf("failed to get default service account: %v\n", err)
			os.Exit(1)
		}

		// The integration tests defaults the log level to INFO. Note that
		// RUBRIK_POLARIS_LOGLEVEL can be used to override this.
		logger := polaris_log.NewStandardLogger()
		logger.SetLogLevel(polaris_log.Info)
		client, err = polaris.NewClient(context.Background(), polAccount, logger)
		if err != nil {
			fmt.Printf("failed to create polaris client: %v\n", err)
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}

func TestRoleManagement(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	accessClient := Wrap(client)

	// Add role.
	id, err := accessClient.AddRole(ctx, "Integration Test Role", "Test Role Description", []Permission{{
		Operation: "VIEW_CLUSTER",
		Hierarchies: []SnappableHierarchy{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"CLUSTER_ROOT"},
		}},
	}}, NoProtectableClusters)
	if err != nil {
		t.Fatal(err)
	}
	if id == uuid.Nil {
		t.Fatal("id is nil")
	}

	// Get role by role id.
	role, err := accessClient.Role(ctx, id)
	if err != nil {
		t.Error(err)
	}
	if role.ID != id {
		t.Errorf("invalid role id: %v", role.ID)
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
	if !reflect.DeepEqual(role.AssignedPermissions, []Permission{{
		Operation: "VIEW_CLUSTER",
		Hierarchies: []SnappableHierarchy{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"CLUSTER_ROOT"},
		}},
	}}) {
		t.Errorf("invalid role permissions: %v", role.AssignedPermissions)
	}

	// Update role.
	err = accessClient.UpdateRole(ctx, id, "Integration Test Role Updated", "Test Role Description Updated", []Permission{{
		Operation: "VIEW_CLUSTER",
		Hierarchies: []SnappableHierarchy{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"CLUSTER_ROOT"},
		}},
	}, {
		Operation: "REMOVE_CLUSTER",
		Hierarchies: []SnappableHierarchy{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"CLUSTER_ROOT"},
		}},
	}}, NoProtectableClusters)
	if err != nil {
		t.Error(err)
	}

	// Get role by filter.
	roles, err := accessClient.Roles(ctx, "Integration Test")
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) != 1 {
		t.Errorf("")
	}
	if roles[0].ID != id {
		t.Errorf("invalid role id: %v", roles[0].ID)
	}
	if roles[0].Name != "Integration Test Role Updated" {
		t.Errorf("invalid role name: %v", roles[0].Name)
	}
	if roles[0].Description != "Test Role Description Updated" {
		t.Errorf("invalid role description: %v", roles[0].Description)
	}
	if roles[0].IsOrgAdmin {
		t.Error("is org admin is true")
	}
	if !reflect.DeepEqual(roles[0].AssignedPermissions, []Permission{{
		Operation: "REMOVE_CLUSTER",
		Hierarchies: []SnappableHierarchy{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"CLUSTER_ROOT"},
		}},
	}, {
		Operation: "VIEW_CLUSTER",
		Hierarchies: []SnappableHierarchy{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"CLUSTER_ROOT"},
		}},
	}}) {
		t.Errorf("invalid role permissions: %#v", roles[0].AssignedPermissions)
	}

	// Remove role.
	if err := accessClient.RemoveRole(ctx, id); err != nil {
		t.Fatal(err)
	}

	// List role templates.
	roleTemplates, err := accessClient.RoleTemplates(ctx, "Office")
	if err != nil {
		t.Fatal(err)
	}
	if n := len(roleTemplates); n != 1 {
		t.Errorf("invalid number of role templates: %d", n)
	}
	if roleTemplates[0].ID != uuid.MustParse("00000000-0000-0000-0000-000000000006") {
		t.Errorf("invalid role template id: %v", roleTemplates[0].ID)
	}
	if roleTemplates[0].Name != "Compliance Officer" {
		t.Errorf("invalid role template name: %v", roleTemplates[0].Name)
	}
	if roleTemplates[0].Description != "Template for compliance officer" {
		t.Errorf("invalid role template description: %v", roleTemplates[0].Description)
	}
	if !reflect.DeepEqual(roleTemplates[0].AssignedPermissions, []Permission{{
		Operation: "CONFIGURE_DATA_CLASS_GLOBAL",
		Hierarchies: []SnappableHierarchy{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"GlobalResource"},
		}},
	}, {
		Operation: "EXPORT_DATA_CLASS_GLOBAL",
		Hierarchies: []SnappableHierarchy{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"GlobalResource"},
		}},
	}, {
		Operation: "VIEW_DATA_CLASS_GLOBAL",
		Hierarchies: []SnappableHierarchy{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"GlobalResource"},
		}},
	}}) {
		t.Errorf("invalid role template permissions: %#v", roleTemplates[0].AssignedPermissions)
	}
}
