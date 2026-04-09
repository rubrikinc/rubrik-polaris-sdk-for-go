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
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Category represents the category of a data security policy.
type Category string

const (
	CategoryMisplaced   Category = "MISPLACED"
	CategoryOverexposed Category = "OVEREXPOSED"
	CategoryRedundant   Category = "REDUNDANT"
	CategoryUnprotected Category = "UNPROTECTED"
)

// Severity represents the severity of a data security policy.
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// Relationship represents the comparison operator for a filter condition.
type Relationship string

const (
	RelAfter       Relationship = "AFTER"
	RelBefore      Relationship = "BEFORE"
	RelBetween     Relationship = "BETWEEN"
	RelContains    Relationship = "CONTAINS"
	RelNotContains Relationship = "DOES_NOT_CONTAIN"
	RelEquals      Relationship = "EQUALS"
	RelExists      Relationship = "EXISTS"
	RelGreaterThan Relationship = "GREATER_THAN"
	RelIs          Relationship = "IS"
	RelIsEmpty     Relationship = "IS_EMPTY"
	RelIsNot       Relationship = "IS_NOT"
	RelIsNotEmpty  Relationship = "IS_NOT_EMPTY"
	RelLessThan    Relationship = "LESS_THAN"
	RelNoneOf      Relationship = "NONE_OF"
	RelNotEquals   Relationship = "NOT_EQUALS"
	RelOtherThan   Relationship = "OTHER_THAN"
)

// FilterType represents the type/dimension of a filter condition.
type FilterType string

const (
	FilterTypeDocumentDataCategory           FilterType = "SECURITY_DOCUMENT_DATA_CATEGORY"
	FilterTypeDocumentDataType               FilterType = "SECURITY_DOCUMENT_DATA_TYPE"
	FilterTypeDocumentDocumentType           FilterType = "SECURITY_DOCUMENT_DOCUMENT_TYPE"
	FilterTypeDocumentExposure               FilterType = "SECURITY_DOCUMENT_EXPOSURE"
	FilterTypeDocumentHitCount               FilterType = "SECURITY_DOCUMENT_HIT_COUNT"
	FilterTypeDocumentLastAccess             FilterType = "SECURITY_DOCUMENT_LAST_ACCESS"
	FilterTypeDocumentLastModified           FilterType = "SECURITY_DOCUMENT_LAST_MODIFIED"
	FilterTypeDocumentMIPLabel               FilterType = "SECURITY_DOCUMENT_MIP_LABEL"
	FilterTypeDocumentSensitivity            FilterType = "SECURITY_DOCUMENT_SENSITIVITY"
	FilterTypeIdentityCAPMetadata            FilterType = "SECURITY_IDENTITY_CAP_METADATA"
	FilterTypeIdentityDepartment             FilterType = "SECURITY_IDENTITY_DEPARTMENT"
	FilterTypeIdentityDirectDescendantCount  FilterType = "SECURITY_IDENTITY_DIRECT_DESCENDANT_COUNT"
	FilterTypeIdentityEventActionType        FilterType = "SECURITY_IDENTITY_EVENT_ACTION_TYPE"
	FilterTypeIdentityEventActor             FilterType = "SECURITY_IDENTITY_EVENT_ACTOR"
	FilterTypeIdentityEventChangedAttribute  FilterType = "SECURITY_IDENTITY_EVENT_CHANGED_ATTRIBUTE"
	FilterTypeIdentityEventChangedAttNewVal  FilterType = "SECURITY_IDENTITY_EVENT_CHANGED_ATT_NEW_VAL"
	FilterTypeIdentityEventChangedAttOldVal  FilterType = "SECURITY_IDENTITY_EVENT_CHANGED_ATT_OLD_VAL"
	FilterTypeIdentityEventDCName            FilterType = "SECURITY_IDENTITY_EVENT_DC_NAME"
	FilterTypeIdentityEventSourceEntityID    FilterType = "SECURITY_IDENTITY_EVENT_SOURCE_ENTITY_ID"
	FilterTypeIdentityEventTargetEntity      FilterType = "SECURITY_IDENTITY_EVENT_TARGET_ENTITY"
	FilterTypeIdentityEventTargetEntityType  FilterType = "SECURITY_IDENTITY_EVENT_TARGET_ENTITY_TYPE"
	FilterTypeIdentityEventTimestamp         FilterType = "SECURITY_IDENTITY_EVENT_TIMESTAMP"
	FilterTypeIdentityEventTitle             FilterType = "SECURITY_IDENTITY_EVENT_TITLE"
	FilterTypeIdentityGroupMembership        FilterType = "SECURITY_IDENTITY_GROUP_MEMBERSHIP"
	FilterTypeIdentityIDPMetadataLabel       FilterType = "SECURITY_IDENTITY_IDP_METADATA_LABEL"
	FilterTypeIdentityIDPType                FilterType = "SECURITY_IDENTITY_IDP_TYPE"
	FilterTypeIdentityMetadata               FilterType = "SECURITY_IDENTITY_METADATA"
	FilterTypeIdentityMetadataBitmask        FilterType = "SECURITY_IDENTITY_METADATA_BITMASK"
	FilterTypeIdentityMetadataLabel          FilterType = "SECURITY_IDENTITY_METADATA_LABEL"
	FilterTypeIdentityMetadataListLength     FilterType = "SECURITY_IDENTITY_METADATA_LIST_LENGTH"
	FilterTypeIdentityMetadataValues         FilterType = "SECURITY_IDENTITY_METADATA_VALUES"
	FilterTypeIdentityMFAStrength            FilterType = "SECURITY_IDENTITY_MFA_STRENGTH"
	FilterTypeIdentityName                   FilterType = "SECURITY_IDENTITY_NAME"
	FilterTypeIdentityNativeCreationTime     FilterType = "SECURITY_IDENTITY_NATIVE_CREATION_TIME"
	FilterTypeIdentityNumberOfSecrets        FilterType = "SECURITY_IDENTITY_NUMBER_OF_SECRETS"
	FilterTypeIdentityOrigin                 FilterType = "SECURITY_IDENTITY_ORIGIN"
	FilterTypeIdentityPrivilegeType          FilterType = "SECURITY_IDENTITY_PRIVILEGE_TYPE"
	FilterTypeIdentityPwdAge                 FilterType = "SECURITY_IDENTITY_PWD_AGE"
	FilterTypeIdentitySecretCreationTime     FilterType = "SECURITY_IDENTITY_SECRET_CREATION_TIME"
	FilterTypeIdentitySecretExpiryTime       FilterType = "SECURITY_IDENTITY_SECRET_EXPIRY_TIME"
	FilterTypeIdentityServicePrincipalName   FilterType = "SECURITY_IDENTITY_SERVICE_PRINCIPAL_NAME"
	FilterTypeIdentityStatus                 FilterType = "SECURITY_IDENTITY_STATUS"
	FilterTypeIdentitySyncedOnpremPrivileged FilterType = "SECURITY_IDENTITY_SYNCED_ONPREM_PRIVILEGED_ACCOUNT"
	FilterTypeIdentityTitle                  FilterType = "SECURITY_IDENTITY_TITLE"
	FilterTypeIdentityType                   FilterType = "SECURITY_IDENTITY_TYPE"
	FilterTypeIdentityUniqueIdentifier       FilterType = "SECURITY_IDENTITY_UNIQUE_IDENTIFIER"
	FilterTypeIDPCAPMetadata                 FilterType = "SECURITY_IDP_CAP_METADATA"
	FilterTypeIDPCriticalAppCoverage         FilterType = "SECURITY_IDP_CRITICAL_APP_COVERAGE"
	FilterTypeIDPHasGroupWithLabel           FilterType = "SECURITY_IDP_HAS_GROUP_WITH_LABEL"
	FilterTypeIDPHasUserWithLabel            FilterType = "SECURITY_IDP_HAS_USER_WITH_LABEL"
	FilterTypeIDPMetadataLabel               FilterType = "SECURITY_IDP_METADATA_LABEL"
	FilterTypeIDPMetadataNumericComparison   FilterType = "SECURITY_IDP_METADATA_NUMERIC_COMPARISON"
	FilterTypeIDPMetadataValueLength         FilterType = "SECURITY_IDP_METADATA_VALUE_LENGTH"
	FilterTypeIDPPrivilegedAccountCount      FilterType = "SECURITY_IDP_PRIVILEGED_ACCOUNT_COUNT"
	FilterTypeIDPPrivilegedUserCount         FilterType = "SECURITY_IDP_PRIVILEGED_USER_COUNT"
	FilterTypeIDPType                        FilterType = "SECURITY_IDP_TYPE"
	FilterTypeSnappableBackup                FilterType = "SECURITY_SNAPPABLE_BACKUP"
	FilterTypeSnappableCloudAccount          FilterType = "SECURITY_SNAPPABLE_CLOUD_ACCOUNT"
	FilterTypeSnappableCreatedAt             FilterType = "SECURITY_SNAPPABLE_CREATED_AT"
	FilterTypeSnappableEncryption            FilterType = "SECURITY_SNAPPABLE_ENCRYPTION"
	FilterTypeSnappableLogging               FilterType = "SECURITY_SNAPPABLE_LOGGING"
	FilterTypeSnappableName                  FilterType = "SECURITY_SNAPPABLE_NAME"
	FilterTypeSnappableNetworkAccess         FilterType = "SECURITY_SNAPPABLE_NETWORK_ACCESS"
	FilterTypeSnappableRegion                FilterType = "SECURITY_SNAPPABLE_REGION"
	FilterTypeSnappableTag                   FilterType = "SECURITY_SNAPPABLE_TAG"
	FilterTypeSnappableType                  FilterType = "SECURITY_SNAPPABLE_TYPE"
)

// LogicalOp represents a logical operator for combining filters.
type LogicalOp string

const (
	LogicalAnd LogicalOp = "AND"
	LogicalOr  LogicalOp = "OR"
)

// Config represents a single filter condition (input format).
type Config struct {
	Type         FilterType   `json:"filterType"`
	Values       []string     `json:"values"`
	Relationship Relationship `json:"relationship"`
}

// Node represents either a single filter or a nested group (input format).
type Node struct {
	Config      *Config      `json:"filterConfig,omitempty"`
	GroupConfig *GroupConfig `json:"filterGroupConfig,omitempty"`
}

// GroupConfig represents a group of filters joined by a logical operator
// (input format).
type GroupConfig struct {
	Op      LogicalOp `json:"logicalOp"`
	Filters []Node    `json:"filterList"`
}

// Policy represents a data security policy from RSC.
type Policy struct {
	ID              uuid.UUID    `json:"policyId"`
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	Category        Category     `json:"policyCategory"`
	Severity        Severity     `json:"policySeverity"`
	Enabled         bool         `json:"isEnabled"`
	Predefined      bool         `json:"isPredefined"`
	Filter          *GroupConfig `json:"-"`
	ThresholdFilter *GroupConfig `json:"-"`
	Created         time.Time    `json:"createdAt"`
	Updated         time.Time    `json:"updatedAt"`
}

// UnmarshalJSON implements json.Unmarshaler. It converts the GraphQL response
// filter format (which uses different field names and a __typename discriminator)
// into the SDK's GroupConfig type.
func (p *Policy) UnmarshalJSON(data []byte) error {
	type Alias Policy
	var raw struct {
		Alias
		Filter          json.RawMessage `json:"filter"`
		ThresholdFilter json.RawMessage `json:"thresholdFilter"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = Policy(raw.Alias)

	filter, err := parsePolicyFilter(raw.Filter)
	if err != nil {
		return fmt.Errorf("filter: %v", err)
	}
	p.Filter = filter

	thresholdFilter, err := parsePolicyFilter(raw.ThresholdFilter)
	if err != nil {
		return fmt.Errorf("thresholdFilter: %v", err)
	}
	p.ThresholdFilter = thresholdFilter

	return nil
}

// Response-side types matching the GraphQL response field names. These are
// unexported because __typename discrimination is an internal concern.

type responseFilterGroupConfig struct {
	LogicalOperator LogicalOp            `json:"logicalOperator"`
	FiltersList     []responseFilterNode `json:"filtersList"`
}

type responseFilterNode struct {
	FilterConfig json.RawMessage `json:"filterConfig"`
}

// parsePolicyFilter unmarshals a PolicyFilter response JSON into a
// *GroupConfig. Returns nil if the raw JSON is null or empty.
func parsePolicyFilter(raw json.RawMessage) (*GroupConfig, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var pf struct {
		FilterConfig json.RawMessage `json:"filterConfig"`
	}
	if err := json.Unmarshal(raw, &pf); err != nil {
		return nil, err
	}
	if len(pf.FilterConfig) == 0 || string(pf.FilterConfig) == "null" {
		return nil, nil
	}

	node, err := parseFilterNode(pf.FilterConfig)
	if err != nil {
		return nil, err
	}
	if node.GroupConfig != nil {
		return node.GroupConfig, nil
	}
	// Wrap a lone FilterConfig in a single-element group for consistency.
	return &GroupConfig{Op: LogicalAnd, Filters: []Node{node}}, nil
}

// parseFilterNode unmarshals a raw filterConfig union member into a Node.
// The JSON is either a FilterConfig or FilterGroupConfig, determined by
// __typename.
func parseFilterNode(raw json.RawMessage) (Node, error) {
	var typename struct {
		Typename string `json:"__typename"`
	}
	if err := json.Unmarshal(raw, &typename); err != nil {
		return Node{}, fmt.Errorf("failed to determine filter type: %v", err)
	}

	switch typename.Typename {
	case "FilterConfig":
		var fc struct {
			Type         FilterType   `json:"type"`
			Values       []string     `json:"values"`
			Relationship Relationship `json:"relationship"`
		}
		if err := json.Unmarshal(raw, &fc); err != nil {
			return Node{}, fmt.Errorf("failed to unmarshal FilterConfig: %v", err)
		}
		return Node{
			Config: &Config{
				Type:         fc.Type,
				Values:       fc.Values,
				Relationship: fc.Relationship,
			},
		}, nil
	case "FilterGroupConfig":
		var gc responseFilterGroupConfig
		if err := json.Unmarshal(raw, &gc); err != nil {
			return Node{}, fmt.Errorf("failed to unmarshal FilterGroupConfig: %v", err)
		}
		nodes := make([]Node, 0, len(gc.FiltersList))
		for _, child := range gc.FiltersList {
			node, err := parseFilterNode(child.FilterConfig)
			if err != nil {
				return Node{}, err
			}
			nodes = append(nodes, node)
		}
		return Node{
			GroupConfig: &GroupConfig{
				Op:      gc.LogicalOperator,
				Filters: nodes,
			},
		}, nil
	default:
		return Node{}, fmt.Errorf("unknown filter __typename %q", typename.Typename)
	}
}

// CreateInput holds the parameters for creating a data security policy.
type CreateInput struct {
	Name            string       `json:"policyName"`
	Description     string       `json:"description"`
	Category        Category     `json:"policyCategory"`
	Severity        Severity     `json:"policySeverity"`
	Filter          GroupConfig  `json:"filter"`
	ThresholdFilter *GroupConfig `json:"thresholdFilter,omitempty"`
}

// UpdateInput holds the parameters for updating a data security policy.
type UpdateInput struct {
	ID              uuid.UUID    `json:"policyId"`
	Name            *string      `json:"policyName,omitempty"`
	Description     *string      `json:"description,omitempty"`
	Category        *Category    `json:"policyCategory,omitempty"`
	Severity        *Severity    `json:"policySeverity,omitempty"`
	Enabled         *bool        `json:"isEnabled,omitempty"`
	Filter          *GroupConfig `json:"filter,omitempty"`
	ThresholdFilter *GroupConfig `json:"thresholdFilter,omitempty"`
}

// policyTypeDataGov is the RSC policy type for data governance security
// policies managed by the DSPM API.
const policyTypeDataGov = "POLICY_TYPE_DATAGOV"

// maxFilterDepth is the maximum nesting depth for filter groups, matching the
// FilterFields GraphQL fragment.
const maxFilterDepth = 2

// validateFilterDepth verifies that the filter group does not exceed the
// maximum supported nesting depth.
func validateFilterDepth(gc GroupConfig, depth int) error {
	for _, node := range gc.Filters {
		if node.GroupConfig != nil {
			if depth+1 >= maxFilterDepth {
				return fmt.Errorf("filter nesting exceeds maximum depth of %d", maxFilterDepth)
			}
			if err := validateFilterDepth(*node.GroupConfig, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreatePolicy creates a new data security policy. Returns the ID of the
// created policy.
func CreatePolicy(ctx context.Context, gql *graphql.Client, input CreateInput) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	if err := validateFilterDepth(input.Filter, 0); err != nil {
		return uuid.UUID{}, err
	}
	if input.ThresholdFilter != nil {
		if err := validateFilterDepth(*input.ThresholdFilter, 0); err != nil {
			return uuid.UUID{}, err
		}
	}

	type createInput struct {
		CreateInput
		Type string `json:"policyType"`
	}

	query := createSecurityPolicyQuery
	buf, err := gql.Request(ctx, query, struct {
		Input createInput `json:"input"`
	}{Input: createInput{CreateInput: input, Type: policyTypeDataGov}})
	if err != nil {
		return uuid.UUID{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				PolicyID string `json:"policyId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.UUID{}, graphql.UnmarshalError(query, err)
	}

	id, err := uuid.Parse(payload.Data.Result.PolicyID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to parse policy ID %q: %v", payload.Data.Result.PolicyID, err)
	}

	return id, nil
}

// PolicyByID returns the data security policy with the specified ID.
func PolicyByID(ctx context.Context, gql *graphql.Client, id uuid.UUID) (Policy, error) {
	gql.Log().Print(log.Trace)

	query := securityPolicyQuery
	buf, err := gql.Request(ctx, query, struct {
		PolicyID   uuid.UUID `json:"policyId"`
		PolicyType string    `json:"policyType"`
	}{PolicyID: id, PolicyType: policyTypeDataGov})
	if err != nil {
		// Not-found errors cannot be distinguished from other errors,
		// so fall back to listing all policies to determine whether
		// the policy actually exists.
		gql.Log().Printf(log.Trace, "PolicyByID: primary query failed, falling back to list: %v", err)
		policies, listErr := Policies(ctx, gql)
		if listErr != nil {
			return Policy{}, fmt.Errorf("%w (list fallback also failed: %v)", graphql.RequestError(query, err), listErr)
		}
		for _, p := range policies {
			if p.ID == id {
				return p, nil
			}
		}
		return Policy{}, fmt.Errorf("policy %q %w", id, graphql.ErrNotFound)
	}

	var payload struct {
		Data struct {
			Result struct {
				Policy Policy `json:"policy"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return Policy{}, graphql.UnmarshalError(query, err)
	}

	if payload.Data.Result.Policy.ID == uuid.Nil {
		return Policy{}, fmt.Errorf("policy %q %w", id, graphql.ErrNotFound)
	}

	return payload.Data.Result.Policy, nil
}

// Policies returns all data security policies.
func Policies(ctx context.Context, gql *graphql.Client) ([]Policy, error) {
	gql.Log().Print(log.Trace)

	query := allSecurityPoliciesQuery
	buf, err := gql.Request(ctx, query, struct {
		PolicyTypes []string `json:"policyTypes"`
	}{PolicyTypes: []string{policyTypeDataGov}})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []struct {
				Policy Policy `json:"policy"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	policies := make([]Policy, 0, len(payload.Data.Result))
	for _, r := range payload.Data.Result {
		policies = append(policies, r.Policy)
	}

	return policies, nil
}

// UpdatePolicy updates a data security policy.
func UpdatePolicy(ctx context.Context, gql *graphql.Client, input UpdateInput) error {
	gql.Log().Print(log.Trace)

	if input.Filter != nil {
		if err := validateFilterDepth(*input.Filter, 0); err != nil {
			return err
		}
	}
	if input.ThresholdFilter != nil {
		if err := validateFilterDepth(*input.ThresholdFilter, 0); err != nil {
			return err
		}
	}

	type updateInput struct {
		UpdateInput
		Type string `json:"policyType"`
	}

	query := updateSecurityPolicyQuery
	_, err := gql.Request(ctx, query, struct {
		Input updateInput `json:"input"`
	}{Input: updateInput{UpdateInput: input, Type: policyTypeDataGov}})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	return nil
}

// DeletePolicy deletes the data security policy with the specified ID.
func DeletePolicy(ctx context.Context, gql *graphql.Client, id uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deleteSecurityPolicyQuery
	_, err := gql.Request(ctx, query, struct {
		PolicyID   uuid.UUID `json:"policyId"`
		PolicyType string    `json:"policyType"`
	}{PolicyID: id, PolicyType: policyTypeDataGov})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	return nil
}
