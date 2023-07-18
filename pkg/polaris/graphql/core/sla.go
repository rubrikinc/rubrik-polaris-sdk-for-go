package core

import (
	"context"
	"encoding/json"
	"fmt"

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

// SLADomain represents a Polaris SLA domain.
type SLADomain struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// SLAAssignType represents how an SLA is assigned in Polaris.
type SLAAssignType string

const (
	NoAssignment     SLAAssignType = "noAssignment"
	DoNotProtect     SLAAssignType = "doNotProtect"
	ProtectWithSLAID SLAAssignType = "protectWithSlaId"
)

// SnappableLevelHierarchyType represents snappable types.
type SnappableLevelHierarchyType string

const (
	KuprNamespace             SnappableLevelHierarchyType = "KuprNamespace"
	O365Teams                 SnappableLevelHierarchyType = "O365Teams"
	O365Onedrive              SnappableLevelHierarchyType = "O365Onedrive"
	AwsNativeS3Bucket         SnappableLevelHierarchyType = "AWS_NATIVE_S3_BUCKET"
	AwsNativeRdsInstance      SnappableLevelHierarchyType = "AwsNativeRdsInstance"
	O365SharePointList        SnappableLevelHierarchyType = "O365SharePointList"
	AllSubHierarchyType       SnappableLevelHierarchyType = "AllSubHierarchyType"
	AzureSqlManagedInstanceDb SnappableLevelHierarchyType = "AzureSqlManagedInstanceDb"
	O365SharePointDrive       SnappableLevelHierarchyType = "O365SharePointDrive"
	AwsNativeEc2Instance      SnappableLevelHierarchyType = "AwsNativeEc2Instance"
	O365Mailbox               SnappableLevelHierarchyType = "O365Mailbox"
	AzureStorageAccount       SnappableLevelHierarchyType = "AZURE_STORAGE_ACCOUNT"
	GcpNativeGCEInstance      SnappableLevelHierarchyType = "GcpNativeGCEInstance"
	AzureADDirectory          SnappableLevelHierarchyType = "AZURE_AD_DIRECTORY"
	AwsNativeEbsVolume        SnappableLevelHierarchyType = "AwsNativeEbsVolume"
	AzureSqlDatabaseDb        SnappableLevelHierarchyType = "AzureSqlDatabaseDb"
	AzureNativeManagedDisk    SnappableLevelHierarchyType = "AzureNativeManagedDisk"
	O365Site                  SnappableLevelHierarchyType = "O365Site"
	AzureNativeVirtualMachine SnappableLevelHierarchyType = "AzureNativeVirtualMachine"
)

// GlobalExistingSnapshotRetention represents list of predefined retention types.
type GlobalExistingSnapshotRetention string

const (
	ExpireImmediately GlobalExistingSnapshotRetention = "EXPIRE_IMMEDIATELY"
	KeepForever       GlobalExistingSnapshotRetention = "KEEP_FOREVER"
	NotApplicable     GlobalExistingSnapshotRetention = "NOT_APPLICABLE"
	RetainSnapshots   GlobalExistingSnapshotRetention = "RETAIN_SNAPSHOTS"
)

// AssignSlaForSnappableHierarchies assigns SLA defined by globalSLAOptionalFid
// to ObjectsIDs.
func (a API) AssignSlaForSnappableHierarchies(
	ctx context.Context,
	globalSLAOptionalFid *uuid.UUID,
	globalSLAAssignType SLAAssignType,
	ObjectIDs []uuid.UUID,
	applicableSnappableTypes []SnappableLevelHierarchyType,
	shouldApplyToExistingSnapshots bool,
	shouldApplyToNonPolicySnapshots bool,
	globalExistingSnapshotRetention GlobalExistingSnapshotRetention,
	userNote string,
) ([]bool, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		assignSlaForSnappableHierarchiesQuery,
		struct {
			GlobalSLAOptionalFid            *uuid.UUID                      `json:"globalSlaOptionalFid"`
			GlobalSLAAssignType             SLAAssignType                   `json:"globalSlaAssignType"`
			ObjectIDs                       []uuid.UUID                     `json:"objectIds"`
			ApplicableSnappableTypes        []SnappableLevelHierarchyType   `json:"applicableSnappableTypes"`
			ShouldApplyToExistingSnapshots  bool                            `json:"shouldApplyToExistingSnapshots"`
			ShouldApplyToNonPolicySnapshots bool                            `json:"shouldApplyToNonPolicySnapshots"`
			GlobalExistingSnapshotRetention GlobalExistingSnapshotRetention `json:"globalExistingSnapshotRetention"`
			UserNote                        string                          `json:"userNote"`
		}{
			GlobalSLAOptionalFid:            globalSLAOptionalFid,
			GlobalSLAAssignType:             globalSLAAssignType,
			ObjectIDs:                       ObjectIDs,
			ApplicableSnappableTypes:        applicableSnappableTypes,
			ShouldApplyToExistingSnapshots:  shouldApplyToExistingSnapshots,
			ShouldApplyToNonPolicySnapshots: shouldApplyToNonPolicySnapshots,
			GlobalExistingSnapshotRetention: globalExistingSnapshotRetention,
			UserNote:                        userNote,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to request assignSlaForSnappableHierarchies: %v", err)
	}

	a.log.Printf(log.Debug, "assignSlaForSnappableHierarchies(): %s", string(buf))

	var payload struct {
		Data struct {
			AssignSlaForSnappableHierarchies []struct {
				Success bool `json:"success"`
			} `json:"assignSlasForSnappableHierarchies"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal assignSlaForSnappableHierarchies: %v", err)
	}

	ret := make([]bool, len(payload.Data.AssignSlaForSnappableHierarchies))
	for i, res := range payload.Data.AssignSlaForSnappableHierarchies {
		ret[i] = res.Success
	}
	return ret, nil
}
