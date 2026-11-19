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

package azure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ManagementGroup represents an Azure management group.
type ManagementGroup struct {
	NativeID                  string    `json:"nativeId"`
	Name                      string    `json:"name"`
	CustomerManagementGroupID uuid.UUID `json:"customerManagementGroupId"` // Only populated for onboarded groups.
	IsAuthorized              bool      `json:"isAuthorized"`
}

// ManagementGroupInput holds the management group information passed to RSC
// when subscriptions are added through a management group.
type ManagementGroupInput struct {
	NativeID string `json:"nativeId"`
	Name     string `json:"name"`
}

// SubscriptionInput identifies an Azure subscription.
type SubscriptionInput struct {
	NativeID uuid.UUID `json:"nativeId"`
	Name     string    `json:"name"`
}

// ManagementGroupFeature holds the information for a feature when it's
// onboarded through a management group. Unlike CloudAccountFeature, the
// permission groups are not versioned.
type ManagementGroupFeature struct {
	FeatureType      string         `json:"featureType"`
	PermissionGroups []string       `json:"permissionsGroups,omitempty"`
	ResourceGroup    *ResourceGroup `json:"resourceGroup,omitempty"`
}

// ManagementGroupSubscription holds the subscription, and the features to
// enable on the subscription, when adding subscriptions through a management
// group.
type ManagementGroupSubscription struct {
	Subscription SubscriptionInput        `json:"subscription"`
	Features     []ManagementGroupFeature `json:"features"`
}

// IneligibilityReason is the reason an entity in a management group hierarchy
// cannot be onboarded.
type IneligibilityReason string

const (
	IneligibleUnspecified              IneligibilityReason = "AZURE_ONBOARDING_INELIGIBILITY_REASON_UNSPECIFIED"
	IneligibleAlreadyOnboarded         IneligibilityReason = "AZURE_ONBOARDING_INELIGIBILITY_REASON_ALREADY_ONBOARDED"
	IneligibleAuthTypeMismatch         IneligibilityReason = "AZURE_ONBOARDING_INELIGIBILITY_REASON_AUTH_TYPE_MISMATCH"
	IneligibleDifferentManagementGroup IneligibilityReason = "AZURE_ONBOARDING_INELIGIBILITY_REASON_DIFFERENT_MANAGEMENT_GROUP"
	IneligibleNotInManagementGroup     IneligibilityReason = "AZURE_ONBOARDING_INELIGIBILITY_REASON_NOT_IN_MANAGEMENT_GROUP"
)

// OperationType is the operation a management group hierarchy is listed for.
type OperationType string

const (
	OperationAdd     OperationType = "ADD"
	OperationUpgrade OperationType = "UPGRADE"
	OperationMigrate OperationType = "MIGRATE"
)

// AuthType is the Azure authentication method of a cloud account.
type AuthType string

const (
	AuthTypeNonOAuth         AuthType = "AZURE_AUTH_TYPE_NON_OAUTH"
	AuthTypeOAuthRubrikApp   AuthType = "AZURE_AUTH_TYPE_OAUTH_RUBRIK_APP"
	AuthTypeOAuthCustomerApp AuthType = "AZURE_AUTH_TYPE_OAUTH_CUSTOMER_APP"
)

// HierarchySubscription is a subscription found in a management group
// hierarchy.
type HierarchySubscription struct {
	NativeID            uuid.UUID
	Name                string
	IsAuthorized        bool
	IsEligible          bool
	IneligibilityReason IneligibilityReason
}

// HierarchyManagementGroup is a management group found in a management group
// hierarchy.
type HierarchyManagementGroup struct {
	ManagementGroup
	IsEligible          bool
	IneligibilityReason IneligibilityReason
}

// ManagementGroupHierarchy holds the flattened hierarchy under a management
// group.
type ManagementGroupHierarchy struct {
	ManagementGroups []HierarchyManagementGroup
	Subscriptions    []HierarchySubscription
}

// ManagementGroupHierarchyRequest holds the parameters for listing the
// hierarchy under a management group.
type ManagementGroupHierarchyRequest struct {
	Cloud                    Cloud         `json:"cloudType"`
	TenantDomain             string        `json:"tenantDomainName"`
	SessionID                uuid.UUID     `json:"sessionId"`
	ManagementGroupID        string        `json:"managementGroupId"`
	ScopedManagementGroupIDs []string      `json:"scopedManagementGroupIds,omitempty"`
	ShouldRecurse            bool          `json:"shouldRecurse"`
	Features                 []string      `json:"features,omitempty"`
	Operation                OperationType `json:"operationType"`
	AuthType                 AuthType      `json:"authType"`
}

// InitiateCloudAccountAddWithCustomerApp starts an add cloud account session
// using the customer app credentials. If the app secret is empty, the secret
// stored in RSC for the app is used. Returns the session ID, which expires
// after 60 minutes.
func (a API) InitiateCloudAccountAddWithCustomerApp(ctx context.Context, cloud Cloud, tenantDomain string, appID uuid.UUID, appSecret secret.String) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	// The shared session validator requires the UI callback URL even
	// when customer app authentication does not open a browser.
	endpoint, err := url.Parse(a.GQL.Endpoint())
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid RSC GraphQL URL: %w", err)
	}
	redirectURL := endpoint.ResolveReference(&url.URL{Path: "/setup_azure"}).String()
	query := azureCloudAccountAddWithCustomerAppInitiateQuery
	vars := struct {
		Cloud        Cloud         `json:"cloudType"`
		TenantDomain string        `json:"tenantDomainName"`
		AppID        uuid.UUID     `json:"appId"`
		AppSecret    secret.String `json:"appSecretKey,omitempty"`
		RedirectURL  string        `json:"redirectUrl"`
	}{Cloud: cloud, TenantDomain: tenantDomain, AppID: appID, AppSecret: appSecret, RedirectURL: redirectURL}

	// Request logs the raw response data. Decode it before logging so the
	// returned session ID can be redacted along with the request secret.
	requestLog, err := json.MarshalIndent(secret.Redact(vars), "", "    ")
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal query variables: %w", err)
	}
	a.log.Printf(log.Debug, "%s request:\n%s\n%s", graphql.QueryName(query), query, requestLog)

	buf, err := a.GQL.RequestWithoutLogging(ctx, query, vars)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				SessionID secret.String `json:"sessionId"`
				Success   bool          `json:"success"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	if responseLog, err := json.MarshalIndent(secret.Redact(payload), "", "    "); err == nil {
		a.log.Printf(log.Debug, "%s response:\n%s", graphql.QueryName(query), responseLog)
	}
	if !payload.Data.Result.Success {
		return uuid.Nil, graphql.ResponseError(query, errors.New("initiate add cloud account failed"))
	}
	sessionID, err := uuid.Parse(string(payload.Data.Result.SessionID))
	if err != nil {
		return uuid.Nil, graphql.ResponseError(query, errors.New("invalid session id"))
	}

	if sessionID == uuid.Nil {
		return uuid.Nil, graphql.ResponseError(query, errors.New("empty session id"))
	}
	return sessionID, nil
}

// ListManagementGroups lists the management groups the session has access to.
func (a API) ListManagementGroups(ctx context.Context, cloud Cloud, tenantDomain string, sessionID uuid.UUID) ([]ManagementGroup, error) {
	a.log.Print(log.Trace)

	query := azureListManagementGroupsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Cloud        Cloud         `json:"cloudType"`
		TenantDomain string        `json:"tenantDomainName"`
		SessionID    secret.String `json:"sessionId"`
	}{Cloud: cloud, TenantDomain: tenantDomain, SessionID: secret.String(sessionID.String())})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				ManagementGroups []ManagementGroup `json:"managementGroups"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.ManagementGroups, nil
}

// ListManagementGroupHierarchy lists the flattened hierarchy of management
// groups and subscriptions under a management group.
func (a API) ListManagementGroupHierarchy(ctx context.Context, req ManagementGroupHierarchyRequest) (ManagementGroupHierarchy, error) {
	a.log.Print(log.Trace)

	query := azureListManagementGroupHierarchyQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ManagementGroupHierarchyRequest
		SessionID secret.String `json:"sessionId"`
	}{ManagementGroupHierarchyRequest: req, SessionID: secret.String(req.SessionID.String())})
	if err != nil {
		return ManagementGroupHierarchy{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Entities []struct {
					IsEligible          bool                `json:"isEligible"`
					IneligibilityReason IneligibilityReason `json:"ineligibilityReason"`
					Entity              struct {
						Typename     string `json:"__typename"`
						NativeID     string `json:"nativeId"`
						Name         string `json:"name"`
						IsAuthorized bool   `json:"isAuthorized"`
					} `json:"entity"`
				} `json:"entities"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ManagementGroupHierarchy{}, graphql.UnmarshalError(query, err)
	}

	var hierarchy ManagementGroupHierarchy
	for _, e := range payload.Data.Result.Entities {
		switch e.Entity.Typename {
		case "AzureManagementGroup":
			hierarchy.ManagementGroups = append(hierarchy.ManagementGroups, HierarchyManagementGroup{
				ManagementGroup: ManagementGroup{
					NativeID:     e.Entity.NativeID,
					Name:         e.Entity.Name,
					IsAuthorized: e.Entity.IsAuthorized,
				},
				IsEligible:          e.IsEligible,
				IneligibilityReason: e.IneligibilityReason,
			})
		case "CloudAccountsAzureSubscription":
			nativeID, err := uuid.Parse(e.Entity.NativeID)
			if err != nil {
				return ManagementGroupHierarchy{}, graphql.ResponseError(query, fmt.Errorf("invalid subscription id %q: %s", e.Entity.NativeID, err))
			}
			hierarchy.Subscriptions = append(hierarchy.Subscriptions, HierarchySubscription{
				NativeID:            nativeID,
				Name:                e.Entity.Name,
				IsAuthorized:        e.Entity.IsAuthorized,
				IsEligible:          e.IsEligible,
				IneligibilityReason: e.IneligibilityReason,
			})
		default:
			return ManagementGroupHierarchy{}, graphql.ResponseError(query, fmt.Errorf("unexpected entity type %q", e.Entity.Typename))
		}
	}

	return hierarchy, nil
}

// AddCloudAccountThroughManagementGroup adds the subscriptions to RSC through
// the management group, using the session from
// InitiateCloudAccountAddWithCustomerApp. RSC onboards the subscriptions
// asynchronously, the returned task chain ID can be used to wait for the
// onboarding to complete.
func (a API) AddCloudAccountThroughManagementGroup(ctx context.Context, sessionID uuid.UUID, tenantDomain string,
	regions []azure.Region, managementGroup ManagementGroupInput, subscriptions []ManagementGroupSubscription) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	query := addAzureCloudAccountQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		SessionID       secret.String                  `json:"sessionId"`
		TenantDomain    string                         `json:"tenantDomainName"`
		Regions         []azure.CloudAccountRegionEnum `json:"regions"`
		ManagementGroup ManagementGroupInput           `json:"managementGroup"`
		Subscriptions   []ManagementGroupSubscription  `json:"subscriptions"`
	}{
		SessionID:       secret.String(sessionID.String()),
		TenantDomain:    tenantDomain,
		Regions:         RegionsToCloudAccountRegionEnum(regions),
		ManagementGroup: managementGroup,
		Subscriptions:   subscriptions,
	})
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				TaskChainID string `json:"taskchainUuid"`
				Status      []struct {
					SubscriptionNativeID string `json:"azureSubscriptionNativeId"`
					Error                string `json:"error"`
				} `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	for _, status := range payload.Data.Result.Status {
		if status.Error != "" {
			return uuid.Nil, graphql.ResponseError(query, fmt.Errorf("subscription %s: %s", status.SubscriptionNativeID, status.Error))
		}
	}
	taskChainID, err := uuid.Parse(payload.Data.Result.TaskChainID)
	if err != nil {
		return uuid.Nil, graphql.ResponseError(query, fmt.Errorf("invalid task chain id: %s", err))
	}

	if taskChainID == uuid.Nil {
		return uuid.Nil, graphql.ResponseError(query, errors.New("empty task chain id"))
	}
	return taskChainID, nil
}

// DeleteCloudAccountsWithSession removes the specified feature and its Azure
// role assignments using the customer app session. Use this for features that
// do not support the asynchronous disable step.
func (a API) DeleteCloudAccountsWithSession(ctx context.Context, ids []uuid.UUID, feature core.Feature, sessionID uuid.UUID) error {
	a.log.Print(log.Trace)

	query := deleteAzureCloudAccountQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		IDs       []uuid.UUID   `json:"subscriptionIds"`
		Features  []string      `json:"features"`
		SessionID secret.String `json:"sessionId"`
	}{IDs: ids, Features: []string{feature.Name}, SessionID: secret.String(sessionID.String())})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Status []struct {
					NativeID string `json:"azureSubscriptionNativeId"`
					Success  bool   `json:"isSuccess"`
					Error    string `json:"error"`
				} `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	var failures []error
	for _, status := range payload.Data.Result.Status {
		if !status.Success || status.Error != "" {
			failures = append(failures, fmt.Errorf("delete failed for Azure subscription %s: %s", status.NativeID, status.Error))
		}
	}
	if len(payload.Data.Result.Status) != len(ids) {
		failures = append(failures, fmt.Errorf("expected %d deletion results, got %d", len(ids), len(payload.Data.Result.Status)))
	}
	if err := errors.Join(failures...); err != nil {
		return graphql.ResponseError(query, err)
	}
	return nil
}

// ManagementGroupDisableJob identifies an accepted deletion job and its target.
type ManagementGroupDisableJob struct {
	ID             uuid.UUID `json:"jobId"`
	CloudAccountID uuid.UUID `json:"rubrikObjectId"`
}

// StartDisableCloudAccountsJobWithSession starts task chains to disable the
// feature in the cloud accounts with the specified RSC cloud account IDs, using
// the session from InitiateCloudAccountAddWithCustomerApp. With a session, RSC
// also removes the Azure role assignments. Returns the RSC task chain IDs, one
// for each cloud account.
func (a API) StartDisableCloudAccountsJobWithSession(ctx context.Context, ids []uuid.UUID, feature core.Feature, sessionID uuid.UUID, deleteSnapshots bool) ([]ManagementGroupDisableJob, error) {
	a.log.Print(log.Trace)

	query := startDisableAzureCloudAccountJobWithSessionQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		IDs             []uuid.UUID   `json:"cloudAccountIds"`
		Feature         string        `json:"feature"`
		SessionID       secret.String `json:"sessionId"`
		DeleteSnapshots bool          `json:"deleteSnapshots"`
	}{IDs: ids, Feature: feature.Name, SessionID: secret.String(sessionID.String()), DeleteSnapshots: deleteSnapshots})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				JobIDs []ManagementGroupDisableJob `json:"jobIds"`
				Errors []struct {
					CloudAccountID uuid.UUID `json:"rubrikObjectId"`
					Error          string    `json:"error"`
				} `json:"errors"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}
	var failures []error
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, failure := range payload.Data.Result.Errors {
		seen[failure.CloudAccountID] = true
		failures = append(failures, fmt.Errorf("subscription %s: %s", failure.CloudAccountID, failure.Error))
	}
	jobs := payload.Data.Result.JobIDs
	for _, job := range jobs {
		if job.ID == uuid.Nil || job.CloudAccountID == uuid.Nil || seen[job.CloudAccountID] {
			failures = append(failures, fmt.Errorf("invalid or duplicate job for subscription %s", job.CloudAccountID))
		}
		seen[job.CloudAccountID] = true
	}
	for _, id := range ids {
		if !seen[id] {
			failures = append(failures, fmt.Errorf("no job or error returned for subscription %s", id))
		}
	}
	if err := errors.Join(failures...); err != nil {
		return jobs, graphql.ResponseError(query, err)
	}
	return jobs, nil
}
