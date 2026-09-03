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

// Package tags provides high level functions when working with customer tags.
package tags

import (
	"context"
	"fmt"
	"slices"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/tags"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for customer tags.
type API struct {
	client *graphql.Client
	log    log.Logger
}

// Wrap the RSC client in the customer tags API.
func Wrap(client *polaris.Client) API {
	return API{client: client.GQL, log: client.GQL.Log()}
}

// CustomerTags returns the customer tags matching the specified cloud vendor.
// The customer tags returned are scoped to all cloud accounts of the cloud
// vendor, use CustomerTagsByFilter to scope them to a single cloud account.
func (a API) CustomerTags(ctx context.Context, vendor core.CloudVendor) (tags.CustomerTags, error) {
	a.log.Print(log.Trace)

	return a.CustomerTagsByFilter(ctx, tags.CustomerTagsFilter{CloudVendor: vendor})
}

// CustomerTagsByFilter returns the customer tags matching the specified
// customer tags filter. Note, the scope of a cloud account is independent of
// the scope of all cloud accounts of the cloud vendor, the customer tags of
// one scope are not affected by the customer tags of the other.
func (a API) CustomerTagsByFilter(ctx context.Context, filter tags.CustomerTagsFilter) (tags.CustomerTags, error) {
	a.log.Print(log.Trace)

	cts, err := tags.ListCustomerTags(ctx, a.client, filter)
	if err != nil {
		return tags.CustomerTags{}, fmt.Errorf("failed to get customer tags for cloud vendor %q: %s", filter.CloudVendor, err)
	}

	return cts, nil
}

// AddCustomerTags adds the specified customer tags and excluded tags to the
// existing customer tags and excluded tags of the same scope, eliminating
// duplicates.
//
// Note, the OverrideResourceTags field is not merged, it replaces the value
// configured in RSC. Leaving it unset therefore turns the override off for the
// scope, as the zero value of a bool is false.
func (a API) AddCustomerTags(ctx context.Context, customerTags tags.CustomerTags) error {
	a.log.Print(log.Trace)

	// Read existing customer tags of the same scope.
	cts, err := a.CustomerTagsByFilter(ctx, tags.CustomerTagsFilter{
		CloudVendor:    customerTags.CloudVendor,
		CloudAccountID: customerTags.CloudAccountID,
	})
	if err != nil {
		return err
	}

	// Add new customer tags, eliminating duplicates.
	ctm := make(map[string]string, len(cts.Tags)+len(customerTags.Tags))
	for _, tag := range cts.Tags {
		ctm[tag.Key] = tag.Value
	}
	for _, tag := range customerTags.Tags {
		ctm[tag.Key] = tag.Value
	}
	cts.Tags = make([]core.Tag, 0, len(ctm))
	for k, v := range ctm {
		cts.Tags = append(cts.Tags, core.Tag{Key: k, Value: v})
	}
	cts.OverrideResourceTags = customerTags.OverrideResourceTags

	// Add new excluded tag patterns, eliminating duplicates.
	etm := make(map[string]struct{}, len(cts.ExcludedTags)+len(customerTags.ExcludedTags))
	for _, k := range cts.ExcludedTags {
		etm[k] = struct{}{}
	}
	for _, k := range customerTags.ExcludedTags {
		etm[k] = struct{}{}
	}
	cts.ExcludedTags = make([]string, 0, len(etm))
	for k := range etm {
		cts.ExcludedTags = append(cts.ExcludedTags, k)
	}

	// Replace customer tags.
	if err := a.ReplaceCustomerTags(ctx, cts); err != nil {
		return err
	}

	return nil
}

// RemoveCustomerTags removes the specified customer tags from the existing
// customer tags scoped to all cloud accounts of the cloud vendor. Use
// RemoveCustomerTagsByFilter to scope the operation to a single cloud account.
func (a API) RemoveCustomerTags(ctx context.Context, vendor core.CloudVendor, customerTagKeys []string) error {
	a.log.Print(log.Trace)

	return a.RemoveCustomerTagsByFilter(ctx, tags.CustomerTagsFilter{CloudVendor: vendor}, customerTagKeys)
}

// RemoveCustomerTagsByFilter removes the specified customer tags from the
// existing customer tags matching the specified customer tags filter. Note,
// the excluded tags are left unchanged, use RemoveExcludedTagsByFilter to
// remove excluded tags.
func (a API) RemoveCustomerTagsByFilter(ctx context.Context, filter tags.CustomerTagsFilter, customerTagKeys []string) error {
	a.log.Print(log.Trace)

	// Read existing customer tags.
	cts, err := a.CustomerTagsByFilter(ctx, filter)
	if err != nil {
		return err
	}

	// Filter customer tags.
	ctm := make(map[string]string, len(cts.Tags))
	for _, tag := range cts.Tags {
		ctm[tag.Key] = tag.Value
	}
	for _, tag := range customerTagKeys {
		delete(ctm, tag)
	}
	cts.Tags = make([]core.Tag, 0, len(ctm))
	for k, v := range ctm {
		cts.Tags = append(cts.Tags, core.Tag{Key: k, Value: v})
	}

	// Replace customer tags.
	if err := a.ReplaceCustomerTags(ctx, cts); err != nil {
		return err
	}

	return nil
}

// RemoveExcludedTags removes the specified excluded tags from the existing
// excluded tags scoped to all cloud accounts of the cloud vendor. Use
// RemoveExcludedTagsByFilter to scope the operation to a single cloud account.
func (a API) RemoveExcludedTags(ctx context.Context, vendor core.CloudVendor, excludedTags []string) error {
	a.log.Print(log.Trace)

	return a.RemoveExcludedTagsByFilter(ctx, tags.CustomerTagsFilter{CloudVendor: vendor}, excludedTags)
}

// RemoveExcludedTagsByFilter removes the specified excluded tags from the
// existing excluded tags matching the specified customer tags filter. Note,
// the customer tags are left unchanged, use RemoveCustomerTagsByFilter to
// remove customer tags.
func (a API) RemoveExcludedTagsByFilter(ctx context.Context, filter tags.CustomerTagsFilter, excludedTags []string) error {
	a.log.Print(log.Trace)

	// Read existing customer tags.
	cts, err := a.CustomerTagsByFilter(ctx, filter)
	if err != nil {
		return err
	}

	// Filter excluded tags.
	cts.ExcludedTags = slices.DeleteFunc(cts.ExcludedTags, func(excludedTag string) bool {
		return slices.Contains(excludedTags, excludedTag)
	})

	// Replace customer tags.
	if err := a.ReplaceCustomerTags(ctx, cts); err != nil {
		return err
	}

	return nil
}

// ReplaceCustomerTags replaces the entire customer tags configuration of the
// scope specified in the customer tags. This is a destructive full replace of
// the customer tags, the excluded tags and the override flag, and not a merge,
// so all fields must be populated.
//
// Note, the OverrideResourceTags field is always sent, so leaving it unset
// turns the override off for the scope rather than preserving the value
// configured in RSC. Leaving the Tags field unset fails the request, as RSC
// does not accept a null tag list.
//
// To change part of a configuration, read it with CustomerTagsByFilter and
// modify the result before replacing it.
func (a API) ReplaceCustomerTags(ctx context.Context, customerTags tags.CustomerTags) error {
	a.log.Print(log.Trace)

	if err := tags.SetCustomerTags(ctx, a.client, customerTags); err != nil {
		return fmt.Errorf("failed to set customer tags for cloud vendor %q: %s", customerTags.CloudVendor, err)
	}

	return nil
}
