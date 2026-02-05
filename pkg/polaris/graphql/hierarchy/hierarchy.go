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
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
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
// It corresponds to values in the GraphQL Enum WorkloadLevelHierarchy.
// An empty string is equivalent to AllSubHierarchyType.
type Workload string

const (
	// WorkloadAllSubHierarchyType represents all workload types. This is the
	// default value when an empty string is passed.
	WorkloadAllSubHierarchyType Workload = "AllSubHierarchyType"

	// WorkloadAzureNativeVirtualMachine represents Azure native virtual machines.
	WorkloadAzureNativeVirtualMachine Workload = "AzureNativeVirtualMachine"

	// WorkloadAzureNativeManagedDisk represents Azure native managed disks.
	WorkloadAzureNativeManagedDisk Workload = "AzureNativeManagedDisk"

	// WorkloadAzureSQLDatabaseDB represents Azure SQL databases.
	WorkloadAzureSQLDatabaseDB Workload = "AzureSqlDatabaseDb"

	// WorkloadAzureSQLManagedInstanceDB represents Azure SQL Managed Instance
	// databases.
	WorkloadAzureSQLManagedInstanceDB Workload = "AzureSqlManagedInstanceDb"

	// WorkloadAzureStorageAccount represents Azure storage accounts.
	WorkloadAzureStorageAccount Workload = "AZURE_STORAGE_ACCOUNT"
)

// SLAObject represents an RSC hierarchy object with SLA information.
type SLAObject struct {
	Object
	SLAAssignment       sla.Assignment `json:"slaAssignment"`
	ConfiguredSLADomain sla.DomainRef  `json:"configuredSlaDomain"`
	EffectiveSLADomain  sla.DomainRef  `json:"effectiveSlaDomain"`
}

// DoNotProtectSLAID is the special SLA domain ID used to indicate that an
// object should not be protected. This is returned in configuredSlaDomain.ID
// when "Do Not Protect" is directly assigned to an object.
const DoNotProtectSLAID = "DO_NOT_PROTECT"

// UnprotectedSLAID is the special SLA domain ID used to indicate that an
// object is unprotected (no SLA assigned). This is returned in
// effectiveSlaDomain.ID when the object inherits no protection.
const UnprotectedSLAID = "UNPROTECTED"

// ObjectByID returns the hierarchy object with the specified ID.
// This can be used to query any hierarchy object (VMs, databases, tag rules,
// etc.) and retrieve its SLA assignment information including the configured
// and effective SLA domains.
func (a API) ObjectByID(ctx context.Context, fid uuid.UUID) (SLAObject, error) {
	a.log.Print(log.Trace)

	query := hierarchyObjectQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		FID uuid.UUID `json:"fid"`
	}{FID: fid})
	if err != nil {
		return SLAObject{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result SLAObject `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return SLAObject{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
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
