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
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CreateDomainParams holds the parameters for a global SLA domain create
// operation.
type CreateDomainParams struct {
	ArchivalSpecs       []ArchivalSpec       `json:"archivalSpecs,omitempty"`
	BackupLocationSpecs []BackupLocationSpec `json:"backupLocationSpecs,omitempty"`
	BackupWindows       []BackupWindow       `json:"backupWindows,omitempty"`
	Description         string               `json:"description,omitempty"`

	// If omitted, it will be done at first opportunity.
	FirstFullBackupWindows []BackupWindow `json:"firstFullBackupWindows,omitempty"`

	LocalRetentionLimit   *RetentionDuration     `json:"localRetentionLimit,omitempty"`
	Name                  string                 `json:"name"`
	ObjectSpecificConfigs *ObjectSpecificConfigs `json:"objectSpecificConfigsInput,omitempty"`
	ObjectTypes           []ObjectType           `json:"objectTypes"`
	RetentionLock         bool                   `json:"isRetentionLockedSla"`
	RetentionLockMode     RetentionLockMode      `json:"retentionLockMode,omitempty"`
	ReplicationSpecs      []ReplicationSpec      `json:"replicationSpecsV2,omitempty"`
	SnapshotSchedule      SnapshotSchedule       `json:"snapshotSchedule"`
}

// ArchivalSpec holds the archival specification for an RSC global SLA domain.
type ArchivalSpec struct {
	GroupID                          uuid.UUID                          `json:"archivalGroupId,omitzero"`
	Frequencies                      []RetentionUnit                    `json:"frequencies"`
	Threshold                        int                                `json:"threshold"`
	ThresholdUnit                    RetentionUnit                      `json:"thresholdUnit"`
	ArchivalLocationToClusterMapping []ArchivalLocationToClusterMapping `json:"archivalLocationToClusterMapping,omitempty"`
	ArchivalTieringSpec              *ArchivalTieringSpec               `json:"archivalTieringSpecInput,omitempty"`
}

// BackupLocationSpec holds the backup location specification for an RSC global
// SLA domain.
type BackupLocationSpec struct {
	ArchivalGroupID uuid.UUID `json:"archivalGroupId"`
}

// BackupWindow represents a backup window for an RSC global SLA domain.
type BackupWindow struct {
	DurationInHours int       `json:"durationInHours"`
	StartTime       StartTime `json:"startTimeAttributes"`
}

// StartTime represents the start time for a backup window.
type StartTime struct {
	DayOfWeek DayOfWeek `json:"dayOfWeek,omitzero"`
	Hour      int       `json:"hour"`
	Minute    int       `json:"minute"`
}

// DayOfWeek represents a day of the week.
type DayOfWeek struct {
	Day Day `json:"day"`
}

// ObjectSpecificConfigs holds the object-specific configurations for a global
// RSC SLA domain.
type ObjectSpecificConfigs struct {
	AWSDynamoDBConfig               *AWSDynamoDBConfig          `json:"awsNativeDynamoDbSlaConfigInput,omitempty"`
	AWSS3Config                     *AWSS3Config                `json:"awsNativeS3SlaConfigInput,omitempty"`
	AWSRDSConfig                    *AWSRDSConfig               `json:"awsRdsConfigInput,omitempty"`
	AzureBlobConfig                 *AzureBlobConfig            `json:"azureBlobConfigInput,omitempty"`
	AzureSQLDatabaseDBConfig        *AzureDBConfig              `json:"azureSqlDatabaseDbConfigInput,omitempty"`
	AzureSQLManagedInstanceDBConfig *AzureDBConfig              `json:"azureSqlManagedInstanceDbConfigInput,omitempty"`
	VMwareVMConfig                  *VMwareVMConfig             `json:"vmwareVmConfigInput,omitempty"`
	SapHanaConfig                   *SapHanaConfig              `json:"sapHanaConfigInput,omitempty"`
	DB2Config                       *DB2Config                  `json:"db2ConfigInput,omitempty"`
	MssqlConfig                     *MssqlConfig                `json:"mssqlConfigInput,omitempty"`
	OracleConfig                    *OracleConfig               `json:"oracleConfigInput,omitempty"`
	MongoConfig                     *MongoConfig                `json:"mongoConfigInput,omitempty"`
	ManagedVolumeSlaConfig          *ManagedVolumeSlaConfig     `json:"managedVolumeSlaConfigInput,omitempty"`
	PostgresDbClusterSlaConfig      *PostgresDbClusterSlaConfig `json:"postgresDbClusterSlaConfigInput,omitempty"`
	MysqldbSlaConfig                *MysqldbSlaConfig           `json:"mysqldbConfigInput,omitempty"`
	NcdSlaConfig                    *NcdSlaConfig               `json:"ncdConfigInput,omitempty"`
	InformixSlaConfig               *InformixSlaConfig          `json:"informixConfigInput,omitempty"`
	GcpCloudSqlConfig               *GcpCloudSqlConfig          `json:"gcpCloudSqlConfigInput,omitempty"`
}

// AWSDynamoDBConfig represents the configuration specific for an AWS DynamoDB
// object.
type AWSDynamoDBConfig struct {
	KMSAliasForPrimaryBackup string `json:"cmkAliasForPrimaryBackup"`
}

// AWSS3Config represents the configuration specific for an AWS S3 object.
type AWSS3Config struct {
	ArchivalLocationID              uuid.UUID `json:"archivalLocationId"`
	ContinuousBackupRetentionInDays int       `json:"continuousBackupRetentionInDays"`
}

// AzureBlobConfig represents the configuration specific for an Azure blob
// object.
type AzureBlobConfig struct {
	BackupLocationID                uuid.UUID `json:"backupLocationId"`
	ContinuousBackupRetentionInDays int       `json:"continuousBackupRetentionInDays"`
}

// AWSRDSConfig represents the configuration specific for an AWS RDS object.
type AWSRDSConfig struct {
	LogRetention RetentionDuration `json:"logRetention"`
}

// AzureDBConfig represents the configuration specific for an Azure database
// object.
type AzureDBConfig struct {
	LogRetentionInDays int `json:"logRetentionInDays"`
}

// VMwareVMConfig represents the configuration specific for a VMware vSphere VM
// object.
type VMwareVMConfig struct {
	LogRetentionSeconds int64 `json:"logRetentionSeconds,omitempty"`
}

// SapHanaConfig represents the configuration specific for a SAP HANA database
// object.
type SapHanaConfig struct {
	IncrementalFrequency  RetentionDuration             `json:"incrementalFrequency,omitempty"`
	LogRetention          RetentionDuration             `json:"logRetention,omitempty"`
	DifferentialFrequency RetentionDuration             `json:"differentialFrequency,omitempty"`
	StorageSnapshotConfig *SapHanaStorageSnapshotConfig `json:"storageSnapshotConfig,omitempty"`
}

// SapHanaStorageSnapshotConfig represents the storage snapshot configuration
// for SAP HANA.
type SapHanaStorageSnapshotConfig struct {
	Frequency RetentionDuration `json:"frequency,omitempty"`
	Retention RetentionDuration `json:"retention,omitempty"`
}

// DB2Config represents the configuration specific for a Db2 database object.
type DB2Config struct {
	IncrementalFrequency  RetentionDuration    `json:"incrementalFrequency,omitempty"`
	LogRetention          RetentionDuration    `json:"logRetention,omitempty"`
	DifferentialFrequency RetentionDuration    `json:"differentialFrequency,omitempty"`
	LogArchivalMethod     Db2LogArchivalMethod `json:"logArchivalMethod,omitempty"`
}

// Db2LogArchivalMethod represents the log archival method for Db2 database.
type Db2LogArchivalMethod string

const (
	// Db2LogArchivalMethod1 represents log archival method 1 for Db2.
	Db2LogArchivalMethod1 Db2LogArchivalMethod = "LOGARCHMETH1"
	// Db2LogArchivalMethod2 represents log archival method 2 for Db2.
	Db2LogArchivalMethod2 Db2LogArchivalMethod = "LOGARCHMETH2"
)

// MssqlConfig represents the configuration specific for a SQL Server database
// object.
type MssqlConfig struct {
	Frequency    RetentionDuration `json:"frequency,omitempty"`
	LogRetention RetentionDuration `json:"logRetention,omitempty"`
}

// OracleConfig represents the configuration specific for an Oracle database
// object.
type OracleConfig struct {
	Frequency        RetentionDuration `json:"frequency,omitempty"`
	LogRetention     RetentionDuration `json:"logRetention,omitempty"`
	HostLogRetention RetentionDuration `json:"hostLogRetention,omitempty"`
}

// MongoConfig represents the configuration specific for a MongoDB database
// object.
type MongoConfig struct {
	LogFrequency RetentionDuration `json:"logFrequency,omitempty"`
	LogRetention RetentionDuration `json:"logRetention,omitempty"`
}

// ManagedVolumeSlaConfig represents the configuration specific for a Managed
// Volume object.
type ManagedVolumeSlaConfig struct {
	LogRetention RetentionDuration `json:"logRetention,omitempty"`
}

// PostgresDbClusterSlaConfig represents the configuration specific for a
// Postgres DB Cluster object.
type PostgresDbClusterSlaConfig struct {
	LogRetention RetentionDuration `json:"logRetention,omitempty"`
}

// MysqldbSlaConfig represents the configuration specific for a MySQL object.
type MysqldbSlaConfig struct {
	LogFrequency RetentionDuration `json:"logFrequency,omitempty"`
	LogRetention RetentionDuration `json:"logRetention,omitempty"`
}

// NcdSlaConfig represents the configuration specific for a NAS Cloud Direct
// object.
type NcdSlaConfig struct {
	MinutelyBackupLocations  []uuid.UUID `json:"minutelyBackupLocations,omitempty"`
	HourlyBackupLocations    []uuid.UUID `json:"hourlyBackupLocations,omitempty"`
	DailyBackupLocations     []uuid.UUID `json:"dailyBackupLocations,omitempty"`
	WeeklyBackupLocations    []uuid.UUID `json:"weeklyBackupLocations,omitempty"`
	MonthlyBackupLocations   []uuid.UUID `json:"monthlyBackupLocations,omitempty"`
	QuarterlyBackupLocations []uuid.UUID `json:"quarterlyBackupLocations,omitempty"`
	YearlyBackupLocations    []uuid.UUID `json:"yearlyBackupLocations,omitempty"`
}

// InformixSlaConfig represents the configuration specific for an Informix
// object.
type InformixSlaConfig struct {
	IncrementalFrequency RetentionDuration `json:"incrementalFrequency,omitempty"`
	IncrementalRetention RetentionDuration `json:"incrementalRetention,omitempty"`
	LogFrequency         RetentionDuration `json:"logFrequency,omitempty"`
	LogRetention         RetentionDuration `json:"logRetention,omitempty"`
}

// GcpCloudSqlConfig represents the configuration specific for a GCP Cloud SQL
// object.
type GcpCloudSqlConfig struct {
	LogRetention RetentionDuration `json:"logRetention,omitempty"`
}

// SnapshotSchedule holds the snapshot schedule for an RSC global SLA domain.
type SnapshotSchedule struct {
	Daily     *DailySnapshotSchedule     `json:"daily,omitempty"`
	Hourly    *HourlySnapshotSchedule    `json:"hourly,omitempty"`
	Minute    *MinuteSnapshotSchedule    `json:"minute,omitempty"`
	Monthly   *MonthlySnapshotSchedule   `json:"monthly,omitempty"`
	Quarterly *QuarterlySnapshotSchedule `json:"quarterly,omitempty"`
	Weekly    *WeeklySnapshotSchedule    `json:"weekly,omitempty"`
	Yearly    *YearlySnapshotSchedule    `json:"yearly,omitempty"`
}

// DailySnapshotSchedule holds the snapshot schedule for a daily snapshot.
type DailySnapshotSchedule struct {
	BasicSchedule BasicSnapshotSchedule `json:"basicSchedule"`
}

// HourlySnapshotSchedule holds the snapshot schedule for an hourly snapshot.
type HourlySnapshotSchedule struct {
	BasicSchedule BasicSnapshotSchedule `json:"basicSchedule"`
}

// MinuteSnapshotSchedule holds the snapshot schedule for a minutely snapshot.
type MinuteSnapshotSchedule struct {
	BasicSchedule BasicSnapshotSchedule `json:"basicSchedule"`
}

// MonthlySnapshotSchedule holds the snapshot schedule for a monthly snapshot.
type MonthlySnapshotSchedule struct {
	BasicSchedule BasicSnapshotSchedule `json:"basicSchedule"`
	DayOfMonth    DayOfMonth            `json:"dayOfMonth"`
}

// QuarterlySnapshotSchedule holds the snapshot schedule for a quarterly
// snapshot.
type QuarterlySnapshotSchedule struct {
	BasicSchedule     BasicSnapshotSchedule `json:"basicSchedule"`
	DayOfQuarter      DayOfQuarter          `json:"dayOfQuarter"`
	QuarterStartMonth Month                 `json:"quarterStartMonth"`
}

// WeeklySnapshotSchedule holds the snapshot schedule for a weekly snapshot.
type WeeklySnapshotSchedule struct {
	BasicSchedule BasicSnapshotSchedule `json:"basicSchedule"`
	DayOfWeek     Day                   `json:"dayOfWeek,omitempty"`
}

// YearlySnapshotSchedule holds the snapshot schedule for a yearly snapshot.
type YearlySnapshotSchedule struct {
	BasicSchedule  BasicSnapshotSchedule `json:"basicSchedule"`
	DayOfYear      DayOfYear             `json:"dayOfYear"`
	YearStartMonth Month                 `json:"yearStartMonth"`
}

// BasicSnapshotSchedule represents a basic RSC snapshot schedule.
type BasicSnapshotSchedule struct {
	Frequency     int           `json:"frequency"`
	Retention     int           `json:"retention"`
	RetentionUnit RetentionUnit `json:"retentionUnit"`
}

// ReplicationPair holds the source and target cluster IDs for replication.
type ReplicationPair struct {
	SourceClusterID string `json:"sourceClusterUuid,omitempty"`
	TargetClusterID string `json:"targetClusterUuid,omitempty"`
}

// ColdStorageClass represents the cold storage class for archival tiering.
type ColdStorageClass string

const (
	// ColdStorageClassUnknown represents an unknown cold storage class.
	ColdStorageClassUnknown ColdStorageClass = "COLD_STORAGE_CLASS_UNKNOWN"
	// ColdStorageClassAzureArchive represents Azure Archive cold storage tier.
	ColdStorageClassAzureArchive ColdStorageClass = "AZURE_ARCHIVE"
	// ColdStorageClassAWSGlacier represents AWS Glacier cold storage class.
	ColdStorageClassAWSGlacier ColdStorageClass = "AWS_GLACIER"
	// ColdStorageClassAWSGlacierDeepArchive represents AWS Glacier Deep Archive cold storage class.
	ColdStorageClassAWSGlacierDeepArchive ColdStorageClass = "AWS_GLACIER_DEEP_ARCHIVE"
)

// ArchivalTieringSpec holds the archival tiering specification.
type ArchivalTieringSpec struct {
	InstantTiering                 bool             `json:"isInstantTieringEnabled,omitempty"`
	MinAccessibleDurationInSeconds int64            `json:"minAccessibleDurationInSeconds,omitempty"`
	ColdStorageClass               ColdStorageClass `json:"coldStorageClass,omitempty"`
	TierExistingSnapshots          bool             `json:"shouldTierExistingSnapshots,omitempty"`
}

// ArchivalLocationToClusterMapping holds the mapping between archival location
// and Rubrik cluster.
type ArchivalLocationToClusterMapping struct {
	ClusterID  uuid.UUID `json:"clusterUuid"`
	LocationID uuid.UUID `json:"locationId"`
}

// CascadingArchivalSpec holds the cascading archival specification for
// replication.
type CascadingArchivalSpec struct {
	// Deprecated: use ArchivalLocationToClusterMappings instead.
	ArchivalLocationID                *uuid.UUID                         `json:"archivalLocationId,omitempty"`
	ArchivalThreshold                 *RetentionDuration                 `json:"archivalThreshold,omitempty"`
	ArchivalTieringSpec               *ArchivalTieringSpec               `json:"archivalTieringSpecInput,omitempty"`
	Frequencies                       []RetentionUnit                    `json:"frequency,omitempty"`
	ArchivalLocationToClusterMappings []ArchivalLocationToClusterMapping `json:"archivalLocationToClusterMapping,omitempty"`
}

// ReplicationSpec holds the replication specification for an RSC global SLA
// domain.
type ReplicationSpec struct {
	// AWSAccount is "SAME" or an AWS account id for cross account replication.
	AWSAccount string                       `json:"awsAccount,omitempty"`
	AWSRegion  aws.RegionForReplicationEnum `json:"awsRegion,omitempty,omitzero"`

	// AzureSubscription is "SAME" or an Azure subscription id for cross subscription replication.
	AzureSubscription string                         `json:"azureSubscription,omitempty"`
	AzureRegion       azure.RegionForReplicationEnum `json:"azureRegion,omitempty,omitzero"`

	ReplicationLocalRetentionDuration *RetentionDuration      `json:"replicationLocalRetentionDuration,omitempty"`
	CascadingArchivalSpecs            []CascadingArchivalSpec `json:"cascadingArchivalSpecs,omitempty"`

	ReplicationPairs  []ReplicationPair  `json:"replicationPairs,omitempty"`
	RetentionDuration *RetentionDuration `json:"retentionDuration,omitempty"`
}

// CreateDomain creates a new global SLA domain. Returns the ID of the
// new global SLA domain.
func CreateDomain(ctx context.Context, gql *graphql.Client, params CreateDomainParams) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	query := createGlobalSlaQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				ID string `json:"id"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	id, err := uuid.Parse(payload.Data.Result.ID)
	if err != nil {
		return uuid.Nil, graphql.ResponseError(query, err)
	}

	return id, nil
}

// UpdateDomainParams holds the parameters for an RSC global SLA domain update
// operation.
type UpdateDomainParams struct {
	ID                              uuid.UUID  `json:"id"`
	ShouldApplyToExistingSnapshots  *BoolValue `json:"shouldApplyToExistingSnapshots,omitempty"`
	ShouldApplyToNonPolicySnapshots *BoolValue `json:"shouldApplyToNonPolicySnapshots,omitempty"`
	CreateDomainParams
}

// BoolValue represents a boolean value.
type BoolValue struct {
	Value bool `json:"value"`
}

// UpdateDomain updates an existing global SLA domain.
func UpdateDomain(ctx context.Context, gql *graphql.Client, params UpdateDomainParams) error {
	gql.Log().Print(log.Trace)

	query := updateGlobalSlaQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				ID string `json:"id"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}

// DeleteDomain deletes the global SLA domain with the specified ID.
func DeleteDomain(ctx context.Context, gql *graphql.Client, slaID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deleteGlobalSlaQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"id"`
	}{ID: slaID})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"success"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result.Success {
		return graphql.ResponseError(query, fmt.Errorf("failed to delete SLA domain: %q", slaID))
	}

	return nil
}

// AssignDomainParams holds the parameters for an RSC global SLA domain
// assignment operation.
type AssignDomainParams struct {
	DomainID                  *uuid.UUID                `json:"slaOptionalId,omitempty"`
	DomainAssignType          AssignmentType            `json:"slaDomainAssignType"`
	ObjectIDs                 []uuid.UUID               `json:"objectIds"`
	ApplicableWorkloadType    hierarchy.Workload        `json:"applicableWorkloadType,omitempty"`
	ApplyToExistingSnapshots  *bool                     `json:"shouldApplyToExistingSnapshots,omitempty"`
	ApplyToNonPolicySnapshots *bool                     `json:"shouldApplyToNonPolicySnapshots,omitempty"`
	ExistingSnapshotRetention ExistingSnapshotRetention `json:"existingSnapshotRetention,omitempty"`
}

// AssignDomain assigns the specified RSC global SLA domain to the specified
// objects.
func AssignDomain(ctx context.Context, gql *graphql.Client, params AssignDomainParams) error {
	gql.Log().Print(log.Trace)

	query := assignSlaQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"success"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result.Success {
		return graphql.ResponseError(query, fmt.Errorf("failed to assign SLA domain to objects: %s", params.ObjectIDs))
	}

	return nil
}
