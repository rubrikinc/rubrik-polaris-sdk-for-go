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

// Package cloudcluster provides a high level interface to the AWS part of the RSC
// platform.

package cloudcluster

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/event"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cloudcluster"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	gqlevent "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/event"
	gqlaws "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for cloud cluster management.
type API struct {
	client *graphql.Client
	log    log.Logger
}

// Wrap the RSC client in the cloud cluster API.
func Wrap(client *polaris.Client) API {
	return API{
		client: client.GQL,
		log:    client.GQL.Log(),
	}
}

type CloudCluster struct {
	ID             uuid.UUID
	Name           string
	CdmVersion     string
	CdmProduct     string
	InstanceType   string
	Region         string
	CloudAccountID uuid.UUID
	Status         gqlevent.ActivityStatus
}

// CreateCloudCluster creates a cloud cluster in the specified account AWS account.
func (a API) CreateCloudCluster(ctx context.Context, input cloudcluster.CreateAwsClusterInput, useLatestCdmVersion bool) (cluster CloudCluster, err error) {
	a.log.Print(log.Trace)

	awsClient := aws.WrapGQL(a.client)

	// Ensure account exists and has Server and Apps feature
	account, err := awsClient.AccountByID(ctx, input.CloudAccountID)
	if err != nil {
		return CloudCluster{}, err
	}
	if _, ok := account.Feature(core.FeatureServerAndApps); !ok {
		return CloudCluster{}, fmt.Errorf("account %q missing feature %s", account.ID, core.FeatureServerAndApps.Name)
	}

	// validate region in input
	inputRegion := gqlaws.RegionFromName(input.Region)
	if inputRegion == gqlaws.RegionUnknown {
		return CloudCluster{}, fmt.Errorf("unknown region: %s", input.Region)
	}

	// Get Available CDM versions
	cdmVersions, err := cloudcluster.Wrap(a.client).AllAwsCdmVersions(ctx, input.CloudAccountID, inputRegion)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get cdm versions: %s", err)
	}

	// Validate CDM version is available
	validCdmVersion := false
	var supportedInstanceTypes []cloudcluster.AwsCCInstanceType
	for _, version := range cdmVersions {
		if (version.IsLatest && useLatestCdmVersion) || (version.Version == input.VMConfig.CDMVersion) {
			validCdmVersion = true
			input.VMConfig.CDMVersion = version.Version
			input.VMConfig.CDMProduct = version.ProductCodes[0]
			supportedInstanceTypes = version.SupportedInstanceTypes
			break
		}
	}

	if !validCdmVersion {
		return CloudCluster{}, fmt.Errorf("cdm version %s is not available for account %s", input.VMConfig.CDMVersion, account.ID)
	}

	// ensure specified instance type is supported
	validInstanceType := slices.Contains(supportedInstanceTypes, input.VMConfig.InstanceType)
	if !validInstanceType {
		return CloudCluster{}, fmt.Errorf("instance type %s is not supported for cdm version %s, supported Instance types are: %v", input.VMConfig.InstanceType, input.VMConfig.CDMVersion, supportedInstanceTypes)
	}

	// Get Available configured regions
	regions, err := cloudcluster.Wrap(a.client).AwsCloudAccountRegions(ctx, account.ID)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get cloud account regions from RSC: %s", err)
	}

	// Validate the input region is configured
	validRegion := slices.Contains(regions, inputRegion)
	if !validRegion {
		return CloudCluster{}, fmt.Errorf("region %s is not configured for RSC AWS account %s", input.Region, account.ID)
	}

	// Validate that the VPC exists in RSC metadata via AwsCloudAccountListVpcs
	vpcs, err := cloudcluster.Wrap(a.client).AwsCloudAccountListVpcs(ctx, input.CloudAccountID, inputRegion)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get vpcs from RSC: %s", err)
	}

	vpcSyncedToRsc := slices.ContainsFunc(vpcs, func(vpc cloudcluster.AwsCloudAccountListVpcs) bool {
		return vpc.VpcID == input.VMConfig.VPC
	})
	if !vpcSyncedToRsc {
		return CloudCluster{}, fmt.Errorf("vpc %s does not exist in RSC AWS account %s for region %s. Check the VPC ID and region. If this was recently created, wait a few minutes and try again", input.VMConfig.VPC, account.ID, input.Region)
	}

	// Validate Instance Profile exists in RSC metadata via AllAwsInstanceProfileNames
	instanceProfiles, err := cloudcluster.Wrap(a.client).AllAwsInstanceProfileNames(ctx, account.ID, inputRegion)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get instance profiles: %s", err)
	}
	validInstanceProfile := slices.Contains(instanceProfiles, input.VMConfig.InstanceProfileName)
	if !validInstanceProfile {
		return CloudCluster{}, fmt.Errorf("instance profile %s does not exist in RSC AWS account %s", input.VMConfig.InstanceProfileName, account.ID)
	}

	// Validate Subnet exists in RSC metadata via AwsCloudAccountListSubnets
	subnets, err := cloudcluster.Wrap(a.client).AwsCloudAccountListSubnets(ctx, input.CloudAccountID, inputRegion, input.VMConfig.VPC)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get subnets: %s", err)
	}
	validSubnet := slices.ContainsFunc(subnets, func(subnet cloudcluster.AwsCloudAccountSubnets) bool {
		return subnet.SubnetID == input.VMConfig.Subnet
	})
	if !validSubnet {
		return CloudCluster{}, fmt.Errorf("subnet %s does not exist in RSC AWS account %s", input.VMConfig.Subnet, account.ID)
	}

	// Validate Security Groups
	securityGroups, err := cloudcluster.Wrap(a.client).AwsCloudAccountListSecurityGroups(ctx, input.CloudAccountID, inputRegion, input.VMConfig.VPC)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get security groups: %s", err)
	}
	// Validate Security Groups - check that all provided security groups exist
	for _, inputSG := range input.VMConfig.SecurityGroups {
		validSecurityGroup := slices.ContainsFunc(securityGroups, func(securityGroup cloudcluster.AwsCloudAccountSecurityGroup) bool {
			return securityGroup.SecurityGroupID == inputSG
		})
		if !validSecurityGroup {
			return CloudCluster{}, fmt.Errorf("security group %s does not exist in RSC AWS account %s", inputSG, account.ID)
		}
	}

	// Validate CloudCluster Request
	err = cloudcluster.Wrap(a.client).ValidateCreateAwsClusterInput(ctx, input)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to validate create cloud cluster: %s", err)
	}

	// JobID is ignored here due to a bug in the RSC API
	_, err = cloudcluster.Wrap(a.client).CreateAwsCloudCluster(ctx, input)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to create cloud cluster: %s", err)
	}

	cluster, err = a.monitorCloudClusterEvents(ctx, input.ClusterConfig.ClusterName, input.CloudAccountID, input.VMConfig.CDMVersion, input.VMConfig.CDMProduct, string(input.VMConfig.InstanceType), input.Region)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to monitor cloud cluster events: %s", err)
	}

	return cluster, nil
}

// CreateAzureCloudCluster creates an Azure Cloud Cluster with the specified configuration.
// It validates the cloud account, CDM version, marketplace terms, managed identity,
// resource group, and subnet before creating the cluster. Returns the created cluster
// details after monitoring the creation process.
func (a API) CreateAzureCloudCluster(ctx context.Context, input cloudcluster.CreateAzureClusterInput) (cluster CloudCluster, err error) {
	a.log.Print(log.Trace)

	// Validate Cloud Account exists and has Server and Apps feature
	azureClient := azure.WrapGQL(a.client)
	account, err := azureClient.SubscriptionByID(ctx, input.CloudAccountID)
	if err != nil {
		return CloudCluster{}, err
	}

	if _, ok := account.Feature(core.FeatureServerAndApps); !ok {
		return CloudCluster{}, fmt.Errorf("account %q missing feature %s", account.ID, core.FeatureServerAndApps.Name)
	}

	// Validate CDM version is available
	cdmVersions, err := cloudcluster.Wrap(a.client).AllAzureCdmVersions(ctx, input.CloudAccountID, input.VMConfig.Location)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get cdm versions: %s", err)
	}

	// Validate CDM version is available
	validCdmVersion := false
	cdmVersion := input.VMConfig.CDMVersion
	var supportedInstanceTypes []cloudcluster.AzureCCESSupportedInstanceType
	for _, version := range cdmVersions {
		if version.CDMVersion == cdmVersion {
			validCdmVersion = true
			// We need to clobber the input version because CCES GQL expects the
			// internal version and SKU, not the cdm product version
			input.VMConfig.CDMVersion = version.Version
			input.VMConfig.CDMProduct = version.SKU
			supportedInstanceTypes = version.SupportedInstanceTypes
			break
		}
	}
	if !validCdmVersion {
		return CloudCluster{}, fmt.Errorf("cdm version %s is not available for account %s", cdmVersion, account.ID)
	}

	// ensure specified instance type is supported
	validInstanceType := slices.Contains(supportedInstanceTypes, input.VMConfig.InstanceType)
	if !validInstanceType {
		return CloudCluster{}, fmt.Errorf("instance type %s is not supported for cdm version %s, supported Instance types are: %v", input.VMConfig.InstanceType, input.VMConfig.CDMVersion, supportedInstanceTypes)
	}

	// validate marketplace agreement
	marketplaceTerms, err := cloudcluster.Wrap(a.client).AzureMarketplaceTerms(ctx, input.CloudAccountID, cdmVersion)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("error validating marketplace terms: %s", err)
	}
	// The message when terms are not accepted will include the terms link and specific version of which the terms need to be accepted.
	if !marketplaceTerms.TermsAccepted {
		return CloudCluster{}, fmt.Errorf("%s", marketplaceTerms.Message)
	}
	if marketplaceTerms.MarketplaceSKU == "" {
		return CloudCluster{}, fmt.Errorf("marketplace sku is not available for cdm version %s", cdmVersion)
	}

	// Find ManagedIdentity by name and set client ID and Resource Group
	managedIdentities, err := cloudcluster.Wrap(a.client).AzureCCManagedIdentities(ctx, input.CloudAccountID)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get managed identities: %s", err)
	}
	var managedIdentity []cloudcluster.AzureCCManagedIdentity
	for _, mi := range managedIdentities {
		if mi.Name == input.ClusterConfig.AzureESConfig.ManagedIdentity.Name && mi.ResourceGroup == input.VMConfig.ResourceGroup {
			managedIdentity = append(managedIdentity, mi)
		}
	}
	if n := len(managedIdentity); n != 1 {
		return CloudCluster{}, fmt.Errorf("managed identity %s does not exist in RSC Azure account %s, we found %d matches", input.ClusterConfig.AzureESConfig.ManagedIdentity.Name, account.ID, n)
	}

	// update input managed identity with client ID and resource group
	input.ClusterConfig.AzureESConfig.ManagedIdentity = cloudcluster.AzureManagedIdentityName{
		Name:          input.ClusterConfig.AzureESConfig.ManagedIdentity.Name,
		ResourceGroup: managedIdentity[0].ResourceGroup,
		ClientID:      managedIdentity[0].ClientID,
	}

	// Validate Resource Group exists in RSC metadata via AzureCCResourceGroups
	resourceGroups, err := cloudcluster.Wrap(a.client).AzureCCResourceGroups(ctx, input.CloudAccountID, account.NativeID)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get resource groups: %s", err)
	}
	validResourceGroup := slices.ContainsFunc(resourceGroups, func(resourceGroup cloudcluster.AzureCCResourceGroup) bool {
		return resourceGroup.Name == input.VMConfig.ResourceGroup
	})
	if !validResourceGroup {
		return CloudCluster{}, fmt.Errorf("resource group %s does not exist in RSC Azure account %s", input.VMConfig.ResourceGroup, account.ID)
	}

	// Validate Subnet exists in RSC metadata via AzureCCSubnets
	subnets, err := cloudcluster.Wrap(a.client).AzureCCSubnets(ctx, input.CloudAccountID, input.VMConfig.Location)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get subnets: %s", err)
	}
	validSubnet := slices.ContainsFunc(subnets, func(subnet cloudcluster.AzureCCSubnet) bool {
		return subnet.Name == input.VMConfig.Subnet
	})
	if !validSubnet {
		return CloudCluster{}, fmt.Errorf("subnet %s does not exist in RSC Azure account %s", input.VMConfig.Subnet, account.ID)
	}

	// Validate CloudCluster Request
	err = cloudcluster.Wrap(a.client).ValidateCreateAzureClusterInput(ctx, input)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to validate create cloud cluster: %s", err)
	}

	// JobID is ignored here due to a bug in the RSC API
	_, err = cloudcluster.Wrap(a.client).CreateAzureCloudCluster(ctx, input)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to create cloud cluster: %s", err)
	}

	cluster, err = a.monitorCloudClusterEvents(ctx, input.ClusterConfig.ClusterName, input.CloudAccountID, cdmVersion, input.VMConfig.CDMProduct, string(input.VMConfig.InstanceType), input.VMConfig.Location.Name())
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to monitor cloud cluster events: %s", err)
	}

	return cluster, nil
}

// monitorCloudClusterEvents monitors the events for a cloud cluster create job and returns the cloud cluster object when complete.
func (a API) monitorCloudClusterEvents(ctx context.Context, clusterName string, cloudAccountID uuid.UUID, cdmVersion string, cdmProduct string, instanceType string, region string) (CloudCluster, error) {
	a.log.Print(log.Trace)

	// Poll the event series for the cluster
	eventFilters := gqlevent.EventSeriesFilter{
		ObjectName:        clusterName,
		ObjectType:        []gqlevent.EventObjectType{gqlevent.EventObjectTypeCluster},
		LastUpdatedTimeGt: core.FormatTimestamp(time.Now().Add(-15 * time.Minute)),
	}

	eventSeries, err := gqlevent.Wrap(a.client).EventSeries(ctx, "", eventFilters, 100, gqlevent.EventSeriesSortFieldLastUpdated, core.SortOrderDesc)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get event series: %s", err)
	}
	eventSeriesID := ""
	clusterUUID := ""
	for _, eventSeriesRow := range eventSeries {
		if eventSeriesRow.ObjectName == clusterName {
			if event.InProgress(eventSeriesRow) {
				eventSeriesID = eventSeriesRow.ActivitySeriesID
				clusterUUID = eventSeriesRow.ClusterUUID
				break
			}
		}
	}

	if eventSeriesID == "" {
		return CloudCluster{}, fmt.Errorf("failed to find event series for cluster %s", clusterName)
	}
	if clusterUUID == "" {
		return CloudCluster{}, fmt.Errorf("failed to find cluster UUID for cluster %s", clusterName)
	}

	for {
		activitySeries, err := gqlevent.Wrap(a.client).ActivitySeries(ctx, eventSeriesID, clusterUUID)
		if err != nil {
			return CloudCluster{}, fmt.Errorf("failed to get event series: %s", err)
		}
		switch activitySeries.LastActivityStatus {
		case gqlevent.ActivityStatusQueued:
		case gqlevent.ActivityStatusRunning:
		case gqlevent.ActivityStatusTaskSuccess:
			if len(activitySeries.Activities.Nodes) > 0 {
				a.log.Printf(log.Info, "cloud cluster create in progress: %s\n", activitySeries.Activities.Nodes[0].Message)
			} else {
				a.log.Printf(log.Info, "cloud cluster create in progress: no activity details available")
			}
			time.Sleep(60 * time.Second)
			continue
		case gqlevent.ActivityStatusSuccess:
			clusterID, err := uuid.Parse(activitySeries.ClusterUUID)
			if err != nil {
				return CloudCluster{}, fmt.Errorf("failed to parse cluster UUID: %s", err)
			}
			return CloudCluster{
				ID:             clusterID,
				Name:           activitySeries.Cluster.Name,
				Status:         activitySeries.LastActivityStatus,
				CloudAccountID: cloudAccountID,
				CdmVersion:     cdmVersion,
				CdmProduct:     cdmProduct,
				InstanceType:   instanceType,
				Region:         region,
			}, nil
		case gqlevent.ActivityStatusFailure:
			return CloudCluster{}, fmt.Errorf("cloud cluster create failed: %s", activitySeries.Activities.Nodes[0].Message)
		case gqlevent.ActivityStatusCanceled:
			return CloudCluster{}, fmt.Errorf("cloud cluster create was canceled: %s", activitySeries.Activities.Nodes[0].Message)
		case gqlevent.ActivityStatusCanceling:
			return CloudCluster{}, fmt.Errorf("cloud cluster create is canceling %s", activitySeries.Activities.Nodes[0].Message)
		case gqlevent.ActivityStatusWarning:
			return CloudCluster{}, fmt.Errorf("cloud cluster create has warnings: %s", activitySeries.Activities.Nodes[0].Message)
		case gqlevent.ActivityStatusPartialSuccess:
			return CloudCluster{}, fmt.Errorf("cloud cluster create has partial success: %s", activitySeries.Activities.Nodes[0].Message)
		default:
			return CloudCluster{}, fmt.Errorf("cloud cluster create has unknown status: %s", activitySeries.LastActivityStatus)
		}

		// If we reach here, no matching event was found, wait and try again
		time.Sleep(10 * time.Second)
	}
}
