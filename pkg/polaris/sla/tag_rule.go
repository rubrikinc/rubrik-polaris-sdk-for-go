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
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// TagRuleObjectTypes holds the valid object types for tag rules.
var tagRuleObjectTypes = []string{
	"AWS_S3_BUCKET",
	"AWS_EBS_VOLUME",
	"AZURE_VIRTUAL_MACHINE",
	"AZURE_STORAGE_ACCOUNT",
	"AZURE_SQL_DATABASE_DB",
	"AZURE_SQL_DATABASE_SERVER",
	"AWS_RDS_INSTANCE",
	"AZURE_MANAGED_DISK",
	"AZURE_SQL_MANAGED_INSTANCE_SERVER",
	"AWS_EC2_INSTANCE",
}

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
	for _, objectType := range tagRuleObjectTypes {
		objectTypeTagRules, err := sla.ListTagRules(ctx, a.client, objectType, filter)
		if err != nil {
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
