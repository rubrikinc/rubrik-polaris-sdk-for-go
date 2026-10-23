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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
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
// ProtectionSet. KubernetesClusterID is the RSC FID of the K8s cluster. RSType
// is either "namespace" or "cluster". Definition is a JSON object with
// optional include and exclude resource lists, e.g.
// {"exclude":[{"resource":"Secret"}]}.
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

// PVCStorageClassMappingEntry represents an entry mapping a PVC name to a
// target storage class.
type PVCStorageClassMappingEntry struct {
	PVCName            string `json:"pvcName"`
	TargetStorageClass string `json:"targetStorageClass"`
}

// StorageClassMappingEntry represents an entry mapping a source storage class
// to a target storage class.
type StorageClassMappingEntry struct {
	SourceStorageClass string `json:"sourceStorageClass"`
	TargetStorageClass string `json:"targetStorageClass"`
}

// PVCStorageClassMappings wraps the list of PVC storage class mappings.
type PVCStorageClassMappings struct {
	PVCStorageClassMappingList []PVCStorageClassMappingEntry `json:"pvcStorageClassMappingList,omitempty"`
}

// StorageClassMappings wraps the list of storage class mappings.
type StorageClassMappings struct {
	StorageClassMappingList []StorageClassMappingEntry `json:"storageClassMappingList,omitempty"`
}

// StorageMapping defines storage class mappings for restore and export
// operations.
type StorageMapping struct {
	// PVCStorageClassMappings maps specific PVC names to target storage
	// classes. Takes precedence over StorageClassMappings.
	PVCStorageClassMappings *PVCStorageClassMappings `json:"pvcStorageClassMappings,omitempty"`
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
	CDMID                       string             `json:"cdmId"`
	ClusterUUID                 string             `json:"clusterUuid"`
	ConfiguredSLADomain         sla.DomainRef      `json:"configuredSlaDomain"`
	EffectiveRetentionSLADomain sla.DomainRef      `json:"effectiveRetentionSlaDomain,omitempty"`
	EffectiveSLADomain          sla.DomainRef      `json:"effectiveSlaDomain"`
	EffectiveSLASourceObject    hierarchy.PathNode `json:"effectiveSlaSourceObject"`
	ID                          uuid.UUID          `json:"id"`
	IsRelic                     bool               `json:"isRelic"`
	K8sClusterUUID              uuid.UUID          `json:"k8sClusterUuid"`
	Name                        string             `json:"name"`
	Namespace                   string             `json:"namespace,omitempty"`
	ObjectType                  string             `json:"objectType"`
	PendingSLA                  sla.DomainRef      `json:"pendingSla,omitempty"`
	PrimaryClusterLocation      DataLocation       `json:"primaryClusterLocation"`
	PrimaryClusterUUID          uuid.UUID          `json:"primaryClusterUuid"`
	ReplicatedObjectCount       int                `json:"replicatedObjectCount"`
	RSName                      string             `json:"rsName"`
	RSType                      string             `json:"rsType"`
	SLAAssignment               sla.Assignment     `json:"slaAssignment"`
	SLAPauseStatus              bool               `json:"slaPauseStatus"`
}

// BaseSnapshotSummary picks ID and SLAID fields from the GetSnapshot response.
type BaseSnapshotSummary struct {
	ID    string `json:"id"`
	SLAID string `json:"slaId"`
}

// EKSConfigInput is the input for the EKS config.
type EKSConfigInput struct {
	CloudAccountID string `json:"cloudAccountId"`
	EKSClusterArn  string `json:"eksClusterArn"`
}

// KuprServerProxyConfigInput is the input for the KUPR server proxy config.
type KuprServerProxyConfigInput struct {
	Cert      string `json:"cert"`
	IPAddress string `json:"ipAddress"`
	Port      int    `json:"port,omitempty"`
}

// ClusterAddInput is the input for adding a K8s cluster. Distribution is
// "VANILLA" for upstream Kubernetes.
type ClusterAddInput struct {
	ID                      string                      `json:"id,omitempty"`
	Name                    string                      `json:"name"`
	Distribution            string                      `json:"distribution"`
	Kubeconfig              secret.String               `json:"kubeconfig,omitempty"`
	Transport               string                      `json:"transport,omitempty"`
	Registry                string                      `json:"registry,omitempty"`
	PullSecret              secret.String               `json:"pullSecret,omitempty"`
	Region                  string                      `json:"region,omitempty"`
	EKSConfig               *EKSConfigInput             `json:"eksConfig,omitempty"`
	ServiceAccountName      string                      `json:"serviceAccountName,omitempty"`
	AccessToken             secret.String               `json:"accessToken,omitempty"`
	ClientID                string                      `json:"clientId,omitempty"`
	ClientSecret            secret.String               `json:"clientSecret,omitempty"`
	IsAutoPsCreationEnabled *bool                       `json:"isAutoPsCreationEnabled,omitempty"`
	OnboardingType          string                      `json:"onboardingType,omitempty"`
	KuprServerProxyConfig   *KuprServerProxyConfigInput `json:"kuprServerProxyConfig,omitempty"`
	NADNamespace            string                      `json:"nadNamespace,omitempty"`
	NADName                 string                      `json:"nadName,omitempty"`
}

// ClusterSummary is the response for the addK8sCluster query. ID is the RSC
// FID of the K8s cluster.
type ClusterSummary struct {
	ID                    string                      `json:"id"`
	Name                  string                      `json:"name"`
	Registry              string                      `json:"registry,omitempty"`
	Distribution          string                      `json:"distribution,omitempty"`
	KuprServerProxyConfig *KuprServerProxyConfigInput `json:"kuprServerProxyConfig,omitempty"`
	Transport             string                      `json:"transport,omitempty"`
	LastRefreshTime       time.Time                   `json:"lastRefreshTime,omitempty"`
	Region                string                      `json:"region,omitempty"`
	OnboardingType        string                      `json:"onboardingType,omitempty"`
	Status                string                      `json:"status"`
}

// ClusterUpdateConfigInput is the input for updating a K8s cluster.
type ClusterUpdateConfigInput struct {
	Kubeconfig              secret.String               `json:"kubeconfig,omitempty"`
	Transport               string                      `json:"transport,omitempty"`
	Registry                string                      `json:"registry,omitempty"`
	PullSecret              secret.String               `json:"pullSecret,omitempty"`
	CloudAccountID          string                      `json:"cloudAccountId,omitempty"`
	IsAutoPsCreationEnabled *bool                       `json:"isAutoPsCreationEnabled,omitempty"`
	ServiceAccountName      string                      `json:"serviceAccountName,omitempty"`
	AccessToken             secret.String               `json:"accessToken,omitempty"`
	ClientID                string                      `json:"clientId,omitempty"`
	ClientSecret            secret.String               `json:"clientSecret,omitempty"`
	KuprServerProxyConfig   *KuprServerProxyConfigInput `json:"kuprServerProxyConfig,omitempty"`
	NADNamespace            string                      `json:"nadNamespace,omitempty"`
	NADName                 string                      `json:"nadName,omitempty"`
}

// ObjectType is the type of the Kubernetes object. One of SLA, INVENTORY or
// SNAPSHOT.
type ObjectType string

const (
	ObjectTypeSLA       ObjectType = "SLA"
	ObjectTypeInventory ObjectType = "INVENTORY"
	ObjectTypeSnapshot  ObjectType = "SNAPSHOT"
)

// notFoundError wraps a 404 GraphQL error with graphql.ErrNotFound.
func notFoundError(query string, err error) error {
	var gqlErr graphql.GQLError
	if errors.As(err, &gqlErr) && gqlErr.Code() == 404 {
		err = fmt.Errorf("%s: %w", gqlErr.Error(), graphql.ErrNotFound)
	}
	return graphql.RequestError(query, err)
}

// AddProtectionSet adds the K8s protection set for the given config. The ID of
// the returned AddProtectionSetResponse is the RSC FID of the protection set.
func (a API) AddProtectionSet(ctx context.Context, config AddProtectionSetConfig) (AddProtectionSetResponse, error) {
	a.log.Print(log.Trace)

	query := addK8sProtectionSetQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Config AddProtectionSetConfig `json:"config"`
	}{Config: config})
	if err != nil {
		return AddProtectionSetResponse{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result AddProtectionSetResponse `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AddProtectionSetResponse{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// ProtectionSetByID returns the K8s protection set with the specified FID.
func (a API) ProtectionSetByID(ctx context.Context, fid uuid.UUID) (ProtectionSet, error) {
	a.log.Print(log.Trace)

	query := k8sProtectionSetQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		FID uuid.UUID `json:"fid"`
	}{FID: fid})
	if err != nil {
		return ProtectionSet{}, notFoundError(query, err)
	}

	var payload struct {
		Data struct {
			Result ProtectionSet `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ProtectionSet{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// UpdateProtectionSet updates the K8s protection set with the specified FID.
func (a API) UpdateProtectionSet(ctx context.Context, fid string, config UpdateProtectionSetConfig) error {
	a.log.Print(log.Trace)

	query := updateK8sProtectionSetQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID     string                    `json:"id"`
		Config UpdateProtectionSetConfig `json:"config"`
	}{ID: fid, Config: config})
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
		return graphql.ResponseError(query, fmt.Errorf("failed to update k8s protection set %q", fid))
	}

	return nil
}

// DeleteProtectionSet deletes the K8s protection set with the specified FID.
func (a API) DeleteProtectionSet(ctx context.Context, fid string, preserveSnapshots bool) error {
	a.log.Print(log.Trace)

	query := deleteK8sProtectionSetQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID                string `json:"id"`
		PreserveSnapshots bool   `json:"preserveSnapshots"`
	}{ID: fid, PreserveSnapshots: preserveSnapshots})
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
		return graphql.ResponseError(query, fmt.Errorf("failed to delete k8s protection set %q", fid))
	}

	return nil
}

// CreateProtectionSetSnapshot starts an on-demand snapshot job for the K8s
// protection set with the specified FID.
func (a API) CreateProtectionSetSnapshot(ctx context.Context, protectionSetFID string, config BaseOnDemandSnapshotConfigInput) (core.AsyncRequestStatus, error) {
	a.log.Print(log.Trace)

	query := createK8sProtectionSetSnapshotQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID     string                          `json:"id"`
		Config BaseOnDemandSnapshotConfigInput `json:"config"`
	}{ID: protectionSetFID, Config: config})
	if err != nil {
		return core.AsyncRequestStatus{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result core.AsyncRequestStatus `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return core.AsyncRequestStatus{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// ExportProtectionSetSnapshot starts an export job for the K8s protection set
// snapshot with the specified FID.
func (a API) ExportProtectionSetSnapshot(ctx context.Context, snapshotFID string, jobConfig ExportSnapshotJobConfig) (core.AsyncRequestStatus, error) {
	a.log.Print(log.Trace)

	query := exportK8sProtectionSetSnapshotQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID        string                  `json:"id"`
		JobConfig ExportSnapshotJobConfig `json:"jobConfig"`
	}{ID: snapshotFID, JobConfig: jobConfig})
	if err != nil {
		return core.AsyncRequestStatus{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result core.AsyncRequestStatus `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return core.AsyncRequestStatus{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// RestoreProtectionSetSnapshot starts a restore job for the K8s protection set
// snapshot with the specified FID.
func (a API) RestoreProtectionSetSnapshot(ctx context.Context, snapshotFID string, jobConfig RestoreSnapshotJobConfig) (core.AsyncRequestStatus, error) {
	a.log.Print(log.Trace)

	query := restoreK8sProtectionSetSnapshotQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID        string                   `json:"id"`
		JobConfig RestoreSnapshotJobConfig `json:"jobConfig"`
	}{ID: snapshotFID, JobConfig: jobConfig})
	if err != nil {
		return core.AsyncRequestStatus{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result core.AsyncRequestStatus `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return core.AsyncRequestStatus{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// protectionSetSnapshots returns the snapshot FIDs of the K8s protection set
// with the specified FID. Only used by the integration tests.
func (a API) protectionSetSnapshots(ctx context.Context, fid string) ([]string, error) {
	a.log.Print(log.Trace)

	query := getProtectionSetSnapshotQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID string `json:"id"`
	}{ID: fid})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Data []struct {
					BaseSnapshotSummary BaseSnapshotSummary `json:"baseSnapshotSummary"`
				} `json:"data"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	snapshots := make([]string, 0, len(payload.Data.Result.Data))
	for _, snapshot := range payload.Data.Result.Data {
		snapshots = append(snapshots, snapshot.BaseSnapshotSummary.ID)
	}

	return snapshots, nil
}

// K8sJobInstance returns the K8s CDM job with the specified job ID.
func (a API) K8sJobInstance(ctx context.Context, jobID string, cdmClusterID uuid.UUID) (K8sJobInstanceDetail, error) {
	a.log.Print(log.Trace)

	query := k8sJobInstanceQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		JobID       string    `json:"k8sJobId"`
		ClusterUUID uuid.UUID `json:"clusterUuid"`
	}{JobID: jobID, ClusterUUID: cdmClusterID})
	if err != nil {
		return K8sJobInstanceDetail{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result K8sJobInstanceDetail `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return K8sJobInstanceDetail{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// ObjectFIDByType returns the RSC FID of the K8s object with the specified CDM
// internal ID.
func (a API) ObjectFIDByType(ctx context.Context, internalID, cdmClusterID uuid.UUID, objectType ObjectType) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	query := k8sObjectFidByTypeQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		InternalID  uuid.UUID  `json:"k8SObjectInternalIdArg"`
		ClusterUUID uuid.UUID  `json:"clusterUuid"`
		ObjectType  ObjectType `json:"kubernetesObjectType"`
	}{InternalID: internalID, ClusterUUID: cdmClusterID, ObjectType: objectType})
	if err != nil {
		return uuid.Nil, notFoundError(query, err)
	}

	var payload struct {
		Data struct {
			Result uuid.UUID `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// ObjectInternalIDByType returns the CDM internal ID of the K8s object with
// the specified RSC FID.
func (a API) ObjectInternalIDByType(ctx context.Context, fid, cdmClusterID uuid.UUID, objectType ObjectType) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	query := k8sObjectInternalIdByTypeQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		FID         uuid.UUID  `json:"fid"`
		ClusterUUID uuid.UUID  `json:"clusterUuid"`
		ObjectType  ObjectType `json:"kubernetesObjectType"`
	}{FID: fid, ClusterUUID: cdmClusterID, ObjectType: objectType})
	if err != nil {
		return uuid.Nil, notFoundError(query, err)
	}

	var payload struct {
		Data struct {
			Result uuid.UUID `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// AddCluster adds the K8s cluster to the Rubrik cluster with the specified
// ID. The ID of the returned ClusterSummary is the RSC FID of the K8s cluster.
func (a API) AddCluster(ctx context.Context, cdmClusterID uuid.UUID, config ClusterAddInput) (ClusterSummary, error) {
	a.log.Print(log.Trace)

	query := addK8sClusterQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ClusterUUID uuid.UUID       `json:"clusterUuid"`
		Config      ClusterAddInput `json:"config"`
	}{ClusterUUID: cdmClusterID, Config: config})
	if err != nil {
		return ClusterSummary{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result ClusterSummary `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ClusterSummary{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// UpdateCluster updates the K8s cluster with the specified FID.
func (a API) UpdateCluster(ctx context.Context, k8sClusterFID uuid.UUID, config ClusterUpdateConfigInput) error {
	a.log.Print(log.Trace)

	query := updateK8sClusterQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID     uuid.UUID                `json:"id"`
		Config ClusterUpdateConfigInput `json:"config"`
	}{ID: k8sClusterFID, Config: config})
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
		return graphql.ResponseError(query, fmt.Errorf("failed to update k8s cluster %q", k8sClusterFID))
	}

	return nil
}

// DeleteCluster starts a job deleting the K8s cluster with the specified FID.
// When forceDelete is true, the cluster is deleted even if it is not
// reachable. Note, RSC translates the ID of the returned AsyncRequestStatus as
// if it were a cluster ID, so it cannot be passed to K8sJobInstance.
func (a API) DeleteCluster(ctx context.Context, k8sClusterFID uuid.UUID, forceDelete, preserveSnapshots bool) (core.AsyncRequestStatus, error) {
	a.log.Print(log.Trace)

	query := deleteK8sClusterQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID                uuid.UUID `json:"id"`
		ForceDelete       bool      `json:"forceDelete"`
		PreserveSnapshots bool      `json:"preserveSnapshots"`
	}{ID: k8sClusterFID, ForceDelete: forceDelete, PreserveSnapshots: preserveSnapshots})
	if err != nil {
		return core.AsyncRequestStatus{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result core.AsyncRequestStatus `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return core.AsyncRequestStatus{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}
