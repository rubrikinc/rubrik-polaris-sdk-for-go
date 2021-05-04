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

	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris"
	polaris_log "github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to manage a GCP project with the Polaris Go SDK using a
// previously set GCP service account. The configuration file should contain:
//
//   {
//     "default": {
//       "username": "<your-polaris-username>",
//       "password": "<your-polaris-password>",
//       "url": "<your-polaris-url>",
//       "loglevel": "trace"
//     }
//   }
func main() {
	ctx := context.Background()

	// Load configuration and create client.
	account, err := polaris.DefaultAccount("default")
	if err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClient(account, &polaris_log.StandardLogger{})
	if err != nil {
		log.Fatal(err)
	}

	// Add the service account to Polaris.
	if err = client.GcpServiceAccountSet(ctx, polaris.FromGcpDefault()); err != nil {
		log.Fatal(err)
	}

	// Add the GCP project to Polaris. Usually resolved using the environment
	// variable GOOGLE_APPLICATION_CREDENTIALS.
	err = client.GcpProjectAdd(ctx,
		polaris.FromGcpProject("trinity-fdse", "Trinity-FDSE", 994761414559, "Trinity Org"))
	if err != nil {
		log.Fatal(err)
	}

	// Lookup the newly added project.
	project, err := client.GcpProject(ctx, polaris.WithGcpProjectNumber(994761414559))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %v, ProjectID: %v, ProjectNumber: %v, OrgName: %v\n", project.Name, project.ProjectID,
		project.ProjectNumber, project.OrganizationName)
	for _, feature := range project.Features {
		fmt.Printf("Feature: %v, Status: %v\n", feature.Feature, feature.Status)
	}

	// Remove the GCP account from Polaris.
	if err := client.GcpProjectRemove(ctx, polaris.WithGcpProjectNumber(994761414559), false); err != nil {
		log.Fatal(err)
	}
}
