// Copyright 2026 Rubrik, Inc.
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
	"log"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/devops"
	gqldevops "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/devops"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to read GitHub organizations and their repositories with
// the SDK. GitHub organizations cannot be onboarded through the SDK; onboard
// them via the RSC UI and use the SDK to enumerate the resulting hierarchy.
//
// The RSC service account key file, identifying the RSC account, should be
// pointed out by the RUBRIK_POLARIS_SERVICEACCOUNT_FILE environment variable.
func main() {
	ctx := context.Background()

	// Load configuration and create the client.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	logger := polarislog.NewStandardLogger()
	if err := polaris.SetLogLevelFromEnv(logger); err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClientWithLogger(polAccount, logger)
	if err != nil {
		log.Fatal(err)
	}

	// List every GitHub organization in the account, and for each list its
	// repositories.
	orgs, err := devops.Wrap(client).GitHubOrganizations(ctx, gqldevops.QueryTypeChildren, hierarchy.GitHubRoot)
	if err != nil {
		log.Fatal(err)
	}
	for _, org := range orgs {
		log.Printf("GitHubOrganization: %s (native id: %s, id: %s, repos: %d)\n", org.Name, org.NativeID, org.ID, org.RepoCount)

		repos, err := devops.Wrap(client).GitHubRepositories(ctx, gqldevops.QueryTypeDescendants, org.ID.String())
		if err != nil {
			log.Fatal(err)
		}
		for _, repo := range repos {
			log.Printf("  GitHubRepository: %s (id: %s, size: %d)\n", repo.Name, repo.ID, repo.Size)
		}
	}
}
