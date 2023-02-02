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

// Example showing how to manage an AWS account with the Polaris Go SDK.
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

	// List the roles available in RSC.
	roles, err := accessClient.Roles(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	for _, role := range roles {
		fmt.Printf("Name: %q, Description: %q\n", role.Name, role.Description)
	}

	fmt.Println()

	// List the role templates available in RSC.
	templates, err := accessClient.RoleTemplates(ctx, "compliance")
	if err != nil {
		log.Fatal(err)
	}
	for _, template := range templates {
		fmt.Printf("Name: %q, Description: %q\n", template.Name, template.Description)
		for _, permission := range template.AssignedPermissions {
			fmt.Printf("  Operation: %q\n", permission.Operation)
			for _, hierarchy := range permission.Hierarchies {
				fmt.Printf("    SnappableType: %s\n", hierarchy.SnappableType)
				for _, objectID := range hierarchy.ObjectIDs {
					fmt.Printf("      ObjectID: %s\n", objectID)
				}
			}
		}
	}
}
