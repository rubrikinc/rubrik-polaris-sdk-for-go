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

package sla

import (
	"context"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

var client *graphql.Client

func TestMain(m *testing.M) {
	if testsetup.BoolEnvSet("TEST_INTEGRATION") {
		polAccount, err := polaris.DefaultServiceAccount(true)
		if err != nil {
			fmt.Printf("failed to get default service account: %v\n", err)
			os.Exit(1)
		}

		logger := log.NewStandardLogger()
		logger.SetLogLevel(log.Info)
		if err := polaris.SetLogLevelFromEnv(logger); err != nil {
			fmt.Printf("failed to get log level from env: %v\n", err)
			os.Exit(1)
		}

		polClient, err := polaris.NewClientWithLogger(polAccount, logger)
		if err != nil {
			fmt.Printf("failed to create polaris client: %v\n", err)
			os.Exit(1)
		}
		client = polClient.GQL
	}

	os.Exit(m.Run())
}

// TestDomainByID ensures that the two graphql queries to fetch an SLA domain
// (by ID and by list) return the same result. If it fails, it's likely that the
// graphql queries needs to be updated.
func TestDomainByID(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	slaID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	domainByID, err := DomainByID(ctx, client, slaID)
	if err != nil {
		t.Fatalf("DomainByID failed: %v", err)
	}

	domains, err := ListDomains(ctx, client, nil)
	if err != nil {
		t.Fatalf("ListDomains failed: %v", err)
	}

	var domainFromList *Domain
	for i := range domains {
		if domains[i].ID == slaID {
			domainFromList = &domains[i]
			break
		}
	}
	if domainFromList == nil {
		t.Fatalf("SLA domain %s not found in list", slaID)
	}

	// Sort ObjectType in both as the order is not guaranteed.
	slices.Sort(domainFromList.ObjectTypes)
	slices.Sort(domainByID.ObjectTypes)

	if diff := cmp.Diff(*domainFromList, domainByID); diff != "" {
		t.Errorf("DomainByID mismatch (-list +byID):\n%s", diff)
	}
}
