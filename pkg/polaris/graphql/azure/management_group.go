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

// ManagementGroupHierarchyEntityType identifies a member of the hierarchy entity union.
type ManagementGroupHierarchyEntityType string

const (
	ManagementGroupEntityType ManagementGroupHierarchyEntityType = "AzureManagementGroup"
	SubscriptionEntityType    ManagementGroupHierarchyEntityType = "CloudAccountsAzureSubscription"
)

// ManagementGroupHierarchyEntity is an entry in the flat hierarchy response.
type ManagementGroupHierarchyEntity struct {
	IsEligible          bool                           `json:"isEligible"`
	IneligibilityReason IneligibilityReason            `json:"ineligibilityReason"`
	Entity              ManagementGroupHierarchyMember `json:"entity"`
}

// ManagementGroupHierarchyMember is the projection shared by both union members.
// NativeID is a string because only subscription IDs are always UUIDs.
type ManagementGroupHierarchyMember struct {
	Type         ManagementGroupHierarchyEntityType `json:"__typename"`
	NativeID     string                             `json:"nativeId"`
	Name         string                             `json:"name"`
	IsAuthorized bool                               `json:"isAuthorized"`
}

// ListManagementGroupHierarchyParams holds the parameters for listing the
// hierarchy under a management group.
type ListManagementGroupHierarchyParams struct {
	Cloud             Cloud     `json:"cloudType"`
	TenantDomain      string    `json:"tenantDomainName"`
	SessionID         uuid.UUID `json:"sessionId"`
	ManagementGroupID string    `json:"managementGroupId"`
	// ScopedManagementGroupIDs contains the Azure-native IDs of descendant
	// management groups. An empty list starts the lookup at ManagementGroupID.
	ScopedManagementGroupIDs []string      `json:"scopedManagementGroupIds,omitempty"`
	ShouldRecurse            bool          `json:"shouldRecurse"`
	Features                 []string      `json:"features,omitempty"`
	Operation                OperationType `json:"operationType"`
	AuthType                 AuthType      `json:"authType"`
}

// InitiateCloudAccountAddWithCustomerAppParams contains the credentials used to
// start an add cloud account session. If AppSecret is empty, RSC uses its stored secret.
type InitiateCloudAccountAddWithCustomerAppParams struct {
	Cloud        Cloud         `json:"cloudType"`
	TenantDomain string        `json:"tenantDomainName"`
	AppID        uuid.UUID     `json:"appId"`
	AppSecret    secret.String `json:"appSecretKey,omitempty"`
}

// InitiateCloudAccountAddWithCustomerApp starts an add cloud account session
// using the customer app credentials. The session expires after 60 minutes;
// callers do not need to close it.
func (a API) InitiateCloudAccountAddWithCustomerApp(ctx context.Context, params InitiateCloudAccountAddWithCustomerAppParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	// The shared session validator requires the UI callback URL even
	// when customer app authentication does not open a browser.
	endpoint, err := url.Parse(a.GQL.Endpoint())
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid RSC GraphQL URL: %s", err)
	}
	query := azureCloudAccountAddWithCustomerAppInitiateQuery
	vars := struct {
		InitiateCloudAccountAddWithCustomerAppParams
		RedirectURL string `json:"redirectUrl"`
	}{
		InitiateCloudAccountAddWithCustomerAppParams: params,
		RedirectURL: endpoint.ResolveReference(&url.URL{Path: "/setup_azure"}).String(),
	}
	// The response contains a session ID; don't log the raw payload.
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

// ListManagementGroupHierarchy returns the flat list of RSC hierarchy entities
// under the requested management group or scopes.
func (a API) ListManagementGroupHierarchy(ctx context.Context, params ListManagementGroupHierarchyParams) ([]ManagementGroupHierarchyEntity, error) {
	a.log.Print(log.Trace)

	query := azureListManagementGroupHierarchyQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ListManagementGroupHierarchyParams
		SessionID secret.String `json:"sessionId"`
	}{ListManagementGroupHierarchyParams: params, SessionID: secret.String(params.SessionID.String())})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Entities []ManagementGroupHierarchyEntity `json:"entities"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Entities, nil
}

// AddCloudAccountThroughManagementGroupParams contains the subscriptions to add
// through a management group and the customer app session used to add them.
type AddCloudAccountThroughManagementGroupParams struct {
	SessionID       uuid.UUID
	TenantDomain    string
	Regions         []azure.Region
	ManagementGroup ManagementGroupInput
	Subscriptions   []ManagementGroupSubscription
}

// AddCloudAccountThroughManagementGroup adds the subscriptions to RSC through
// the management group, using the session from
// InitiateCloudAccountAddWithCustomerApp. RSC onboards the subscriptions
// asynchronously, the returned task chain ID can be used to wait for the
// onboarding to complete.
func (a API) AddCloudAccountThroughManagementGroup(ctx context.Context, params AddCloudAccountThroughManagementGroupParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	query := addAzureCloudAccountQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		SessionID       secret.String                  `json:"sessionId"`
		TenantDomain    string                         `json:"tenantDomainName"`
		Regions         []azure.CloudAccountRegionEnum `json:"regions"`
		ManagementGroup ManagementGroupInput           `json:"managementGroup"`
		Subscriptions   []ManagementGroupSubscription  `json:"subscriptions"`
	}{
		SessionID:       secret.String(params.SessionID.String()),
		TenantDomain:    params.TenantDomain,
		Regions:         RegionsToCloudAccountRegionEnum(params.Regions),
		ManagementGroup: params.ManagementGroup,
		Subscriptions:   params.Subscriptions,
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
// do not need a disable job or after a disable-only job has completed.
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

// ManagementGroupDisableJob identifies an accepted disable job and its target.
type ManagementGroupDisableJob struct {
	ID             uuid.UUID `json:"jobId"`
	CloudAccountID uuid.UUID `json:"rubrikObjectId"`
}

// StartDisableCloudAccountsJobWithSessionParams contains the feature, targets,
// and customer app session for a disable job.
type StartDisableCloudAccountsJobWithSessionParams struct {
	CloudAccountIDs []uuid.UUID
	Feature         core.Feature
	SessionID       uuid.UUID
	DeleteSnapshots bool
}

// StartDisableCloudAccountsJobWithSession starts one disable task chain per
// cloud account. If combined disable and delete is enabled, the task also
// deletes the feature and uses the session to remove Azure role assignments.
// Accepted jobs are returned even if other targets fail submission.
func (a API) StartDisableCloudAccountsJobWithSession(ctx context.Context, params StartDisableCloudAccountsJobWithSessionParams) ([]ManagementGroupDisableJob, error) {
	a.log.Print(log.Trace)

	query := startDisableAzureCloudAccountJobWithSessionQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		IDs             []uuid.UUID   `json:"cloudAccountIds"`
		Feature         string        `json:"feature"`
		SessionID       secret.String `json:"sessionId"`
		DeleteSnapshots bool          `json:"deleteSnapshots"`
	}{IDs: params.CloudAccountIDs, Feature: params.Feature.Name, SessionID: secret.String(params.SessionID.String()), DeleteSnapshots: params.DeleteSnapshots})
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
	requested := make(map[uuid.UUID]bool, len(params.CloudAccountIDs))
	for _, id := range params.CloudAccountIDs {
		requested[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(requested))
	for _, failure := range payload.Data.Result.Errors {
		seen[failure.CloudAccountID] = true
		failures = append(failures, fmt.Errorf("subscription %s: %s", failure.CloudAccountID, failure.Error))
	}
	var jobs []ManagementGroupDisableJob
	for _, job := range payload.Data.Result.JobIDs {
		if job.ID == uuid.Nil || job.CloudAccountID == uuid.Nil || !requested[job.CloudAccountID] || seen[job.CloudAccountID] {
			failures = append(failures, fmt.Errorf("invalid or duplicate job for subscription %s", job.CloudAccountID))
			continue
		}
		seen[job.CloudAccountID] = true
		jobs = append(jobs, job)
	}
	for _, id := range params.CloudAccountIDs {
		if !seen[id] {
			failures = append(failures, fmt.Errorf("no job or error returned for subscription %s", id))
		}
	}
	if err := errors.Join(failures...); err != nil {
		return jobs, graphql.ResponseError(query, err)
	}
	return jobs, nil
}
