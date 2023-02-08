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

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/access"
	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to manage roles with the Polaris Go SDK.
//
// The Polaris service account key file identifying the Polaris account should
// either be placed at ~/.rubrik/polaris-service-account.json or pointed out by
// the RUBRIK_POLARIS_SERVICEACCOUNT_FILE environment variable.
func main() {
	ctx := context.Background()

	// Load configuration and create client.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClient(ctx, polAccount, polaris_log.NewStandardLogger())
	if err != nil {
		log.Fatal(err)
	}

	accessClient := access.Wrap(client)

	// Add role to RSC.
	roleID, err := accessClient.AddRole(ctx, "Test Role", "Test Role Description",
		[]access.Permission{{
			Operation: "VIEW_CLUSTER",
			Hierarchies: []access.SnappableHierarchy{{
				SnappableType: "AllSubHierarchyType",
				ObjectIDs:     []string{"CLUSTER_ROOT"},
			}},
		}}, access.NoProtectableClusters)
	if err != nil {
		log.Fatal(err)
	}

	// List roles available in RSC.
	roles, err := accessClient.Roles(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	for _, role := range roles {
		fmt.Printf("ID: %v, Name: %q, Description: %q\n", role.ID, role.Name, role.Description)
	}

	// Add role to user.
	if err := accessClient.AssignRole(ctx, roleID, "firstname.lastname@example.com"); err != nil {
		log.Fatal(err)
	}

	// Remove role from user.
	if err := accessClient.UnassignRole(ctx, roleID, "firstname.lastname@example.com"); err != nil {
		log.Fatal(err)
	}

	// Remove role from RSC.
	if err := accessClient.RemoveRole(ctx, roleID); err != nil {
		log.Fatal(err)
	}
}
