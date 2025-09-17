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

package aws

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/event"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cloudcluster"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	gqlevent "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/event"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

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

	// Ensure account exists and has Server and Apps feature
	account, err := a.AccountByID(ctx, core.FeatureAll, input.CloudAccountID)
	if err != nil {
		return CloudCluster{}, err
	}
	if _, ok := account.Feature(core.FeatureServerAndApps); !ok {
		return CloudCluster{}, fmt.Errorf("account %q missing feature %s", account.ID, core.FeatureServerAndApps.Name)
	}

	// validate region in input
	inputRegion := aws.RegionFromName(input.Region)
	if inputRegion == aws.RegionUnknown {
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
		if (version.IsLatest && useLatestCdmVersion) || (version.Version == input.VmConfig.CdmVersion) {
			validCdmVersion = true
			input.VmConfig.CdmVersion = version.Version
			input.VmConfig.CdmProduct = version.ProductCodes[0]
			supportedInstanceTypes = version.SupportedInstanceTypes
			break
		}
	}

	if !validCdmVersion {
		return CloudCluster{}, fmt.Errorf("cdm version %s is not available for account %s", input.VmConfig.CdmVersion, account.ID)
	}

	// ensure specified instance type is supported
	validInstanceType := slices.Contains(supportedInstanceTypes, input.VmConfig.InstanceType)
	if !validInstanceType {
		return CloudCluster{}, fmt.Errorf("instance type %s is not supported for cdm version %s, supported Instance types are: %v", input.VmConfig.InstanceType, input.VmConfig.CdmVersion, supportedInstanceTypes)
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
		return vpc.VpcID == input.VmConfig.Vpc
	})
	if !vpcSyncedToRsc {
		return CloudCluster{}, fmt.Errorf("vpc %s does not exist in RSC AWS account %s for region %s. Check the VPC ID and region. If this was recently created, wait a few minutes and try again", input.VmConfig.Vpc, account.ID, input.Region)
	}

	// Validate Instance Profile exists in RSC metadata via AllAwsInstanceProfileNames
	instanceProfiles, err := cloudcluster.Wrap(a.client).AllAwsInstanceProfileNames(ctx, account.ID, inputRegion)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get instance profiles: %s", err)
	}
	validInstanceProfile := slices.Contains(instanceProfiles, input.VmConfig.InstanceProfileName)
	if !validInstanceProfile {
		return CloudCluster{}, fmt.Errorf("instance profile %s does not exist in RSC AWS account %s", input.VmConfig.InstanceProfileName, account.ID)
	}

	// Validate Subnet exists in RSC metadata via AwsCloudAccountListSubnets
	subnets, err := cloudcluster.Wrap(a.client).AwsCloudAccountListSubnets(ctx, input.CloudAccountID, inputRegion, input.VmConfig.Vpc)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get subnets: %s", err)
	}
	validSubnet := slices.ContainsFunc(subnets, func(subnet cloudcluster.AwsCloudAccountSubnets) bool {
		return subnet.SubnetID == input.VmConfig.Subnet
	})
	if !validSubnet {
		return CloudCluster{}, fmt.Errorf("subnet %s does not exist in RSC AWS account %s", input.VmConfig.Subnet, account.ID)
	}

	// Validate Security Groups
	securityGroups, err := cloudcluster.Wrap(a.client).AwsCloudAccountListSecurityGroups(ctx, input.CloudAccountID, inputRegion, input.VmConfig.Vpc)
	if err != nil {
		return CloudCluster{}, fmt.Errorf("failed to get security groups: %s", err)
	}
	// Validate Security Groups - check that all provided security groups exist
	for _, inputSG := range input.VmConfig.SecurityGroups {
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

	// Poll the event series for the cluster
	eventFilters := gqlevent.EventSeriesFilter{
		ObjectName:        input.ClusterConfig.ClusterName,
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
		if eventSeriesRow.ObjectName == input.ClusterConfig.ClusterName {
			if event.InProgress(eventSeriesRow) {
				eventSeriesID = eventSeriesRow.ActivitySeriesID
				clusterUUID = eventSeriesRow.ClusterUUID
				break
			}
		}
	}

	if eventSeriesID == "" {
		return CloudCluster{}, fmt.Errorf("failed to find event series for cluster %s", input.ClusterConfig.ClusterName)
	}
	if clusterUUID == "" {
		return CloudCluster{}, fmt.Errorf("failed to find cluster UUID for cluster %s", input.ClusterConfig.ClusterName)
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
			a.log.Printf(log.Info, "AWS cloud cluster create in progress: %s\n", activitySeries.Activities.Nodes[0].Message)
			time.Sleep(60 * time.Second)
			continue
		case gqlevent.ActivityStatusSuccess:
			return CloudCluster{
				ID:             uuid.MustParse(activitySeries.ClusterUUID),
				Name:           activitySeries.Cluster.Name,
				Status:         activitySeries.LastActivityStatus,
				CloudAccountID: input.CloudAccountID,
				CdmVersion:     input.VmConfig.CdmVersion,
				CdmProduct:     input.VmConfig.CdmProduct,
				InstanceType:   string(input.VmConfig.InstanceType),
				Region:         input.Region,
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
