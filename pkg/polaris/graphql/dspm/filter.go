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

package dspm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ResourceType represents a policy resource type used to query available
// filter types.
type ResourceType string

const (
	ResourceTypeIdentity    ResourceType = "RESOURCE_TYPE_IDENTITY"
	ResourceTypeIDP         ResourceType = "RESOURCE_TYPE_IDP"
	ResourceTypeObject      ResourceType = "RESOURCE_TYPE_OBJECT"
	ResourceTypeUnspecified ResourceType = "RESOURCE_TYPE_UNSPECIFIED"
)

// FilterValue represents a possible value for a filter condition.
type FilterValue struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// FilterTreeChild represents a leaf node in a hierarchical filter value tree.
type FilterTreeChild struct {
	Value FilterValue `json:"value"`
}

// FilterTreeValue represents a hierarchical filter value with children.
// Children are leaf nodes — the RSC API and query only support one level of
// nesting.
type FilterTreeValue struct {
	Value    FilterValue       `json:"value"`
	Children []FilterTreeChild `json:"children"`
}

// FilterValueWithProvider represents a filter value associated with a cloud
// provider.
type FilterValueWithProvider struct {
	FilterValue FilterValue `json:"filterValue"`
	Provider    string      `json:"provider"`
}

// FilterMetadata holds the possible relationships and values for a given
// filter type. Exactly one of Values, TreeValues, or ProviderValues will be
// populated depending on the filter type.
type FilterMetadata struct {
	Relationships  []Relationship
	Values         []FilterValue
	TreeValues     []FilterTreeValue
	ProviderValues []FilterValueWithProvider
}

// FilterTypes returns the available filter types for the specified resource
// type.
func FilterTypes(ctx context.Context, gql *graphql.Client, resourceType ResourceType) ([]FilterType, error) {
	gql.Log().Print(log.Trace)

	query := allPolicyFilterTypesQuery
	buf, err := gql.Request(ctx, query, struct {
		ResourceType ResourceType `json:"resourceType"`
		PolicyType   string       `json:"policyType"`
	}{ResourceType: resourceType, PolicyType: policyTypeDataGov})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []FilterType `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// FilterValues returns the possible relationships and values for the
// specified filter type.
func FilterValues(ctx context.Context, gql *graphql.Client, filterType FilterType, searchTerm string) (FilterMetadata, error) {
	gql.Log().Print(log.Trace)

	query := allPolicyFilterValuesQuery
	buf, err := gql.Request(ctx, query, struct {
		FilterType FilterType `json:"policyFilterType"`
		PolicyType string     `json:"policyType"`
		SearchTerm string     `json:"searchTerm,omitempty"`
	}{FilterType: filterType, PolicyType: policyTypeDataGov, SearchTerm: searchTerm})
	if err != nil {
		return FilterMetadata{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Relationships []Relationship  `json:"possibleRelationships"`
				Values        json.RawMessage `json:"possibleValues"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return FilterMetadata{}, graphql.UnmarshalError(query, err)
	}

	result := FilterMetadata{
		Relationships: payload.Data.Result.Relationships,
	}

	var typename struct {
		Typename string `json:"__typename"`
	}
	if err := json.Unmarshal(payload.Data.Result.Values, &typename); err != nil {
		return FilterMetadata{}, graphql.UnmarshalError(query, err)
	}

	switch typename.Typename {
	case "FilterValues":
		var v struct {
			FilterValues []FilterValue `json:"filterValues"`
		}
		if err := json.Unmarshal(payload.Data.Result.Values, &v); err != nil {
			return FilterMetadata{}, graphql.UnmarshalError(query, err)
		}
		result.Values = v.FilterValues
	case "FilterTreeValues":
		var v struct {
			FilterValues []FilterTreeValue `json:"filterValues"`
		}
		if err := json.Unmarshal(payload.Data.Result.Values, &v); err != nil {
			return FilterMetadata{}, graphql.UnmarshalError(query, err)
		}
		result.TreeValues = v.FilterValues
	case "FilterValuesWithProvider":
		var v struct {
			FilterValuesWithProvider []FilterValueWithProvider `json:"filterValuesWithProvider"`
		}
		if err := json.Unmarshal(payload.Data.Result.Values, &v); err != nil {
			return FilterMetadata{}, graphql.UnmarshalError(query, err)
		}
		result.ProviderValues = v.FilterValuesWithProvider
	default:
		return FilterMetadata{}, fmt.Errorf("unknown filter values type %q", typename.Typename)
	}

	return result, nil
}
