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
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// GCPTargetMappingByID returns the GCP target mapping with the specified ID.
// If no target mapping with the specified ID is found, graphql.ErrNotFound is
// returned.
func (a API) GCPTargetMappingByID(ctx context.Context, targetMappingID uuid.UUID) (archival.GCPTargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []archival.ListTargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_ID",
		Text:  targetMappingID.String(),
	}}
	targets, err := archival.ListTargetMappings[archival.GCPTargetMapping](ctx, a.client, filter)
	if err != nil {
		return archival.GCPTargetMapping{}, fmt.Errorf("failed to get target mappings: %s", err)
	}

	for _, target := range targets {
		if target.ID == targetMappingID {
			return target, nil
		}
	}

	return archival.GCPTargetMapping{}, fmt.Errorf("target mapping for %q %w", targetMappingID, graphql.ErrNotFound)
}

// GCPTargetMappingByName returns the GCP target mapping with the specified
// name. If no target mapping with the specified name is found,
// graphql.ErrNotFound is returned.
func (a API) GCPTargetMappingByName(ctx context.Context, name string) (archival.GCPTargetMapping, error) {
	a.log.Print(log.Trace)

	targets, err := a.GCPTargetMappings(ctx, name)
	if err != nil {
		return archival.GCPTargetMapping{}, err
	}

	name = strings.ToLower(name)
	for _, target := range targets {
		if strings.ToLower(target.Name) == name {
			return target, nil
		}
	}

	return archival.GCPTargetMapping{}, fmt.Errorf("target mapping for %q %w", name, graphql.ErrNotFound)
}

// GCPTargetMappings returns all GCP target mappings matching the specified
// name filter. The name filter can be used to search for prefixes of a name.
// If the name filter is empty, it will match all names.
func (a API) GCPTargetMappings(ctx context.Context, nameFilter string) ([]archival.GCPTargetMapping, error) {
	a.log.Print(log.Trace)

	var filter []archival.ListTargetMappingFilter
	if nameFilter != "" {
		filter = append(filter, archival.ListTargetMappingFilter{
			Field: "NAME",
			Text:  nameFilter,
		})
	}
	targets, err := archival.ListTargetMappings[archival.GCPTargetMapping](ctx, a.client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get target mappings: %s", err)
	}

	return targets, nil
}

// CreateGCPStorageSetting creates a cloud native archival location.
func (a API) CreateGCPStorageSetting(ctx context.Context, createParams archival.CreateGCPStorageSettingParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if createParams.LocTemplate == "" {
		createParams.LocTemplate = "SPECIFIC_REGION"
		if createParams.Region.Region == gcp.RegionUnknown {
			createParams.LocTemplate = "SOURCE_REGION"
		}
	}

	targetMappingID, err := archival.CreateCloudNativeStorageSetting[archival.CreateGCPStorageSettingResult](ctx, a.client, createParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create cloud native storage setting: %s", err)
	}

	return targetMappingID, nil
}

// UpdateGCPStorageSetting updates the cloud native archival location with the
// specified target mapping ID.
func (a API) UpdateGCPStorageSetting(ctx context.Context, targetMappingID uuid.UUID, updateParams archival.UpdateGCPStorageSettingParams) error {
	a.log.Print(log.Trace)

	err := archival.UpdateCloudNativeStorageSetting[archival.UpdateGCPStorageSettingResult](ctx, a.client, targetMappingID, updateParams)
	if err != nil {
		return fmt.Errorf("failed to update cloud native storage setting: %s", err)
	}

	return nil
}
