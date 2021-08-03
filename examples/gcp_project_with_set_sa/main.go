// Copyright 2021 Rubrik, Inc.
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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to manage a GCP project with the Polaris Go SDK using a
// previously set GCP service account.
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
	client, err := polaris.NewClientFromServiceAccount(ctx, polAccount, &polaris_log.StandardLogger{})
	if err != nil {
		log.Fatal(err)
	}

	// Add the service account to Polaris.
	err = client.GCP().SetServiceAccount(ctx, gcp.Default(), gcp.Name("global"))
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve service account name.
	name, err := client.GCP().ServiceAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Service Account Name: %v\n", name)

	// Add the GCP project to Polaris without any GCP credentials.
	id, err := client.GCP().AddProject(ctx, gcp.Project("my-project", 123456789012), gcp.Name("My Project"))
	if err != nil {
		log.Fatal(err)
	}

	// Lookup the newly added project.
	account, err := client.GCP().Project(ctx, gcp.CloudAccountID(id), core.CloudNativeProtection)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %v, ProjectID: %v, ProjectNumber: %v\n", account.Name, account.ID, account.ProjectNumber)
	for _, feature := range account.Features {
		fmt.Printf("Feature: %v, Status: %v\n", feature.Name, feature.Status)
	}

	// Remove the GCP account from Polaris.
	err = client.GCP().RemoveProject(ctx, gcp.CloudAccountID(id), false)
	if err != nil {
		log.Fatal(err)
	}
}
