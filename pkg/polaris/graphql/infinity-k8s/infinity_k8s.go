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

// Package infinityk8s provides a low level interface to the Infinity K8s GraphQL queries
// provided by the Polaris platform.
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

type JobInstanceDetail struct {
	Status             string  `json:"status"`
	StartTime          string  `json:"startTime,omitempty"`
	Result             string  `json:"result,omitempty"`
	OpentracingContext string  `json:"opentracingContext,omitempty"`
	NodeId             string  `json:"nodeId"`
	JobType            string  `json:"jobType"`
	JobProgress        float64 `json:"jobProgress,omitempty"`
	IsDisabled         bool    `json:"isDisabled"`
	Id                 string  `json:"id"`
	ErrorInfo          string  `json:"errorInfo,omitempty"`
	EndTime            string  `json:"endTime,omitempty"`
	ChildJobDebugInfo  string  `json:"childJobDebugInfo,omitempty"`
	Archived           bool    `json:"archived"`
}

type AddK8sResourceSetConfig struct {
	Definition            string   `json:"definition"`
	HookConfigs           []string `json:"hookConfigs,omitempty"`
	K8sClusterUuid        string   `json:"k8SClusterUuid,omitempty"`
	K8sNamespace          string   `json:"k8SNamespace,omitempty"`
	KubernetesClusterUuid string   `json:"kubernetesClusterUuid,omitempty"`
	KubernetesNamespace   string   `json:"kubernetesNamespace,omitempty"`
	Name                  string   `json:"name"`
	RSType                string   `json:"rsType"`
}

type AddK8sResourceSetResponse struct {
	Id                    string   `json:"id"`
	Definition            string   `json:"definition"`
	HookConfigs           []string `json:"hookConfigs,omitempty"`
	K8sClusterUuid        string   `json:"k8SClusterUuid,omitempty"`
	K8sNamespace          string   `json:"k8SNamespace,omitempty"`
	KubernetesClusterUuid string   `json:"kubernetesClusterUuid,omitempty"`
	KubernetesNamespace   string   `json:"kubernetesNamespace,omitempty"`
	Name                  string   `json:"name"`
	RSType                string   `json:"rsType"`
}

type ExportK8sResourceSetSnapshotJobConfig struct {
	TargetNamespaceName string `json:"targetNamespaceName"`
	TargetClusterFid    string `json:"targetClusterId"`
	IgnoreErrors        bool   `json:"ignoreErrors,omitempty"`
	Filter              string `json:"filter,omitempty"`
}

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type RequestErrorInfo struct {
	Message string `json:"message"`
}

type AsyncRequestStatus struct {
	Status    string           `json:"status"`
	StartTime time.Time        `json:"startTime,omitempty"`
	Progress  float64          `json:"progress,omitempty"`
	NodeId    string           `json:"nodeId,omitempty"`
	Links     []Link           `json:"links"`
	Id        string           `json:"id"`
	Error     RequestErrorInfo `json:"error,omitempty"`
	EndTime   time.Time        `json:"endTime,omitempty"`
}

type PathNode struct {
	Fid  uuid.UUID `json:"fid"`
	Name string    `json:"name"`
	// ObjectType corresponds to HierarchyObjectTypeEnum.
	ObjectType string `json:"objectType"`
}

type DataLocation struct {
	ClusterUuid uuid.UUID `json:"clusterUuid"`
	CreateDate  string    `json:"createDate"`
	Id          string    `json:"id"`
	IsActive    bool      `json:"isActive"`
	IsArchived  bool      `json:"isArchived"`
	Name        string    `json:"name"`
	// Type corresponds to the DataLocationName enum.
	Type string `json:"type"`
}

type KubernetesResourceSet struct {
	CdmId                       string         `json:"cdmId"`
	ClusterUuid                 string         `json:"clusterUuid"`
	ConfiguredSlaDomain         core.SLADomain `json:"configuredSlaDomain"`
	EffectiveRetentionSlaDomain core.SLADomain `json:"effectiveRetentionSlaDomain,omitempty"`
	EffectiveSlaDomain          core.SLADomain `json:"effectiveSlaDomain"`
	EffectiveSlaSourceObject    PathNode       `json:"effectiveSlaSourceObject"`
	Id                          uuid.UUID      `json:"id"`
	IsRelic                     bool           `json:"isRelic"`
	K8sClusterUuid              uuid.UUID      `json:"k8sClusterUuid"`
	Name                        string         `json:"Name"`
	Namespace                   string         `json:"namespace,omitempty"`
	// ObjectType corresponds to HierarchyObjectTypeEnum.
	ObjectType             string             `json:"objectType"`
	PendingSla             core.SLADomain     `json:"pendingSla,omitempty"`
	PrimaryClusterLocation DataLocation       `json:"primaryClusterLocation"`
	PrimaryClusterUuid     uuid.UUID          `json:"primaryClusterUuid"`
	ReplicatedObjectCount  int                `json:"replicatedObjectCount"`
	RsName                 string             `json:"rsName"`
	RsType                 string             `json:"rsType"`
	SlaAssignment          core.SLAAssignment `json:"slaAssignment"`
	SlaPauseStatus         bool               `json:"slaPauseStatus"`
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

	buf, err := a.GQL.Request(ctx, addK8sResourcesetQuery, struct {
		Config AddK8sResourceSetConfig `json:"config"`
	}{Config: config})
	if err != nil {
		return AddK8sResourceSetResponse{}, fmt.Errorf("failed to request addK8sResourceSet: %w", err)
	}
	a.log.Printf(log.Debug, "addK8sResourceSet(%v): %s", config, string(buf))

	var payload struct {
		Data struct {
			Config AddK8sResourceSetResponse `json:"addK8sResourceSet"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AddK8sResourceSetResponse{}, fmt.Errorf("failed to unmarshal addK8sResourceSet: %v", err)
	}

	return payload.Data.Config, nil
}

// GetK8sResourceSet get the K8s resource set corresponding to the given fid.
func (a API) GetK8sResourceSet(
	ctx context.Context,
	fid uuid.UUID,
) (KubernetesResourceSet, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, k8sResourcesetQuery, struct {
		Fid uuid.UUID `json:"fid"`
	}{Fid: fid})
	if err != nil {
		return KubernetesResourceSet{}, fmt.Errorf("failed to request kubernetesResourceSet: %w", err)
	}
	a.log.Printf(log.Debug, "kubernetesResourceSet(%v): %s", fid, string(buf))

	var payload struct {
		Data struct {
			KubernetesResourceSet KubernetesResourceSet `json:"kubernetesResourceSet"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return KubernetesResourceSet{}, fmt.Errorf("failed to unmarshal kubernetesResourceSet: %v", err)
	}

	return payload.Data.KubernetesResourceSet, nil
}

// DeleteK8sResourceSet deletes the K8s resource set corresponding to the provided fid.
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
			Id                string `json:"id"`
			PreserveSnapshots bool   `json:"preserveSnapshots"`
		}{
			Id:                fid,
			PreserveSnapshots: preserveSnapshots,
		},
	)

	if err != nil {
		return false, fmt.Errorf("failed to request deleteK8sResourceset: %w", err)
	}
	a.log.Printf(log.Debug, "deleteK8sResourceset(%q, %q): %s", fid, preserveSnapshots, string(buf))

	var payload struct {
		Data struct {
			ResponseSuccess struct {
				Success bool `json:"success"`
			} `json:"deleteK8sResourceset"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return false, fmt.Errorf("failed to unmarshal deleteK8sResourceset response: %v", err)
	}
	if !payload.Data.ResponseSuccess.Success {
		return false, fmt.Errorf("failed to delete k8s resource set with fid %q", fid)
	}

	return true, nil
}

// GetJobInstance fetches information about the CDM job corresponding to the given
// jobId and cdmClusterId
func (a API) GetJobInstance(
	ctx context.Context,
	jobId string,
	cdmClusterId string,
) (JobInstanceDetail, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		jobInstanceQuery,
		struct {
			JobId        string `json:"id"`
			CDMClusterId string `json:"clusterUuid"`
		}{
			JobId:        jobId,
			CDMClusterId: cdmClusterId,
		},
	)

	if err != nil {
		return JobInstanceDetail{}, fmt.Errorf("failed to request jobInstance: %w", err)
	}
	a.log.Printf(log.Debug, "jobInstance(%q, %q): %s", jobId, cdmClusterId, string(buf))

	var payload struct {
		Data struct {
			Response JobInstanceDetail `json:"jobInstance"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return JobInstanceDetail{}, fmt.Errorf("failed to unmarshal jobInstance response: %v", err)
	}

	return payload.Data.Response, nil
}

// ExportK8sResourceSetSnapshot takes a snapshot FID, the export job config and starts
// an on-demand export job in CDM.
func (a API) ExportK8sResourceSetSnapshot(
	ctx context.Context,
	snapshotFid string,
	jobConfig ExportK8sResourceSetSnapshotJobConfig,
) (AsyncRequestStatus, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		exportK8sResourcesetSnapshotQuery,
		struct {
			SnapshotFid string                                `json:"id"`
			JobConfig   ExportK8sResourceSetSnapshotJobConfig `json:"jobConfig"`
		}{
			SnapshotFid: snapshotFid,
			JobConfig:   jobConfig,
		},
	)

	if err != nil {
		return AsyncRequestStatus{}, fmt.Errorf("failed to request exportK8sResourceSetSnapshot: %w", err)
	}
	a.log.Printf(log.Debug, "exportK8sResourceSetSnapshot(%q, %q): %s", snapshotFid, jobConfig, string(buf))

	var payload struct {
		Data struct {
			Response AsyncRequestStatus `json:"exportK8sResourceSetSnapshot"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return AsyncRequestStatus{}, fmt.Errorf("failed to unmarshal exportK8sResourceSetSnapshot response: %v", err)
	}

	return payload.Data.Response, nil
}

// GetK8sObjectFid fetches the RSC Fid for the object corresponding to the
// provided internal id and CDM cluster id.
func (a API) GetK8sObjectFid(
	ctx context.Context,
	internalId uuid.UUID,
	cdmClusterId uuid.UUID,
) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		k8sObjectFidQuery,
		struct {
			InternalId  uuid.UUID `json:"k8SObjectInternalIdArg"`
			ClusterUuid uuid.UUID `json:"clusterUuid"`
		}{
			InternalId:  internalId,
			ClusterUuid: cdmClusterId,
		},
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to request k8sObjectFid: %w", err)
	}
	a.log.Printf(log.Debug, "k8sObjectFid(%v, %v): %s", internalId, cdmClusterId, string(buf))

	var payload struct {
		Data struct {
			ObjectFid uuid.UUID `json:"k8sObjectFid"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal k8sObjectFid: %v", err)
	}

	return payload.Data.ObjectFid, nil
}

// GetK8sObjectInternalId fetches the object Internal ID on CDM for the
// given RSC Fid.
func (a API) GetK8sObjectInternalId(
	ctx context.Context,
	fid uuid.UUID,
) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		k8sObjectInternalIdQuery,
		struct {
			Fid uuid.UUID `json:"fid"`
		}{
			Fid: fid,
		},
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to request k8sObjectInternalId: %w", err)
	}
	a.log.Printf(log.Debug, "k8sObjectInternalId(%v): %s", fid, string(buf))

	var payload struct {
		Data struct {
			ObjectInternalId uuid.UUID `json:"k8sObjectInternalId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal k8sObjectInternalId: %v", err)
	}

	return payload.Data.ObjectInternalId, nil
}
