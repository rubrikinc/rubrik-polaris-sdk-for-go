package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/cloudcluster"
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

	client, err := polaris.NewClientWithLogger(polAccount, logger)
	if err != nil {
		log.Fatal(err)
	}

	azureClient := azure.Wrap(client)
	cloudClusterClient := cloudcluster.Wrap(client)

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
	account, err := azureClient.Subscription(ctx, azure.CloudAccountID(id), core.FeatureAll)
	if err != nil {
		log.Fatal(err)
	}

	// Print the subscription details.
	fmt.Printf("Name: %v, NativeID: %v\n", account.Name, account.NativeID)
	for _, feature := range account.Features {
		fmt.Printf("Feature: %v, Regions: %v, Status: %v\n", feature.Name, feature.Regions, feature.Status)
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
			DNSSearchDomains: []string{},
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
	fmt.Printf("Cloud Cluster: %v", cluster)
}
