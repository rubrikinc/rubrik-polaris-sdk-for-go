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

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/cdm"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func main() {
	ctx := context.Background()

	// CDM client.
	cdmClient, err := cdm.NewClientFromEnv(true)
	if err != nil {
		log.Fatal(err)
	}

	// Cluster entitlements.
	clusterDetails, err := cdmClient.OfflineEntitle(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, nodeDetails := range clusterDetails {
		fmt.Printf("%#v\n", nodeDetails)
	}

	// RSC client.
	logger := polarislog.NewStandardLogger()
	if err := polaris.SetLogLevelFromEnv(logger); err != nil {
		log.Fatal(err)
	}
	account, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	rscClient, err := polaris.NewClientWithLogger(account, logger)
	if err != nil {
		log.Fatal(err)
	}

	// Register cluster with RSC.
	var regConfig []core.NodeRegistrationConfig
	for _, nodeDetails := range clusterDetails {
		regConfig = append(regConfig, nodeDetails.ToNodeRegistrationConfig())
	}
	authToken, productType, err := core.Wrap(rscClient.GQL).RegisterCluster(ctx, true, regConfig, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("AuthToken: %s, Product Type: %s\n", authToken, productType)

	// Finalize cluster registration.
	mode, err := cdmClient.SetRegisteredMode(ctx, authToken)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Registered Mode: %s\n", mode)
}
