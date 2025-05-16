package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// SLAAssignment represents the type of SLA assignment in Polaris.
type SLAAssignment string

const (
	// Derived denotes derived SLA Assignment.
	Derived SLAAssignment = "Derived"
	// Direct denotes direct SLA Assignment.
	Direct SLAAssignment = "Direct"
	// Unassigned denotes no SLA Assignment.
	Unassigned SLAAssignment = "Unassigned"
)

// SLADomain represents a Polaris SLA domain.
type SLADomain struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// SLAAssignType represents how an SLA is assigned in Polaris.
type SLAAssignType string

// SLA Assignment types.
const (
	NoAssignment     SLAAssignType = "noAssignment"
	DoNotProtect     SLAAssignType = "doNotProtect"
	ProtectWithSLAID SLAAssignType = "protectWithSlaId"
)

// SnappableLevelHierarchyType represents snappable types.
type SnappableLevelHierarchyType string

// Snappable Hierarchy Types.
const (
	KuprNamespace             SnappableLevelHierarchyType = "KuprNamespace"
	O365Teams                 SnappableLevelHierarchyType = "O365Teams"
	O365Onedrive              SnappableLevelHierarchyType = "O365Onedrive"
	AWSNativeS3Bucket         SnappableLevelHierarchyType = "AWS_NATIVE_S3_BUCKET"
	AWSNativeRDSInstance      SnappableLevelHierarchyType = "AwsNativeRdsInstance"
	O365SharePointList        SnappableLevelHierarchyType = "O365SharePointList"
	AllSubHierarchyType       SnappableLevelHierarchyType = "AllSubHierarchyType"
	AzureSQLManagedInstanceDB SnappableLevelHierarchyType = "AzureSqlManagedInstanceDb"
	O365SharePointDrive       SnappableLevelHierarchyType = "O365SharePointDrive"
	AWSNativeEC2Instance      SnappableLevelHierarchyType = "AwsNativeEc2Instance"
	O365Mailbox               SnappableLevelHierarchyType = "O365Mailbox"
	AzureStorageAccount       SnappableLevelHierarchyType = "AZURE_STORAGE_ACCOUNT"
	GCPNativeGCEInstance      SnappableLevelHierarchyType = "GcpNativeGCEInstance"
	AzureADDirectory          SnappableLevelHierarchyType = "AZURE_AD_DIRECTORY"
	AwsNativeEBSVolume        SnappableLevelHierarchyType = "AwsNativeEbsVolume"
	AzureSQLDatabaseDB        SnappableLevelHierarchyType = "AzureSqlDatabaseDb"
	AzureNativeManagedDisk    SnappableLevelHierarchyType = "AzureNativeManagedDisk"
	O365Site                  SnappableLevelHierarchyType = "O365Site"
	AzureNativeVirtualMachine SnappableLevelHierarchyType = "AzureNativeVirtualMachine"
)

// GlobalExistingSnapshotRetention represents list of predefined retention types.
type GlobalExistingSnapshotRetention string

// Retention types.
const (
	ExpireImmediately GlobalExistingSnapshotRetention = "EXPIRE_IMMEDIATELY"
	KeepForever       GlobalExistingSnapshotRetention = "KEEP_FOREVER"
	NotApplicable     GlobalExistingSnapshotRetention = "NOT_APPLICABLE"
	RetainSnapshots   GlobalExistingSnapshotRetention = "RETAIN_SNAPSHOTS"
)

// AssignSLAForSnappableHierarchies assigns SLA defined by globalSLAOptionalFid
// to ObjectsIDs.
func (a API) AssignSLAForSnappableHierarchies(
	ctx context.Context,
	globalSLAOptionalFID *uuid.UUID,
	globalSLAAssignType SLAAssignType,
	ObjectIDs []uuid.UUID,
	applicableSnappableTypes []SnappableLevelHierarchyType,
	shouldApplyToExistingSnapshots *bool,
	shouldApplyToNonPolicySnapshots *bool,
	globalExistingSnapshotRetention *GlobalExistingSnapshotRetention,
	userNote *string,
) ([]bool, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		assignSlaForSnappableHierarchiesQuery,
		struct {
			GlobalSLAOptionalFID            *uuid.UUID                       `json:"globalSlaOptionalFid"`
			GlobalSLAAssignType             SLAAssignType                    `json:"globalSlaAssignType"`
			ObjectIDs                       []uuid.UUID                      `json:"objectIds"`
			ApplicableSnappableTypes        []SnappableLevelHierarchyType    `json:"applicableSnappableTypes"`
			ShouldApplyToExistingSnapshots  *bool                            `json:"shouldApplyToExistingSnapshots"`
			ShouldApplyToNonPolicySnapshots *bool                            `json:"shouldApplyToNonPolicySnapshots"`
			GlobalExistingSnapshotRetention *GlobalExistingSnapshotRetention `json:"globalExistingSnapshotRetention"`
			UserNote                        *string                          `json:"userNote"`
		}{
			GlobalSLAOptionalFID:            globalSLAOptionalFID,
			GlobalSLAAssignType:             globalSLAAssignType,
			ObjectIDs:                       ObjectIDs,
			ApplicableSnappableTypes:        applicableSnappableTypes,
			ShouldApplyToExistingSnapshots:  shouldApplyToExistingSnapshots,
			ShouldApplyToNonPolicySnapshots: shouldApplyToNonPolicySnapshots,
			GlobalExistingSnapshotRetention: globalExistingSnapshotRetention,
			UserNote:                        userNote,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to request assignSlaForSnappableHierarchies: %v",
			err,
		)
	}

	a.log.Printf(log.Debug, "assignSlaForSnappableHierarchies(): %s", string(buf))

	var payload struct {
		Data struct {
			AssignSLAForSnappableHierarchies []struct {
				Success bool `json:"success"`
			} `json:"assignSlasForSnappableHierarchies"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal assignSlaForSnappableHierarchies: %v",
			err,
		)
	}

	ret := make([]bool, len(payload.Data.AssignSLAForSnappableHierarchies))
	for i, res := range payload.Data.AssignSLAForSnappableHierarchies {
		ret[i] = res.Success
	}
	return ret, nil
}
