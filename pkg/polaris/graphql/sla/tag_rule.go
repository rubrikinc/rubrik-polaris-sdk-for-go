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
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// TagRule represents an RSC tag rule. Note, the ID field of the EffectiveSLA
// field is either a UUID or one of the strings: doNotProtect, noAssignment.
type TagRule struct {
	ID                uuid.UUID         `json:"id"`
	Name              string            `json:"name"`
	ObjectType        ManagedObjectType `json:"objectType"`
	Tag               Tag               `json:"tag"`
	AllACloudAccounts bool              `json:"applyToAllCloudAccounts"`
	CloudAccounts     []struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	} `json:"cloudNativeAccounts"`
	EffectiveSLA struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"effectiveSla"`
}

// Tag represents the tag of an RSC tag rule.
type Tag struct {
	Key       string `json:"tagKey"`
	Value     string `json:"tagValue"`
	AllValues bool   `json:"matchAllValues"`
}

// TagRuleFilter holds the filter for a tag rules list operation.
type TagRuleFilter struct {
	Field  string   `json:"field"`
	Values []string `json:"texts"`
}

// ListTagRules returns all RSC tag rules of the specified object type matching
// the specified tag rule filters.
func ListTagRules(ctx context.Context, gql *graphql.Client, objectType string, filters []TagRuleFilter) ([]TagRule, error) {
	gql.Log().Print(log.Trace)

	query := cloudNativeTagRulesQuery
	buf, err := gql.Request(ctx, query, struct {
		ObjectType string          `json:"objectType"`
		Filters    []TagRuleFilter `json:"filters"`
	}{ObjectType: objectType, Filters: filters})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result struct {
				TagRules []TagRule
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.TagRules, nil
}

// CreateTagRuleParams holds the parameters for a tag rule create operation.
type CreateTagRuleParams struct {
	Name             string                   `json:"tagRuleName"`
	ObjectType       CloudNativeTagObjectType `json:"objectType"`
	Tag              Tag                      `json:"tag"`
	CloudAccounts    *TagRuleCloudAccounts    `json:"cloudNativeAccountIds,omitempty"`
	AllCloudAccounts bool                     `json:"applyToAllCloudAccounts,omitempty"`
}

// TagRuleCloudAccounts holds the cloud accounts for a tag rule.
type TagRuleCloudAccounts struct {
	AWSAccountIDs        []uuid.UUID `json:"awsNativeAccountIds,omitempty"`
	AzureSubscriptionIDs []uuid.UUID `json:"azureNativeSubscriptionIds,omitempty"`
	GCPProjectIDs        []uuid.UUID `json:"gcpNativeProjectIds,omitempty"`
}

// CreateTagRule creates a new tag rule. Returns the ID of the new tag rule.
func CreateTagRule(ctx context.Context, gql *graphql.Client, params CreateTagRuleParams) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	query := createCloudNativeTagRuleQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result struct {
				ID string `json:"tagRuleId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	id, err := uuid.Parse(payload.Data.Result.ID)
	if err != nil {
		return uuid.Nil, graphql.ResponseError(query, err)
	}

	return id, nil
}

// UpdateTagRuleParams holds the parameters for a tag rule update operation.
type UpdateTagRuleParams struct {
	Name             string                `json:"tagRuleName"`
	CloudAccounts    *TagRuleCloudAccounts `json:"cloudNativeAccountIds,omitempty"`
	AllCloudAccounts bool                  `json:"applyToAllCloudAccounts,omitempty"`
}

// UpdateTagRule updates the tag rule with the specified ID.
func UpdateTagRule(ctx context.Context, gql *graphql.Client, tagRuleID uuid.UUID, params UpdateTagRuleParams) error {
	gql.Log().Print(log.Trace)

	query := updateCloudNativeTagRuleQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"tagRuleId"`
		UpdateTagRuleParams
	}{ID: tagRuleID, UpdateTagRuleParams: params})
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

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

// DeleteTagRule deletes the tag rule with the specified ID.
func DeleteTagRule(ctx context.Context, gql *graphql.Client, tagRuleID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deleteCloudNativeTagRuleQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"ruleId"`
	}{ID: tagRuleID})
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

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
