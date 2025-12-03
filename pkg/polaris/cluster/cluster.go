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

// Package cluster provides a high level interface to the cluster part of the RSC
// platform.
package cluster

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlcluster "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cluster"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for cluster management.
type API struct {
	client gqlcluster.API
	log    log.Logger
}

// Wrap the RSC client in the cluster API.
func Wrap(client *polaris.Client) API {
	return API{
		client: gqlcluster.Wrap(client.GQL),
		log:    client.GQL.Log(),
	}
}

// SLASourceClusters returns all SLA source clusters.
func (a API) SLASourceClusters(ctx context.Context) ([]gqlcluster.SLADataLocationCluster, error) {
	a.log.Print(log.Trace)

	clusters, err := gqlcluster.SLASourceClusters(ctx, a.client.GQL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list SLA source clusters: %s", err)
	}

	return clusters, nil
}

// SLASourceClusterByName returns the SLA source cluster with the specified name.
func (a API) SLASourceClusterByName(ctx context.Context, name string) (gqlcluster.SLADataLocationCluster, error) {
	a.log.Print(log.Trace)

	var filters []gqlcluster.ClusterFilter
	if name != "" {
		filters = append(filters, gqlcluster.ClusterFilter{
			Field:  "CLUSTER_NAME",
			Values: []string{name},
		})
	}
	clusters, err := gqlcluster.SLASourceClusters(ctx, a.client.GQL, filters)
	if err != nil {
		return gqlcluster.SLADataLocationCluster{}, fmt.Errorf("failed to list SLA source clusters: %s", err)
	}

	for _, cluster := range clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}

	return gqlcluster.SLADataLocationCluster{}, fmt.Errorf("cluster %q %w", name, graphql.ErrNotFound)
}

// CanIgnoreClusterRemovalPrechecks returns whether the cluster removal prechecks can be ignored
// for the specified cluster.
func (a API) CanIgnoreClusterRemovalPrechecks(ctx context.Context, clusterUUID uuid.UUID) (gqlcluster.ClusterRemovalPrechecks, error) {
	a.log.Print(log.Trace)

	prechecks, err := gqlcluster.CanIgnoreClusterRemovalPrechecks(ctx, a.client.GQL, clusterUUID)
	if err != nil {
		return gqlcluster.ClusterRemovalPrechecks{}, fmt.Errorf("failed to get cluster removal prechecks: %s", err)
	}

	return prechecks, nil
}

// RCVLocations returns all Recovery Cloud Vault locations for the specified cluster.
func (a API) RCVLocations(ctx context.Context, clusterUUID uuid.UUID) ([]gqlcluster.RCVLocation, error) {
	a.log.Print(log.Trace)

	locations, err := gqlcluster.ClusterRCVLocations(ctx, a.client.GQL, clusterUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster RCV locations: %s", err)
	}

	return locations, nil
}

// GlobalSLAs returns all global SLAs for the specified cluster.
func (a API) GlobalSLAs(ctx context.Context, clusterUUID uuid.UUID) ([]gqlcluster.GlobalSLA, error) {
	a.log.Print(log.Trace)

	slas, err := gqlcluster.AllClusterGlobalSLAs(ctx, a.client.GQL, clusterUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster global SLAs: %s", err)
	}

	return slas, nil
}

// VerifySLAReplication verifies if there are active SLAs with replication to the specified cluster.
// The includeArchived parameter determines whether to include archived SLAs in the check.
func (a API) VerifySLAReplication(ctx context.Context, clusterUUID uuid.UUID, includeArchived bool) (gqlcluster.SLAReplicationInfo, error) {
	a.log.Print(log.Trace)

	info, err := gqlcluster.VerifySLAWithReplicationToCluster(ctx, a.client.GQL, clusterUUID, includeArchived)
	if err != nil {
		return gqlcluster.SLAReplicationInfo{}, fmt.Errorf("failed to verify SLA replication: %s", err)
	}

	return info, nil
}

// RemoveCDMCluster removes the specified CDM cluster. The expireInDays parameter
// specifies the number of days before the cluster data expires. If nil, the default
// expiration is used. The isForce parameter forces removal even if prechecks fail.
func (a API) RemoveCDMCluster(ctx context.Context, clusterUUID uuid.UUID, expireInDays *int64, isForce bool) (bool, error) {
	a.log.Print(log.Trace)

	result, err := gqlcluster.RemoveCDMCluster(ctx, a.client.GQL, clusterUUID, expireInDays, isForce)
	if err != nil {
		return false, fmt.Errorf("failed to remove CDM cluster: %s", err)
	}

	return result, nil
}

// ClusterRemovalInfo contains information about a cluster's removal status and RCV locations.
type ClusterRemovalInfo struct {
	Prechecks            gqlcluster.ClusterRemovalPrechecks
	RCVLocations         []gqlcluster.RCVLocation
	BlockingConditions   bool
	ForceRemovalEligible bool
	ForceRemovable       bool
}

// RemoveCluster performs a comprehensive cluster removal operation
// It first checks if the cluster removal prechecks can be ignored, retrieves RCV locations,
// validates force removal eligibility, and then removes the cluster.
//
// The expireInDays parameter specifies the number of days before the cluster data expires.
// If nil, the default expiration is used.
//
// The isForce parameter requests force removal. Force removal is ONLY available when ALL of these are true:
//   - isDisconnected == true
//   - hasBlockingConditions == true (has replication SLAs, global SLAs, or RCV locations)
//   - ignorePrecheckTime != nil
//   - isAirGapped == false
//   - canIgnorePrecheck == true (cluster disconnected for ≥7 days)
//
// Air-gapped clusters are NEVER eligible for force removal - they must resolve blocking conditions manually.
//
// Returns ClusterRemovalInfo containing precheck results, RCV locations, and eligibility flags,
// along with a boolean indicating if the removal was successful.
func (a API) RemoveCluster(ctx context.Context, clusterUUID uuid.UUID, expireInDays *int64, forceRemoval bool) (ClusterRemovalInfo, bool, error) {
	a.log.Print(log.Trace)

	var info ClusterRemovalInfo

	// Check cluster removal prechecks
	prechecks, err := a.CanIgnoreClusterRemovalPrechecks(ctx, clusterUUID)
	if err != nil {
		return info, false, fmt.Errorf("failed to check cluster removal prechecks: %s", err)
	}
	info.Prechecks = prechecks

	// Get RCV locations
	rcvLocations, err := a.RCVLocations(ctx, clusterUUID)
	if err != nil {
		return info, false, fmt.Errorf("failed to get RCV locations: %s", err)
	}
	info.RCVLocations = rcvLocations

	// Check for SLA replication (excluding archived SLAs)
	slaReplication, err := a.VerifySLAReplication(ctx, clusterUUID, false)
	if err != nil {
		return info, false, fmt.Errorf("failed to verify SLA replication: %s", err)
	}

	// Get global SLAs
	globalSLAs, err := a.GlobalSLAs(ctx, clusterUUID)
	if err != nil {
		return info, false, fmt.Errorf("failed to get global SLAs: %s", err)
	}

	// Calculate blocking conditions
	// A cluster has blocking conditions if it has:
	// - Active SLAs with replication, OR
	// - Global SLAs, OR
	// - RCV locations
	info.BlockingConditions = slaReplication.IsActiveSLA || len(globalSLAs) > 0 || len(rcvLocations) > 0

	// Determine force removal eligibility (when force removal option is available)
	// Force removal is ONLY available when ALL of these are true:
	// - isDisconnected == true
	// - hasBlockingConditions == true
	// - ignorePrecheckTime != nil (not empty string)
	// - isAirGapped == false
	info.ForceRemovalEligible = prechecks.Disconnected &&
		info.BlockingConditions &&
		prechecks.IgnorePrecheckTime != "" &&
		!prechecks.AirGapped

	// Determine if force removal can actually be executed
	// Force removal can ONLY be executed when:
	// - All eligibility conditions are met, AND
	// - canIgnorePrecheck == true (cluster disconnected for ≥7 days)
	info.ForceRemovable = info.ForceRemovalEligible && prechecks.IgnorePrecheck

	// Validate force removal request
	if forceRemoval {
		// Air-gapped clusters are NEVER eligible for force removal
		if prechecks.AirGapped {
			return info, false, fmt.Errorf("force removal is not available for air-gapped clusters; blocking conditions must be resolved manually")
		}

		// Check if force removal is eligible
		if !info.ForceRemovalEligible {
			return info, false, fmt.Errorf("force removal is not available: cluster must be disconnected with blocking conditions")
		}

		// Check if force removal can be executed (cluster disconnected for ≥7 days)
		if !info.ForceRemovable {
			return info, false, fmt.Errorf("force removal cannot be executed: cluster has not been disconnected for at least 7 days")
		}
	}

	// Remove the CDM cluster
	success, err := a.RemoveCDMCluster(ctx, clusterUUID, expireInDays, forceRemoval)
	if err != nil {
		return info, false, fmt.Errorf("failed to remove cluster: %s", err)
	}

	return info, success, nil
}
