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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/devops"
	gqldevops "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/devops"
	azureregions "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to onboard an Azure DevOps organization with the SDK.
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

	// Register the customer application for Azure DevOps. This is a separate
	// credential store from the cloud native protection application, selected
	// by the AppUseCaseDevOps use case.
	const tenantDomain = "my-domain.onmicrosoft.com"
	if _, err := azure.Wrap(client).AddServicePrincipalForUseCase(ctx, azure.Default(tenantDomain), azure.AppUseCaseDevOps, false); err != nil {
		log.Fatal(err)
	}

	// Onboard an Azure DevOps organization using Rubrik-hosted exocompute and
	// Rubrik Cloud Vault storage. AddAzureCloudAccount resolves and returns the
	// onboarded organizations.
	const organizationID = "my-azure-devops-org"
	exocomputeRegion := azureregions.RegionEastUS
	orgs, err := devops.Wrap(client).AddAzureCloudAccount(ctx, gqldevops.AddAzureCloudAccountParams{
		OrganizationNativeIDs: []string{organizationID},
		TenantDomain:          tenantDomain,
		Features:              devops.AzureSupportedFeatures(),
		HostType:              gqldevops.HostTypeRubrik,
		StorageType:           gqldevops.StorageTypeRCV,
		ExocomputeRegion:      &exocomputeRegion,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, org := range orgs {
		log.Printf("AzureOrganization: %s (native id: %s, id: %s)\n", org.Name, org.NativeID, org.ID)
	}

	// Remove the onboarded organization, keeping its snapshots.
	if err := devops.Wrap(client).DeleteAzureCloudAccount(ctx, orgs[0].ID, false); err != nil {
		log.Fatal(err)
	}
}
