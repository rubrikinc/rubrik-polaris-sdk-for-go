// Copyright 2025 Rubrik, Inc.
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
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	gqlaccess "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/access"
)

func TestRoleTemplates(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	accessClient := Wrap(client)

	roleTemplate, err := accessClient.RoleTemplateByName(ctx, "Compliance Auditor")
	if err != nil {
		t.Fatal(err)
	}
	if roleTemplate.ID != uuid.MustParse("00000000-0000-0000-0000-000000000007") {
		t.Errorf("invalid role template id: %s", roleTemplate.ID)
	}
	if roleTemplate.Name != "Compliance Auditor" {
		t.Errorf("invalid role template name: %s", roleTemplate.Name)
	}
	if roleTemplate.Description != "Template for compliance auditor" {
		t.Errorf("invalid role template description: %s", roleTemplate.Description)
	}

	// Sort permissions in ascending order before asserting.
	permissions := roleTemplate.AssignedPermissions
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i].Operation < permissions[j].Operation
	})
	if !reflect.DeepEqual(permissions, []gqlaccess.Permission{{
		Operation: "EXPORT_DATA_CLASS_GLOBAL",
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"GlobalResource"},
		}},
	}, {
		Operation: "VIEW_DATA_CLASS_GLOBAL",
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"GlobalResource"},
		}},
	}}) {
		t.Errorf("invalid role template permissions: %#v", permissions)
	}

	// List role templates using a name filter. A number of role templates comes
	// bundled with RSC. They can be used to create roles with a predefined set
	// of permissions.
	roleTemplates, err := accessClient.RoleTemplates(ctx, "Officer")
	if err != nil {
		t.Fatal(err)
	}
	if n := len(roleTemplates); n != 1 {
		t.Errorf("invalid number of role templates: %d", n)
	}
	if roleTemplates[0].ID != uuid.MustParse("00000000-0000-0000-0000-000000000006") {
		t.Errorf("invalid role template id: %s", roleTemplates[0].ID)
	}
	if roleTemplates[0].Name != "Compliance Officer" {
		t.Errorf("invalid role template name: %s", roleTemplates[0].Name)
	}
	if roleTemplates[0].Description != "Template for compliance officer" {
		t.Errorf("invalid role template description: %s", roleTemplates[0].Description)
	}

	// Sort permissions in ascending order before asserting.
	permissions = roleTemplates[0].AssignedPermissions
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i].Operation < permissions[j].Operation
	})
	if !reflect.DeepEqual(permissions, []gqlaccess.Permission{{
		Operation: "CONFIGURE_DATA_CLASS_GLOBAL",
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"GlobalResource"},
		}},
	}, {
		Operation: "EXPORT_DATA_CLASS_GLOBAL",
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"GlobalResource"},
		}},
	}, {
		Operation: "VIEW_DATA_CLASS_GLOBAL",
		ObjectsForHierarchyTypes: []gqlaccess.ObjectsForHierarchyType{{
			SnappableType: "AllSubHierarchyType",
			ObjectIDs:     []string{"GlobalResource"},
		}},
	}}) {
		t.Errorf("invalid role template permissions: %#v", permissions)
	}
}
