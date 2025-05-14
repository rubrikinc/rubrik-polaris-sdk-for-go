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

package tags

import (
	"context"
	"fmt"

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
func (a API) CustomerTags(ctx context.Context, vendor core.CloudVendor) (tags.CustomerTags, error) {
	a.log.Print(log.Trace)

	cts, err := tags.ListCustomerTags(ctx, a.client, tags.CustomerTagsFilter{
		CloudVendor: vendor,
	})
	if err != nil {
		return tags.CustomerTags{}, fmt.Errorf("failed to get customer tags for cloud vendor %q: %s", vendor, err)
	}

	return cts, nil
}

// AddCustomerTags adds the specified customer tags to the existing customer
// tags.
func (a API) AddCustomerTags(ctx context.Context, customerTags tags.CustomerTags) error {
	a.log.Print(log.Trace)

	// Read existing customer tags.
	cts, err := a.CustomerTags(ctx, customerTags.CloudVendor)
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

	// Replace customer tags.
	if err := a.ReplaceCustomerTags(ctx, cts); err != nil {
		return err
	}

	return nil
}

// RemoveCustomerTags removes the specified customer tags from the existing
// customer tags.
func (a API) RemoveCustomerTags(ctx context.Context, vendor core.CloudVendor, customerTagKeys []string) error {
	a.log.Print(log.Trace)

	// Read existing customer tags.
	cts, err := a.CustomerTags(ctx, vendor)
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

// ReplaceCustomerTags replaces all the customer tags.
func (a API) ReplaceCustomerTags(ctx context.Context, customerTags tags.CustomerTags) error {
	a.log.Print(log.Trace)

	if err := tags.SetCustomerTags(ctx, a.client, customerTags); err != nil {
		return fmt.Errorf("failed to set customer tags for cloud vendor %q: %s", customerTags.CloudVendor, err)
	}

	return nil
}
