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

package azure

import (
	"context"
	"fmt"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

type options struct {
	name                string
	regions             []azure.Region
	resourceGroup       *azure.ResourceGroup
	featureSpecificInfo *azure.FeatureSpecificInfo
}

// OptionFunc gives the value passed to the function creating the OptionFunc
// to the specified options instance.
type OptionFunc func(ctx context.Context, opts *options) error

// Name returns an OptionFunc that gives the specified name to the options
// instance.
func Name(name string) OptionFunc {
	return func(ctx context.Context, opts *options) error {
		opts.name = name
		return nil
	}
}

// Region returns an OptionFunc that gives the specified region to the options
// instance.
func Region(region string) OptionFunc {
	return func(ctx context.Context, opts *options) error {
		r, err := azure.ParseRegion(region)
		if err != nil {
			return fmt.Errorf("failed to parse region: %v", err)
		}

		opts.regions = append(opts.regions, r)
		return nil
	}
}

// Regions returns an OptionFunc that gives the specified regions to the
// options instance.
func Regions(regions ...string) OptionFunc {
	return func(ctx context.Context, opts *options) error {
		set := make(map[azure.Region]struct{}, len(regions)+len(opts.regions))
		for _, region := range opts.regions {
			set[region] = struct{}{}
		}

		for _, r := range regions {
			region, err := azure.ParseRegion(r)
			if err != nil {
				return fmt.Errorf("failed to parse region: %v", err)
			}

			if _, ok := set[region]; !ok {
				opts.regions = append(opts.regions, region)
				set[region] = struct{}{}
			}
		}

		return nil
	}
}

// ResourceGroup returns an OptionFunc that gives the specified resource group
// to the options instance. This is currently only needed for archival feature.
// Support will be added for other features soon.
func ResourceGroup(name, region string, tags map[string]string) OptionFunc {
	return func(ctx context.Context, opts *options) error {
		if name == "" {
			return fmt.Errorf("invalid name for resource group")
		}
		r, err := azure.ParseRegion(region)
		if err != nil {
			return fmt.Errorf("failed to parse region: %v", err)
		}

		tagList := make([]azure.Tag, 0, len(tags))
		for key, value := range tags {
			tagList = append(tagList, azure.Tag{Key: key, Value: value})
		}

		opts.resourceGroup = &azure.ResourceGroup{
			Name:    name,
			TagList: &azure.TagList{Tags: tagList},
			Region:  r,
		}
		return nil
	}
}

// ManagedIdentity returns an OptionFunc that gives the specified managed
// identity to the options instance. This is currently only needed for archival
// encryption feature.
func ManagedIdentity(name, resourceGroup, principalID, region string) OptionFunc {
	return func(ctx context.Context, opts *options) error {
		if name == "" {
			return fmt.Errorf("invalid name for managed identity")
		}
		if resourceGroup == "" {
			return fmt.Errorf("invalid resource group for managed identity")
		}
		if principalID == "" {
			return fmt.Errorf("invalid principal ID for managed identity")
		}

		r, err := azure.ParseRegion(region)
		if err != nil {
			return fmt.Errorf("failed to parse region: %v", err)
		}

		if opts.featureSpecificInfo == nil {
			opts.featureSpecificInfo = &azure.FeatureSpecificInfo{}
		}
		opts.featureSpecificInfo.UserAssignedManagedIdentity = &azure.UserAssignedManagedIdentity{
			Name:              name,
			ResourceGroupName: resourceGroup,
			PrincipalID:       principalID,
			Region:            r,
		}
		return nil
	}
}

func verifyOptionsForFeature(opts options, feature core.Feature) error {
	switch {
	case feature.Equal(core.FeatureCloudNativeArchivalEncryption):
		if opts.featureSpecificInfo == nil ||
			opts.featureSpecificInfo.UserAssignedManagedIdentity == nil {
			return fmt.Errorf("managed identity should be added for archival encryption")
		}
		if opts.resourceGroup == nil {
			return fmt.Errorf("resource group should be added for archival encryption")
		}
	case feature.Equal(core.FeatureCloudNativeArchival):
		if opts.resourceGroup == nil {
			return fmt.Errorf("resource group should be added for archival")
		}
	}
	return nil
}
