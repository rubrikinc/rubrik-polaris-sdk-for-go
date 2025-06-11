//go:generate go run ../queries_gen.go sla

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
	"fmt"
	"time"
)

const (
	Monday    Day = "MONDAY"
	Tuesday   Day = "TUESDAY"
	Wednesday Day = "WEDNESDAY"
	Thursday  Day = "THURSDAY"
	Friday    Day = "FRIDAY"
	Saturday  Day = "SATURDAY"
	Sunday    Day = "SUNDAY"
)

// Day represents a day of the week.
type Day string

func (d Day) ToWeekday() (time.Weekday, error) {
	switch d {
	case Monday:
		return time.Monday, nil
	case Tuesday:
		return time.Tuesday, nil
	case Wednesday:
		return time.Wednesday, nil
	case Thursday:
		return time.Thursday, nil
	case Friday:
		return time.Friday, nil
	case Saturday:
		return time.Saturday, nil
	case Sunday:
		return time.Sunday, nil
	default:
		return time.Sunday, fmt.Errorf("invalid day: %s", d)
	}
}

// AllDaysAsStrings returns all days of the week as a slice of strings.
func AllDaysAsStrings() []string {
	return []string{
		string(Monday),
		string(Tuesday),
		string(Wednesday),
		string(Thursday),
		string(Friday),
		string(Saturday),
		string(Sunday),
	}
}

// DayOfMonth represents a day of the month.
type DayOfMonth string

// DayOfQuarter represents a day of the quarter.
type DayOfQuarter string

// DayOfYear represents a day of the year.
type DayOfYear string

const (
	FirstDay                = "FIRST_DAY" // Valid for DayOfMonth, DayOfQuarter and DayOfYear.
	LastDay                 = "LAST_DAY"  // Valid for DayOfMonth, DayOfQuarter and DayOfYear.
	FifteenthDay DayOfMonth = "FIFTEENTH" // Valid for DayOfMonth.
)

// Month represents a month of the year.
type Month string

const (
	January   Month = "JANUARY"
	February  Month = "FEBRUARY"
	March     Month = "MARCH"
	April     Month = "APRIL"
	May       Month = "MAY"
	June      Month = "JUNE"
	July      Month = "JULY"
	August    Month = "AUGUST"
	September Month = "SEPTEMBER"
	October   Month = "OCTOBER"
	November  Month = "NOVEMBER"
	December  Month = "DECEMBER"
)

// AllMonthsAsStrings returns all months of the year as a slice of strings.
func AllMonthsAsStrings() []string {
	return []string{
		string(January),
		string(February),
		string(March),
		string(April),
		string(May),
		string(June),
		string(July),
		string(August),
		string(September),
		string(October),
		string(November),
		string(December),
	}
}

// ObjectType represents the object type an SLA domain is applicable to.
type ObjectType string

const (
	ObjectAWSEC2EBS               ObjectType = "AWS_EC2_EBS_OBJECT_TYPE"
	ObjectAWSRDS                  ObjectType = "AWS_RDS_OBJECT_TYPE"
	ObjectAWSS3                   ObjectType = "AWS_S3_OBJECT_TYPE"
	ObjectAzure                   ObjectType = "AZURE_OBJECT_TYPE"
	ObjectAzureSQLDatabase        ObjectType = "AZURE_SQL_DATABASE_OBJECT_TYPE"
	ObjectAzureSQLManagedInstance ObjectType = "AZURE_SQL_MANAGED_INSTANCE_OBJECT_TYPE"
	ObjectAzureBlob               ObjectType = "AZURE_BLOB_OBJECT_TYPE"
	ObjectGCP                     ObjectType = "GCP_OBJECT_TYPE"
)

// AllObjectTypesAsStrings returns all SLA object types as a slice of strings.
func AllObjectTypesAsStrings() []string {
	return []string{
		string(ObjectAWSEC2EBS),
		string(ObjectAWSRDS),
		string(ObjectAWSS3),
		string(ObjectAzure),
		string(ObjectAzureSQLDatabase),
		string(ObjectAzureSQLManagedInstance),
		string(ObjectAzureBlob),
		string(ObjectGCP),
	}
}

// RetentionLockMode represents the retention lock mode for an SLA domain.
type RetentionLockMode string

const (
	NoLock     RetentionLockMode = "NO_LOCK"
	Compliance RetentionLockMode = "COMPLIANCE"
	Protection RetentionLockMode = "GOVERNANCE"
)

// RetentionDuration holds a duration of time for retention.
type RetentionDuration struct {
	Duration int           `json:"duration"`
	Unit     RetentionUnit `json:"unit"`
}

// RetentionUnit represents a unit of time for retention.
type RetentionUnit string

const (
	Minute   RetentionUnit = "MINUTES"
	Hours    RetentionUnit = "HOURS"
	Days     RetentionUnit = "DAYS"
	Weeks    RetentionUnit = "WEEKS"
	Months   RetentionUnit = "MONTHS"
	Quarters RetentionUnit = "QUARTERS"
	Years    RetentionUnit = "YEARS"
)

// AllRetentionUnitsAsStrings returns all retention units as a slice of strings.
func AllRetentionUnitsAsStrings() []string {
	units := []string{
		string(Minute),
		string(Hours),
		string(Days),
		string(Weeks),
		string(Months),
		string(Quarters),
		string(Years),
	}

	return units
}

// Assignment represents the type of SLA assignment in Polaris.
type Assignment string

const (
	Derived    Assignment = "Derived"
	Direct     Assignment = "Direct"
	Unassigned Assignment = "Unassigned"
)

// AssignmentType represents the type of assignment for an SLA domain.
type AssignmentType string

const (
	NoAssignment   AssignmentType = "noAssignment"
	DoNotProtect   AssignmentType = "doNotProtect"
	ProtectWithSLA AssignmentType = "protectWithSlaId"
)

type ExistingSnapshotRetention string

const (
	NotApplicable     ExistingSnapshotRetention = "NOT_APPLICABLE"
	RetainSnapshots   ExistingSnapshotRetention = "RETAIN_SNAPSHOTS"
	KeepForever       ExistingSnapshotRetention = "KEEP_FOREVER"
	ExpireImmediately ExistingSnapshotRetention = "EXPIRE_IMMEDIATELY"
)

// ProtectionStatus represents the protection status of an object protected by
// an RSC global SLA domain.
type ProtectionStatus string

const (
	StatusUnspecified ProtectionStatus = "PROTECTION_STATUS_UNSPECIFIED"
	StatusProtected   ProtectionStatus = "PROTECTED"
	StatusRelic       ProtectionStatus = "RELIC"
	StatusUnprotected ProtectionStatus = "UNPROTECTED"
)
