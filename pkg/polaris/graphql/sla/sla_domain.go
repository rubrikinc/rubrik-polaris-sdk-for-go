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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CreateGlobalSLAParams holds the parameters for a global SLA domain create
// operation.
type CreateGlobalSLAParams struct {
	ArchivalSpecs []ArchivalSpec `json:"archivalSpecs,omitempty"`
	BackupWindows []BackupWindow `json:"backupWindows,omitempty"`
	Description   string         `json:"description,omitempty"`

	// If omitted, it will be done at first opportunity.
	FirstFullBackupWindows []BackupWindow `json:"firstFullBackupWindows,omitempty"`

	LocalRetentionLimit   *RetentionDuration     `json:"localRetentionLimit,omitempty"`
	Name                  string                 `json:"name"`
	ObjectSpecificConfigs *ObjectSpecificConfigs `json:"objectSpecificConfigsInput,omitempty"`
	ObjectTypes           []SLAObjectType        `json:"objectTypes"`
	RetentionLock         bool                   `json:"isRetentionLockedSla"`
	RetentionLockMode     RetentionLockMode      `json:"retentionLockMode,omitempty"`
	SnapshotSchedule      GlobalSnapshotSchedule `json:"snapshotSchedule"`
}

// ArchivalSpec holds the archival specification for an RSC global SLA domain.
type ArchivalSpec struct {
	GroupID       uuid.UUID       `json:"archivalGroupId"`
	Frequencies   []RetentionUnit `json:"frequencies"`
	Threshold     int             `json:"threshold"`
	ThresholdUnit RetentionUnit   `json:"thresholdUnit"`
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
	AWSS3Config                     *CloudNativeObjectConfig `json:"awsNativeS3SlaConfig,omitempty"`
	AWSRDSConfig                    *AWSRDSConfig            `json:"awsRdsConfig,omitempty"`
	AzureBlobConfig                 *CloudNativeObjectConfig `json:"azureBlobConfig,omitempty"`
	AzureSQLDatabaseDBConfig        *AzureDBConfig           `json:"azureSqlDatabaseDbConfig,omitempty"`
	AzureSQLManagedInstanceDBConfig *AzureDBConfig           `json:"azureSqlManagedInstanceDbConfig,omitempty"`
}

// CloudNativeObjectConfig represents the configuration specific for cloud
// native objects.
type CloudNativeObjectConfig struct {
	ArchivalLocationID              uuid.UUID `json:"archivalLocationId"`
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

// GlobalSnapshotSchedule holds the snapshot schedule for an RSC global SLA
// domain.
type GlobalSnapshotSchedule struct {
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

// CreateGlobalSLADomain creates a new global SLA domain. Returns the ID of the
// new global SLA domain.
func CreateGlobalSLADomain(ctx context.Context, gql *graphql.Client, createParams CreateGlobalSLAParams) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	query := createGlobalSlaQuery
	buf, err := gql.Request(ctx, query, createParams)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

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

// UpdateGlobalSLAParams holds the parameters for an RSC global SLA domain
// update operation.
type UpdateGlobalSLAParams struct {
	ID                              uuid.UUID  `json:"id"`
	ShouldApplyToExistingSnapshots  *BoolValue `json:"shouldApplyToExistingSnapshots,omitempty"`
	ShouldApplyToNonPolicySnapshots *BoolValue `json:"shouldApplyToNonPolicySnapshots,omitempty"`
	CreateGlobalSLAParams
}

// BoolValue represents a boolean value.
type BoolValue struct {
	Value bool `json:"value"`
}

// UpdateGlobalSLADomain updates an existing global SLA domain.
func UpdateGlobalSLADomain(ctx context.Context, gql *graphql.Client, updateParams UpdateGlobalSLAParams) error {
	gql.Log().Print(log.Trace)

	query := updateGlobalSlaQuery
	buf, err := gql.Request(ctx, query, updateParams)
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result struct{} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}

// DeleteGlobalSLADomain deletes the global SLA domain with the specified ID.
func DeleteGlobalSLADomain(ctx context.Context, gql *graphql.Client, slaID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deleteGlobalSlaQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"id"`
	}{ID: slaID})
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

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

// AssignSLAParams holds the parameters for an RSC global SLA domain assignment
// operation.
type AssignSLAParams struct {
	SLAID                  *uuid.UUID          `json:"slaOptionalId,omitempty"`
	SLADomainAssignType    SLADomainAssignType `json:"slaDomainAssignType"`
	ObjectIDs              []uuid.UUID         `json:"objectIds"`
	ApplicableWorkloadType string              `json:"applicableWorkloadType,omitempty"`
}

// AssignSLADomain assigns the specified RSC global SLA domain to the specified
// objects.
func AssignSLADomain(ctx context.Context, gql *graphql.Client, assignParams AssignSLAParams) error {
	gql.Log().Print(log.Trace)

	query := assignSlaQuery
	buf, err := gql.Request(ctx, query, assignParams)
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

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
		return graphql.ResponseError(query, fmt.Errorf("failed to assign SLA domain to objects: %s", assignParams.ObjectIDs))
	}

	return nil
}
