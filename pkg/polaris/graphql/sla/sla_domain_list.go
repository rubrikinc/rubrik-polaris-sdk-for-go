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
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Domain represents a global SLA domain.
type Domain struct {
	ArchivalSpecs []struct {
		Frequencies    []RetentionUnit `json:"frequencies"`
		Threshold      int             `json:"threshold"`
		ThresholdUnit  RetentionUnit   `json:"thresholdUnit"`
		StorageSetting struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"storageSetting"`
		ArchivalLocationToClusterMapping []struct {
			Cluster struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"cluster"`
			Location struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"location"`
		} `json:"archivalLocationToClusterMapping"`
		ArchivalTieringSpec *struct {
			InstantTiering                 bool             `json:"isInstantTieringEnabled"`
			MinAccessibleDurationInSeconds int64            `json:"minAccessibleDurationInSeconds"`
			ColdStorageClass               ColdStorageClass `json:"coldStorageClass"`
			TierExistingSnapshots          bool             `json:"shouldTierExistingSnapshots"`
		} `json:"archivalTieringSpec"`
	} `json:"archivalSpecs"`
	Archived            bool `json:"isArchived"`
	BackupLocationSpecs []struct {
		ArchivalGroup struct {
			ID string `json:"id"`
		} `json:"archivalGroup"`
	} `json:"backupLocationSpecs"`
	BackupWindows          []BackupWindow `json:"backupWindows"`
	Description            string         `json:"description"`
	FirstFullBackupWindows []BackupWindow `json:"firstFullBackupWindows"`
	ID                     uuid.UUID      `json:"id"`
	LocalRetentionLimit    *struct {
		Duration int           `json:"duration"`
		Unit     RetentionUnit `json:"unit"`
	} `json:"localRetentionLimit"`
	Name                  string `json:"name"`
	ObjectSpecificConfigs struct {
		AWSDynamoDBConfig *AWSDynamoDBConfig `json:"awsNativeDynamoDbSlaConfig"`
		AWSS3Config       *struct {
			AWSS3Config
			ArchivalLocationName string `json:"archivalLocationName"`
		} `json:"awsNativeS3SlaConfig"`
		AWSRDSConfig    *AWSRDSConfig `json:"awsRdsConfig"`
		AzureBlobConfig *struct {
			AzureBlobConfig
			BackupLocationName string `json:"backupLocationName"`
		} `json:"azureBlobConfig"`
		AzureSQLDatabaseDBConfig        *AzureDBConfig              `json:"azureSqlDatabaseDbConfig"`
		AzureSQLManagedInstanceDBConfig *AzureDBConfig              `json:"azureSqlManagedInstanceDbConfig"`
		VMwareVMConfig                  *VMwareVMConfig             `json:"vmwareVmConfig"`
		SapHanaConfig                   *SapHanaConfig              `json:"sapHanaConfig"`
		DB2Config                       *DB2Config                  `json:"db2Config"`
		MssqlConfig                     *MssqlConfig                `json:"mssqlConfig"`
		OracleConfig                    *OracleConfig               `json:"oracleConfig"`
		MongoConfig                     *MongoConfig                `json:"mongoConfig"`
		ManagedVolumeSlaConfig          *ManagedVolumeSlaConfig     `json:"managedVolumeSlaConfig"`
		PostgresDbClusterSlaConfig      *PostgresDbClusterSlaConfig `json:"postgresDbClusterSlaConfig"`
		MysqldbSlaConfig                *MysqldbSlaConfig           `json:"mysqldbSlaConfig"`
		NcdSlaConfig                    *NcdSlaConfig               `json:"ncdSlaConfig"`
		InformixSlaConfig               *InformixSlaConfig          `json:"informixSlaConfig"`
		GcpCloudSqlConfig               *GcpCloudSqlConfig          `json:"gcpCloudSqlConfig"`
	} `json:"objectSpecificConfigs"`
	ObjectTypes      []ObjectType `json:"objectTypes"`
	ReplicationSpecs []struct {
		AWSRegion aws.RegionForReplicationEnum `json:"awsRegion"`
		AWS       struct {
			AccountID string                       `json:"accountId"`
			Region    aws.RegionForReplicationEnum `json:"region"`
		} `json:"awsTarget"`
		AzureRegion azure.RegionForReplicationEnum `json:"azureRegion"`
		Azure       struct {
			SubscriptionID string                         `json:"subscriptionId"`
			Region         azure.RegionForReplicationEnum `json:"region"`
		} `json:"azureTarget"`
		ReplicationPairs []struct {
			SourceCluster struct {
				ID string `json:"id"`
			} `json:"sourceCluster"`
			TargetCluster struct {
				ID string `json:"id"`
			} `json:"targetCluster"`
		} `json:"replicationPairs"`

		RetentionDuration                 RetentionDuration  `json:"retentionDuration"`
		ReplicationLocalRetentionDuration *RetentionDuration `json:"replicationLocalRetentionDuration"`
		CascadingArchivalSpecs            []struct {
			ArchivalThreshold   *RetentionDuration `json:"archivalThreshold"`
			ArchivalTieringSpec *struct {
				InstantTiering                 bool             `json:"isInstantTieringEnabled"`
				MinAccessibleDurationInSeconds int64            `json:"minAccessibleDurationInSeconds"`
				ColdStorageClass               ColdStorageClass `json:"coldStorageClass"`
				TierExistingSnapshots          bool             `json:"shouldTierExistingSnapshots"`
			} `json:"archivalTieringSpec"`
			Frequencies      []RetentionUnit `json:"frequency"`
			ArchivalLocation *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"archivalLocation"`
			ArchivalLocationToClusterMapping []struct {
				Cluster struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"cluster"`
				Location struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"location"`
			} `json:"archivalLocationToClusterMapping"`
		} `json:"cascadingArchivalSpecs"`
	} `json:"replicationSpecsV2"`
	RetentionLock     bool              `json:"isRetentionLockedSla"`
	RetentionLockMode RetentionLockMode `json:"retentionLockMode"`
	SnapshotSchedule  SnapshotSchedule  `json:"snapshotSchedule"`
	Version           string            `json:"version"`
}

// DomainByID returns the global SLA domain with the specified ID.
func DomainByID(ctx context.Context, gql *graphql.Client, id uuid.UUID) (Domain, error) {
	gql.Log().Print(log.Trace)

	domain, err := domainByID(ctx, gql, id)
	if err == nil {
		return domain, nil
	}

	// Fallback to listing all domains if querying by ID fails with a 403 error.
	// RSC returns a 403 error when the SLA domain isn't found.
	var gqlErr graphql.GQLError
	if !errors.As(err, &gqlErr) || gqlErr.Code() != 403 {
		return Domain{}, err
	}

	domains, err := ListDomains(ctx, gql, nil)
	if err != nil {
		return Domain{}, err
	}
	for _, domain := range domains {
		if domain.ID == id {
			return domain, nil
		}
	}

	return Domain{}, fmt.Errorf("global SLA domain %q %w", id, graphql.ErrNotFound)
}

func domainByID(ctx context.Context, gql *graphql.Client, id uuid.UUID) (Domain, error) {
	query := slaDomainQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"slaDomainId"`
	}{ID: id})
	if err != nil {
		return Domain{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result Domain `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return Domain{}, graphql.UnmarshalError(query, err)
	}
	if payload.Data.Result.Archived {
		return Domain{}, graphql.ErrNotFound
	}

	return payload.Data.Result, nil
}

// DomainFilter holds the filter parameters for an SLA domain list operation.
type DomainFilter struct {
	Field string `json:"field"`
	Value string `json:"text"`
}

// ListDomains returns all global SLA domains matching the specified SLA domain
// filters.
func ListDomains(ctx context.Context, gql *graphql.Client, filters []DomainFilter) ([]Domain, error) {
	gql.Log().Print(log.Trace)

	var cursor string
	var nodes []Domain
	for {
		query := slaDomainsQuery
		buf, err := gql.Request(ctx, query, struct {
			After  string         `json:"after,omitempty"`
			Filter []DomainFilter `json:"filter,omitempty"`
		}{After: cursor, Filter: filters})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []Domain `json:"nodes"`
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

// Object represents an object protected by an RSC global SLA domain.
type Object struct {
	ID                uuid.UUID        `json:"id"`
	Name              string           `json:"name"`
	ObjectType        string           `json:"objectType"`
	EffectiveDomainID string           `json:"effectiveSla"`
	ProtectionStatus  ProtectionStatus `json:"protectionStatus"`
}

// ObjectFilter holds the filter parameters for a protected object list
// operation.
type ObjectFilter struct {
	ObjectName                  string           `json:"objectName"`
	ProtectionStatus            ProtectionStatus `json:"protectionStatus"`
	OnlyDirectlyAssignedObjects bool             `json:"showOnlyDirectlyAssignedObjects"`
}

// ListDomainObjects returns all objects protected by the specified global SLA
// domain.
func ListDomainObjects(ctx context.Context, gql *graphql.Client, slaID uuid.UUID, filter ObjectFilter) ([]Object, error) {
	gql.Log().Print(log.Trace)

	var cursor string
	var nodes []Object
	for {
		query := slaProtectedObjectsQuery
		buf, err := gql.Request(ctx, query, struct {
			DomainID []uuid.UUID  `json:"slaIds"`
			After    string       `json:"after,omitempty"`
			Filter   ObjectFilter `json:"filter"`
		}{DomainID: []uuid.UUID{slaID}, After: cursor, Filter: filter})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []Object `json:"nodes"`
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
