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

// Example showing how to manage an AWS account with the Polaris Go SDK. The
// configuration file should contain:
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
	polAccount, err := polaris.DefaultAccount("default")
	if err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClient(polAccount, &polaris_log.StandardLogger{})
	if err != nil {
		log.Fatal(err)
	}

	// Add the AWS default account to Polaris. Usually resolved using the
	// environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and
	// AWS_DEFAULT_REGION.
	err = client.AwsAccountAdd(ctx, polaris.FromAwsDefault(),
		polaris.WithName("Trinity-TPM-DevOps"), polaris.WithRegion("us-east-2"))
	if err != nil {
		log.Fatal(err)
	}

	// Lookup the newly added account.
	account, err := client.AwsAccount(ctx, polaris.FromAwsDefault())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %v, NativeID: %v\n", account.Name, account.NativeID)
	for _, feature := range account.Features {
		fmt.Printf("Feature: %v, Regions: %v, Status: %v\n", feature.Feature, feature.AwsRegions, feature.Status)
	}

	// Remove the AWS account from Polaris.
	if err := client.AwsAccountRemove(ctx, polaris.FromAwsDefault(), false); err != nil {
		log.Fatal(err)
	}
}
