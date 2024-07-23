// Copyright 2023 Rubrik, Inc.
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

package aws

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// TargetMapping represents an AWS cloud archival location.
type TargetMapping struct {
	ID               uuid.UUID
	Name             string
	ArchivalGroup    string
	ArchivalTarget   string
	ConnectionStatus string
	BucketPrefix     string
	StorageClass     string
	Region           string
	KMSMasterKey     string
	LocTemplate      string
	BucketTags       map[string]string
}

// TargetMappingByID returns the AWS target mapping with the specified ID.
// If no target mapping with the specified ID is found, graphql.ErrNotFound is
// returned.
func (a API) TargetMappingByID(ctx context.Context, id uuid.UUID) (TargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []aws.TargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_ID",
		Text:  id.String(),
	}}
	targets, err := archival.ListTargetMappings[aws.TargetMapping](ctx, a.client, filter)
	if err != nil {
		return TargetMapping{}, fmt.Errorf("failed to get target mappings: %s", err)
	}

	for _, target := range targets {
		if target.ID == id {
			return toTargetMapping(target), nil
		}
	}

	return TargetMapping{}, fmt.Errorf("target mapping for %q %w", id, graphql.ErrNotFound)
}

// TargetMappingByName returns the AWS target mapping with the specified name.
// If no target mapping with the specified ID is found, graphql.ErrNotFound is
// returned.
func (a API) TargetMappingByName(ctx context.Context, name string) (TargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []aws.TargetMappingFilter{{
		Field: "NAME",
		Text:  name,
	}}
	targets, err := archival.ListTargetMappings[aws.TargetMapping](ctx, a.client, filter)
	if err != nil {
		return TargetMapping{}, fmt.Errorf("failed to get target mappings: %s", err)
	}

	for _, target := range targets {
		if target.Name == name {
			return toTargetMapping(target), nil
		}
	}

	return TargetMapping{}, fmt.Errorf("target mapping for %q %w", name, graphql.ErrNotFound)
}

// TargetMappings returns all AWS target mappings that match the specified
// archival group and name filter. The name filter can be used to search for
// prefixes of a name. If the name filter is empty, is will match all names.
// In RSC cloud, archival locations are also referred to as target mappings.
func (a API) TargetMappings(ctx context.Context, nameFilter string) ([]TargetMapping, error) {
	a.log.Print(log.Trace)

	filter := []aws.TargetMappingFilter{{
		Field: "NAME",
		Text:  nameFilter,
	}}
	targets, err := archival.ListTargetMappings[aws.TargetMapping](ctx, a.client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get target mappings: %s", err)
	}

	targetMappings := make([]TargetMapping, 0, len(targets))
	for _, target := range targets {
		targetMappings = append(targetMappings, toTargetMapping(target))
	}

	return targetMappings, nil
}

// DeleteTargetMapping deletes the target mapping with the specified ID. In RSC
// cloud archival locations are also referred to as target mappings.
func (a API) DeleteTargetMapping(ctx context.Context, id uuid.UUID) error {
	a.log.Print(log.Trace)

	return archival.DeleteTargetMapping(ctx, a.client, id)
}

// CreateStorageSetting creates a cloud native archival location.
// The KMS master key can be either a key alias or a key ID. Region, KMS master
// key and bucket tags are optional.
func (a API) CreateStorageSetting(ctx context.Context, id IdentityFunc, name, bucketPrefix, storageClass, region, kmsMasterKey string, bucketTags map[string]string) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	cloudAccountID, err := a.toCloudAccountID(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}

	var tags *struct {
		TagList []aws.Tag `json:"tagList"`
	}
	if len(bucketTags) > 0 {
		tags = &struct {
			TagList []aws.Tag `json:"tagList"`
		}{TagList: make([]aws.Tag, 0, len(bucketTags))}
		for key, value := range bucketTags {
			tags.TagList = append(tags.TagList, aws.Tag{Key: key, Value: value})
		}
	}

	reg := aws.RegionUnknown
	if region != "" && region != "UNKNOWN_AWS_REGION" {
		reg = aws.ParseRegionNoValidation(region)
	}

	locTemplate := "SPECIFIC_REGION"
	if reg == aws.RegionUnknown {
		locTemplate = "SOURCE_REGION"
	}

	targetMappingID, err := archival.CreateCloudNativeStorageSetting[aws.StorageSettingCreateResult](ctx, a.client, cloudAccountID, aws.StorageSettingCreateParams{
		Name:         name,
		BucketPrefix: bucketPrefix,
		StorageClass: storageClass,
		Region:       reg,
		KmsMasterKey: kmsMasterKey,
		LocTemplate:  locTemplate,
		BucketTags:   tags,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create cloud native storage setting: %s", err)
	}

	return targetMappingID, nil
}

// UpdateStorageSetting updates the cloud native archival location with the
// specified ID. The KMS master key can be either a key alias or a key ID. The
// bucket tags replace all existing tags. Note that not all properties can be
// updated, only the name, storage class, KMS master key and bucket tags can be
// updated.
func (a API) UpdateStorageSetting(ctx context.Context, targetMappingID uuid.UUID, name, storageClass, kmsMasterKey string, bucketTags map[string]string) error {
	a.log.Print(log.Trace)

	var deleteTags bool
	var tags *struct {
		TagList []aws.Tag `json:"tagList"`
	}
	if len(bucketTags) > 0 {
		tags = &struct {
			TagList []aws.Tag `json:"tagList"`
		}{TagList: make([]aws.Tag, 0, len(bucketTags))}
		for key, value := range bucketTags {
			tags.TagList = append(tags.TagList, aws.Tag{Key: key, Value: value})
		}
	} else {
		deleteTags = true
	}

	err := archival.UpdateCloudNativeStorageSetting[aws.StorageSettingUpdateResult](ctx, a.client, targetMappingID, aws.StorageSettingUpdateParams{
		Name:                name,
		StorageClass:        storageClass,
		KmsMasterKey:        kmsMasterKey,
		DeleteAllBucketTags: deleteTags,
		BucketTags:          tags,
	})
	if err != nil {
		return fmt.Errorf("failed to update cloud native storage setting: %s", err)
	}

	return nil
}

// toTargetMapping converts an aws.TargetMapping to a TargetMapping.
func toTargetMapping(target aws.TargetMapping) TargetMapping {
	bucketTags := make(map[string]string, len(target.TargetTemplate.BucketTags))
	for _, tag := range target.TargetTemplate.BucketTags {
		bucketTags[tag.Key] = tag.Value
	}

	region := ""
	if target.TargetTemplate.Region != aws.RegionUnknown {
		region = aws.FormatRegion(target.TargetTemplate.Region)
	}

	return TargetMapping{
		ID:               target.ID,
		Name:             target.Name,
		ArchivalGroup:    target.GroupType,
		ArchivalTarget:   target.TargetType,
		ConnectionStatus: target.ConnectionStatus.Status,
		BucketPrefix:     target.TargetTemplate.BucketPrefix,
		StorageClass:     target.TargetTemplate.StorageClass,
		Region:           region,
		KMSMasterKey:     target.TargetTemplate.KMSMasterKey,
		LocTemplate:      target.TargetTemplate.LocTemplate,
		BucketTags:       bucketTags,
	}
}
