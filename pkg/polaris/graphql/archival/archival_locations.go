// Copyright 2025 Rubrik, Inc.
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

package archival

import (
	"github.com/google/uuid"
)

// ImmutabilitySettings represents immutability settings for AWS targets.
type ImmutabilitySettings struct {
	LockDurationDays    int  `json:"lockDurationDays"`
	IsObjectLockEnabled bool `json:"isObjectLockEnabled,omitempty"`
}

// S3CompatibleImmutabilitySetting represents immutability settings for S3 compatible targets.
type S3CompatibleImmutabilitySetting struct {
	BucketLockDurationDays int `json:"bucketLockDurationDays"`
}

// ArchivalLocation represents an archival location with cluster information.
type ArchivalLocation struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Cluster struct {
		ID      uuid.UUID `json:"id"`
		Name    string    `json:"name"`
		Status  string    `json:"status"`
		Version string    `json:"version"`
		State   struct {
			ConnectedState string `json:"connectedState"`
		} `json:"state"`
	} `json:"cluster"`
	IsActive   bool   `json:"isActive"`
	Status     string `json:"status"`
	TargetType string `json:"targetType"`

	// Fields for RubrikManagedAwsTarget
	SyncStatus           string                `json:"syncStatus,omitempty"`
	StorageClass         string                `json:"storageClass,omitempty"`
	ImmutabilitySettings *ImmutabilitySettings `json:"immutabilitySettings,omitempty"`

	// Fields for RubrikManagedS3CompatibleTarget
	ImmutabilitySetting *S3CompatibleImmutabilitySetting `json:"immutabilitySetting,omitempty"`

	// Fields for RubrikManagedRcsTarget
	ImmutabilityPeriodDays int    `json:"immutabilityPeriodDays,omitempty"`
	Tier                   string `json:"tier,omitempty"`
}

// ListQuery implements the ListTargetResult interface for ArchivalLocation.
func (r ArchivalLocation) ListQuery(cursor string, filters []ListTargetFilter) (string, any) {
	return targetsQuery, struct {
		After   string             `json:"after,omitempty"`
		Filters []ListTargetFilter `json:"filters,omitempty"`
	}{After: cursor, Filters: filters}
}

// Validate implements the ListTargetResult interface for ArchivalLocation.
func (r ArchivalLocation) Validate() error {
	return nil
}
