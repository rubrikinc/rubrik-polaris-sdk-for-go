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

package sla

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// GlobalSLADomain represents an RSC global SLA domain.
type GlobalSLADomain struct {
	ArchivalSpecs []struct {
		Frequencies    []RetentionUnit `json:"frequencies"`
		Threshold      int             `json:"threshold"`
		ThresholdUnit  RetentionUnit   `json:"thresholdUnit"`
		StorageSetting struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"storageSetting"`
	} `json:"archivalSpecs"`
	BackupWindows          []BackupWindow `json:"backupWindows"`
	Description            string         `json:"description"`
	FirstFullBackupWindows []BackupWindow `json:"firstFullBackupWindows"`
	ID                     uuid.UUID      `json:"id"`
	LocalRetentionLimit    *struct {
		Duration int    `json:"duration"`
		Unit     string `json:"unit"`
	} `json:"localRetentionLimit"`
	Name                  string `json:"name"`
	ObjectSpecificConfigs struct {
		AWSS3Config *struct {
			CloudNativeObjectConfig
			ArchivalLocationName string `json:"archivalLocationName"`
		} `json:"awsNativeS3SlaConfig"`
		AWSRDSConfig    *AWSRDSConfig `json:"awsRdsConfig"`
		AzureBlobConfig *struct {
			CloudNativeObjectConfig
			ArchivalLocationName string `json:"archivalLocationName"`
		} `json:"azureBlobConfig"`
		AzureSQLDatabaseDBConfig        *AzureDBConfig `json:"azureSqlDatabaseDbConfig"`
		AzureSQLManagedInstanceDBConfig *AzureDBConfig `json:"azureSqlManagedInstanceDbConfig"`
	} `json:"objectSpecificConfigs"`
	ObjectTypes       []SLAObjectType        `json:"objectTypes"`
	RetentionLock     bool                   `json:"isRetentionLockedSla"`
	RetentionLockMode RetentionLockMode      `json:"retentionLockMode"`
	SnapshotSchedule  GlobalSnapshotSchedule `json:"snapshotSchedule"`
	Version           string                 `json:"version"`
}

// SLADomainFilter holds the filter parameters for an SLA domain list operation.
type SLADomainFilter struct {
	Field string `json:"field"`
	Value string `json:"text"`
}

// ListSLADomains returns all RSC global SLA domains matching the specified SLA
// domain filters.
func ListSLADomains(ctx context.Context, gql *graphql.Client, filters []SLADomainFilter) ([]GlobalSLADomain, error) {
	gql.Log().Print(log.Trace)

	var cursor string
	var nodes []GlobalSLADomain
	for {
		query := slaDomainsQuery
		buf, err := gql.Request(ctx, query, struct {
			After  string            `json:"after,omitempty"`
			Filter []SLADomainFilter `json:"filter,omitempty"`
		}{After: cursor, Filter: filters})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}
		graphql.LogResponse(gql.Log(), query, buf)

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []GlobalSLADomain `json:"nodes"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, graphql.UnmarshalError(query, err)
		}
		nodes = append(nodes, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return nodes, nil
}

// ProtectedObject represents an object protected by an RSC global SLA domain.
type ProtectedObject struct {
	ID                 uuid.UUID        `json:"id"`
	Name               string           `json:"name"`
	ObjectType         string           `json:"objectType"`
	EffectiveSLADomain string           `json:"effectiveSla"`
	ProtectionStatus   ProtectionStatus `json:"protectionStatus"`
}

// ProtectedObjectFilter holds the filter parameters for a protected object list
// operation.
type ProtectedObjectFilter struct {
	ObjectName                      string           `json:"objectName"`
	ProtectionStatus                ProtectionStatus `json:"protectionStatus"`
	ShowOnlyDirectlyAssignedObjects bool             `json:"showOnlyDirectlyAssignedObjects"`
}

// ListSLADomainProtectedObjects returns all objects protected by the specified
// RSC global SLA domain.
func ListSLADomainProtectedObjects(ctx context.Context, gql *graphql.Client, slaID uuid.UUID, filter ProtectedObjectFilter) ([]ProtectedObject, error) {
	gql.Log().Print(log.Trace)

	var cursor string
	var nodes []ProtectedObject
	for {
		query := slaProtectedObjectsQuery
		buf, err := gql.Request(ctx, query, struct {
			SLAID  []uuid.UUID           `json:"slaIds"`
			After  string                `json:"after,omitempty"`
			Filter ProtectedObjectFilter `json:"filter,omitempty"`
		}{SLAID: []uuid.UUID{slaID}, After: cursor, Filter: filter})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}
		graphql.LogResponse(gql.Log(), query, buf)

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []ProtectedObject `json:"nodes"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, graphql.UnmarshalError(query, err)
		}
		nodes = append(nodes, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return nodes, nil
}
