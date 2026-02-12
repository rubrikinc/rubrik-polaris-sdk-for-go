// Copyright 2024 Rubrik, Inc.
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

package sla

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// TagRuleByID returns the tag rule with the specified ID.
func (a API) TagRuleByID(ctx context.Context, tagRuleID uuid.UUID) (sla.TagRule, error) {
	a.log.Print(log.Trace)

	tagRules, err := a.TagRules(ctx, "")
	if err != nil {
		return sla.TagRule{}, err
	}

	for _, tagRule := range tagRules {
		if tagRule.ID == tagRuleID {
			return tagRule, nil
		}
	}

	return sla.TagRule{}, fmt.Errorf("tag rule %q %w", tagRuleID, graphql.ErrNotFound)
}

// TagRuleByName returns the tag rule with the specified name.
func (a API) TagRuleByName(ctx context.Context, name string) (sla.TagRule, error) {
	a.log.Print(log.Trace)

	tagRules, err := a.TagRules(ctx, name)
	if err != nil {
		return sla.TagRule{}, err
	}

	name = strings.ToLower(name)
	for _, tagRule := range tagRules {
		if strings.ToLower(tagRule.Name) == name {
			return tagRule, nil
		}
	}

	return sla.TagRule{}, fmt.Errorf("tag rule %q %w", name, graphql.ErrNotFound)
}

// TagRules returns all tag rules matching the specified name filter.
// Note, object types the service account doesn't have permissions to access
// are silently ignored.
func (a API) TagRules(ctx context.Context, nameFilter string) ([]sla.TagRule, error) {
	a.log.Print(log.Trace)

	var filter []sla.TagRuleFilter
	if nameFilter != "" {
		filter = append(filter, sla.TagRuleFilter{
			Field:  "NAME",
			Values: []string{nameFilter},
		})
	}

	var tagRules []sla.TagRule
	for _, objectType := range sla.AllCloudNativeTagObjectTypes() {
		objectTypeTagRules, err := sla.ListTagRules(ctx, a.client, objectType, filter)
		if err != nil {
			var gqlErr graphql.GQLError
			if errors.As(err, &gqlErr) {
				if gqlErr.Code() == 403 {
					continue
				}
			}
			return nil, fmt.Errorf("failed to list tag rules: %s", err)
		}
		tagRules = append(tagRules, objectTypeTagRules...)
	}
	slices.SortFunc(tagRules, func(i, j sla.TagRule) int {
		return cmp.Compare(i.Name, j.Name)
	})

	return tagRules, nil
}

// CreateTagRule creates a new tag rule with the specified parameters.
func (a API) CreateTagRule(ctx context.Context, params sla.CreateTagRuleParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	// Check if the multiple key-value pairs feature flag is enabled.
	flag, err := core.Wrap(a.client).FeatureFlag(ctx, core.FeatureFlagMultipleKeyValuePairsInTagRules)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to check feature flag: %s", err)
	}

	// Update params based on the feature flag.
	if flag.Enabled {
		// Feature flag enabled: use tagConditions.
		// Convert Tag to TagConditions if needed.
		//lint:ignore SA1019 internal use of deprecated field for feature flag compatibility
		if params.TagConditions == nil && params.Tag.Key != "" {
			//lint:ignore SA1019 internal use of deprecated field for feature flag compatibility
			tagConditions := params.Tag.ToTagConditions()
			params.TagConditions = &tagConditions
		}
		// Clear the deprecated Tag field.
		//lint:ignore SA1019 internal use of deprecated field for feature flag compatibility
		params.Tag = sla.Tag{}
	} else {
		// Feature flag disabled: use the deprecated tag field.
		// Convert TagConditions to Tag if needed.
		//lint:ignore SA1019 internal use of deprecated field for feature flag compatibility
		if params.Tag.Key == "" && params.TagConditions != nil {
			if len(params.TagConditions.TagPairs) != 1 {
				return uuid.Nil, fmt.Errorf("multiple tag pairs require feature flag to be enabled")
			}
			pair := params.TagConditions.TagPairs[0]
			if len(pair.Values) > 1 {
				return uuid.Nil, fmt.Errorf("multiple tag values require feature flag to be enabled")
			}
			value := ""
			if len(pair.Values) > 0 {
				value = pair.Values[0]
			}
			//lint:ignore SA1019 internal use of deprecated field for feature flag compatibility
			params.Tag = sla.Tag{
				Key:       pair.Key,
				Value:     value,
				AllValues: pair.MatchAllTagValues,
			}
		}
		// Clear the TagConditions field.
		params.TagConditions = nil
	}

	id, err := sla.CreateTagRule(ctx, a.client, params)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create tag rule: %s", err)
	}

	return id, nil
}

// UpdateTagRule updates the tag rule with the specified ID.
func (a API) UpdateTagRule(ctx context.Context, tagRuleID uuid.UUID, params sla.UpdateTagRuleParams) error {
	a.log.Print(log.Trace)

	err := sla.UpdateTagRule(ctx, a.client, tagRuleID, params)
	if err != nil {
		return fmt.Errorf("failed to update tag rule: %s", err)
	}

	return nil
}

// DeleteTagRule deletes the tag rule with the specified ID.
func (a API) DeleteTagRule(ctx context.Context, tagRuleID uuid.UUID) error {
	a.log.Print(log.Trace)

	err := sla.DeleteTagRule(ctx, a.client, tagRuleID)
	if err != nil {
		return fmt.Errorf("failed to delete tag rule: %s", err)
	}

	return nil
}

// HierarchyObjectByID returns the hierarchy object with the specified ID.
// This can be used to query any hierarchy object (VMs, databases, tag rules,
// etc.) and retrieve its SLA assignment information including the configured
// and effective SLA domains.
//
// This function uses AllSubHierarchyType as the workload hierarchy, which
// returns the generic SLA assignment. Use HierarchyObjectByIDAndWorkload to
// specify a specific workload hierarchy for workload-specific SLA resolution.
func (a API) HierarchyObjectByID(ctx context.Context, fid uuid.UUID) (sla.HierarchyObject, error) {
	return a.HierarchyObjectByIDAndWorkload(ctx, fid, hierarchy.WorkloadAllSubHierarchyType)
}

// HierarchyObjectByIDAndWorkload returns the hierarchy object with the
// specified ID and workload hierarchy type.
// This can be used to query any hierarchy object (VMs, databases, tag rules,
// etc.) and retrieve its SLA assignment information including the configured
// and effective SLA domains.
//
// The workloadHierarchy parameter determines which workload type to use for
// SLA Domain resolution. Different workload types can have different SLA
// assignments on the same parent object. Pass hierarchy.WorkloadAllSubHierarchyType
// for the generic view, or a specific workload type (e.g.,
// hierarchy.WorkloadAzureVM) for workload-specific SLA resolution.
func (a API) HierarchyObjectByIDAndWorkload(ctx context.Context, fid uuid.UUID, workloadHierarchy hierarchy.Workload) (sla.HierarchyObject, error) {
	a.log.Print(log.Trace)

	obj, err := sla.ObjectByIDAndWorkload(ctx, a.client, fid, workloadHierarchy)
	if err != nil {
		return sla.HierarchyObject{}, fmt.Errorf("failed to get hierarchy object: %s", err)
	}

	return obj, nil
}
