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

package sla

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the common RSC client used for tests. By reusing the same client
// we reduce the risk of hitting rate limits when access tokens are created.
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
		logger := log.NewStandardLogger()
		logger.SetLogLevel(log.Info)
		if err := polaris.SetLogLevelFromEnv(logger); err != nil {
			fmt.Printf("failed to get log level from env: %v\n", err)
			os.Exit(1)
		}

		client, err = polaris.NewClientWithLogger(polAccount, logger)
		if err != nil {
			fmt.Printf("failed to create polaris client: %v\n", err)
			os.Exit(1)
		}

		version, err := client.GQL.DeploymentVersion(context.Background())
		if err != nil {
			fmt.Printf("failed to get deployment version: %v\n", err)
			os.Exit(1)
		}
		logger.Printf(log.Info, "Polaris version: %s", version)
	}

	os.Exit(m.Run())
}

// TestTagRuleWithDeprecatedTag verifies that the deprecated Tag field in
// CreateTagRuleParams works correctly for backward compatibility.
//
// To run this test against an RSC instance the following environment variables
// need to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
func TestTagRuleWithDeprecatedTag(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	slaAPI := Wrap(client)

	// Create a tag rule using the deprecated Tag field for backward
	// compatibility testing.
	tagRuleID, err := slaAPI.CreateTagRule(ctx, sla.CreateTagRuleParams{
		Name:             "sdk-test-tag-rule-deprecated",
		ObjectType:       sla.TagObjectAWSEC2Instance,
		AllCloudAccounts: true,
		Tag: sla.Tag{
			Key:       "env",
			Value:     "test",
			AllValues: false,
		},
	})
	if err != nil {
		t.Fatalf("failed to create tag rule: %v", err)
	}

	// Ensure cleanup happens even if test fails.
	t.Cleanup(func() {
		if err := slaAPI.DeleteTagRule(ctx, tagRuleID); err != nil {
			t.Errorf("failed to delete tag rule: %v", err)
		}
	})

	// Read the tag rule back and verify both TagConditions and deprecated
	// Tag field are populated correctly.
	tagRule, err := slaAPI.TagRuleByID(ctx, tagRuleID)
	if err != nil {
		t.Fatalf("failed to get tag rule: %v", err)
	}

	// Verify TagConditions is populated correctly.
	if len(tagRule.TagConditions.TagPairs) != 1 {
		t.Fatalf("expected 1 tag pair, got %d", len(tagRule.TagConditions.TagPairs))
	}
	pair := tagRule.TagConditions.TagPairs[0]
	if pair.Key != "env" {
		t.Errorf("expected tag key 'env', got %q", pair.Key)
	}
	if len(pair.Values) != 1 || pair.Values[0] != "test" {
		t.Errorf("expected tag values ['test'], got %v", pair.Values)
	}
	if pair.MatchAllTagValues != false {
		t.Errorf("expected MatchAllTagValues=false, got %v", pair.MatchAllTagValues)
	}

	// Verify deprecated Tag field is populated for backward compatibility
	// (single tag pair with single value).
	//lint:ignore SA1019 testing deprecated field for backward compatibility
	if tagRule.Tag.Key != "env" {
		//lint:ignore SA1019 testing deprecated field for backward compatibility
		t.Errorf("expected deprecated Tag.Key 'env', got %q", tagRule.Tag.Key)
	}
	//lint:ignore SA1019 testing deprecated field for backward compatibility
	if tagRule.Tag.Value != "test" {
		//lint:ignore SA1019 testing deprecated field for backward compatibility
		t.Errorf("expected deprecated Tag.Value 'test', got %q", tagRule.Tag.Value)
	}
	//lint:ignore SA1019 testing deprecated field for backward compatibility
	if tagRule.Tag.AllValues != false {
		//lint:ignore SA1019 testing deprecated field for backward compatibility
		t.Errorf("expected deprecated Tag.AllValues=false, got %v", tagRule.Tag.AllValues)
	}
}

// TestTagRuleWithTagConditionsMultipleValues verifies that when TagConditions
// has multiple values, the deprecated Tag field is NOT populated.
//
// To run this test against an RSC instance the following environment variables
// need to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
func TestTagRuleWithTagConditionsMultipleValues(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	slaAPI := Wrap(client)

	// Create a tag rule using the new TagConditions field with multiple
	// values.
	tagRuleID, err := slaAPI.CreateTagRule(ctx, sla.CreateTagRuleParams{
		Name:             "sdk-test-tag-rule-multi-value",
		ObjectType:       sla.TagObjectAWSEC2Instance,
		AllCloudAccounts: true,
		TagConditions: &sla.TagConditions{
			TagPairs: []sla.TagPair{
				{
					Key:               "environment",
					MatchAllTagValues: true,
					Values:            []string{"prod", "staging", "dev"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create tag rule: %v", err)
	}

	// Ensure cleanup happens even if test fails.
	t.Cleanup(func() {
		if err := slaAPI.DeleteTagRule(ctx, tagRuleID); err != nil {
			t.Errorf("failed to delete tag rule: %v", err)
		}
	})

	// Read the tag rule back and verify TagConditions.
	tagRule, err := slaAPI.TagRuleByID(ctx, tagRuleID)
	if err != nil {
		t.Fatalf("failed to get tag rule: %v", err)
	}

	// Verify TagConditions is populated correctly.
	if len(tagRule.TagConditions.TagPairs) != 1 {
		t.Fatalf("expected 1 tag pair, got %d", len(tagRule.TagConditions.TagPairs))
	}
	pair := tagRule.TagConditions.TagPairs[0]
	if pair.Key != "environment" {
		t.Errorf("expected tag key 'environment', got %q", pair.Key)
	}
	if len(pair.Values) != 3 {
		t.Errorf("expected 3 tag values, got %d", len(pair.Values))
	}
	if pair.MatchAllTagValues != true {
		t.Errorf("expected MatchAllTagValues=true, got %v", pair.MatchAllTagValues)
	}

	// Verify deprecated Tag field is NOT populated when there are multiple
	// values (more than 1).
	//lint:ignore SA1019 testing deprecated field for backward compatibility
	if tagRule.Tag.Key != "" {
		//lint:ignore SA1019 testing deprecated field for backward compatibility
		t.Errorf("expected deprecated Tag.Key to be empty, got %q", tagRule.Tag.Key)
	}
	//lint:ignore SA1019 testing deprecated field for backward compatibility
	if tagRule.Tag.Value != "" {
		//lint:ignore SA1019 testing deprecated field for backward compatibility
		t.Errorf("expected deprecated Tag.Value to be empty, got %q", tagRule.Tag.Value)
	}
}

// TestTagRuleWithMatchAllTagValuesAndEmptyValues tests creating a tag rule
// with MatchAllTagValues set to true and an empty Values slice.
//
// To run this test against an RSC instance the following environment variables
// need to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
func TestTagRuleWithMatchAllTagValuesAndEmptyValues(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	slaAPI := Wrap(client)

	// Create a tag rule with MatchAllTagValues=true and an empty Values slice.
	tagRuleID, err := slaAPI.CreateTagRule(ctx, sla.CreateTagRuleParams{
		Name:             "sdk-test-tag-rule-match-all-empty",
		ObjectType:       sla.TagObjectAWSEC2Instance,
		AllCloudAccounts: true,
		TagConditions: &sla.TagConditions{
			TagPairs: []sla.TagPair{
				{
					Key:               "environment",
					MatchAllTagValues: true,
					Values:            []string{},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create tag rule: %v", err)
	}

	// Ensure cleanup happens even if test fails.
	t.Cleanup(func() {
		if err := slaAPI.DeleteTagRule(ctx, tagRuleID); err != nil {
			t.Errorf("failed to delete tag rule: %v", err)
		}
	})

	// Read the tag rule back and verify TagConditions.
	tagRule, err := slaAPI.TagRuleByID(ctx, tagRuleID)
	if err != nil {
		t.Fatalf("failed to get tag rule: %v", err)
	}

	// Verify TagConditions is populated correctly.
	if len(tagRule.TagConditions.TagPairs) != 1 {
		t.Fatalf("expected 1 tag pair, got %d", len(tagRule.TagConditions.TagPairs))
	}
	pair := tagRule.TagConditions.TagPairs[0]
	if pair.Key != "environment" {
		t.Errorf("expected tag key 'environment', got %q", pair.Key)
	}
	if pair.MatchAllTagValues != true {
		t.Errorf("expected MatchAllTagValues=true, got %v", pair.MatchAllTagValues)
	}
	if len(pair.Values) != 0 {
		t.Errorf("expected empty Values slice, got %v", pair.Values)
	}

	// Verify deprecated Tag field is populated correctly for a single pair
	// with no values and MatchAllTagValues=true.
	//lint:ignore SA1019 testing deprecated field for backward compatibility
	if tagRule.Tag.Key != "environment" {
		//lint:ignore SA1019 testing deprecated field for backward compatibility
		t.Errorf("expected deprecated Tag.Key 'environment', got %q", tagRule.Tag.Key)
	}
	//lint:ignore SA1019 testing deprecated field for backward compatibility
	if tagRule.Tag.Value != "" {
		//lint:ignore SA1019 testing deprecated field for backward compatibility
		t.Errorf("expected deprecated Tag.Value to be empty, got %q", tagRule.Tag.Value)
	}
	//lint:ignore SA1019 testing deprecated field for backward compatibility
	if tagRule.Tag.AllValues != true {
		//lint:ignore SA1019 testing deprecated field for backward compatibility
		t.Errorf("expected deprecated Tag.AllValues=true, got %v", tagRule.Tag.AllValues)
	}
}
