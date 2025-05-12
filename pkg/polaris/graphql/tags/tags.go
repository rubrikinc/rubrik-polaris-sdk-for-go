//go:generate go run ../queries_gen.go tags

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
	"encoding/json"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CustomerTags holds the customer tags for a specific cloud vendor.
type CustomerTags struct {
	CloudVendor          core.CloudVendor `json:"cloudVendor"`
	Tags                 []core.Tag       `json:"customerTags"`
	OverrideResourceTags bool             `json:"shouldOverrideResourceTags"`
}

// CustomerTagsFilter holds the filter for a customer tags list operation.
type CustomerTagsFilter struct {
	CloudVendor    core.CloudVendor `json:"cloudVendor"`
	CloudAccountID string           `json:"cloudAccountId,omitempty"`
}

// ListCustomerTags returns all customer tags matching the specified customer
// tags filter. Note, ALL_VENDORS cannot be specified in the filter.
func ListCustomerTags(ctx context.Context, gql *graphql.Client, filter CustomerTagsFilter) (CustomerTags, error) {
	gql.Log().Print(log.Trace)

	query := cloudNativeCustomerTagsQuery
	buf, err := gql.Request(ctx, query, filter)
	if err != nil {
		return CustomerTags{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result CustomerTags `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return CustomerTags{}, graphql.UnmarshalError(query, err)
	}
	payload.Data.Result.CloudVendor = filter.CloudVendor

	return payload.Data.Result, nil
}

// SetCustomerTags sets the customer tags for the specified cloud vendor.
func SetCustomerTags(ctx context.Context, gql *graphql.Client, params CustomerTags) error {
	gql.Log().Print(log.Trace)

	query := setCustomerTagsQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct{} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}
