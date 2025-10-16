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
	GroupID       uuid.UUID       `json:"archivalGroupId"`
	Frequencies   []RetentionUnit `json:"frequencies"`
	Threshold     int             `json:"threshold"`
	ThresholdUnit RetentionUnit   `json:"thresholdUnit"`
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
	DayOfWeek DayOfWeek `json:"dayOfWeek"`
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
	AWSS3Config                     *AWSS3Config     `json:"awsNativeS3SlaConfigInput,omitempty"`
	AWSRDSConfig                    *AWSRDSConfig    `json:"awsRdsConfigInput,omitempty"`
	AzureBlobConfig                 *AzureBlobConfig `json:"azureBlobConfigInput,omitempty"`
	AzureSQLDatabaseDBConfig        *AzureDBConfig   `json:"azureSqlDatabaseDbConfigInput,omitempty"`
	AzureSQLManagedInstanceDBConfig *AzureDBConfig   `json:"azureSqlManagedInstanceDbConfigInput,omitempty"`
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
	DayOfWeek     Day                   `json:"dayOfWeek"`
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

// ReplicationSpec holds the replication specification for an RSC global SLA
// domain.
type ReplicationSpec struct {
	// AWSAccount is "SAME" or an AWS account id for cross account replication.
	AWSAccount string                       `json:"awsAccount,omitempty"`
	AWSRegion  aws.RegionForReplicationEnum `json:"awsRegion,omitempty,omitzero"`

	// AzureSubscription is "SAME" or an Azure subscription id for cross subscription replication.
	AzureSubscription string                         `json:"azureSubscription,omitempty"`
	AzureRegion       azure.RegionForReplicationEnum `json:"azureRegion,omitempty,omitzero"`

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
	ApplicableWorkloadType    string                    `json:"applicableWorkloadType,omitempty"`
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
