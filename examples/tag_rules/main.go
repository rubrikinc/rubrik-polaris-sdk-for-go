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

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	gqlsla "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/sla"
)

// Example showing how to manage tag rules.
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

	// Remove the tag rule from RSC.
	if err := slaClient.DeleteTagRule(ctx, tagRuleID); err != nil {
		log.Fatal(err)
	}
}
