// Copyright 2024 Rubrik, Inc.
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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	gqlsla "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/sla"
)

// Example showing how to manage SLA domains and tag rules.
//
// The RSC service account key file identifying the RSC account should be
// pointed out by the RUBRIK_POLARIS_SERVICEACCOUNT_FILE environment variable.
func main() {
	ctx := context.Background()

	// Load configuration and create client.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClientWithLogger(polAccount, polarislog.NewStandardLogger())
	if err != nil {
		log.Fatal(err)
	}

	slaClient := sla.Wrap(client)

	// Create SLA domain in RSC.
	slaID, err := slaClient.CreateDomain(ctx, gqlsla.CreateDomainParams{
		Name:        "test-sla-42",
		Description: "test-description",
		ObjectTypes: []gqlsla.ObjectType{gqlsla.ObjectAWSEC2EBS},
		BackupWindows: []gqlsla.BackupWindow{{
			DurationInHours: 5,
			StartTime: gqlsla.StartTime{
				DayOfWeek: gqlsla.DayOfWeek{Day: gqlsla.Monday},
				Hour:      1,
				Minute:    13,
			},
		}},
		FirstFullBackupWindows: []gqlsla.BackupWindow{{
			DurationInHours: 1,
			StartTime: gqlsla.StartTime{
				DayOfWeek: gqlsla.DayOfWeek{Day: gqlsla.Wednesday},
				Hour:      18,
				Minute:    5,
			},
		}},
		ArchivalSpecs: []gqlsla.ArchivalSpec{{
			GroupID:       uuid.MustParse("b5f060e4-1025-4989-a4ac-29cd34fe40f6"),
			Frequencies:   []gqlsla.RetentionUnit{gqlsla.Days},
			Threshold:     1,
			ThresholdUnit: gqlsla.Days,
		}},
		SnapshotSchedule: gqlsla.SnapshotSchedule{
			Daily: &gqlsla.DailySnapshotSchedule{
				BasicSchedule: gqlsla.BasicSnapshotSchedule{
					Frequency:     1,
					Retention:     2,
					RetentionUnit: gqlsla.Days,
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create tag rule in RSC.
	tagRuleID, err := slaClient.CreateTagRule(ctx, gqlsla.CreateTagRuleParams{
		Name:       "test-tag-rule",
		ObjectType: gqlsla.TagObjectAWSEC2Instance,
		Tag: gqlsla.Tag{
			Key:   "test-key",
			Value: "test-value",
		},
		AllCloudAccounts: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// List tag rules available in RSC using the tag rules name filter.
	tagRules, err := slaClient.TagRules(ctx, "test-")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Tag rules with \"test-\" as a name prefix:")
	for _, tagRule := range tagRules {
		fmt.Printf("ID: %s, Name: %q, Object Type: %q\n", tagRule.ID, tagRule.Name, tagRule.ObjectType)
	}

	// Assign SLA domain to the tag rule.
	err = slaClient.AssignDomain(ctx, gqlsla.AssignDomainParams{
		DomainID:         &slaID,
		DomainAssignType: gqlsla.ProtectWithSLA,
		ObjectIDs:        []uuid.UUID{tagRuleID},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Remove the tag rule from RSC.
	if err := slaClient.DeleteTagRule(ctx, tagRuleID); err != nil {
		log.Fatal(err)
	}

	// Remove the SLA domain from RSC.
	if err := slaClient.DeleteDomain(ctx, slaID); err != nil {
		log.Fatal(err)
	}
}
