//go:generate go run ../queries_gen.go gcp

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

// Package k8s provides a low level interface to the K8s GraphQL queries
// provided by the Polaris platform.
package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API wraps around GraphQL clients to give them the Polaris K8s API.
type API struct {
	Version string
	GQL     *graphql.Client
}

// Wrap the GraphQL client in the GCP API.
func Wrap(gql *graphql.Client) API {
	return API{Version: gql.Version, GQL: gql}
}

// NewAPI returns a new API instance. Note that this is a very cheap call to
// make.
func NewAPI(gql *graphql.Client) API {
	return API{Version: gql.Version, GQL: gql}
}

type Filter struct {
	Field string   `json:"field"`
	Texts []string `json:"texts"`
}

type ActivitySeriesConnectionFilter struct {
	ObjectType []string   `json:"objectType,omitempty"`
	LastUpdatedGt time.Time   `json:"lastUpdated_gt,omitempty"`
}

type TimeRangeInput struct {
	Start time.Time   `json:"start,omitempty"`
	End time.Time   `json:"end,omitempty"`
}

type PolarisSnapshotFilterInput struct {
	SnappableId string   `json:"snappableId,omitempty"`
	IsOnDemandSnapshot bool   `json:"lastUpdated_gt,omitempty"`
	TimeRange TimeRangeInput   `json:"timeRange,omitempty"`
}

type K8sNamespace struct {
	ID                  uuid.UUID      `json:"id"`
	K8sClusterID        uuid.UUID      `json:"k8sClusterID"`
	NamespaceName       string         `json:"namespaceName"`
	IsRelic             bool           `json:"isRelic"`
	ConfiguredSLADomain core.SLADomain `json:"configuredSlaDomain"`
	EffectiveSLADomain  core.SLADomain `json:"effectiveSlaDomain"`
}

type PvcInformation struct {
	ID	string	`json:"id"`
	Name	string	`json:"name"`
	Capacity	string	`json:"capacity"`
	AccessMode	string	`json:"accessMode"`
	StorageClass	string	`json:"storageClass"`
	Volume	string	`json:"volume"`
	Labels	string	`json:"labels"`
	Phase	string	`json:"phase"`
}

type ActivityConnection struct {
	Message string   `json:"message,omitempty"`
}

type ActivitySeriesConnection struct {

	// Activity series ID.
	// Required: true
	ID int64 `json:"id"`

	// Last activity type of the activity series.
	// Required: true
	LastActivityType string `json:"lastActivityType"`

	// Last activity status of the activity series.
	// Required: true
	LastActivityStatus string `json:"lastActivityStatus"`

	// The severity of the activity series.
	// Required: true
	Severity string `json:"severity"`

	// The ID of the associated object.
	// Required: true
	ObjectID string `json:"objectId"`

	// The name of the associated object.
	ObjectName string `json:"objectName,omitempty"`

	// The type of the associated object.
	// Required: true
	ObjectType string `json:"objectType"`

	// ID of the activity series.
	// Required: true
	ActivitySeriesID string `json:"activitySeriesId"`

	// The progress of the activity series.
	Progress string `json:"progress,omitempty"`

	ActivityConnection struct {
		Node[] ActivityConnection `json:"nodes"`
	} `json:"activityConnection"`
}

type ActivitySeries struct {
	// Activity Info.
	// Required: true
	ActivityInfo string `json:"activityInfo"`

	// Message of the activity series.
	// Required: true
	Message string `json:"message"`

	// Status of the activity series.
	// Required: true
	Status string `json:"status"`

	// The time of the activity series.
	// Required: true
	Time time.Time `json:"time"`

	// Severity of the activity series.
	// Required: true
	Severity string `json:"severity"`
}

type Snapshot struct {
	Id	uuid.UUID `json:"id"`
	Date	time.Time `json:"date"`
	IsOnDemandSnapshot bool `json:"isOnDemandSnapshot"`
	ExpirationDate time.Time `json:"expirationDate"`
	IsCorrupted bool `json:"isCorrupted"`
	IsDeletedFromSource bool `json:"isDeletedFromSource"`
	IsReplicated bool `json:"isReplicated"`
	IsArchived bool `json:"isArchived"`
	IsReplica bool `json:"isReplica"`
	IsExpired bool `json:"isExpired"`
}

type TaskchainInfo struct {
	TaskchainId	string `json:"taskchainId"`
	State string `json:"state"`
	StartTime time.Time `json:"startTime"`
	EndTime time.Time `json:"endTime"`
	Progress int64 `json:"progress"`
	JobId int64 `json:"jobId"`
	JobType string `json:"jobType"`
	Error string `json:"error"`
	Account string `json:"account"`
}

type NamespaceSnapshot struct {
	NamespaceId	uuid.UUID `json:"namespaceId"`
	OnDemandSnapshotSlaId string `json:"onDemandSnapshotSlaId"`
}

type CreateK8sNamespaceSnapshots struct {
	SnapshotInput	[]NamespaceSnapshot `json:"snapshotInput"`
}

type NamespaceRestoreRequest struct {
	SnapshotUUID uuid.UUID `json:"snapshotUUID"`
	TargetClusterUUID uuid.UUID `json:"targetClusterUUID"`
	TargetNamespaceName string `json:"targetNamespaceName"`
}

type NamespaceSnaphotInfo struct {
	TaskchainId	string `json:"taskchainId"`
	JobId int64 `json:"jobId"`
}

func (a API) ListSLA(ctx context.Context) ([]core.GlobalSLA, error) {
	a.GQL.Log().Print(log.Info, "polaris/graphql/k8s.ListSLA")

	slaDomains := make([]core.GlobalSLA, 0, 10)
	cursor := ""
	for {
		buf, err := a.GQL.Request(
			ctx,
			listSlaQuery,
			struct {
				After  string                      `json:"after,omitempty"`
				Filter []core.GlobalSLAFilterInput `json:"filter"`
			}{
				After: cursor,
				Filter: []core.GlobalSLAFilterInput{
					{
						Field: core.ObjectType,
						Text:  "",
						ObjectTypeList: []core.SLAObjectType{
							core.KuprObjectType,
						},
					},
				},
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
						Node core.GlobalSLA `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"globalSlaConnection"`
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

func (a API) ListK8sNamespace(ctx context.Context, clusterID uuid.UUID) ([]K8sNamespace, error) {
	a.GQL.Log().Print(log.Info, "polaris/graphql/k8s.GetKuprNamespaces")

	namespaces := make([]K8sNamespace, 0, 10)
	cursor := ""
	for {
		buf, err := a.GQL.Request(
			ctx,
			getNamespacesQuery,
			struct {
				After     string    `json:"after,omitempty"`
				Filter    []Filter  `json:"filter,omitempty"`
				ClusterID uuid.UUID `json:"clusterID"`
			}{
				After:     cursor,
				Filter:    []Filter{},
				ClusterID: clusterID,
			},
		)
		if err != nil {
			return nil, err
		}

		a.GQL.Log().Printf(log.Debug, "GetNamespaces(%s): %s", clusterID.String(), string(buf))

		var payload struct {
			Data struct {
				Query struct {
					Count int `json:"count"`
					Edges []struct {
						Node K8sNamespace `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"k8sNamespaces"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, err
		}
		for _, edge := range payload.Data.Query.Edges {
			namespaces = append(namespaces, edge.Node)
		}

		if !payload.Data.Query.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Query.PageInfo.EndCursor
	}
	return namespaces, nil
}

func (a API) GetTaskchainInfo(
	ctx context.Context,
	taskchainId string,
	jobType string,
	) (TaskchainInfo, error) {
	a.GQL.Log().Print(log.Info, "polaris/graphql/k8s.getTaskchainInfo")

	buf, err := a.GQL.Request(ctx, getTaskchainInfoQuery, struct {
		TaskchainId     string    `json:"taskchainId"`
		JobType    string  `json:"jobType"`
	}{TaskchainId:	taskchainId, JobType:	jobType})
	if err != nil {
		return TaskchainInfo{}, err
	}

	a.GQL.Log().Printf(
		log.Debug,
		"GetTaskchainInfo for jobType %s and taskchainId %s is %s",
		jobType, taskchainId, string(buf),
		)
	var payload struct {
		Data struct {
			Info TaskchainInfo `json:"getTaskchainInfo"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return TaskchainInfo{}, fmt.Errorf(
			"failed to unmarshal GetTaskchainInfo: %v",
			err,
			)
	}

	return payload.Data.Info, nil
}

func (a API) TakeK8NamespaceSnapshot(
	ctx context.Context,
	namespaceId uuid.UUID,
	onDemandSnapshotSlaId string,
	) (NamespaceSnaphotInfo, error) {
	a.GQL.Log().Print(log.Info, "polaris/graphql/k8s.takeK8NamespaceSnapshot")

	namespaceSnapshot := NamespaceSnapshot{
		NamespaceId: namespaceId,
		OnDemandSnapshotSlaId: onDemandSnapshotSlaId,
	}

	input := CreateK8sNamespaceSnapshots{SnapshotInput: []NamespaceSnapshot{namespaceSnapshot}}
	buf, err := a.GQL.Request(ctx, snapshotK8sNamespaceQuery, struct {
		Input CreateK8sNamespaceSnapshots    `json:"input"`
	}{Input:	input})
	if err != nil {
		return NamespaceSnaphotInfo{}, err
	}

	a.GQL.Log().Printf(
		log.Debug,
		"TakeK8NamespaceSnapshot for namespaceId %s and onDemandSnapshotSlaId %s is %s",
		namespaceId, onDemandSnapshotSlaId, string(buf),
	)
	var payload struct {
		Data struct {
			Info []NamespaceSnaphotInfo `json:"createK8sNamespaceSnapshots"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return NamespaceSnaphotInfo{}, fmt.Errorf(
			"failed to unmarshal TakeK8NamespaceSnapshot: %v",
			err,
		)
	}

	return payload.Data.Info[0], nil
}

func (a API) GetActivitySeriesConnection(
	ctx context.Context,
	objectType []string,
	lastUpdatedGtInUTC time.Time,
	) ([]ActivitySeriesConnection, error) {
	a.GQL.Log().Print(log.Info, "polaris/graphql/k8s.getActivitySeriesConnection")
	activities := make([]ActivitySeriesConnection, 0, 10)
	cursor := ""
	for {
		buf, err := a.GQL.Request(
			ctx,
			getActivitySeriesConnectionQuery,
			struct {
				After     string    `json:"after,omitempty"`
				Filters    ActivitySeriesConnectionFilter  `json:"filters,omitempty"`
			}{
				After:     cursor,
				Filters:    ActivitySeriesConnectionFilter{
					ObjectType: objectType, LastUpdatedGt: lastUpdatedGtInUTC,
				},
			},
		)
		if err != nil {
			return nil, err
		}

		var payload struct {
			Data struct {
				ActivitySeriesConnection struct {
					Count int `json:"count"`
					Edges []struct {
						Node ActivitySeriesConnection `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"activitySeriesConnection"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, err
		}
		for _, edge := range payload.Data.ActivitySeriesConnection.Edges {
			activities = append(activities, edge.Node)
		}

		if !payload.Data.ActivitySeriesConnection.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.ActivitySeriesConnection.PageInfo.EndCursor
	}
	return activities, nil
}

func (a API) RestoreK8NamespaceSnapshot(
	ctx context.Context,
	snapshotUUID uuid.UUID,
	targetClusterUUID uuid.UUID,
	targetNamespaceName string,
) (NamespaceSnaphotInfo, error) {
	a.GQL.Log().Print(log.Info, "polaris/graphql/k8s.RestoreK8NamespaceSnapshot")

	buf, err := a.GQL.Request(ctx, restoreK8sNamespaceQuery, struct {
		K8sNamespaceRestoreRequest NamespaceRestoreRequest    `json:"k8sNamespaceRestoreRequest"`
	}{
		K8sNamespaceRestoreRequest:	NamespaceRestoreRequest{
			SnapshotUUID: snapshotUUID,
			TargetClusterUUID: targetClusterUUID,
			TargetNamespaceName: targetNamespaceName,
		},
	})
	if err != nil {
		return NamespaceSnaphotInfo{}, err
	}

	a.GQL.Log().Printf(
		log.Debug,
		"RestoreK8NamespaceSnapshot for (snapshotUUID: %s), " +
			"(targetClusterUUID: %s) and (targetNamespaceName: %s)",
		snapshotUUID, targetClusterUUID, targetNamespaceName,
	)
	var payload struct {
		Data struct {
			Info NamespaceSnaphotInfo
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return NamespaceSnaphotInfo{}, fmt.Errorf(
			"failed to unmarshal RestoreK8NamespaceSnapshot: %v",
			err,
		)
	}

	return payload.Data.Info, nil
}

func (a API) GetActivitySeries(
	ctx context.Context,
	activitySeriesId uuid.UUID,
	clusterUuid uuid.UUID,
	cursor string,
) ([]ActivitySeries, string, error) {
	a.GQL.Log().Print(log.Info, "polaris/graphql/k8s.getActivitySeries")
	buf, err := a.GQL.Request(
		ctx,
		getActivitySeriesQuery,
		struct {
			ActivitySeriesId     uuid.UUID    `json:"activitySeriesId"`
			ClusterUuid    uuid.UUID  `json:"clusterUuid"`
			After     string    `json:"after,omitempty"`
		}{
			ActivitySeriesId:     activitySeriesId,
			ClusterUuid:    clusterUuid,
			After:     cursor,
		},
	)
	if err != nil {
		return nil, "", err
	}
	var payload struct {
		Data struct {
			ActivitySeriesData struct {
				ActivityConnection struct {
					Nodes []ActivitySeries `json:"nodes"`
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
		return nil, "", err
	}

	if !payload.Data.ActivitySeriesData.ActivityConnection.PageInfo.HasNextPage {
		return payload.Data.ActivitySeriesData.ActivityConnection.Nodes, "", nil
	}
	cursor = payload.Data.ActivitySeriesData.ActivityConnection.PageInfo.EndCursor
	return payload.Data.ActivitySeriesData.ActivityConnection.Nodes, cursor, nil
}

func (a API) GetK8sNamespace(
	ctx context.Context,
	snappableId uuid.UUID,
	startTime time.Time,
	endTime time.Time,
	cursor string,
) ([]Snapshot, string, error) {
	a.GQL.Log().Print(log.Info, "polaris/graphql/k8s.k8sNamespace")
	snapshots := make([]Snapshot, 0, 10)
	buf, err := a.GQL.Request(
		ctx,
		k8sNamespaceQuery,
		struct {
			After     string    `json:"after,omitempty"`
			Filters    PolarisSnapshotFilterInput  `json:"filters,omitempty"`
			Fid    uuid.UUID  `json:"fid,omitempty"`
		}{
			After:     cursor,
			Filters:    PolarisSnapshotFilterInput{
				TimeRange: TimeRangeInput{Start:startTime, End:endTime},
			},
			Fid: snappableId,
		},
	)
	if err != nil {
		return nil, "", err
	}

	var payload struct {
		Data struct {
			K8sNamespace struct {
				SnapshotConnection struct {
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
					Nodes []Snapshot `json:"nodes"`
				} `json:"snapshotConnection"`
			} `json:"k8sNamespace"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, "", err
	}

	for _, s := range payload.Data.K8sNamespace.SnapshotConnection.Nodes {
		snapshots = append(snapshots, s)
	}

	if !payload.Data.K8sNamespace.SnapshotConnection.PageInfo.HasNextPage {
		return snapshots, "", nil
	}
	cursor = payload.Data.K8sNamespace.SnapshotConnection.PageInfo.EndCursor
	return snapshots, cursor, nil
}

func (a API) GetAllSnapshotPVCS(
	ctx context.Context,
	snappableId uuid.UUID,
	snapshotId string,
) ([]PvcInformation, error) {
	a.GQL.Log().Print(log.Info, "polaris/graphql/k8s.allSnapshotPvcs")
	buf, err := a.GQL.Request(
		ctx,
		allSnapshotPvcsQuery,
		struct {
			SnapshotId     string    `json:"snapshotId"`
			SnappableId    uuid.UUID  `json:"snappableId"`
		}{
			SnapshotId:     snapshotId,
			SnappableId:    snappableId,
		},
	)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Data struct {
			PvcInfo[] PvcInformation `json:"allSnapshotPvcs"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	return payload.Data.PvcInfo, nil
}