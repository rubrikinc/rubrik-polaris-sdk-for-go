//go:generate go run ../queries_gen.go hierarchy

// Copyright 2026 Rubrik, Inc.
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

// Package hierarchy provides a low-level interface to hierarchy GraphQL queries
// provided by the Polaris platform.
package hierarchy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API wraps around GraphQL clients to give them the Polaris Hierarchy API.
type API struct {
	GQL *graphql.Client
	log log.Logger
}

// Wrap the GraphQL client in the Hierarchy API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

// ObjectType represents the type of a hierarchy object.
type ObjectType string

// Workload represents the workload hierarchy type for SLA Domain assignments.
// A zero value is equivalent to WorkloadAllSubHierarchyType. Use ToWorkload to
// convert a string to a Workload constant. Custom JSON marshaling converts
// between the Workload constants and the GraphQL WorkloadLevelHierarchy enum
// values.
type Workload int

const (
	// WorkloadAllSubHierarchyType represents all workload types. This is the
	// default value.
	WorkloadAllSubHierarchyType Workload = iota

	// WorkloadAzureVM represents Azure native virtual machines.
	WorkloadAzureVM

	// WorkloadAzureManagedDisk represents Azure native managed disks.
	WorkloadAzureManagedDisk

	// WorkloadAzureSQLDB represents Azure SQL databases.
	WorkloadAzureSQLDB

	// WorkloadAzureSQLMIDB represents Azure SQL Managed Instance databases.
	WorkloadAzureSQLMIDB

	// WorkloadAzureStorageAccount represents Azure storage accounts.
	WorkloadAzureStorageAccount
)

// workloadMap maps Workload values to their GraphQL and string
// representations.
var workloadMap = map[Workload]struct {
	GraphQL string
	Value   string
}{
	WorkloadAllSubHierarchyType: {GraphQL: "AllSubHierarchyType", Value: "ALL_SUB_HIERARCHY_TYPE"},
	WorkloadAzureVM:             {GraphQL: "AzureNativeVirtualMachine", Value: "AZURE_NATIVE_VIRTUAL_MACHINE"},
	WorkloadAzureManagedDisk:    {GraphQL: "AzureNativeManagedDisk", Value: "AZURE_NATIVE_MANAGED_DISK"},
	WorkloadAzureSQLDB:          {GraphQL: "AzureSqlDatabaseDb", Value: "AZURE_SQL_DATABASE_DB"},
	WorkloadAzureSQLMIDB:        {GraphQL: "AzureSqlManagedInstanceDb", Value: "AZURE_SQL_MANAGED_INSTANCE_DB"},
	WorkloadAzureStorageAccount: {GraphQL: "AZURE_STORAGE_ACCOUNT", Value: "AZURE_STORAGE_ACCOUNT"},
}

// ToWorkload converts a string to a Workload constant.
func ToWorkload(s string) (Workload, error) {
	for workload, strings := range workloadMap {
		if strings.Value == s {
			return workload, nil
		}
	}
	return 0, fmt.Errorf("unknown workload string: %s", s)
}

// AllWorkloadsAsStrings returns a slice of all valid workload string values.
func AllWorkloadsAsStrings() []string {
	strs := make([]string, 0, len(workloadMap))
	for _, s := range workloadMap {
		strs = append(strs, s.Value)
	}
	return strs
}

// String returns the string representation of the Workload constant.
func (w Workload) String() string {
	if s, ok := workloadMap[w]; ok {
		return s.Value
	}
	panic(fmt.Sprintf("unknown workload value: %d", w))
}

// MarshalJSON converts the Workload constant to its JSON string value.
func (w Workload) MarshalJSON() ([]byte, error) {
	if s, ok := workloadMap[w]; ok {
		return json.Marshal(s.GraphQL)
	}
	return nil, fmt.Errorf("unknown workload value: %d", w)
}

// UnmarshalJSON converts a JSON string value to the Workload constant.
func (w *Workload) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	for workload, strings := range workloadMap {
		if strings.GraphQL == s {
			*w = workload
			return nil
		}
	}
	return fmt.Errorf("unknown workload string: %s", s)
}

// Status represents the status of a cloud account or feature.
type Status string

const (
	StatusRefreshing     Status = "REFRESHING"
	StatusRefreshFailed  Status = "REFRESH_FAILED"
	StatusAdded          Status = "ADDED"
	StatusDeletionFailed Status = "DELETION_FAILED"
	StatusRefreshed      Status = "REFRESHED"
	StatusDeleting       Status = "DELETING"
	StatusDeleted        Status = "DELETED"
	StatusDisconnected   Status = "DISCONNECTED"
)

// Feature represents a feature enabled for a cloud account.
type Feature struct {
	Name            string    `json:"featureName"`
	Status          Status    `json:"status"`
	LastRefreshedAt time.Time `json:"lastRefreshedAt"`
}

// InventoryObject is a constraint for types that can be returned from the
// inventory root query.
type InventoryObject interface {
	AWSNativeAccount | AzureNativeSubscription
	// typeFilter returns the object type filter to use for the inventory root
	// query. It corresponds to values in the GraphQL Enum HierarchyObjectTypeEnum.
	typeFilter() ObjectType
}

// Object contains the common fields for all hierarchy objects.
type Object struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	ObjectType ObjectType `json:"objectType"`
}

// AWSNativeAccount represents an AWS native account from the inventory root.
type AWSNativeAccount struct {
	Object
	Status   Status    `json:"status"`
	Features []Feature `json:"awsFeatures"`
}

func (AWSNativeAccount) typeFilter() ObjectType {
	return "AwsNativeAccount"
}

// AzureNativeSubscription represents an Azure native subscription from the
// inventory root.
type AzureNativeSubscription struct {
	Object
	Status   Status    `json:"azureSubscriptionStatus"`
	NativeID uuid.UUID `json:"azureSubscriptionNativeId"`
	Features []Feature `json:"azureFeatures"`
}

func (AzureNativeSubscription) typeFilter() ObjectType {
	return "AzureNativeSubscription"
}

// ObjectsByName returns hierarchy objects from the inventory root matching
// the specified exact name. The type parameter determines which object type
// to query for.
//
// The workloadHierarchy parameter specifies which workload type to use for SLA
// Domain resolution. This affects the effective SLA Domain returned for objects.
// Pass an empty string to use AllSubHierarchyType (the default), which returns
// a generic view without workload-specific SLA assignments. For Azure Native
// workloads, use one of the Workload constants (e.g., WorkloadAzureNativeVirtualMachine)
// to get the correct effective SLA for that specific workload type.
//
// Note that name isn't unique for all types, so the query can return multiple
// objects with the same name. An example of this is AWS accounts that return
// multiple objects for the same account if it has been added to RSC multiple
// times. In that case the caller must inspect the features of the returned
// objects to determine which one to use.
func ObjectsByName[T InventoryObject](ctx context.Context, a API, name string, workloadHierarchy Workload) ([]T, error) {
	a.log.Print(log.Trace)

	var zero T
	typeFilter := zero.typeFilter()

	var objects []T
	var cursor string
	for {
		query := inventoryRootQuery
		buf, err := a.GQL.Request(ctx, query, struct {
			After             string           `json:"after,omitempty"`
			Filter            []map[string]any `json:"filter,omitempty"`
			First             int              `json:"first,omitempty"`
			TypeFilter        []ObjectType     `json:"typeFilter,omitempty"`
			WorkloadHierarchy Workload         `json:"workloadHierarchy,omitempty"`
		}{
			After: cursor,
			Filter: []map[string]any{
				{"field": "NAME_EXACT_MATCH", "texts": []string{name}},
			},
			First:             100,
			TypeFilter:        []ObjectType{typeFilter},
			WorkloadHierarchy: workloadHierarchy,
		})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					DescendantConnection struct {
						Count    int `json:"count"`
						Nodes    []T `json:"nodes"`
						PageInfo struct {
							EndCursor   string `json:"endCursor"`
							HasNextPage bool   `json:"hasNextPage"`
						} `json:"pageInfo"`
					} `json:"descendantConnection"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, graphql.UnmarshalError(query, err)
		}
		objects = append(objects, payload.Data.Result.DescendantConnection.Nodes...)

		if !payload.Data.Result.DescendantConnection.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.DescendantConnection.PageInfo.EndCursor
	}

	return objects, nil
}

// ObjectByIDAndWorkload returns the hierarchy object with the specified ID
// and workload hierarchy type. The type parameter T allows callers to specify
// a custom return type that can include additional fields beyond the base
// Object fields.
//
// This can be used to query any hierarchy object (VMs, databases, tag rules,
// etc.) and retrieve its information. The workloadHierarchy parameter
// determines which workload type to use for resolution.
func ObjectByIDAndWorkload[T any](ctx context.Context, gql *graphql.Client, fid uuid.UUID, workloadHierarchy Workload) (T, error) {
	gql.Log().Print(log.Trace)

	var zero T
	query := hierarchyObjectQuery
	buf, err := gql.Request(ctx, query, struct {
		FID               uuid.UUID `json:"fid"`
		WorkloadHierarchy Workload  `json:"workloadHierarchy,omitempty"`
	}{FID: fid, WorkloadHierarchy: workloadHierarchy})
	if err != nil {
		return zero, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result T `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return zero, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}
