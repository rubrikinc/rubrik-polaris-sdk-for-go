package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/cloudcluster"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/cluster"
	cloudclustergql "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cloudcluster"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	azuregqlregions "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func main() {
	ctx := context.Background()

	// Load configuration and create client.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}

	logger := polarislog.NewStandardLogger()
	logger.SetLogLevel(polarislog.Trace)
	if err := polaris.SetLogLevelFromEnv(logger); err != nil {
		log.Fatal(err)
	}

	client, err := polaris.NewClientWithLogger(polAccount, logger)
	if err != nil {
		log.Fatal(err)
	}

	azureClient := azure.Wrap(client)
	cloudClusterClient := cloudcluster.Wrap(client)
	clusterClient := cluster.Wrap(client)

	azureFqdn := "my-domain.onmicrosoft.com"
	azureSubscriptionID := "abcdefg-a123-12ab-1a23-1a2b3c45de6f"

	// Add default Azure service principal to Polaris. Usually resolved using
	// the environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	_, err = azureClient.SetServicePrincipal(ctx, azure.Default(azureFqdn))
	if err != nil {
		log.Fatal(err)
	}

	// Add Azure subscription to Polaris.
	subscription := azure.Subscription(uuid.MustParse(azureSubscriptionID),
		azureFqdn)
	id, err := azureClient.AddSubscription(ctx, subscription, core.FeatureServerAndApps, azure.Regions("westus"))
	if err != nil {
		log.Fatal(err)
	}

	// Lookup the newly added subscription.
	account, err := azureClient.SubscriptionByID(ctx, id)
	if err != nil {
		log.Fatal(err)
	}

	// Print the subscription details.
	log.Printf("Name: %v, NativeID: %v\n", account.Name, account.NativeID)
	for _, feature := range account.Features {
		log.Printf("Feature: %v, Regions: %v, Status: %v\n", feature.Name, feature.Regions, feature.Status)
	}

	// Create the Cloud Cluster
	resourceGroup := "resource-group"
	cluster, err := cloudClusterClient.CreateAzureCloudCluster(ctx, cloudclustergql.CreateAzureClusterInput{
		CloudAccountID:       account.ID,
		IsESType:             true,
		KeepClusterOnFailure: false,
		ClusterConfig: cloudclustergql.AzureClusterConfig{
			ClusterName:      "cces-cluster",
			UserEmail:        "hello@domain.com",
			AdminPassword:    secret.String("RubrikGoForward!"),
			DNSNameServers:   []string{"8.8.8.8"},
			DNSSearchDomains: []string{"rubrikdemo.com"},
			NTPServers:       []string{"pool.ntp.org"},
			NumNodes:         3,
			AzureESConfig: cloudclustergql.AzureEsConfigInput{
				ResourceGroup:         resourceGroup,
				StorageAccount:        "storage-account",
				ContainerName:         "container-name",
				ShouldCreateContainer: false,
				EnableImmutability:    false,
				ManagedIdentity: cloudclustergql.AzureManagedIdentityName{
					Name: "managed-identity",
				},
			},
		},
		Validations: []cloudclustergql.ClusterCreateValidations{
			cloudclustergql.AllChecks,
		},
		VMConfig: cloudclustergql.AzureVMConfig{
			CDMVersion:                   "9.2.3-p7-29713",
			InstanceType:                 cloudclustergql.AzureInstanceTypeStandardD8SV5,
			Location:                     azuregqlregions.RegionWestUS,
			ResourceGroup:                resourceGroup,
			NetworkResourceGroup:         resourceGroup,
			VnetResourceGroup:            resourceGroup,
			Subnet:                       "subnet-id",
			Vnet:                         "vnet-id",
			NetworkSecurityGroup:         "nsg-id",
			NetworkSecurityResourceGroup: resourceGroup,
			VMType:                       cloudclustergql.CCVmConfigExtraDense,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Cloud Cluster ID: %v\n", cluster.ID)
	log.Printf("Cloud Cluster Name: %v\n", cluster.Name)
	log.Printf("Cloud Cluster Status: %v\n", cluster.Status)

	// Attempt normal removal first (isForce = false)
	// If the cluster has blocking conditions and is eligible for force removal,
	// you can set isForce = true
	info, err := clusterClient.RemoveCluster(ctx, cluster.ID, false, 0)
	if err != nil {
		log.Fatal(err)
	}

	// Display cluster removal information
	log.Print("\nCluster Removal Prechecks:\n")
	log.Printf("  Can Ignore Precheck: %v\n", info.Prechecks.IgnorePrecheck)
	log.Printf("  Is Disconnected: %v\n", info.Prechecks.Disconnected)
	log.Printf("  Is Air Gapped: %v\n", info.Prechecks.AirGapped)
	log.Printf("  Last Connection Time: %v\n", info.Prechecks.LastConnectionTime)
	log.Printf("  Ignore Precheck Time: %v\n", info.Prechecks.IgnorePrecheckTime)

	log.Print("\nRCV Locations for cluster:\n")
	for _, location := range info.RCVLocations {
		log.Printf("  ID: %v, Name: %v\n", location.ID, location.Name)
	}

	log.Print("\nForce Removal Eligibility:\n")
	log.Printf("  Blocking Conditions: %v\n", info.BlockingConditions)
	log.Printf("  Force Removal Eligible: %v\n", info.ForceRemovalEligible)
	log.Printf("  Force Removable: %v\n", info.ForceRemovable)
}
