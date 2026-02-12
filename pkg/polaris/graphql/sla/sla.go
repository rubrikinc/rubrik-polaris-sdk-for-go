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
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
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
	ObjectActiveDirectory         ObjectType = "ACTIVE_DIRECTORY_OBJECT_TYPE"
	ObjectAtlassianJira           ObjectType = "ATLASSIAN_JIRA_OBJECT_TYPE"
	ObjectAWSDynamoDB             ObjectType = "AWS_DYNAMODB_OBJECT_TYPE"
	ObjectAWSEC2EBS               ObjectType = "AWS_EC2_EBS_OBJECT_TYPE"
	ObjectAWSRDS                  ObjectType = "AWS_RDS_OBJECT_TYPE"
	ObjectAWSS3                   ObjectType = "AWS_S3_OBJECT_TYPE"
	ObjectAzure                   ObjectType = "AZURE_OBJECT_TYPE"
	ObjectAzureAD                 ObjectType = "AZURE_AD_OBJECT_TYPE"
	ObjectAzureBlob               ObjectType = "AZURE_BLOB_OBJECT_TYPE"
	ObjectAzureDevOps             ObjectType = "AZURE_DEVOPS_OBJECT_TYPE"
	ObjectAzureSQLDatabase        ObjectType = "AZURE_SQL_DATABASE_OBJECT_TYPE"
	ObjectAzureSQLManagedInstance ObjectType = "AZURE_SQL_MANAGED_INSTANCE_OBJECT_TYPE"
	ObjectCassandra               ObjectType = "CASSANDRA_OBJECT_TYPE"
	ObjectD365                    ObjectType = "D365_OBJECT_TYPE"
	ObjectDB2                     ObjectType = "DB2_OBJECT_TYPE"
	ObjectExchange                ObjectType = "EXCHANGE_OBJECT_TYPE"
	ObjectFileset                 ObjectType = "FILESET_OBJECT_TYPE"
	ObjectGCP                     ObjectType = "GCP_OBJECT_TYPE"
	ObjectGCPCloudSQL             ObjectType = "GCP_CLOUD_SQL_OBJECT_TYPE"
	ObjectGoogleWorkspace         ObjectType = "GOOGLE_WORKSPACE_OBJECT_TYPE"
	ObjectHyperV                  ObjectType = "HYPERV_OBJECT_TYPE"
	ObjectInformixInstance        ObjectType = "INFORMIX_INSTANCE_OBJECT_TYPE"
	ObjectK8s                     ObjectType = "K8S_OBJECT_TYPE"
	ObjectKupr                    ObjectType = "KUPR_OBJECT_TYPE"
	ObjectM365BackupStorage       ObjectType = "M365_BACKUP_STORAGE_OBJECT_TYPE"
	ObjectManagedVolume           ObjectType = "MANAGED_VOLUME_OBJECT_TYPE"
	ObjectMicrosoft365            ObjectType = "O365_OBJECT_TYPE"
	ObjectMongo                   ObjectType = "MONGO_OBJECT_TYPE"
	ObjectMongoDB                 ObjectType = "MONGODB_OBJECT_TYPE"
	ObjectMSSQL                   ObjectType = "MSSQL_OBJECT_TYPE"
	ObjectMySQLDB                 ObjectType = "MYSQLDB_OBJECT_TYPE"
	ObjectNAS                     ObjectType = "NAS_OBJECT_TYPE"
	ObjectNCD                     ObjectType = "NCD_OBJECT_TYPE"
	ObjectNutanix                 ObjectType = "NUTANIX_OBJECT_TYPE"
	ObjectOkta                    ObjectType = "OKTA_OBJECT_TYPE"
	ObjectOLVM                    ObjectType = "OLVM_OBJECT_TYPE"
	ObjectOpenStack               ObjectType = "OPENSTACK_OBJECT_TYPE"
	ObjectOracle                  ObjectType = "ORACLE_OBJECT_TYPE"
	ObjectPostgresDBCluster       ObjectType = "POSTGRES_DB_CLUSTER_OBJECT_TYPE"
	ObjectProxmox                 ObjectType = "PROXMOX_OBJECT_TYPE"
	ObjectSalesforce              ObjectType = "SALESFORCE_OBJECT_TYPE"
	ObjectSAPHANA                 ObjectType = "SAP_HANA_OBJECT_TYPE"
	ObjectSnapMirrorCloud         ObjectType = "SNAPMIRROR_CLOUD_OBJECT_TYPE"
	ObjectVCD                     ObjectType = "VCD_OBJECT_TYPE"
	ObjectVolumeGroup             ObjectType = "VOLUME_GROUP_OBJECT_TYPE"
	ObjectVSphereVM               ObjectType = "VSPHERE_OBJECT_TYPE"
)

// AllObjectTypesAsStrings returns all SLA object types as a slice of strings.
func AllObjectTypesAsStrings() []string {
	return []string{
		string(ObjectActiveDirectory),
		string(ObjectAtlassianJira),
		string(ObjectAWSDynamoDB),
		string(ObjectAWSEC2EBS),
		string(ObjectAWSRDS),
		string(ObjectAWSS3),
		string(ObjectAzure),
		string(ObjectAzureAD),
		string(ObjectAzureBlob),
		string(ObjectAzureDevOps),
		string(ObjectAzureSQLDatabase),
		string(ObjectAzureSQLManagedInstance),
		string(ObjectCassandra),
		string(ObjectD365),
		string(ObjectDB2),
		string(ObjectExchange),
		string(ObjectFileset),
		string(ObjectGCP),
		string(ObjectGCPCloudSQL),
		string(ObjectGoogleWorkspace),
		string(ObjectHyperV),
		string(ObjectInformixInstance),
		string(ObjectK8s),
		string(ObjectKupr),
		string(ObjectM365BackupStorage),
		string(ObjectManagedVolume),
		string(ObjectMicrosoft365),
		string(ObjectMongo),
		string(ObjectMongoDB),
		string(ObjectMSSQL),
		string(ObjectMySQLDB),
		string(ObjectNAS),
		string(ObjectNCD),
		string(ObjectNutanix),
		string(ObjectOkta),
		string(ObjectOLVM),
		string(ObjectOpenStack),
		string(ObjectOracle),
		string(ObjectPostgresDBCluster),
		string(ObjectProxmox),
		string(ObjectSalesforce),
		string(ObjectSAPHANA),
		string(ObjectSnapMirrorCloud),
		string(ObjectVCD),
		string(ObjectVolumeGroup),
		string(ObjectVSphereVM),
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

// DomainRef is a reference to a global SLA domain in RSC. A DomainRef holds
// the ID and name of an SLA domain.
type DomainRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DoNotProtectSLAID is the special SLA domain ID used to indicate that an
// object should not be protected. This is returned in configuredSlaDomain.ID
// when "Do Not Protect" is directly assigned to an object.
const DoNotProtectSLAID = "DO_NOT_PROTECT"

// UnprotectedSLAID is the special SLA domain ID used to indicate that an
// object is unprotected (no SLA assigned). This is returned in
// effectiveSlaDomain.ID when the object inherits no protection.
const UnprotectedSLAID = "UNPROTECTED"

// HierarchyObject represents an RSC hierarchy object with SLA information.
type HierarchyObject struct {
	hierarchy.Object
	SLAAssignment       Assignment `json:"slaAssignment"`
	ConfiguredSLADomain DomainRef  `json:"configuredSlaDomain"`
	EffectiveSLADomain  DomainRef  `json:"effectiveSlaDomain"`
}

// ObjectByID returns the hierarchy object with the specified ID.
// This can be used to query any hierarchy object (VMs, databases, tag rules,
// etc.) and retrieve its SLA assignment information including the configured
// and effective SLA domains.
//
// This function uses AllSubHierarchyType as the workload hierarchy, which
// returns the generic SLA assignment. Use ObjectByIDAndWorkload to specify
// a specific workload hierarchy for workload-specific SLA resolution.
func ObjectByID(ctx context.Context, gql *graphql.Client, fid uuid.UUID) (HierarchyObject, error) {
	return ObjectByIDAndWorkload(ctx, gql, fid, hierarchy.WorkloadAllSubHierarchyType)
}

// ObjectByIDAndWorkload returns the hierarchy object with the specified ID
// and workload hierarchy type.
// This can be used to query any hierarchy object (VMs, databases, tag rules,
// etc.) and retrieve its SLA assignment information including the configured
// and effective SLA domains.
//
// The workloadHierarchy parameter determines which workload type to use for
// SLA Domain resolution. Different workload types can have different SLA
// assignments on the same parent object. Pass hierarchy.WorkloadAllSubHierarchyType
// for the generic view, or a specific workload type (e.g.,
// hierarchy.WorkloadAzureVM) for workload-specific SLA resolution.
func ObjectByIDAndWorkload(ctx context.Context, gql *graphql.Client, fid uuid.UUID, workloadHierarchy hierarchy.Workload) (HierarchyObject, error) {
	return hierarchy.ObjectByIDAndWorkload[HierarchyObject](ctx, gql, fid, workloadHierarchy)
}
