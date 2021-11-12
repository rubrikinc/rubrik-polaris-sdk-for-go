package core

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// SLAAssignment represents the type of a SLA assignment in Polaris.
type SLAAssignment string

const (
	Derived    SLAAssignment = "Derived"
	Direct     SLAAssignment = "Direct"
	Unassigned SLAAssignment = "Unassigned"
)

// SLAAssignType represents how an SLA is assigned in Polaris.
type SLAAssignType string

const (
	NoAssignment     SLAAssignType = "noAssignment"
	DoNotProtect                   = "doNotProtect"
	ProtectWithSLAID               = "protectWithSlaId"
)

type SLAObjectType string

const (
	AWSEC2EBSObjectType               SLAObjectType = "AWS_EC2_EBS_OBJECT_TYPE"
	AWSRDSObjectType                                = "AWS_RDS_OBJECT_TYPE"
	AzureObjectType                                 = "AZURE_OBJECT_TYPE"
	AzureSQLDatabaseObjectType                      = "AZURE_SQL_DATABASE_OBJECT_TYPE"
	AzureSQLManagedInstanceObjectType               = "AZURE_SQL_MANAGED_INSTANCE_OBJECT_TYPE"
	CassandraObjectType                             = "CASSANDRA_OBJECT_TYPE"
	FilesetObjectType                               = "FILESET_OBJECT_TYPE"
	GCPObjectType                                   = "GCP_OBJECT_TYPE"
	KuprObjectType                                  = "KUPR_OBJECT_TYPE"
	MSSQLObjectType                                 = "MSSQL_OBJECT_TYPE"
	O365ObjectType                                  = "O365_OBJECT_TYPE"
	SAPHanaObjectType                               = "SAP_HANA_OBJECT_TYPE"
	SnapmirrorCloudObjectType                       = "SNAPMIRROR_CLOUD_OBJECT_TYPE"
	UnknownObjectType                               = "UNKNOWN_OBJECT_TYPE"
	VolumeGroupObjectType                           = "VOLUME_GROUP_OBJECT_TYPE"
	VsphereObjectType                               = "VSPHERE_OBJECT_TYPE"
)

type SLAQuerySortByField string

const (
	SortByName           SLAQuerySortByField = "NAME"
	ProtectedObjectCount                     = "PROTECTED_OBJECT_COUNT"
)

type SLAQuerySortByOrder string

const (
	ASC  SLAQuerySortByOrder = "ASC"
	DESC                     = "DESC"
)

type GlobalSLAQueryFilterInput string

const (
	ClusterUUID            GlobalSLAQueryFilterInput = "CLUSTER_UUID"
	IsEligibleForMigration                           = "IS_ELIGIBLE_FOR_MIGRATION"
	MigrationStatus                                  = "MIGRATION_STATUS"
	FilterByName                                     = "NAME"
	ObjectType                                       = "OBJECT_TYPE"
	ShowClusterSlasOnly                              = "SHOW_CLUSTER_SLAS_ONLY"
)

// SnappableLevelHierarchyType represents snappable types.
type SnappableLevelHierarchyType string

const (
	AllSubHierarchyType       SnappableLevelHierarchyType = "AllSubHierarchyType"
	AWSNativeEBSVolume                                    = "AwsNativeEbsVolume"
	AWSNativeEC2Instance                                  = "AwsNativeEc2Instance"
	AWSNativeRDSInstance                                  = "AwsNativeRdsInstance"
	AzureNativeManagedDisk                                = "AzureNativeManagedDisk"
	AzureNativeVirtualMachine                             = "AzureNativeVirtualMachine"
	AzureSQLDatabaseDb                                    = "AzureSqlDatabaseDb"
	AzureSQLManagedInstanceDb                             = "AzureSqlManagedInstanceDb"
	GCPNativeGCEInstance                                  = "GcpNativeGCEInstance"
	KuprNamespace                                         = "KuprNamespace"
	O365Mailbox                                           = "O365Mailbox"
	O365Onedrive                                          = "O365Onedrive"
	O365SharePointDrive                                   = "O365SharePointDrive"
	O365SharePointList                                    = "O365SharePointList"
	O365Site                                              = "O365Site"
	O365Teams                                             = "O365Teams"
)

type GlobalSLAFilterInput struct {
	Field          GlobalSLAQueryFilterInput `json:"field"`
	Text           string                    `json:"text"`
	ObjectTypeList []SLAObjectType           `json:"objectTypeList"`
}

// GlobalExistingSnapshotRetention represents list of predefined retention types.
type GlobalExistingSnapshotRetention string

const (
	ExpireImmediately GlobalExistingSnapshotRetention = "EXPIRE_IMMEDIATELY"
	KeepForever                                       = "KEEP_FOREVER"
	NotApplicable                                     = "NOT_APPLICABLE"
	RetainSnapshots                                   = "RETAIN_SNAPSHOTS"
)

type ContextFilterType string

const (
	AppflowsFailoverToAWS ContextFilterType = "APPFLOWS_FAILOVER_TO_AWS"
	AppflowsFailoverToCDM                   = "APPFLOWS_FAILOVER_TO_CDM"
	Default                                 = "DEFAULT"
)

type ContextFilterInputField struct {
	Field string `json:"field"`
	Text  string `json:"text"`
}

// SLADomain represents a Polaris SLA domain.
type SLADomain struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type RetentionUnit string

const (
	Days     RetentionUnit = "DAYS"
	Hours                  = "HOURS"
	Minutes                = "HOURS"
	Months                 = "MONTHS"
	Quarters               = "QUARTERS"
	Weeks                  = "WEEKS"
	Years                  = "YEARS"
)

type SLADuration struct {
	Duration int
	Unit RetentionUnit
}

type GlobalSLA struct {
	SLADomain
	ObjectTypeList []SLAObjectType
	BaseFrequency SLADuration
}

// AssignSlaForSnappableHierarchies assigns SLA defined by globalSLAOptionalFid
// to ObjectsIDs.
func (a API) AssignSlaForSnappableHierarchies(
	ctx context.Context,
	globalSLAOptionalFid uuid.UUID,
	globalSLAAssignType SLAAssignType,
	ObjectIDs []uuid.UUID,
	applicableSnappableTypes []SnappableLevelHierarchyType,
	shouldApplyToExistingSnapshots bool,
	shouldApplyToNonPolicySnapshots bool,
	globalExistingSnapshotRetention GlobalExistingSnapshotRetention,
) ([]bool, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/core.AssignSlaForSnappableHierarchies")

	buf, err := a.GQL.Request(
		ctx,
		assignSlaForSnappableHierarchiesQuery,
		struct {
			GlobalSLAOptionalFid uuid.UUID
			GlobalSLAAssignType SLAAssignType
			ObjectIDs []uuid.UUID
			ApplicableSnappableTypes []SnappableLevelHierarchyType
			ShouldApplyToExistingSnapshots bool
			ShouldApplyToNonPolicySnapshots bool
			GlobalExistingSnapshotRetention GlobalExistingSnapshotRetention
		} {
			GlobalSLAOptionalFid: globalSLAOptionalFid,
			GlobalSLAAssignType: globalSLAAssignType,
			ObjectIDs: ObjectIDs,
			ApplicableSnappableTypes: applicableSnappableTypes,
			ShouldApplyToExistingSnapshots: shouldApplyToExistingSnapshots,
			ShouldApplyToNonPolicySnapshots: shouldApplyToNonPolicySnapshots,
			GlobalExistingSnapshotRetention: globalExistingSnapshotRetention,
		})
	if err != nil {
		return nil, err
	}

	a.GQL.Log().Printf(log.Debug, "assignSlaForSnappableHierarchies: %s", string(buf))

	var payload struct {
		Data struct {
			AssignSlaForSnappableHierarchies []struct {
				Success bool `json:"success"`
			} `json:"assignSlaForSnappableHierarchiesQuery"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	ret := make([]bool, len(payload.Data.AssignSlaForSnappableHierarchies))
	for i, res := range payload.Data.AssignSlaForSnappableHierarchies {
		ret[i] = res.Success
	}
	return ret, nil
}

// ListSLA lists available SLAs on Polaris
func (a API) ListSLA(
	ctx context.Context,
	sortBy SLAQuerySortByField,
	sortOrder SLAQuerySortByOrder,
	filter []GlobalSLAFilterInput,
	contextFilter ContextFilterType,
	contextFilterInput []ContextFilterInputField,
	showSyncStatus bool,
	showProtectedObjectCount bool,
	showUpgradeInfo bool,
) ([]GlobalSLA, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/core.ListSLA")

	slaDomains := make([]GlobalSLA, 0, 10)
	cursor := ""
	for {
		buf, err := a.GQL.Request(
			ctx,
			globalSlaConnectionQuery,
			struct {
				After string
				SortBy SLAQuerySortByField
				SortOrder SLAQuerySortByOrder
				Filter []GlobalSLAFilterInput
				ContextFilter ContextFilterType
				ContextFilterInput []ContextFilterInputField
				ShowSyncStatus bool
				ShowProtectedObjectCount bool
				ShowUpgradeInfo bool
			} {
				After: cursor,
				SortBy: sortBy,
				SortOrder: sortOrder,
				Filter: filter,
				ContextFilter: contextFilter,
				ContextFilterInput: contextFilterInput,
				ShowSyncStatus: showSyncStatus,
				ShowProtectedObjectCount: showProtectedObjectCount,
				ShowUpgradeInfo: showUpgradeInfo,
			},
        )
		if err != nil {
			return nil, err
		}

		a.GQL.Log().Printf(log.Debug, "globalSlaConnection(): %s", string(buf))

		var payload struct {
			Data struct {
				Query struct {
					Count int `json:"count"`
					Edges []struct {
						Node GlobalSLA `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				}`json:"globalSlaConnection"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, err
		}
		for _, edge := range payload.Data.Query.Edges {
			slaDomains = append(slaDomains, edge.Node)
		}

		if !payload.Data.Query.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Query.PageInfo.EndCursor
	}
	return slaDomains, nil
}

