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
	EndTime            string  `json:"endTime,omitempty"`
	ChildJobDebugInfo  string  `json:"childJobDebugInfo,omitempty"`
	Archived           bool    `json:"archived"`
}

// AddK8sResourceSetConfig defines the input parameters required to a
// ResourceSet.
type AddK8sResourceSetConfig struct {
	Definition            string   `json:"definition"`
	HookConfigs           []string `json:"hookConfigs,omitempty"`
	K8sClusterUUID        string   `json:"k8SClusterUuid,omitempty"`
	K8sNamespace          string   `json:"k8SNamespace,omitempty"`
	KubernetesClusterUUID string   `json:"kubernetesClusterUuid,omitempty"`
	KubernetesNamespace   string   `json:"kubernetesNamespace,omitempty"`
	Name                  string   `json:"name"`
	RSType                string   `json:"rsType"`
}

// AddK8sResourceSetResponse is the response received from adding a K8s cluster.
type AddK8sResourceSetResponse struct {
	ID                    string   `json:"id"`
	Definition            string   `json:"definition"`
	HookConfigs           []string `json:"hookConfigs,omitempty"`
	K8sClusterUUID        string   `json:"k8SClusterUuid,omitempty"`
	K8sNamespace          string   `json:"k8SNamespace,omitempty"`
	KubernetesClusterUUID string   `json:"kubernetesClusterUuid,omitempty"`
	KubernetesNamespace   string   `json:"kubernetesNamespace,omitempty"`
	Name                  string   `json:"name"`
	RSType                string   `json:"rsType"`
}

// ExportK8sResourceSetSnapshotJobConfig defines parameters required to
// export a snapshot.
type ExportK8sResourceSetSnapshotJobConfig struct {
	TargetNamespaceName string `json:"targetNamespaceName"`
	TargetClusterFID    string `json:"targetClusterId"`
	IgnoreErrors        bool   `json:"ignoreErrors,omitempty"`
	Filter              string `json:"filter,omitempty"`
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
	FID  uuid.UUID `json:"fid"`
	Name string    `json:"name"`
	// ObjectType corresponds to HierarchyObjectTypeEnum.
	ObjectType string `json:"objectType"`
}

// DataLocation defines the primary location of a Snappable.
type DataLocation struct {
	ClusterUUID uuid.UUID `json:"clusterUuid"`
	CreateDate  string    `json:"createDate"`
	ID          string    `json:"id"`
	IsActive    bool      `json:"isActive"`
	IsArchived  bool      `json:"isArchived"`
	Name        string    `json:"name"`
	// Type corresponds to the DataLocationName enum.
	Type string `json:"type"`
}

// KubernetesResourceSet contains fields contained in a ResourceSet snappable.
type KubernetesResourceSet struct {
	CDMID                       string         `json:"cdmId"`
	ClusterUUID                 string         `json:"clusterUuid"`
	ConfiguredSLADomain         core.SLADomain `json:"configuredSlaDomain"`
	EffectiveRetentionSLADomain core.SLADomain `json:"effectiveRetentionSlaDomain,omitempty"`
	EffectiveSLADomain          core.SLADomain `json:"effectiveSlaDomain"`
	EffectiveSLASourceObject    PathNode       `json:"effectiveSlaSourceObject"`
	ID                          uuid.UUID      `json:"id"`
	IsRelic                     bool           `json:"isRelic"`
	K8sClusterUUID              uuid.UUID      `json:"k8sClusterUuid"`
	Name                        string         `json:"Name"`
	Namespace                   string         `json:"namespace,omitempty"`
	// ObjectType corresponds to HierarchyObjectTypeEnum.
	ObjectType             string             `json:"objectType"`
	PendingSLA             core.SLADomain     `json:"pendingSla,omitempty"`
	PrimaryClusterLocation DataLocation       `json:"primaryClusterLocation"`
	PrimaryClusterUUID     uuid.UUID          `json:"primaryClusterUuid"`
	ReplicatedObjectCount  int                `json:"replicatedObjectCount"`
	RSName                 string             `json:"rsName"`
	RSType                 string             `json:"rsType"`
	SLAAssignment          core.SLAAssignment `json:"slaAssignment"`
	SLAPauseStatus         bool               `json:"slaPauseStatus"`
}

// BaseSnapshotSummary picks ID and SLAID fields from the GetSnapshot response.
type BaseSnapshotSummary struct {
	ID    string `json:"id"`
	SLAID string `json:"slaId"`
}

// API wraps around GraphQL clients to give them the RSC Infinity K8s API.
type API struct {
	GQL *graphql.Client
	log log.Logger
}

// Wrap the RSC client in the Infinity K8s API.
func Wrap(client *polaris.Client) API {
	return API{GQL: client.GQL, log: client.GQL.Log()}
}

// AddK8sResourceSet adds the K8s resource set for the given config.
func (a API) AddK8sResourceSet(
	ctx context.Context,
	config AddK8sResourceSetConfig,
) (AddK8sResourceSetResponse, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx, addK8sResourcesetQuery, struct {
			Config AddK8sResourceSetConfig `json:"config"`
		}{Config: config},
	)
	if err != nil {
		return AddK8sResourceSetResponse{}, fmt.Errorf(
			"failed to request addK8sResourceSet: %w",
			err,
		)
	}
	a.log.Printf(log.Debug, "addK8sResourceSet(%v): %s", config, string(buf))

	var payload struct {
		Data struct {
			Config AddK8sResourceSetResponse `json:"addK8sResourceSet"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AddK8sResourceSetResponse{}, fmt.Errorf(
			"failed to unmarshal addK8sResourceSet: %v",
			err,
		)
	}

	return payload.Data.Config, nil
}

// GetK8sResourceSet get the K8s resource set corresponding to the given fid.
func (a API) GetK8sResourceSet(
	ctx context.Context,
	fid uuid.UUID,
) (KubernetesResourceSet, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx, k8sResourcesetQuery, struct {
			FID uuid.UUID `json:"fid"`
		}{FID: fid},
	)
	if err != nil {
		return KubernetesResourceSet{}, fmt.Errorf(
			"failed to request kubernetesResourceSet: %w",
			err,
		)
	}
	a.log.Printf(log.Debug, "kubernetesResourceSet(%v): %s", fid, string(buf))

	var payload struct {
		Data struct {
			KubernetesResourceSet KubernetesResourceSet `json:"kubernetesResourceSet"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return KubernetesResourceSet{}, fmt.Errorf(
			"failed to unmarshal kubernetesResourceSet: %v",
			err,
		)
	}

	return payload.Data.KubernetesResourceSet, nil
}

// DeleteK8sResourceSet deletes the K8s resource set corresponding to the
// provided fid.
func (a API) DeleteK8sResourceSet(
	ctx context.Context,
	fid string,
	preserveSnapshots bool,
) (bool, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		deleteK8sResourcesetQuery,
		struct {
			ID                string `json:"id"`
			PreserveSnapshots bool   `json:"preserveSnapshots"`
		}{
			ID:                fid,
			PreserveSnapshots: preserveSnapshots,
		},
	)

	if err != nil {
		return false, fmt.Errorf("failed to request deleteK8sResourceset: %w", err)
	}
	a.log.Printf(
		log.Debug,
		"deleteK8sResourceset(%q, %q): %s",
		fid,
		preserveSnapshots,
		string(buf),
	)

	var payload struct {
		Data struct {
			ResponseSuccess struct {
				Success bool `json:"success"`
			} `json:"deleteK8sResourceset"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return false, fmt.Errorf(
			"failed to unmarshal deleteK8sResourceset response: %v",
			err,
		)
	}
	if !payload.Data.ResponseSuccess.Success {
		return false, fmt.Errorf(
			"failed to delete k8s resource set with fid %q",
			fid,
		)
	}

	return true, nil
}

// GetJobInstance fetches information about the CDM job corresponding to the
// given jobID and cdmClusterID.
func (a API) GetJobInstance(
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
			"failed to request jobInstance: %w",
			err,
		)
	}
	a.log.Printf(
		log.Debug,
		"jobInstance(%q, %q): %s",
		jobID,
		cdmClusterID,
		string(buf),
	)

	var payload struct {
		Data struct {
			Response JobInstanceDetail `json:"jobInstance"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return JobInstanceDetail{}, fmt.Errorf(
			"failed to unmarshal jobInstance response: %v",
			err,
		)
	}

	return payload.Data.Response, nil
}

// ExportK8sResourceSetSnapshot takes a snapshot FID, the export job config and
// starts an on-demand export job in CDM.
func (a API) ExportK8sResourceSetSnapshot(
	ctx context.Context,
	snapshotFID string,
	jobConfig ExportK8sResourceSetSnapshotJobConfig,
) (AsyncRequestStatus, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		exportK8sResourcesetSnapshotQuery,
		struct {
			SnapshotFID string                                `json:"id"`
			JobConfig   ExportK8sResourceSetSnapshotJobConfig `json:"jobConfig"`
		}{
			SnapshotFID: snapshotFID,
			JobConfig:   jobConfig,
		},
	)

	if err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to request exportK8sResourceSetSnapshot: %w",
			err,
		)
	}
	a.log.Printf(
		log.Debug,
		"exportK8sResourceSetSnapshot(%q, %q): %s",
		snapshotFID,
		jobConfig,
		string(buf),
	)

	var payload struct {
		Data struct {
			Response AsyncRequestStatus `json:"exportK8sResourceSetSnapshot"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to unmarshal exportK8sResourceSetSnapshot response: %v",
			err,
		)
	}

	return payload.Data.Response, nil
}

// CreateK8sResourceSnapshot takes a ResourceSetFID, the snapshot job config and
// starts an on-demand snapshot job in CDM.
func (a API) CreateK8sResourceSnapshot(
	ctx context.Context,
	resourceSetFID string,
	jobConfig BaseOnDemandSnapshotConfigInput,
) (AsyncRequestStatus, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		createK8sResourceSnapshotQuery,
		struct {
			ResourceSetID string                          `json:"resourceSetId"`
			JobConfig     BaseOnDemandSnapshotConfigInput `json:"jobConfig"`
		}{
			ResourceSetID: resourceSetFID,
			JobConfig:     jobConfig,
		},
	)

	if err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to request createK8sResourceSetSnapshot: %w",
			err,
		)
	}
	a.log.Printf(
		log.Debug,
		"createK8sResourceSetSnapshot(%q, %q): %s",
		resourceSetFID,
		jobConfig,
		string(buf),
	)

	var payload struct {
		Data struct {
			Response AsyncRequestStatus `json:"createK8sResourceSetSnapshot"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return AsyncRequestStatus{}, fmt.Errorf(
			"failed to unmarshal createK8sResourceSetSnapshot response: %v",
			err,
		)
	}

	return payload.Data.Response, nil
}

// GetK8sObjectFID fetches the RSC FID for the object corresponding to the
// provided internal id and CDM cluster id.
func (a API) GetK8sObjectFID(
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
		return uuid.Nil, fmt.Errorf("failed to request k8sObjectFid: %w", err)
	}
	a.log.Printf(
		log.Debug,
		"k8sObjectFid(%v, %v): %s",
		internalID,
		cdmClusterID,
		string(buf),
	)

	var payload struct {
		Data struct {
			ObjectFID uuid.UUID `json:"k8sObjectFid"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal k8sObjectFid: %v", err)
	}

	return payload.Data.ObjectFID, nil
}

// GetK8sObjectInternalID fetches the object Internal ID on CDM for the
// given RSC FID.
func (a API) GetK8sObjectInternalID(
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
			"failed to request k8sObjectInternalId: %w",
			err,
		)
	}
	a.log.Printf(log.Debug, "k8sObjectInternalId(%v): %s", fid, string(buf))

	var payload struct {
		Data struct {
			ObjectInternalID uuid.UUID `json:"k8sObjectInternalId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf(
			"failed to unmarshal k8sObjectInternalId: %v",
			err,
		)
	}

	return payload.Data.ObjectInternalID, nil
}

// getResourceSetSnapshots get initial snapshots for a given fid. Only used
// internally for testing puporses.
func (a API) getResourceSetSnapshots(
	ctx context.Context,
	fid string,
) ([]string, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		getResourcesetSnapshotQuery,
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
	a.log.Printf(log.Debug, "k8sResourceSetSnapshots(%v): %s", fid, string(buf))

	var payload struct {
		Data struct {
			K8sResourceSnapshots struct {
				Data []struct {
					BaseSnapshotSummary BaseSnapshotSummary `json:"baseSnapshotSummary"`
				} `json:"data"`
			} `json:"k8sResourceSetSnapshots"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal k8sResourceSetSnapshots: %v",
			err,
		)
	}
	snaps := make([]string, len(payload.Data.K8sResourceSnapshots.Data))
	for i, item := range payload.Data.K8sResourceSnapshots.Data {
		snaps[i] = item.BaseSnapshotSummary.ID
	}
	return snaps, nil
}
