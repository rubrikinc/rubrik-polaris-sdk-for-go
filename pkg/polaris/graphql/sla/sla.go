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

// CloudNativeTagObjectType represents the valid object type values when
// creating a cloud native tag rule.
type CloudNativeTagObjectType string

const (
	TagObjectAWSEBSVolume                  CloudNativeTagObjectType = "AWS_EBS_VOLUME"
	TagObjectAWSEC2Instance                CloudNativeTagObjectType = "AWS_EC2_INSTANCE"
	TagObjectAWSRDSInstance                CloudNativeTagObjectType = "AWS_RDS_INSTANCE"
	TagObjectAWSS3Bucket                   CloudNativeTagObjectType = "AWS_S3_BUCKET"
	TagObjectAzureManagedDisk              CloudNativeTagObjectType = "AZURE_MANAGED_DISK"
	TagObjectAzureSQLDatabaseDB            CloudNativeTagObjectType = "AZURE_SQL_DATABASE_DB"
	TagObjectAzureSQLDatabaseServer        CloudNativeTagObjectType = "AZURE_SQL_DATABASE_SERVER"
	TagObjectAzureSQLManagedInstanceServer CloudNativeTagObjectType = "AZURE_SQL_MANAGED_INSTANCE_SERVER"
	TagObjectAzureStorageAccount           CloudNativeTagObjectType = "AZURE_STORAGE_ACCOUNT"
	TagObjectAzureVirtualMachine           CloudNativeTagObjectType = "AZURE_VIRTUAL_MACHINE"
)

// AllCloudNativeTagObjectTypesAsStrings returns all cloud native tag object
// types as a slice of strings.
func AllCloudNativeTagObjectTypesAsStrings() []string {
	return []string{
		string(TagObjectAWSEBSVolume),
		string(TagObjectAWSEC2Instance),
		string(TagObjectAWSRDSInstance),
		string(TagObjectAWSS3Bucket),
		string(TagObjectAzureManagedDisk),
		string(TagObjectAzureSQLDatabaseDB),
		string(TagObjectAzureSQLDatabaseServer),
		string(TagObjectAzureSQLManagedInstanceServer),
		string(TagObjectAzureStorageAccount),
		string(TagObjectAzureVirtualMachine),
	}
}

var managedObjectTypeMap = map[ManagedObjectType]CloudNativeTagObjectType{
	AWSNativeEBSVolume:            TagObjectAWSEBSVolume,
	AWSNativeEC2Instance:          TagObjectAWSEC2Instance,
	AWSNativeRDSInstance:          TagObjectAWSRDSInstance,
	AWSNativeS3Bucket:             TagObjectAWSS3Bucket,
	AzureManagedDisk:              TagObjectAzureManagedDisk,
	AzureSQLDatabaseDB:            TagObjectAzureSQLDatabaseDB,
	AzureSQLDatabaseServer:        TagObjectAzureSQLDatabaseServer,
	AzureSQLManagedInstanceServer: TagObjectAzureSQLManagedInstanceServer,
	AzureStorageAccount:           TagObjectAzureStorageAccount,
	AzureVirtualMachine:           TagObjectAzureVirtualMachine,
}

// FromManagedObjectType returns the corresponding CloudNativeTagObjectType for
// the given ManagedObjectType.
func FromManagedObjectType(objectType ManagedObjectType) (CloudNativeTagObjectType, error) {
	if tagObjectType, ok := managedObjectTypeMap[objectType]; ok {
		return tagObjectType, nil
	}

	return "", fmt.Errorf("unsupported managed object type: %s", objectType)
}

// ManagedObjectType represents the object type of a managed object.
type ManagedObjectType string

const (
	AWSNativeAccount                   ManagedObjectType = "AWS_NATIVE_ACCOUNT"
	AWSNativeEBSVolume                 ManagedObjectType = "AWS_NATIVE_EBS_VOLUME"
	AWSNativeEC2Instance               ManagedObjectType = "AWS_NATIVE_EC2_INSTANCE"
	AWSNativeRDSInstance               ManagedObjectType = "AWS_NATIVE_RDS_INSTANCE"
	AWSNativeS3Bucket                  ManagedObjectType = "AWS_NATIVE_S3_BUCKET"
	AzureManagedDisk                   ManagedObjectType = "AZURE_MANAGED_DISK"
	AzureResourceGroup                 ManagedObjectType = "AZURE_RESOURCE_GROUP"
	AzureResourceGroupForVMHierarchy   ManagedObjectType = "AZURE_RESOURCE_GROUP_FOR_VM_HIERARCHY"
	AzureResourceGroupFprDiskHierarchy ManagedObjectType = "AZURE_RESOURCE_GROUP_FOR_DISK_HIERARCHY"
	AzureSQLDatabaseDB                 ManagedObjectType = "AZURE_SQL_DATABASE_DB"
	AzureSQLDatabaseServer             ManagedObjectType = "AZURE_SQL_DATABASE_SERVER"
	AzureSQLManagedInstanceDB          ManagedObjectType = "AZURE_SQL_MANAGED_INSTANCE_DB"
	AzureSQLManagedInstanceServer      ManagedObjectType = "AZURE_SQL_MANAGED_INSTANCE_SERVER"
	AzureStorageAccount                ManagedObjectType = "AZURE_STORAGE_ACCOUNT"
	AzureSubscription                  ManagedObjectType = "AZURE_SUBSCRIPTION"
	AzureUnmanagedDisk                 ManagedObjectType = "AZURE_UNMANAGED_DISK"
	AzureVirtualMachine                ManagedObjectType = "AZURE_VIRTUAL_MACHINE"
	CloudNativeTagRule                 ManagedObjectType = "CLOUD_NATIVE_TAG_RULE"
	GCPNativeDisk                      ManagedObjectType = "GCP_NATIVE_DISK"
	GCPNativeGCEInstance               ManagedObjectType = "GCP_NATIVE_GCE_INSTANCE"
	GCPNativeProject                   ManagedObjectType = "GCP_NATIVE_PROJECT"
)

// SLAObjectType represents the object type an SLA domain is applicable to.
type SLAObjectType string

const (
	SLAObjectAWSEC2EBS               SLAObjectType = "AWS_EC2_EBS_OBJECT_TYPE"
	SLAObjectAWSRDS                  SLAObjectType = "AWS_RDS_OBJECT_TYPE"
	SLAObjectAWSS3                   SLAObjectType = "AWS_S3_OBJECT_TYPE"
	SLAObjectAzure                   SLAObjectType = "AZURE_OBJECT_TYPE"
	SLAObjectAzureSQLDatabase        SLAObjectType = "AZURE_SQL_DATABASE_OBJECT_TYPE"
	SLAObjectAzureSQLManagedInstance SLAObjectType = "AZURE_SQL_MANAGED_INSTANCE_OBJECT_TYPE"
	SLAObjectAzureBlob               SLAObjectType = "AZURE_BLOB_OBJECT_TYPE"
	SLAObjectGCP                     SLAObjectType = "GCP_OBJECT_TYPE"
)

// AllSLAObjectTypesAsStrings returns all SLA object types as a slice of
// strings.
func AllSLAObjectTypesAsStrings() []string {
	return []string{
		string(SLAObjectAWSEC2EBS),
		string(SLAObjectAWSRDS),
		string(SLAObjectAWSS3),
		string(SLAObjectAzure),
		string(SLAObjectAzureSQLDatabase),
		string(SLAObjectAzureSQLManagedInstance),
		string(SLAObjectAzureBlob),
		string(SLAObjectGCP),
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

// SLADomainAssignType represents the type of assignment for an SLA domain.
type SLADomainAssignType string

const (
	NoAssignment   SLADomainAssignType = "noAssignment"
	DoNotProtect   SLADomainAssignType = "doNotProtect"
	ProtectWithSLA SLADomainAssignType = "protectWithSlaId"
)

// ProtectionStatus represents the protection status of an object protected by
// an RSC global SLA domain.
type ProtectionStatus string

const (
	ProtectionStatusUnspecified ProtectionStatus = "PROTECTION_STATUS_UNSPECIFIED"
	ProtectionStatusProtected   ProtectionStatus = "PROTECTED"
	ProtectionStatusRelic       ProtectionStatus = "RELIC"
	ProtectionStatusUnprotected ProtectionStatus = "UNPROTECTED"
)

type ExistingSnapshotRetention string

const (
	NotApplicable     ExistingSnapshotRetention = "NOT_APPLICABLE"
	RetainSnapshots   ExistingSnapshotRetention = "RETAIN_SNAPSHOTS"
	KeepForever       ExistingSnapshotRetention = "KEEP_FOREVER"
	ExpireImmediately ExistingSnapshotRetention = "EXPIRE_IMMEDIATELY"
)
