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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/cluster"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to list SLA source clusters.
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
	logger := polarislog.NewStandardLogger()
	if err := polaris.SetLogLevelFromEnv(logger); err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClientWithLogger(polAccount, logger)
	if err != nil {
		log.Fatal(err)
	}

	clusterClient := cluster.Wrap(client)

	// First, list all available clusters to see what's in the environment
	fmt.Println("Listing all available SLA source clusters:")
	allClusters, err := clusterClient.SLASourceClusters(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if len(allClusters) == 0 {
		fmt.Println("No SLA source clusters found in this environment.")
		return
	}
	for i, c := range allClusters {
		fmt.Printf("%d. ID: %s, Name: %q, Version: %q, IsConnected: %v\n",
			i+1, c.ID, c.Name, c.Version, c.ClusterInfo.IsConnected)
	}

	// Get a specific SLA source cluster by name.
	clusterName := allClusters[0].Name
	cluster, err := clusterClient.SLASourceClusterByName(ctx, clusterName)
	if err != nil {
		fmt.Printf("\nError finding cluster %q: %v\n", clusterName, err)
		return
	}
	fmt.Printf("\nSLA source cluster found by name:\nID: %s, Name: %q, Version: %q, IsConnected: %v\n",
		cluster.ID, cluster.Name, cluster.Version, cluster.ClusterInfo.IsConnected)
}
