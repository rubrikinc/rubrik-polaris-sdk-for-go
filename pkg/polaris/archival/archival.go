// Copyright 2024 Rubrik, Inc.
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

// Package archival provides high level interfaces when working with archival
// locations for SLAs.
package archival

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for archival and storage management.
type API struct {
	client *graphql.Client
	log    log.Logger
}

// Wrap the RSC client in the archival API.
func Wrap(client *polaris.Client) API {
	return API{client: client.GQL, log: client.GQL.Log()}
}

// DeleteTarget deletes the target with the specified target ID. Data center
// archival locations are also referred to as targets.
func (a API) DeleteTarget(ctx context.Context, targetID uuid.UUID) error {
	a.log.Print(log.Trace)

	if err := archival.DisableTarget(ctx, a.client, targetID); err != nil {
		return fmt.Errorf("failed to disable target: %s", err)
	}

	if err := a.waitForTargetStatus(ctx, targetID, archival.TargetDisabled); err != nil {
		return fmt.Errorf("failed to wait for target to become disabled: %s", err)
	}

	if err := archival.DeleteTarget(ctx, a.client, targetID); err != nil {
		return fmt.Errorf("failed to delete target: %s", err)
	}

	if err := a.waitForTargetStatus(ctx, targetID, archival.TargetDeleted); err != nil {
		return fmt.Errorf("failed to wait for target to become deleted: %s", err)
	}

	return nil
}

// DeleteTargetMapping deletes the target mapping with the specified target
// mapping ID. Cloud archival locations are also referred to as target mappings.
func (a API) DeleteTargetMapping(ctx context.Context, targetMappingID uuid.UUID) error {
	a.log.Print(log.Trace)

	if err := archival.DeleteTargetMapping(ctx, a.client, targetMappingID); err != nil {
		return fmt.Errorf("failed to delete target mapping: %s", err)
	}

	return nil
}

// ClusterArchivalLocationByName returns archival locations for a specific
// cluster and name pattern for data center use case.
func (a API) ClusterArchivalLocationByName(ctx context.Context, clusterID uuid.UUID, name string) ([]archival.ArchivalLocation, error) {
	a.log.Print(log.Trace)

	filters := []archival.ListTargetFilter{
		{Field: "IS_MANAGED_BY_AUTO_AG", TextList: []string{"false"}},
		{Field: "STATUS", TextList: []string{"READ_WRITE", "PAUSED", "DISABLED"}},
		{Field: "ARCHIVAL_ENTITY_USE_CASE_TYPE", Text: "DATA_CENTER"},
		{Field: "CLUSTER_ID", TextList: []string{clusterID.String()}},
		{Field: "NAME", Text: name},
	}

	locations, err := archival.ListTargets[archival.ArchivalLocation](ctx, a.client, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list archival locations: %s", err)
	}

	return locations, nil
}

// waitForTargetStatus wait for the target's status to become the specified
// status.
func (a API) waitForTargetStatus(ctx context.Context, targetID uuid.UUID, status string) error {
	a.log.Print(log.Trace)

	for {
		target, err := archival.GetTarget[archival.Target](ctx, a.client, targetID)
		if err != nil {
			return fmt.Errorf("failed to get target: %s", err)
		}
		if target.Status == status {
			return nil
		}

		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
