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

package devops

import (
	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// QueryType selects the hierarchy traversal mode for the list queries. It maps
// to the schema QueryType enum.
type QueryType string

const (
	// QueryTypeChildren returns only the direct children of the ancestor.
	QueryTypeChildren QueryType = "CHILDREN"
	// QueryTypeDescendants returns all descendants of the ancestor.
	QueryTypeDescendants QueryType = "DESCENDANTS"
)

// BackupLocation represents the backup location associated with a DevOps
// organization. Nil on the organization when no backup location is configured.
type BackupLocation struct {
	ID              uuid.UUID   `json:"id"`
	ArchivalGroupID uuid.UUID   `json:"archivalGroupId"`
	Name            string      `json:"name"`
	StorageType     StorageType `json:"storageType"`
	Region          struct {
		AzureRegion azure.RegionEnum `json:"azureRegion"`
	} `json:"cloudSpecificRegion"`
}

// CloudNativeExocompute represents the customer cloud-native exocompute
// associated with a DevOps organization. Nil on the organization when no
// cloud-native exocompute is configured.
type CloudNativeExocompute struct {
	ID       uuid.UUID `json:"id"`
	HostName string    `json:"hostName"`
	Region   struct {
		Region struct {
			AzureRegion azure.CommonRegionEnum `json:"azureRegion"`
		} `json:"region"`
	} `json:"region"`
}

// RubrikHostedExocompute represents the Rubrik-hosted exocompute associated
// with a DevOps organization. Nil on the organization when no Rubrik-hosted
// exocompute is configured.
type RubrikHostedExocompute struct {
	ExocomputeClusterID uuid.UUID    `json:"exocomputeClusterId"`
	Region              azure.Region `json:"region"`
}
