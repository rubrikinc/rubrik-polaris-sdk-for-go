//go:generate go run ../queries_gen.go infinityk8s

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

// Package infinityk8s provides a low level interface to the Infinity K8s
// GraphQL queries provided by the Polaris platform.
package infinityk8s

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// JobInstanceDetail describes the state of a CDM job.
type JobInstanceDetail struct {
	// Status gives the current status of the job.
	Status string `json:"status"`
	// StartTime gives the time when the job started.
	StartTime string `json:"startTime,omitempty"`
	// Result gives the result of the job.
	Result string `json:"result,omitempty"`
	// OpentracingContext is the opentracing context of the job.
	OpentracingContext string `json:"opentracingContext,omitempty"`
	// NodeID is the ID of the CDM node where the job is running.
	NodeID string `json:"nodeId"`
	// JobType is the type of the job.
	JobType string `json:"jobType"`
	// JobProgress is the progress of the job.
	JobProgress float64 `json:"jobProgress,omitempty"`
	// IsDisabled is true if the job is disabled.
	IsDisabled bool `json:"isDisabled"`
	// ID is the ID of the job.
	ID string `json:"id"`
	// ErrorInfo is the error information of the job.
	ErrorInfo string `json:"errorInfo,omitempty"`
	// EventSeriesID is the ID of the event series to identify events related
	// to the job.
	EventSeriesID string `json:"eventSeriesId,omitempty"`
	// EndTime is the time when the job ended.
	EndTime string `json:"endTime,omitempty"`
	// ChildJobDebugInfo is the debug information of the child job.
	ChildJobDebugInfo string `json:"childJobDebugInfo,omitempty"`
	// Archived is true if the job is archived.
	Archived bool `json:"archived"`
}

// K8sJobInstanceDetail describes the state of a K8s CDM job. It is limited to
// the fields that are used by the K8s job CRDs.
type K8sJobInstanceDetail struct {
	// ID is the ID of the job.
	ID string `json:"id"`
	// EventSeriesID is the ID of the event series to identify events related
	// to the job.
	EventSeriesID string `json:"eventSeriesId"`
	// StartTime gives the time when the job started.
	StartTime string `json:"startTime,omitempty"`
	// EndTime is the time when the job ended.
	EndTime string `json:"endTime,omitempty"`
	// JobStatus gives the current status of the job.
	JobStatus string `json:"jobStatus"`
}

// AddK8sProtectionSetConfig defines the input parameters required to create a
// ProtectionSet.
type AddK8sProtectionSetConfig struct {
	// Definition is the definition containing filters for the protection set.
	Definition string `json:"definition"`
	// HookConfigs are the hook configurations for the protection set.
	HookConfigs []string `json:"hookConfigs,omitempty"`
	// KubernetesClusterID is the ID of the Kubernetes cluster.
	KubernetesClusterID string `json:"kubernetesClusterId,omitempty"`
	// KubernetesNamespace is the namespace of the Kubernetes cluster.
	KubernetesNamespace string `json:"kubernetesNamespace,omitempty"`
	// Name is the name of the protection set.
	Name string `json:"name"`
	// RSType is the type of the protection set.
	RSType string `json:"rsType"`
}

// UpdateK8sProtectionSetConfig defines the input parameters required to
// update a ProtectionSet.
type UpdateK8sProtectionSetConfig struct {
	// Definition is the filter definition of the protection set.
	Definition string `json:"definition,omitempty"`
	// HookConfigs are the hook configurations for the protection set.
	HookConfigs []string `json:"hookConfigs,omitempty"`
}

// AddK8sProtectionSetResponse is the response received from adding a K8s
// cluster.
type AddK8sProtectionSetResponse struct {
	// ID is the ID of the protection set.
	ID string `json:"id"`
	// Definition is the filter definition of the protection set.
	Definition string `json:"definition"`
	// HookConfigs are the hook configurations for the protection set.
	HookConfigs []string `json:"hookConfigs,omitempty"`
	// KubernetesClusterUUID is the UUID of the source Kubernetes cluster.
	KubernetesClusterUUID string `json:"kubernetesClusterUuid,omitempty"`
	// KubernetesNamespace is the namespace of the Kubernetes cluster.
	KubernetesNamespace string `json:"kubernetesNamespace,omitempty"`
	// Name is the name of the protection set.
	Name string `json:"name"`
	// RSType is the type of the protection set.
	RSType string `json:"rsType"`
}

// ExportK8sProtectionSetSnapshotJobConfig defines parameters required to
// export a snapshot.
type ExportK8sProtectionSetSnapshotJobConfig struct {
	// TargetNamespaceName is the name of the target namespace.
	TargetNamespaceName string `json:"targetNamespaceName"`
	// TargetClusterFID is the FID of the target cluster.
	TargetClusterFID string `json:"targetClusterId"`
	// IgnoreErrors indicates whether to ignore errors during the export.
	IgnoreErrors bool `json:"ignoreErrors,omitempty"`
	// Filter is the filter to apply during the export.
	Filter string `json:"filter,omitempty"`
	// PVCNames are the names of the PVCs to export.
	PVCNames []string `json:"pvcNames,omitempty"`
}

// RestoreK8sProtectionSetSnapshotJobConfig defines parameters required to
// export a snapshot.
type RestoreK8sProtectionSetSnapshotJobConfig struct {
	// IgnoreErrors indicates whether to ignore errors during the restore.
	IgnoreErrors bool `json:"ignoreErrors,omitempty"`
	// Filter is the filter to apply during the restore.
	Filter string `json:"filter,omitempty"`
	// PVCNames are the names of the PVCs to restore.
	PVCNames []string `json:"pvcNames,omitempty"`
}

// BaseOnDemandSnapshotConfigInput defines parameters required to take an
// on-demand snapshot.
type BaseOnDemandSnapshotConfigInput struct {
	// SLAID is the ID of the SLA.
	SLAID string `json:"slaId"`
}

// Link has the REST link to the job.
type Link struct {
	// Rel is the relation of the link.
	Rel string `json:"rel"`
	// Href is the URL of the link.
	Href string `json:"href"`
}

// RequestErrorInfo contains request error message if any.
type RequestErrorInfo struct {
	// Message is the error message.
	Message string `json:"message"`
}

// AsyncRequestStatus is the status of the submitted jobs.
type AsyncRequestStatus struct {
	// Status is the current status of the request.
	Status string `json:"status"`
	// StartTime is the time when the request started.
	StartTime time.Time `json:"startTime,omitempty"`
	// Progress is the progress of the request.
	Progress float64 `json:"progress,omitempty"`
	// NodeID is the ID of the node handling the request.
	NodeID string `json:"nodeId,omitempty"`
	// Links are the REST links related to the request.
	Links []Link `json:"links"`
	// ID is the ID of the request.
	ID string `json:"id"`
	// Error contains error information if any.
	Error RequestErrorInfo `json:"error,omitempty"`
	// EndTime is the time when the request ended.
	EndTime time.Time `json:"endTime,omitempty"`
}

// PathNode is the path description of the Snappable.
type PathNode struct {
	// FID is the FID of the path node.
	FID uuid.UUID `json:"fid"`
	// Name is the name of the path node.
	Name string `json:"name"`
	// ObjectType corresponds to HierarchyObjectTypeEnum.
	ObjectType string `json:"objectType"`
}

// DataLocation defines the primary location of a Snappable.
type DataLocation struct {
	// ClusterUUID is the UUID of the cluster.
	ClusterUUID uuid.UUID `json:"clusterUuid"`
	// CreateDate is the creation date of the data location.
	CreateDate string `json:"createDate"`
	// ID is the ID of the data location.
	ID string `json:"id"`
	// IsActive indicates whether the data location is active.
	IsActive bool `json:"isActive"`
	// IsArchived indicates whether the data location is archived.
	IsArchived bool `json:"isArchived"`
	// Name is the name of the data location.
	Name string `json:"name"`
	// Type corresponds to the DataLocationName enum.
	Type string `json:"type"`
}

// KubernetesProtectionSet contains fields contained in a ProtectionSet
// snappable.
type KubernetesProtectionSet struct {
	// CDMID is the ID of the CDM.
	CDMID string `json:"cdmId"`
	// ClusterUUID is the UUID of the cluster.
	ClusterUUID string `json:"clusterUuid"`
	// ConfiguredSLADomain is the configured SLA domain.
	ConfiguredSLADomain core.SLADomain `json:"configuredSlaDomain"`
	// EffectiveRetentionSLADomain is the effective retention SLA domain.
	EffectiveRetentionSLADomain core.SLADomain `json:"effectiveRetentionSlaDomain,omitempty"`
	// EffectiveSLADomain is the effective SLA domain.
	EffectiveSLADomain core.SLADomain `json:"effectiveSlaDomain"`
	// EffectiveSLASourceObject is the source object of the effective SLA.
	EffectiveSLASourceObject PathNode `json:"effectiveSlaSourceObject"`
	// ID is the ID of the protection set.
	ID uuid.UUID `json:"id"`
	// IsRelic indicates whether the protection set is a relic.
	IsRelic bool `json:"isRelic"`
	// K8sClusterUUID is the UUID of the Kubernetes cluster.
	K8sClusterUUID uuid.UUID `json:"k8sClusterUuid"`
	// Name is the name of the protection set.
	Name string `json:"Name"`
	// Namespace is the namespace of the protection set.
	Namespace string `json:"namespace,omitempty"`
	// ObjectType corresponds to HierarchyObjectTypeEnum.
	ObjectType string `json:"objectType"`
	// PendingSLA is the pending SLA domain.
	PendingSLA core.SLADomain `json:"pendingSla,omitempty"`
	// PrimaryClusterLocation is the primary location of the cluster.
	PrimaryClusterLocation DataLocation `json:"primaryClusterLocation"`
	// PrimaryClusterUUID is the UUID of the primary cluster.
	PrimaryClusterUUID uuid.UUID `json:"primaryClusterUuid"`
	// ReplicatedObjectCount is the count of replicated objects.
	ReplicatedObjectCount int `json:"replicatedObjectCount"`
	// RSName is the name of the protection set.
	RSName string `json:"rsName"`
	// RSType is the type of the protection set.
	RSType string `json:"rsType"`
	// SLAAssignment is the SLA assignment.
	SLAAssignment core.SLAAssignment `json:"slaAssignment"`
	// SLAPauseStatus indicates whether the SLA is paused.
	SLAPauseStatus bool `json:"slaPauseStatus"`
}

// BaseSnapshotSummary picks ID and SLAID fields from the GetSnapshot response.
type BaseSnapshotSummary struct {
	// ID is the ID of the snapshot.
	ID string `json:"id"`
	// SLAID is the ID of the SLA.
	SLAID string `json:"slaId"`
}

// ActivitySeriesInput is the input for the activitySeries query.
type ActivitySeriesInput struct {
	// ActivitySeriesID is the ID of the activity series.
	ActivitySeriesID uuid.UUID `json:"activitySeriesId"`
	// CDMClusterUUID is the UUID of the CDM cluster.
	CDMClusterUUID uuid.UUID `json:"clusterUuid,omitempty"`
}

// EventSeverity is the severity of the event.
type EventSeverity string

const (
	// EventSeverityCritical is the critical severity.
	EventSeverityCritical EventSeverity = "Critical"
	// EventSeverityWarning is the warning severity.
	EventSeverityWarning EventSeverity = "Warning"
	// EventSeverityInfo is the info severity.
	EventSeverityInfo EventSeverity = "Info"
)

// ActivitySeries is the response for the activitySeries query.
type ActivitySeries struct {
	// ActivityInfo is the information about the activity.
	ActivityInfo string `json:"activityInfo"`
	// Message is the message of the activity.
	Message string `json:"message"`
	// Status is the status of the activity.
	Status string `json:"status"`
	// Time is the time of the activity.
	Time time.Time `json:"time"`
	// Severity is the severity of the activity.
	Severity EventSeverity `json:"severity"`
}

// EksConfigInput is the input for the EKS config.
type EksConfigInput struct {
	// CloudAccountId is the ID of the cloud account.
	CloudAccountId string `json:"cloudAccountId"`
	// EKSClusterArn is the ARN of the EKS cluster.
	EKSClusterArn string `json:"eksClusterArn"`
}

// KuprServerProxyConfigInput is the input for the KUPR server proxy config.
type KuprServerProxyConfigInput struct {
	// Cert is the certificate for the proxy.
	Cert string `json:"cert"`
	// IPAddress is the IP address of the proxy.
	IPAddress string `json:"ipAddress"`
	// Port is the port of the proxy.
	Port int `json:"port,omitempty"`
}

// K8sClusterAddInput is the input for adding a K8s cluster.
type K8sClusterAddInput struct {
	// ID is the ID of the cluster.
	ID string `json:"id,omitempty"`
	// Name is the name of the cluster.
	Name string `json:"name"`
	// Distribution is the distribution of the cluster.
	Distribution string `json:"distribution"`
	// Kubeconfig is the kubeconfig of the cluster.
	Kubeconfig string `json:"kubeconfig,omitempty"`
	// Transport is the transport method for the cluster.
	Transport string `json:"transport,omitempty"`
	// Registry is the registry for the cluster.
	Registry string `json:"registry,omitempty"`
	// PullSecret is the pull secret for the cluster.
	PullSecret string `json:"pullSecret,omitempty"`
	// Region is the region of the cluster.
	Region string `json:"region,omitempty"`
	// EKSConfig is the EKS configuration for the cluster.
	EKSConfig EksConfigInput `json:"eksConfig,omitempty"`
	// ServiceAccountName is the name of the service account.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// AccessToken is the access token for the cluster.
	AccessToken string `json:"accessToken,omitempty"`
	// ClientId is the client ID for the cluster.
	ClientId string `json:"clientId,omitempty"`
	// ClientSecret is the client secret for the cluster.
	ClientSecret string `json:"clientSecret,omitempty"`
	// IsAutoPsCreationEnabled indicates whether auto protection set creation is enabled.
	IsAutoPsCreationEnabled bool `json:"isAutoPsCreationEnabled,omitempty"`
	// OnboardingType is the onboarding type for the cluster.
	OnboardingType string `json:"onboardingType,omitempty"`
	// KuprServerProxyConfig is the proxy configuration for the KUPR server.
	KuprServerProxyConfig KuprServerProxyConfigInput `json:"kuprServerProxyConfig,omitempty"`
}

// K8sClusterSummary is the response for the addK8sCluster query.
type K8sClusterSummary struct {
	// ID is the ID of the cluster.
	ID string `json:"id"`
	// Name is the name of the cluster.
	Name string `json:"name"`
	// Registry is the registry for the cluster.
	Registry string `json:"registry,omitempty"`
	// Distribution is the distribution of the cluster.
	Distribution string `json:"distribution,omitempty"`
	// KuprServerProxyConfig is the proxy configuration for the KUPR server.
	KuprServerProxyConfig KuprServerProxyConfigInput `json:"kuprServerProxyConfig,omitempty"`
	// Transport is the transport method for the cluster.
	Transport string `json:"transport,omitempty"`
	// LastRefreshTime is the last refresh time of the cluster.
	LastRefreshTime time.Time `json:"lastRefreshTime,omitempty"`
	// Region is the region of the cluster.
	Region string `json:"region,omitempty"`
	// OnboardingType is the onboarding type for the cluster.
	OnboardingType string `json:"onboardingType,omitempty"`
	// Status is the status of the cluster.
	Status string `json:"status"`
}

// K8sClusterUpdateConfigInput is the input for updating a K8s cluster.
type K8sClusterUpdateConfigInput struct {
	// Kubeconfig is the kubeconfig of the cluster.
	Kubeconfig string `json:"kubeconfig,omitempty"`
	// Transport is the transport method for the cluster.
	Transport string `json:"transport,omitempty"`
	// Registry is the registry for the cluster.
	Registry string `json:"registry,omitempty"`
	// PullSecret is the pull secret for the cluster.
	PullSecret string `json:"pullSecret,omitempty"`
	// CloudAccountId is the ID of the cloud account.
	CloudAccountId string `json:"cloudAccountId,omitempty"`
	// IsAutoPsCreationEnabled indicates whether auto protection set creation is enabled.
	IsAutoPsCreationEnabled bool `json:"isAutoPsCreationEnabled,omitempty"`
	// ServiceAccountName is the name of the service account.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// AccessToken is the access token for the cluster.
	AccessToken string `json:"accessToken,omitempty"`
	// ClientId is the client ID for the cluster.
	ClientId string `json:"clientId,omitempty"`
	// ClientSecret is the client secret for the cluster.
	ClientSecret string `json:"clientSecret,omitempty"`
	// KuprServerProxyConfig is the proxy configuration for the KUPR server.
	KuprServerProxyConfig KuprServerProxyConfigInput `json:"kuprServerProxyConfig,omitempty"`
}

type API struct {
	GQL *graphql.Client
	log log.Logger
}

// Wrap the RSC client in the Infinity K8s API.
func Wrap(client *polaris.Client) API {
	return API{GQL: client.GQL, log: client.GQL.Log()}
}

// K8sObjectType is the type of the Kubernetes object. One between SLA,
// INVENTORY or SNAPSHOTS.
type K8sObjectType string

const (
	// K8sObjectTypeSLA is the SLA type.
	K8sObjectTypeSLA K8sObjectType = "SLA"
	// K8sObjectTypeInventory is the Inventory type.
	K8sObjectTypeInventory K8sObjectType = "INVENTORY"
	// K8sObjectTypeSnapshot is the Snapshots type.
	K8sObjectTypeSnapshot K8sObjectType = "SNAPSHOT"
)

// AddK8sProtectionSet adds the K8s protection set for the given config.
func (a API) AddK8sProtectionSet(
	ctx context.Context,
	config AddK8sProtectionSetConfig,
) (AddK8sProtectionSetResponse, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx, addK8sProtectionSetQuery, struct {
			Config AddK8sProtectionSetConfig `json:"config"`
		}{Config: config},
	)
	if err != nil {
		return AddK8sProtectionSetResponse{}, fmt.Errorf(
			"failed to request addK8sProtectionSet: %w",
			err,
		)
	}
	a.log.Printf(log.Debug, "addK8sProtectionSet(%v): %s", config, string(buf))

	var payload struct {
		Data struct {
			Config AddK8sProtectionSetResponse `json:"addK8sProtectionSet"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AddK8sProtectionSetResponse{}, fmt.Errorf(
			"failed to unmarshal addK8sProtectionSet: %v",
			err,
		)
	}

	return payload.Data.Config, nil
}

// K8sProtectionSet get the K8s protection set corresponding to the given fid.
func (a API) K8sProtectionSet(
	ctx context.Context,
	fid uuid.UUID,
) (KubernetesProtectionSet, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx, k8sProtectionSetQuery, struct {
			FID uuid.UUID `json:"fid"`
		}{FID: fid},
	)
	if err != nil {
		return KubernetesProtectionSet{}, fmt.Errorf(
			"failed to request kubernetesProtectionSet: %w",
			err,
		)
	}
	a.log.Printf(log.Debug, "kubernetesProtectionSet(%v): %s", fid, string(buf))

	var payload struct {
		Data struct {
			KubernetesProtectionSet KubernetesProtectionSet `json:"kubernetesProtectionSet"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return KubernetesProtectionSet{}, fmt.Errorf(
			"failed to unmarshal kubernetesProtectionSet: %v",
			err,
		)
	}

	return payload.Data.KubernetesProtectionSet, nil
}

// DeleteK8sProtectionSet deletes the K8s protection set corresponding to the
// provided fid.
func (a API) DeleteK8sProtectionSet(
	ctx context.Context,
	fid string,
	preserveSnapshots bool,
) (bool, error) {
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
		return false, fmt.Errorf(
			"failed to request deleteK8sProtectionSet: %w",
			err,
		)
	}
	a.log.Printf(
		log.Debug,
		"deleteK8sProtectionSet(%q, %q): %s",
		fid,
		preserveSnapshots,
		string(buf),
	)

	var payload struct {
		Data struct {
			ResponseSuccess struct {
				Success bool `json:"success"`
			} `json:"deleteK8sProtectionSet"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return false, fmt.Errorf(
			"failed to unmarshal deleteK8sProtectionSet response: %v",
			err,
		)
	}
	if !payload.Data.ResponseSuccess.Success {
		return false, fmt.Errorf(
			"failed to delete k8s protection set with fid %q",
			fid,
		)
	}

	return true, nil
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
			"failed to request k8sJobInstance: %w",
			err,
		)
	}

	a.log.Printf(
		log.Debug,
		"k8sJobInstance(%q, %q): %s",
		jobID,
		cdmClusterID,
		string(buf),
	)

	var payload struct {
		Data struct {
			Response K8sJobInstanceDetail `json:"k8sJobInstance"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return K8sJobInstanceDetail{}, fmt.Errorf(
			"failed to unmarshal k8sJobInstance response: %v",
			err,
		)
	}

	return payload.Data.Response, nil
}

// ExportK8sProtectionSetSnapshot takes a snapshot FID, the export job config
// and starts an on-demand export job in CDM.
func (a API) ExportK8sProtectionSetSnapshot(
	ctx context.Context,
	snapshotFID string,
	jobConfig ExportK8sProtectionSetSnapshotJobConfig,
) (AsyncRequestStatus, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		exportK8sProtectionSetSnapshotQuery,
		struct {
			SnapshotFID string                                  `json:"id"`
			JobConfig   ExportK8sProtectionSetSnapshotJobConfig `json:"jobConfig"`
		}{
			SnapshotFID: snapshotFID,
			JobConfig:   jobConfig,
		},
	)

	if err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to request exportK8sProtectionSetSnapshot: %w",
			err,
		)
	}
	a.log.Printf(
		log.Debug,
		"exportK8sProtectionSetSnapshot(%q, %q): %s",
		snapshotFID,
		jobConfig,
		string(buf),
	)

	var payload struct {
		Data struct {
			Response AsyncRequestStatus `json:"exportK8sProtectionSetSnapshot"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to unmarshal exportK8sProtectionSetSnapshot response: %v",
			err,
		)
	}

	return payload.Data.Response, nil
}

// RestoreK8sProtectionSetSnapshot takes a snapshot FID, the restore job config
// and starts an on-demand restore job in CDM.
func (a API) RestoreK8sProtectionSetSnapshot(
	ctx context.Context,
	snapshotFID string,
	jobConfig RestoreK8sProtectionSetSnapshotJobConfig,
) (AsyncRequestStatus, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		restoreK8sProtectionSetSnapshotQuery,
		struct {
			SnapshotFID string                                   `json:"id"`
			JobConfig   RestoreK8sProtectionSetSnapshotJobConfig `json:"jobConfig"`
		}{
			SnapshotFID: snapshotFID,
			JobConfig:   jobConfig,
		},
	)

	if err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to request restoreK8sProtectionSetSnapshot: %w",
			err,
		)
	}
	a.log.Printf(
		log.Debug,
		"restoreK8sProtectionSetSnapshot(%q, %q): %s",
		snapshotFID,
		jobConfig,
		string(buf),
	)

	var payload struct {
		Data struct {
			Response AsyncRequestStatus `json:"restoreK8sProtectionSetSnapshot"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to unmarshal restoreK8sProtectionSetSnapshot response: %v",
			err,
		)
	}

	return payload.Data.Response, nil
}

// CreateK8sProtectionSetSnapshot takes a ProtectionSetFID, the snapshot job
// config and starts an on-demand snapshot job in CDM.
func (a API) CreateK8sProtectionSetSnapshot(
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
			"failed to request createK8sProtectionSetSnapshot: %w",
			err,
		)
	}
	a.log.Printf(
		log.Debug,
		"createK8sProtectionSetSnapshot(%q, %q): %s",
		protectionSetFID,
		jobConfig,
		string(buf),
	)

	var payload struct {
		Data struct {
			Response AsyncRequestStatus `json:"createK8sProtectionSetSnapshot"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to unmarshal createK8sProtectionSetSnapshot response: %v",
			err,
		)
	}

	return payload.Data.Response, nil
}

// K8sObjectFIDByType fetches the RSC FID for the object corresponding to the
// provided internal id, CDM cluster id and object type.
func (a API) K8sObjectFIDByType(
	ctx context.Context,
	internalID uuid.UUID,
	cdmClusterID uuid.UUID,
	objectType K8sObjectType,
) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		k8sObjectFidByTypeQuery,
		struct {
			InternalID    uuid.UUID     `json:"k8SObjectInternalIdArg"`
			ClusterUUID   uuid.UUID     `json:"clusterUuid"`
			K8sObjectType K8sObjectType `json:"kubernetesObjectType"`
		}{
			InternalID:    internalID,
			ClusterUUID:   cdmClusterID,
			K8sObjectType: objectType,
		},
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to request k8sObjectFidByType: %w", err)
	}
	a.log.Printf(
		log.Debug,
		"k8sObjectFidByType(%v, %v, %v): %s",
		internalID,
		cdmClusterID,
		objectType,
		string(buf),
	)

	var payload struct {
		Data struct {
			ObjectFID uuid.UUID `json:"k8sObjectFidByType"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf(
			"failed to unmarshal k8sObjectFidByType: %v",
			err,
		)
	}

	return payload.Data.ObjectFID, nil
}

// K8sObjectInternalIDByType fetches the object Internal ID on CDM for the
// given RSC FID by type.
func (a API) K8sObjectInternalIDByType(
	ctx context.Context,
	fid uuid.UUID,
	clusterUUID uuid.UUID,
	objectType K8sObjectType,
) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		k8sObjectInternalIdByTypeQuery,
		struct {
			ClusterUUID          uuid.UUID     `json:"clusterUuid"`
			FID                  uuid.UUID     `json:"fid"`
			KubernetesObjectType K8sObjectType `json:"kubernetesObjectType"`
		}{
			FID:                  fid,
			ClusterUUID:          clusterUUID,
			KubernetesObjectType: objectType,
		},
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf(
			"failed to request k8sObjectInternalIdByType: %w",
			err,
		)
	}
	a.log.Printf(
		log.Debug,
		"k8sObjectInternalIdByType(%v, %v, %v): %s",
		fid,
		clusterUUID,
		objectType,
		string(buf),
	)

	var payload struct {
		Data struct {
			ObjectInternalID uuid.UUID `json:"k8sObjectInternalIdByType"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf(
			"failed to unmarshal k8sObjectInternalIdByType: %v",
			err,
		)
	}

	return payload.Data.ObjectInternalID, nil
}

// protectionSetSnapshots get initial snapshots for a given fid. Only used
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
			"failed to request k8sObjectInternalId: %w",
			err,
		)
	}
	a.log.Printf(log.Debug, "k8sProtectionSetSnapshots(%v): %s", fid, string(buf))

	var payload struct {
		Data struct {
			K8sProtectionSetSnapshots struct {
				Data []struct {
					BaseSnapshotSummary BaseSnapshotSummary `json:"baseSnapshotSummary"`
				} `json:"data"`
			} `json:"k8sProtectionSetSnapshots"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal k8sProtectionSetSnapshots: %v",
			err,
		)
	}
	snaps := make([]string, len(payload.Data.K8sProtectionSetSnapshots.Data))
	for i, item := range payload.Data.K8sProtectionSetSnapshots.Data {
		snaps[i] = item.BaseSnapshotSummary.ID
	}
	return snaps, nil
}

// ActivitySeries fetches the activity series for the
// given activity series id.
func (a API) ActivitySeries(
	ctx context.Context,
	activitySeriesID uuid.UUID,
	cdmClusterID uuid.UUID,
) ([]ActivitySeries, error) {

	var ret []ActivitySeries
	// Currently the pagination is handled within the call, but we may want to
	// expose that outside the API.
	var cursor string
	for {
		a.GQL.Log().Print(log.Info, "polaris/graphql/k8s.getActivitySeries")
		buf, err := a.GQL.Request(
			ctx,
			activitySeriesQuery,
			struct {
				Input ActivitySeriesInput `json:"input"`
				After string              `json:"after,omitempty"`
			}{
				Input: ActivitySeriesInput{
					ActivitySeriesID: activitySeriesID, CDMClusterUUID: cdmClusterID,
				},
				After: cursor,
			},
		)
		if err != nil {
			return nil, err
		}
		var payload struct {
			Data struct {
				ActivitySeriesData struct {
					ActivityConnection struct {
						Nodes    []ActivitySeries `json:"nodes"`
						PageInfo struct {
							EndCursor   string `json:"endCursor"`
							HasNextPage bool   `json:"hasNextPage"`
						} `json:"pageInfo"`
						Count int `json:"count"`
					} `json:"activityConnection"`
				} `json:"activitySeries"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, err
		}

		cursor = payload.Data.ActivitySeriesData.ActivityConnection.PageInfo.EndCursor
		ret = append(
			ret,
			payload.Data.ActivitySeriesData.ActivityConnection.Nodes...,
		)

		if !payload.Data.ActivitySeriesData.ActivityConnection.PageInfo.HasNextPage {
			break
		}
	}
	return ret, nil
}

// UpdateK8sProtectionSet updates the K8s protection set corresponding to fid
// with the config.
func (a API) UpdateK8sProtectionSet(
	ctx context.Context,
	fid string,
	config UpdateK8sProtectionSetConfig,
) (bool, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		updateK8sProtectionSetQuery,
		struct {
			ID           string                       `json:"id"`
			UpdateConfig UpdateK8sProtectionSetConfig `json:"updateConfig"`
		}{
			ID:           fid,
			UpdateConfig: config,
		},
	)
	if err != nil {
		return false, fmt.Errorf(
			"failed to request updateK8sProtectionSet: %w",
			err,
		)
	}
	a.log.Printf(
		log.Debug,
		"updateK8sProtectionSet(%q, %q): %s",
		fid,
		config,
		string(buf),
	)

	var payload struct {
		Data struct {
			ResponseSuccess struct {
				Success bool `json:"success"`
			} `json:"updateK8sProtectionSet"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return false, fmt.Errorf(
			"failed to unmarshal updateK8sProtectionSet: %v",
			err,
		)
	}

	return payload.Data.ResponseSuccess.Success, nil
}

// AddK8sCluster adds the K8s cluster for the given config.
func (a API) AddK8sCluster(
	ctx context.Context,
	cdmClusterID uuid.UUID,
	config K8sClusterAddInput,
) (K8sClusterSummary, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		addK8sClusterQuery,
		struct {
			ClusterUUID string             `json:"clusterUuid"`
			Config      K8sClusterAddInput `json:"config"`
		}{
			ClusterUUID: cdmClusterID.String(),
			Config:      config,
		},
	)
	if err != nil {
		return K8sClusterSummary{}, fmt.Errorf(
			"failed to request addK8sCluster: %w",
			err,
		)
	}
	a.log.Printf(
		log.Debug,
		"addK8sCluster(%s, %q): %s",
		cdmClusterID.String(),
		config,
		string(buf),
	)
	var payload struct {
		Data struct {
			Response K8sClusterSummary `json:"addK8sCluster"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return K8sClusterSummary{}, fmt.Errorf(
			"failed to unmarshal addK8sCluster: %v",
			err,
		)
	}
	return payload.Data.Response, nil
}

// UpdateK8sCluster updates the K8s cluster for the given config.
func (a API) UpdateK8sCluster(
	ctx context.Context,
	k8sClusterFID uuid.UUID,
	config K8sClusterUpdateConfigInput,
) (bool, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		updateK8sClusterQuery,
		struct {
			ID     string                      `json:"id"`
			Config K8sClusterUpdateConfigInput `json:"config"`
		}{
			ID:     k8sClusterFID.String(),
			Config: config,
		},
	)
	if err != nil {
		return false, fmt.Errorf(
			"failed to request updateK8sCluster: %w",
			err,
		)
	}
	a.log.Printf(
		log.Debug,
		"updateK8sCluster(%s, %q): %s",
		k8sClusterFID.String(),
		config,
		string(buf),
	)
	var payload struct {
		Data struct {
			ResponseSuccess struct {
				Success bool `json:"success"`
			} `json:"updateK8sCluster"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return false, fmt.Errorf(
			"failed to unmarshal updateK8sCluster: %v",
			err,
		)
	}
	return payload.Data.ResponseSuccess.Success, nil
}
