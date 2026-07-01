//go:generate go run ../queries_gen.go kubernetes

// Copyright 2021 Rubrik, Inc.
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

// Package kubernetes provides a low-level interface to the Kubernetes GraphQL
// queries provided by the Polaris platform.
package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API wraps around GraphQL client to give it the RSC Kubernetes API.
type API struct {
	GQL *graphql.Client
	log log.Logger
}

// Wrap the GraphQL client in the Kubernetes API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

// JobInstanceDetail describes the state of a CDM job.
type JobInstanceDetail struct {
	Status             string  `json:"status"`
	StartTime          string  `json:"startTime,omitempty"`
	Result             string  `json:"result,omitempty"`
	OpentracingContext string  `json:"opentracingContext,omitempty"`
	NodeID             string  `json:"nodeId"`
	JobType            string  `json:"jobType"`
	JobProgress        float64 `json:"jobProgress,omitempty"`
	IsDisabled         bool    `json:"isDisabled"`
	ID                 string  `json:"id"`
	ErrorInfo          string  `json:"errorInfo,omitempty"`
	EventSeriesID      string  `json:"eventSeriesId,omitempty"`
	EndTime            string  `json:"endTime,omitempty"`
	ChildJobDebugInfo  string  `json:"childJobDebugInfo,omitempty"`
	Archived           bool    `json:"archived"`
}

// K8sJobInstanceDetail describes the state of a K8s CDM job. It is limited to
// the fields that are used by the K8s job CRDs.
type K8sJobInstanceDetail struct {
	ID            string `json:"id"`
	EventSeriesID string `json:"eventSeriesId"`
	StartTime     string `json:"startTime,omitempty"`
	EndTime       string `json:"endTime,omitempty"`
	JobStatus     string `json:"jobStatus"`
}

// AddProtectionSetConfig defines the input parameters required to create a
// ProtectionSet.
type AddProtectionSetConfig struct {
	Definition          string   `json:"definition"`
	HookConfigs         []string `json:"hookConfigs,omitempty"`
	KubernetesClusterID string   `json:"kubernetesClusterId,omitempty"`
	KubernetesNamespace string   `json:"kubernetesNamespace,omitempty"`
	Name                string   `json:"name"`
	RSType              string   `json:"rsType"`
}

// UpdateProtectionSetConfig defines the input parameters required to update a
// ProtectionSet.
type UpdateProtectionSetConfig struct {
	Definition  string   `json:"definition,omitempty"`
	HookConfigs []string `json:"hookConfigs,omitempty"`
}

// AddProtectionSetResponse is the response received from adding a K8s
// protection set.
type AddProtectionSetResponse struct {
	ID                    string   `json:"id"`
	Definition            string   `json:"definition"`
	HookConfigs           []string `json:"hookConfigs,omitempty"`
	KubernetesClusterUUID string   `json:"kubernetesClusterUuid,omitempty"`
	KubernetesNamespace   string   `json:"kubernetesNamespace,omitempty"`
	Name                  string   `json:"name"`
	RSType                string   `json:"rsType"`
}

// PvcStorageClassMappingEntry represents an entry mapping a PVC name to a
// target storage class.
type PvcStorageClassMappingEntry struct {
	PvcName            string `json:"pvcName"`
	TargetStorageClass string `json:"targetStorageClass"`
}

// StorageClassMappingEntry represents an entry mapping a source storage class
// to a target storage class.
type StorageClassMappingEntry struct {
	SourceStorageClass string `json:"sourceStorageClass"`
	TargetStorageClass string `json:"targetStorageClass"`
}

// PvcStorageClassMappings wraps the list of PVC storage class mappings.
type PvcStorageClassMappings struct {
	PvcStorageClassMappingList []PvcStorageClassMappingEntry `json:"pvcStorageClassMappingList,omitempty"`
}

// StorageClassMappings wraps the list of storage class mappings.
type StorageClassMappings struct {
	StorageClassMappingList []StorageClassMappingEntry `json:"storageClassMappingList,omitempty"`
}

// StorageMapping defines storage class mappings for restore and export
// operations.
type StorageMapping struct {
	// PvcStorageClassMappings maps specific PVC names to target storage
	// classes. Takes precedence over StorageClassMappings.
	PvcStorageClassMappings *PvcStorageClassMappings `json:"pvcStorageClassMappings,omitempty"`
	// StorageClassMappings maps source storage classes to target storage
	// classes.
	StorageClassMappings *StorageClassMappings `json:"storageClassMappings,omitempty"`
}

// ExportSnapshotJobConfig defines parameters required to export a snapshot.
type ExportSnapshotJobConfig struct {
	TargetNamespaceName string          `json:"targetNamespaceName"`
	TargetClusterFID    string          `json:"targetClusterId"`
	IgnoreErrors        bool            `json:"ignoreErrors,omitempty"`
	Filter              string          `json:"filter,omitempty"`
	PVCNames            []string        `json:"pvcNames,omitempty"`
	StorageMapping      *StorageMapping `json:"storageMapping,omitempty"`
}

// RestoreSnapshotJobConfig defines parameters required to restore a snapshot.
type RestoreSnapshotJobConfig struct {
	IgnoreErrors   bool            `json:"ignoreErrors,omitempty"`
	Filter         string          `json:"filter,omitempty"`
	PVCNames       []string        `json:"pvcNames,omitempty"`
	StorageMapping *StorageMapping `json:"storageMapping,omitempty"`
}

// BaseOnDemandSnapshotConfigInput defines parameters required to take an
// on-demand snapshot.
type BaseOnDemandSnapshotConfigInput struct {
	SLAID string `json:"slaId"`
}

// Link has the REST link to the job.
type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

// RequestErrorInfo contains request error message if any.
type RequestErrorInfo struct {
	Message string `json:"message"`
}

// AsyncRequestStatus is the status of the submitted jobs.
type AsyncRequestStatus struct {
	Status    string           `json:"status"`
	StartTime time.Time        `json:"startTime,omitempty"`
	Progress  float64          `json:"progress,omitempty"`
	NodeID    string           `json:"nodeId,omitempty"`
	Links     []Link           `json:"links"`
	ID        string           `json:"id"`
	Error     RequestErrorInfo `json:"error,omitempty"`
	EndTime   time.Time        `json:"endTime,omitempty"`
}

// PathNode is the path description of the Snappable.
type PathNode struct {
	FID        uuid.UUID `json:"fid"`
	Name       string    `json:"name"`
	ObjectType string    `json:"objectType"`
}

// DataLocation defines the primary location of a Snappable.
type DataLocation struct {
	ClusterUUID uuid.UUID `json:"clusterUuid"`
	CreateDate  string    `json:"createDate"`
	ID          string    `json:"id"`
	IsActive    bool      `json:"isActive"`
	IsArchived  bool      `json:"isArchived"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
}

// ProtectionSet contains fields contained in a ProtectionSet snappable.
type ProtectionSet struct {
	CDMID                       string         `json:"cdmId"`
	ClusterUUID                 string         `json:"clusterUuid"`
	ConfiguredSLADomain         sla.DomainRef  `json:"configuredSlaDomain"`
	EffectiveRetentionSLADomain sla.DomainRef  `json:"effectiveRetentionSlaDomain,omitempty"`
	EffectiveSLADomain          sla.DomainRef  `json:"effectiveSlaDomain"`
	EffectiveSLASourceObject    PathNode       `json:"effectiveSlaSourceObject"`
	ID                          uuid.UUID      `json:"id"`
	IsRelic                     bool           `json:"isRelic"`
	K8sClusterUUID              uuid.UUID      `json:"k8sClusterUuid"`
	Name                        string         `json:"Name"`
	Namespace                   string         `json:"namespace,omitempty"`
	ObjectType                  string         `json:"objectType"`
	PendingSLA                  sla.DomainRef  `json:"pendingSla,omitempty"`
	PrimaryClusterLocation      DataLocation   `json:"primaryClusterLocation"`
	PrimaryClusterUUID          uuid.UUID      `json:"primaryClusterUuid"`
	ReplicatedObjectCount       int            `json:"replicatedObjectCount"`
	RSName                      string         `json:"rsName"`
	RSType                      string         `json:"rsType"`
	SLAAssignment               sla.Assignment `json:"slaAssignment"`
	SLAPauseStatus              bool           `json:"slaPauseStatus"`
}

// BaseSnapshotSummary picks ID and SLAID fields from the GetSnapshot response.
type BaseSnapshotSummary struct {
	ID    string `json:"id"`
	SLAID string `json:"slaId"`
}

// EksConfigInput is the input for the EKS config.
type EksConfigInput struct {
	CloudAccountID string `json:"cloudAccountId"`
	EKSClusterArn  string `json:"eksClusterArn"`
}

// KuprServerProxyConfigInput is the input for the KUPR server proxy config.
type KuprServerProxyConfigInput struct {
	Cert      string `json:"cert"`
	IPAddress string `json:"ipAddress"`
	Port      int    `json:"port,omitempty"`
}

// ClusterAddInput is the input for adding a K8s cluster.
type ClusterAddInput struct {
	ID                      string                     `json:"id,omitempty"`
	Name                    string                     `json:"name"`
	Distribution            string                     `json:"distribution"`
	Kubeconfig              string                     `json:"kubeconfig,omitempty"`
	Transport               string                     `json:"transport,omitempty"`
	Registry                string                     `json:"registry,omitempty"`
	PullSecret              string                     `json:"pullSecret,omitempty"`
	Region                  string                     `json:"region,omitempty"`
	EKSConfig               EksConfigInput             `json:"eksConfig,omitempty"`
	ServiceAccountName      string                     `json:"serviceAccountName,omitempty"`
	AccessToken             string                     `json:"accessToken,omitempty"`
	ClientID                string                     `json:"clientId,omitempty"`
	ClientSecret            string                     `json:"clientSecret,omitempty"`
	IsAutoPsCreationEnabled bool                       `json:"isAutoPsCreationEnabled,omitempty"`
	OnboardingType          string                     `json:"onboardingType,omitempty"`
	KuprServerProxyConfig   KuprServerProxyConfigInput `json:"kuprServerProxyConfig,omitempty"`
	NadNamespace            string                     `json:"nadNamespace,omitempty"`
	NadName                 string                     `json:"nadName,omitempty"`
}

// ClusterSummary is the response for the addK8sCluster query.
type ClusterSummary struct {
	ID                    string                     `json:"id"`
	Name                  string                     `json:"name"`
	Registry              string                     `json:"registry,omitempty"`
	Distribution          string                     `json:"distribution,omitempty"`
	KuprServerProxyConfig KuprServerProxyConfigInput `json:"kuprServerProxyConfig,omitempty"`
	Transport             string                     `json:"transport,omitempty"`
	LastRefreshTime       time.Time                  `json:"lastRefreshTime,omitempty"`
	Region                string                     `json:"region,omitempty"`
	OnboardingType        string                     `json:"onboardingType,omitempty"`
	Status                string                     `json:"status"`
}

// ClusterUpdateConfigInput is the input for updating a K8s cluster.
type ClusterUpdateConfigInput struct {
	Kubeconfig              string                     `json:"kubeconfig,omitempty"`
	Transport               string                     `json:"transport,omitempty"`
	Registry                string                     `json:"registry,omitempty"`
	PullSecret              string                     `json:"pullSecret,omitempty"`
	CloudAccountID          string                     `json:"cloudAccountId,omitempty"`
	IsAutoPsCreationEnabled bool                       `json:"isAutoPsCreationEnabled,omitempty"`
	ServiceAccountName      string                     `json:"serviceAccountName,omitempty"`
	AccessToken             string                     `json:"accessToken,omitempty"`
	ClientID                string                     `json:"clientId,omitempty"`
	ClientSecret            string                     `json:"clientSecret,omitempty"`
	KuprServerProxyConfig   KuprServerProxyConfigInput `json:"kuprServerProxyConfig,omitempty"`
	NadNamespace            string                     `json:"nadNamespace,omitempty"`
	NadName                 string                     `json:"nadName,omitempty"`
}

// handleGraphQLError processes GraphQL errors and wraps 404 errors with
// graphql.ErrNotFound. It returns a formatted error with the operation name.
func handleGraphQLError(err error, operation string) error {
	var gqlErr graphql.GQLError
	if errors.As(err, &gqlErr) && gqlErr.Code() == 404 {
		err = fmt.Errorf("%s: %w", gqlErr.Error(), graphql.ErrNotFound)
	}
	return fmt.Errorf("failed to request %s: %w", operation, err)
}

// ObjectType is the type of the Kubernetes object. One of SLA, INVENTORY or
// SNAPSHOT.
type ObjectType string

const (
	ObjectTypeSLA       ObjectType = "SLA"
	ObjectTypeInventory ObjectType = "INVENTORY"
	ObjectTypeSnapshot  ObjectType = "SNAPSHOT"
)

// AddProtectionSet adds the K8s protection set for the given config.
func (a API) AddProtectionSet(
	ctx context.Context,
	config AddProtectionSetConfig,
) (AddProtectionSetResponse, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx, addK8sProtectionSetQuery, struct {
			Config AddProtectionSetConfig `json:"config"`
		}{Config: config},
	)
	if err != nil {
		return AddProtectionSetResponse{}, fmt.Errorf(
			"failed to request addK8sProtectionSet: %w", err,
		)
	}
	a.log.Printf(log.Debug, "addK8sProtectionSet(%v): %s", config, string(buf))

	var payload struct {
		Data struct {
			Config AddProtectionSetResponse `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AddProtectionSetResponse{}, fmt.Errorf(
			"failed to unmarshal addK8sProtectionSet: %v", err,
		)
	}

	return payload.Data.Config, nil
}

// ProtectionSetByID gets the K8s protection set corresponding to the given fid.
func (a API) ProtectionSetByID(
	ctx context.Context,
	fid uuid.UUID,
) (ProtectionSet, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx, k8sProtectionSetQuery, struct {
			FID uuid.UUID `json:"fid"`
		}{FID: fid},
	)
	if err != nil {
		return ProtectionSet{}, handleGraphQLError(err, "kubernetesProtectionSet")
	}
	a.log.Printf(log.Debug, "kubernetesProtectionSet(%v): %s", fid, string(buf))

	var payload struct {
		Data struct {
			ProtectionSet ProtectionSet `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ProtectionSet{}, fmt.Errorf(
			"failed to unmarshal kubernetesProtectionSet: %v", err,
		)
	}

	return payload.Data.ProtectionSet, nil
}

// DeleteProtectionSet deletes the K8s protection set corresponding to the
// provided fid.
func (a API) DeleteProtectionSet(
	ctx context.Context,
	fid string,
	preserveSnapshots bool,
) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		deleteK8sProtectionSetQuery,
		struct {
			ID                string `json:"id"`
			PreserveSnapshots bool   `json:"preserveSnapshots"`
		}{
			ID:                fid,
			PreserveSnapshots: preserveSnapshots,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to request deleteK8sProtectionSet: %w", err)
	}
	a.log.Printf(
		log.Debug, "deleteK8sProtectionSet(%q, %v): %s",
		fid, preserveSnapshots, string(buf),
	)

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"success"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal deleteK8sProtectionSet response: %v", err)
	}
	if !payload.Data.Result.Success {
		return fmt.Errorf("failed to delete k8s protection set with fid %q", fid)
	}

	return nil
}

// K8sJobInstance fetches information about the K8s CDM job corresponding to
// the given jobID and cdmClusterID.
func (a API) K8sJobInstance(
	ctx context.Context,
	jobID string,
	cdmClusterID uuid.UUID,
) (K8sJobInstanceDetail, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		k8sJobInstanceQuery,
		struct {
			K8sJobID     string    `json:"k8sJobId"`
			CDMClusterID uuid.UUID `json:"clusterUuid"`
		}{
			K8sJobID:     jobID,
			CDMClusterID: cdmClusterID,
		},
	)
	if err != nil {
		return K8sJobInstanceDetail{}, fmt.Errorf(
			"failed to request k8sJobInstance: %w", err,
		)
	}
	a.log.Printf(
		log.Debug, "k8sJobInstance(%q, %q): %s",
		jobID, cdmClusterID, string(buf),
	)

	var payload struct {
		Data struct {
			Response K8sJobInstanceDetail `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return K8sJobInstanceDetail{}, fmt.Errorf(
			"failed to unmarshal k8sJobInstance response: %v", err,
		)
	}

	return payload.Data.Response, nil
}

// ExportProtectionSetSnapshot takes a snapshot FID, the export job config
// and starts an on-demand export job in CDM.
func (a API) ExportProtectionSetSnapshot(
	ctx context.Context,
	snapshotFID string,
	jobConfig ExportSnapshotJobConfig,
) (AsyncRequestStatus, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		exportK8sProtectionSetSnapshotQuery,
		struct {
			SnapshotFID string                  `json:"id"`
			JobConfig   ExportSnapshotJobConfig `json:"jobConfig"`
		}{
			SnapshotFID: snapshotFID,
			JobConfig:   jobConfig,
		},
	)
	if err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to request exportK8sProtectionSetSnapshot: %w", err,
		)
	}
	a.log.Printf(
		log.Debug, "exportK8sProtectionSetSnapshot(%q, %v): %s",
		snapshotFID, jobConfig, string(buf),
	)

	var payload struct {
		Data struct {
			Response AsyncRequestStatus `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to unmarshal exportK8sProtectionSetSnapshot response: %v", err,
		)
	}

	return payload.Data.Response, nil
}

// RestoreProtectionSetSnapshot takes a snapshot FID, the restore job config
// and starts an on-demand restore job in CDM.
func (a API) RestoreProtectionSetSnapshot(
	ctx context.Context,
	snapshotFID string,
	jobConfig RestoreSnapshotJobConfig,
) (AsyncRequestStatus, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		restoreK8sProtectionSetSnapshotQuery,
		struct {
			SnapshotFID string                   `json:"id"`
			JobConfig   RestoreSnapshotJobConfig `json:"jobConfig"`
		}{
			SnapshotFID: snapshotFID,
			JobConfig:   jobConfig,
		},
	)
	if err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to request restoreK8sProtectionSetSnapshot: %w", err,
		)
	}
	a.log.Printf(
		log.Debug, "restoreK8sProtectionSetSnapshot(%q, %v): %s",
		snapshotFID, jobConfig, string(buf),
	)

	var payload struct {
		Data struct {
			Response AsyncRequestStatus `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to unmarshal restoreK8sProtectionSetSnapshot response: %v", err,
		)
	}

	return payload.Data.Response, nil
}

// CreateProtectionSetSnapshot takes a ProtectionSetFID, the snapshot job
// config and starts an on-demand snapshot job in CDM.
func (a API) CreateProtectionSetSnapshot(
	ctx context.Context,
	protectionSetFID string,
	jobConfig BaseOnDemandSnapshotConfigInput,
) (AsyncRequestStatus, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		createK8sProtectionSetSnapshotQuery,
		struct {
			ProtectionSetID string                          `json:"protectionSetId"`
			JobConfig       BaseOnDemandSnapshotConfigInput `json:"jobConfig"`
		}{
			ProtectionSetID: protectionSetFID,
			JobConfig:       jobConfig,
		},
	)
	if err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to request createK8sProtectionSetSnapshot: %w", err,
		)
	}
	a.log.Printf(
		log.Debug, "createK8sProtectionSetSnapshot(%q, %v): %s",
		protectionSetFID, jobConfig, string(buf),
	)

	var payload struct {
		Data struct {
			Response AsyncRequestStatus `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to unmarshal createK8sProtectionSetSnapshot response: %v", err,
		)
	}

	return payload.Data.Response, nil
}

// ObjectFIDByType fetches the RSC FID for the object corresponding to the
// provided internal id, CDM cluster id and object type.
func (a API) ObjectFIDByType(
	ctx context.Context,
	internalID uuid.UUID,
	cdmClusterID uuid.UUID,
	objectType ObjectType,
) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		k8sObjectFidByTypeQuery,
		struct {
			InternalID    uuid.UUID  `json:"k8SObjectInternalIdArg"`
			ClusterUUID   uuid.UUID  `json:"clusterUuid"`
			K8sObjectType ObjectType `json:"kubernetesObjectType"`
		}{
			InternalID:    internalID,
			ClusterUUID:   cdmClusterID,
			K8sObjectType: objectType,
		},
	)
	if err != nil {
		return uuid.Nil, handleGraphQLError(err, "k8sObjectFidByType")
	}
	a.log.Printf(
		log.Debug, "k8sObjectFidByType(%v, %v, %v): %s",
		internalID, cdmClusterID, objectType, string(buf),
	)

	var payload struct {
		Data struct {
			ObjectFID uuid.UUID `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal k8sObjectFidByType: %v", err)
	}

	return payload.Data.ObjectFID, nil
}

// ObjectInternalIDByType fetches the object Internal ID on CDM for the
// given RSC FID by type.
func (a API) ObjectInternalIDByType(
	ctx context.Context,
	fid uuid.UUID,
	clusterUUID uuid.UUID,
	objectType ObjectType,
) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		k8sObjectInternalIdByTypeQuery,
		struct {
			ClusterUUID          uuid.UUID  `json:"clusterUuid"`
			FID                  uuid.UUID  `json:"fid"`
			KubernetesObjectType ObjectType `json:"kubernetesObjectType"`
		}{
			FID:                  fid,
			ClusterUUID:          clusterUUID,
			KubernetesObjectType: objectType,
		},
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf(
			"failed to request k8sObjectInternalIdByType: %w", err,
		)
	}
	a.log.Printf(
		log.Debug, "k8sObjectInternalIdByType(%v, %v, %v): %s",
		fid, clusterUUID, objectType, string(buf),
	)

	var payload struct {
		Data struct {
			ObjectInternalID uuid.UUID `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf(
			"failed to unmarshal k8sObjectInternalIdByType: %v", err,
		)
	}

	return payload.Data.ObjectInternalID, nil
}

// ObjectFID fetches the RSC FID for the K8s object corresponding to the
// provided internal id and CDM cluster id. Unlike ObjectFIDByType, this
// variant does not require specifying the object type.
func (a API) ObjectFID(
	ctx context.Context,
	internalID uuid.UUID,
	cdmClusterID uuid.UUID,
) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		k8sObjectFidQuery,
		struct {
			InternalID  uuid.UUID `json:"k8SObjectInternalIdArg"`
			ClusterUUID uuid.UUID `json:"clusterUuid"`
		}{
			InternalID:  internalID,
			ClusterUUID: cdmClusterID,
		},
	)
	if err != nil {
		return uuid.Nil, handleGraphQLError(err, "k8sObjectFid")
	}
	a.log.Printf(
		log.Debug, "k8sObjectFid(%v, %v): %s",
		internalID, cdmClusterID, string(buf),
	)

	var payload struct {
		Data struct {
			ObjectFID uuid.UUID `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal k8sObjectFid: %v", err)
	}

	return payload.Data.ObjectFID, nil
}

// ObjectInternalID fetches the CDM internal ID for the K8s object
// corresponding to the provided RSC FID. Unlike ObjectInternalIDByType, this
// variant does not require specifying the cluster UUID or object type.
func (a API) ObjectInternalID(
	ctx context.Context,
	fid uuid.UUID,
) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		k8sObjectInternalIdQuery,
		struct {
			FID uuid.UUID `json:"fid"`
		}{
			FID: fid,
		},
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf(
			"failed to request k8sObjectInternalId: %w", err,
		)
	}
	a.log.Printf(log.Debug, "k8sObjectInternalId(%v): %s", fid, string(buf))

	var payload struct {
		Data struct {
			ObjectInternalID uuid.UUID `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf(
			"failed to unmarshal k8sObjectInternalId: %v", err,
		)
	}

	return payload.Data.ObjectInternalID, nil
}

// JobInstance fetches information about a generic CDM job corresponding to the
// given jobID and cdmClusterID. This returns the full JobInstanceDetail with
// all fields, unlike K8sJobInstance which returns a K8s-specific subset.
func (a API) JobInstance(
	ctx context.Context,
	jobID string,
	cdmClusterID string,
) (JobInstanceDetail, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		jobInstanceQuery,
		struct {
			JobID        string `json:"id"`
			CDMClusterID string `json:"clusterUuid"`
		}{
			JobID:        jobID,
			CDMClusterID: cdmClusterID,
		},
	)
	if err != nil {
		return JobInstanceDetail{}, fmt.Errorf(
			"failed to request jobInstance: %w", err,
		)
	}
	a.log.Printf(
		log.Debug, "jobInstance(%q, %q): %s",
		jobID, cdmClusterID, string(buf),
	)

	var payload struct {
		Data struct {
			Response JobInstanceDetail `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return JobInstanceDetail{}, fmt.Errorf(
			"failed to unmarshal jobInstance response: %v", err,
		)
	}

	return payload.Data.Response, nil
}

// protectionSetSnapshots gets initial snapshots for a given fid. Only used
// internally for testing purposes.
func (a API) protectionSetSnapshots(
	ctx context.Context,
	fid string,
) ([]string, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		getProtectionSetSnapshotQuery,
		struct {
			FID string `json:"fid"`
		}{
			FID: fid,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to request k8sProtectionSetSnapshots: %w", err,
		)
	}
	a.log.Printf(log.Debug, "k8sProtectionSetSnapshots(%v): %s", fid, string(buf))

	var payload struct {
		Data struct {
			K8sProtectionSetSnapshots struct {
				Data []struct {
					BaseSnapshotSummary BaseSnapshotSummary `json:"baseSnapshotSummary"`
				} `json:"data"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal k8sProtectionSetSnapshots: %v", err,
		)
	}
	snaps := make([]string, len(payload.Data.K8sProtectionSetSnapshots.Data))
	for i, item := range payload.Data.K8sProtectionSetSnapshots.Data {
		snaps[i] = item.BaseSnapshotSummary.ID
	}
	return snaps, nil
}

// UpdateProtectionSet updates the K8s protection set corresponding to fid
// with the config.
func (a API) UpdateProtectionSet(
	ctx context.Context,
	fid string,
	config UpdateProtectionSetConfig,
) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		updateK8sProtectionSetQuery,
		struct {
			ID           string                    `json:"id"`
			UpdateConfig UpdateProtectionSetConfig `json:"updateConfig"`
		}{
			ID:           fid,
			UpdateConfig: config,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to request updateK8sProtectionSet: %w", err)
	}
	a.log.Printf(
		log.Debug, "updateK8sProtectionSet(%q, %v): %s",
		fid, config, string(buf),
	)

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"success"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal updateK8sProtectionSet: %v", err)
	}
	if !payload.Data.Result.Success {
		return fmt.Errorf("failed to update k8s protection set with fid %q", fid)
	}

	return nil
}

// AddCluster adds the K8s cluster for the given config.
func (a API) AddCluster(
	ctx context.Context,
	cdmClusterID uuid.UUID,
	config ClusterAddInput,
) (ClusterSummary, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		addK8sClusterQuery,
		struct {
			ClusterUUID string          `json:"clusterUuid"`
			Config      ClusterAddInput `json:"config"`
		}{
			ClusterUUID: cdmClusterID.String(),
			Config:      config,
		},
	)
	if err != nil {
		return ClusterSummary{}, fmt.Errorf(
			"failed to request addK8sCluster: %w", err,
		)
	}
	a.log.Printf(
		log.Debug, "addK8sCluster(%s, %v): %s",
		cdmClusterID.String(), config, string(buf),
	)

	var payload struct {
		Data struct {
			Response ClusterSummary `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ClusterSummary{}, fmt.Errorf(
			"failed to unmarshal addK8sCluster: %v", err,
		)
	}
	return payload.Data.Response, nil
}

// UpdateCluster updates the K8s cluster for the given config.
func (a API) UpdateCluster(
	ctx context.Context,
	k8sClusterFID uuid.UUID,
	config ClusterUpdateConfigInput,
) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		updateK8sClusterQuery,
		struct {
			ID     string                   `json:"id"`
			Config ClusterUpdateConfigInput `json:"config"`
		}{
			ID:     k8sClusterFID.String(),
			Config: config,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to request updateK8sCluster: %w", err)
	}
	a.log.Printf(
		log.Debug, "updateK8sCluster(%s, %v): %s",
		k8sClusterFID.String(), config, string(buf),
	)

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"success"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal updateK8sCluster: %v", err)
	}
	if !payload.Data.Result.Success {
		return fmt.Errorf("failed to update k8s cluster with fid %q", k8sClusterFID)
	}

	return nil
}
