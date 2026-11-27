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
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// managementGroupPollInterval is the time between status checks when waiting
// for a management group operation to complete.
const managementGroupPollInterval = 10 * time.Second

// ManagementGroup represents an Azure management group in RSC.
type ManagementGroup struct {
	NativeID     string    // Azure management group ID.
	Name         string    // Management group display name.
	CustomerID   uuid.UUID // RSC customer management group ID.
	IsAuthorized bool      // True if RSC can onboard the management group.
}

// ManagementGroupResourceGroup is the resource group to use for a feature
// onboarded through a management group.
type ManagementGroupResourceGroup struct {
	Name   string
	Region azure.Region
	Tags   map[string]string
}

// ManagementGroupFeature is a feature to enable on subscriptions onboarded
// through a management group.
type ManagementGroupFeature struct {
	Feature       core.Feature
	ResourceGroup *ManagementGroupResourceGroup
}

// ManagementGroupHierarchySubscription is a subscription found in a management
// group hierarchy.
type ManagementGroupHierarchySubscription struct {
	NativeID            uuid.UUID
	Name                string
	IsEligible          bool
	IneligibilityReason string
}

// ManagementGroupHierarchy is a flat listing. If the query recurses,
// Subscriptions includes descendant subscriptions, without parent IDs.
type ManagementGroupHierarchy struct {
	ManagementGroups []ManagementGroup
	Subscriptions    []ManagementGroupHierarchySubscription
}

// ManagementGroupHierarchyParams contains the inputs for listing a hierarchy.
type ManagementGroupHierarchyParams struct {
	TenantDomain      string
	AppID             uuid.UUID
	ManagementGroupID string
	// ScopedManagementGroupIDs contains Azure-native descendant management group
	// IDs, not display names or ARM resource paths. An empty list uses ManagementGroupID.
	ScopedManagementGroupIDs []string
	// NoRecurse limits the listing to immediate children of each requested scope.
	NoRecurse bool
	// Features controls the eligibility check only. An empty list checks Cloud Discovery.
	Features []core.Feature
}

// ManagementGroupCustomerAppID returns the configured non-OAuth customer app
// for a management group's tenant, not the tenant's primary OAuth app.
func (a API) ManagementGroupCustomerAppID(ctx context.Context, tenantDomain string) (uuid.UUID, error) {
	tenants, err := gqlazure.Wrap(a.client).CloudAccountTenants(ctx, core.FeatureAll, false)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to list Azure tenants: %s", err)
	}
	for _, tenant := range tenants {
		// DNS tenant domains are case-insensitive.
		if !strings.EqualFold(tenant.DomainName, tenantDomain) {
			continue
		}
		for _, app := range tenant.Apps {
			if app.AuthType == gqlazure.AuthTypeNonOAuth && app.ClientID != uuid.Nil {
				return app.ClientID, nil
			}
		}
		// Older, single-app tenants expose only the primary app. Do not
		// fall back when auth types are populated but none are non-OAuth.
		if len(tenant.Apps) == 0 && tenant.ClientID != uuid.Nil {
			return tenant.ClientID, nil
		}
	}
	return uuid.Nil, fmt.Errorf("non-OAuth customer app for %q: %w", tenantDomain, graphql.ErrNotFound)
}

// ManagementGroups returns the management groups visible to the customer app
// with the specified app ID. The secret stored in RSC for the app is used.
func (a API) ManagementGroups(ctx context.Context, tenantDomain string, appID uuid.UUID) ([]ManagementGroup, error) {
	a.log.Print(log.Trace)

	sessionID, err := a.managementGroupSession(ctx, tenantDomain, appID)
	if err != nil {
		return nil, err
	}

	groups, err := gqlazure.Wrap(a.client).ListManagementGroups(ctx, gqlazure.PublicCloud, tenantDomain, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list management groups: %s", err)
	}

	return toManagementGroups(groups), nil
}

// ManagementGroupHierarchy lists management groups and subscriptions under the
// requested scope. The secret stored in RSC for the app is used.
func (a API) ManagementGroupHierarchy(ctx context.Context, params ManagementGroupHierarchyParams) (ManagementGroupHierarchy, error) {
	a.log.Print(log.Trace)

	sessionID, err := a.managementGroupSession(ctx, params.TenantDomain, params.AppID)
	if err != nil {
		return ManagementGroupHierarchy{}, err
	}

	return a.managementGroupHierarchy(ctx, sessionID, params)
}

func (a API) managementGroupSession(ctx context.Context, tenantDomain string, appID uuid.UUID) (uuid.UUID, error) {
	sessionID, err := gqlazure.Wrap(a.client).InitiateCloudAccountAddWithCustomerApp(ctx, gqlazure.InitiateCloudAccountAddWithCustomerAppParams{
		Cloud:        gqlazure.PublicCloud,
		TenantDomain: tenantDomain,
		AppID:        appID,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to initiate session: %s", err)
	}
	return sessionID, nil
}

func (a API) managementGroupHierarchy(ctx context.Context, sessionID uuid.UUID, params ManagementGroupHierarchyParams) (ManagementGroupHierarchy, error) {
	features := core.FeatureNames(params.Features)
	if len(features) == 0 {
		features = []string{core.CloudDiscovery}
	}

	entities, err := gqlazure.Wrap(a.client).ListManagementGroupHierarchy(ctx, gqlazure.ListManagementGroupHierarchyParams{
		Cloud:                    gqlazure.PublicCloud,
		TenantDomain:             params.TenantDomain,
		SessionID:                sessionID,
		ManagementGroupID:        params.ManagementGroupID,
		ScopedManagementGroupIDs: params.ScopedManagementGroupIDs,
		ShouldRecurse:            !params.NoRecurse,
		Features:                 features,
		Operation:                gqlazure.OperationAdd,
		AuthType:                 gqlazure.AuthTypeNonOAuth,
	})
	if err != nil {
		return ManagementGroupHierarchy{}, fmt.Errorf("failed to list management group hierarchy: %s", err)
	}

	result := ManagementGroupHierarchy{Subscriptions: []ManagementGroupHierarchySubscription{}}
	for _, entry := range entities {
		entity := entry.Entity
		switch entity.Type {
		case gqlazure.ManagementGroupEntityType:
			result.ManagementGroups = append(result.ManagementGroups, ManagementGroup{
				NativeID:     entity.NativeID,
				Name:         entity.Name,
				IsAuthorized: entity.IsAuthorized,
			})
		case gqlazure.SubscriptionEntityType:
			nativeID, err := uuid.Parse(entity.NativeID)
			if err != nil {
				return ManagementGroupHierarchy{}, fmt.Errorf("invalid subscription ID %q in management group hierarchy: %s", entity.NativeID, err)
			}
			result.Subscriptions = append(result.Subscriptions, ManagementGroupHierarchySubscription{
				NativeID:            nativeID,
				Name:                entity.Name,
				IsEligible:          entry.IsEligible,
				IneligibilityReason: string(entry.IneligibilityReason),
			})
		default:
			return ManagementGroupHierarchy{}, fmt.Errorf("unexpected management group hierarchy entity type %q", entity.Type)
		}
	}

	return result, nil
}

// ManagementGroupSubscription is a subscription, and the features to enable on
// it, to add through a management group.
type ManagementGroupSubscription struct {
	NativeID uuid.UUID
	Features []ManagementGroupFeature
}

// StartManagementGroupOnboardingParams describes the subscriptions and features
// to add through a management group.
type StartManagementGroupOnboardingParams struct {
	TenantDomain      string
	AppID             uuid.UUID
	ManagementGroupID string
	Regions           []azure.Region
	Subscriptions     []ManagementGroupSubscription
}

// StartManagementGroupOnboarding submits non-OAuth onboarding and returns the
// accepted task ID without waiting. The customer app must already have the
// Contributor and User Access Administrator grants at management group scope.
func (a API) StartManagementGroupOnboarding(ctx context.Context, params StartManagementGroupOnboardingParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if len(params.Subscriptions) == 0 {
		return uuid.Nil, errors.New("at least one subscription is required")
	}
	var features []core.Feature
	for _, sub := range params.Subscriptions {
		if len(sub.Features) == 0 {
			return uuid.Nil, fmt.Errorf("subscription %s requires at least one feature", sub.NativeID)
		}
		for _, feature := range sub.Features {
			if !slices.ContainsFunc(features, feature.Feature.Equal) {
				features = append(features, feature.Feature)
			}
		}
	}

	gql := gqlazure.Wrap(a.client)
	sessionID, err := a.managementGroupSession(ctx, params.TenantDomain, params.AppID)
	if err != nil {
		return uuid.Nil, err
	}

	// RSC requires both the group ID and display name for onboarding.
	groups, err := gql.ListManagementGroups(ctx, gqlazure.PublicCloud, params.TenantDomain, sessionID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to list management groups: %s", err)
	}
	// Azure management group lookups accept case variants; use the returned native ID on subsequent requests.
	groupIdx := slices.IndexFunc(groups, func(g gqlazure.ManagementGroup) bool {
		return strings.EqualFold(g.NativeID, params.ManagementGroupID)
	})
	if groupIdx < 0 {
		return uuid.Nil, fmt.Errorf("management group %q not visible to the app; check the ID and the Contributor and User Access "+
			"Administrator grants at management group scope: %w", params.ManagementGroupID, graphql.ErrNotFound)
	}
	group := groups[groupIdx]

	// The hierarchy gives the subscription names and eligibility.
	hierarchy, err := a.managementGroupHierarchy(ctx, sessionID, ManagementGroupHierarchyParams{
		TenantDomain:      params.TenantDomain,
		ManagementGroupID: group.NativeID,
		Features:          features,
	})
	if err != nil {
		return uuid.Nil, err
	}

	input := make([]gqlazure.ManagementGroupSubscription, 0, len(params.Subscriptions))
	for _, sub := range params.Subscriptions {
		idx := slices.IndexFunc(hierarchy.Subscriptions, func(s ManagementGroupHierarchySubscription) bool {
			return s.NativeID == sub.NativeID
		})
		if idx < 0 {
			return uuid.Nil, fmt.Errorf("subscription %s not found in management group %q: %w", sub.NativeID, group.NativeID, graphql.ErrNotFound)
		}
		if found := hierarchy.Subscriptions[idx]; !found.IsEligible {
			return uuid.Nil, fmt.Errorf("subscription %s is not eligible for onboarding: %s", sub.NativeID, found.IneligibilityReason)
		}

		inputFeatures := make([]gqlazure.ManagementGroupFeature, 0, len(sub.Features))
		for _, feature := range sub.Features {
			permGroups := make([]string, 0, len(feature.Feature.PermissionGroups))
			for _, permGroup := range feature.Feature.PermissionGroups {
				permGroups = append(permGroups, string(permGroup))
			}

			var resourceGroup *gqlazure.ResourceGroup
			if rg := feature.ResourceGroup; rg != nil {
				tags := make([]core.Tag, 0, len(rg.Tags))
				for key, value := range rg.Tags {
					tags = append(tags, core.Tag{Key: key, Value: value})
				}
				resourceGroup = &gqlazure.ResourceGroup{
					Name:    rg.Name,
					TagList: gqlazure.TagList{Tags: tags},
					Region:  rg.Region.ToCloudAccountRegionEnum(),
				}
			}
			inputFeatures = append(inputFeatures, gqlazure.ManagementGroupFeature{
				FeatureType:      feature.Feature.Name,
				PermissionGroups: permGroups,
				ResourceGroup:    resourceGroup,
			})
		}
		input = append(input, gqlazure.ManagementGroupSubscription{
			Subscription: gqlazure.SubscriptionInput{NativeID: sub.NativeID, Name: hierarchy.Subscriptions[idx].Name},
			Features:     inputFeatures,
		})
	}

	taskChainID, err := gql.AddCloudAccountThroughManagementGroup(ctx, gqlazure.AddCloudAccountThroughManagementGroupParams{
		SessionID:       sessionID,
		TenantDomain:    params.TenantDomain,
		Regions:         params.Regions,
		ManagementGroup: gqlazure.ManagementGroupInput{NativeID: group.NativeID, Name: group.Name},
		Subscriptions:   input,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add subscriptions: %s", err)
	}

	return taskChainID, nil
}

// WaitForManagementGroupOnboardingParams specifies the accepted task and the
// subscriptions and features it should connect.
type WaitForManagementGroupOnboardingParams struct {
	TaskChainID       uuid.UUID
	TenantDomain      string
	ManagementGroupID string
	Subscriptions     []ManagementGroupSubscription
}

// WaitForManagementGroupOnboarding waits for the accepted task to finish and
// verifies that each requested feature is connected through the expected group.
func (a API) WaitForManagementGroupOnboarding(ctx context.Context, params WaitForManagementGroupOnboardingParams) error {
	a.log.Print(log.Trace)
	if err := a.WaitForManagementGroupTask(ctx, params.TaskChainID); err != nil {
		return err
	}
	accounts, err := a.Subscriptions(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to verify onboarding task %s: %s", params.TaskChainID, err)
	}
	var failures []error
	for _, sub := range params.Subscriptions {
		idx := slices.IndexFunc(accounts, func(c CloudAccount) bool {
			return c.NativeID == sub.NativeID && strings.EqualFold(c.TenantDomain, params.TenantDomain) &&
				c.ManagementGroup != nil && strings.EqualFold(c.ManagementGroup.NativeID, params.ManagementGroupID) &&
				c.AuthType != gqlazure.AuthTypeOAuthRubrikApp && c.AuthType != gqlazure.AuthTypeOAuthCustomerApp
		})
		for _, feature := range sub.Features {
			if idx < 0 {
				failures = append(failures, fmt.Errorf("subscription %s feature %s was not onboarded", sub.NativeID, feature.Feature.Name))
				continue
			}
			onboarded, ok := accounts[idx].Feature(feature.Feature)
			if !ok || onboarded.Status != core.StatusConnected {
				failures = append(failures, fmt.Errorf("subscription %s feature %s did not connect (status %q)", sub.NativeID, feature.Feature.Name, onboarded.Status))
			}
		}
	}
	return errors.Join(failures...)
}

// WaitForManagementGroupTask waits for the task to succeed, fail or be canceled.
// A terminal failure is reported as an error.
func (a API) WaitForManagementGroupTask(ctx context.Context, taskChainID uuid.UUID) error {
	state, err := core.Wrap(a.client).WaitForTaskChain(ctx, taskChainID, managementGroupPollInterval)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("waiting for management group task %s: %s", taskChainID, err)
	}
	if state != core.TaskChainSucceeded {
		return fmt.Errorf("management group task %s ended with state %s", taskChainID, state)
	}
	return nil
}

// RemoveManagementGroupSubscriptionFeatureParams describes a feature to remove
// using a customer app and the RSC cloud account IDs of its subscriptions.
type RemoveManagementGroupSubscriptionFeatureParams struct {
	TenantDomain    string
	AppID           uuid.UUID
	CloudAccountIDs []uuid.UUID
	Feature         core.Feature
	DeleteSnapshots bool
}

// RemoveManagementGroupSubscriptionFeature removes the feature and its Azure
// role assignments using the customer app's secret stored in RSC. Keep the
// app's Contributor and User Access Administrator grants until removal finishes.
//
// If DeleteSnapshots is true, the disable job deletes the protection snapshots
// rather than retaining them. The function blocks until removal is verified.
func (a API) RemoveManagementGroupSubscriptionFeature(ctx context.Context, params RemoveManagementGroupSubscriptionFeatureParams) error {
	a.log.Print(log.Trace)
	if len(params.CloudAccountIDs) == 0 {
		return nil
	}

	// Only features supported by RSC's Azure disable-job workflow need a job.
	// This includes Discovery and Exocompute, not just workload protection.
	feature := params.Feature
	disable := false
	switch feature.Name {
	case core.CloudNativeProtection, core.AzureSQLDBProtection, core.AzureSQLMIProtection,
		core.CloudNativeBlobProtection, core.AzurePostgresFlexibleServerProtection, core.AzureCosmosNoSQLProtection,
		core.Exocompute, core.CloudDiscovery, core.Archival, core.CloudCostReport:
		disable = true
	}
	combined := false
	if disable {
		flag, err := core.Wrap(a.client).FeatureFlag(ctx, core.FeatureFlagAzureDisableDeleteCombined)
		if err != nil {
			return fmt.Errorf("failed to check Azure combined disable and delete flag: %s", err)
		}
		if flag.Name == "" {
			return errors.New("RSC did not return the Azure combined disable and delete flag")
		}
		combined = flag.Enabled
	}

	sessionStarted := time.Now()
	sessionID, err := a.managementGroupSession(ctx, params.TenantDomain, params.AppID)
	if err != nil {
		return err
	}
	gql := gqlazure.Wrap(a.client)
	verifyAbsent := func(ctx context.Context, ids []uuid.UUID) error {
		accounts, err := a.Subscriptions(ctx, "")
		if err != nil {
			return fmt.Errorf("failed to verify removal of %s: %s", feature, err)
		}
		requested := make(map[uuid.UUID]bool, len(ids))
		for _, id := range ids {
			requested[id] = true
		}
		var failures []error
		for _, account := range accounts {
			if !requested[account.ID] {
				continue
			}
			if _, present := account.Feature(feature); present {
				failures = append(failures, fmt.Errorf("feature %s remains on subscription %s after removal", feature, account.ID))
			}
		}
		return errors.Join(failures...)
	}
	deleteFeature := func(ctx context.Context, ids []uuid.UUID, sessionID uuid.UUID) error {
		err := gql.DeleteCloudAccountsWithSession(ctx, ids, feature, sessionID)
		if err != nil {
			err = fmt.Errorf("failed to delete %s: %s", feature, err)
		}
		return errors.Join(err, verifyAbsent(ctx, ids))
	}
	if !disable {
		return deleteFeature(ctx, params.CloudAccountIDs, sessionID)
	}

	jobs, submitErr := gql.StartDisableCloudAccountsJobWithSession(ctx, gqlazure.StartDisableCloudAccountsJobWithSessionParams{
		CloudAccountIDs: params.CloudAccountIDs,
		Feature:         feature,
		SessionID:       sessionID,
		DeleteSnapshots: params.DeleteSnapshots,
	})
	var failures []error
	if submitErr != nil {
		failures = append(failures, fmt.Errorf("failed to schedule removal of %s: %s", feature, submitErr))
	}
	var toDelete []uuid.UUID
	for _, job := range jobs {
		if err := core.Wrap(a.client).WaitForFeatureDisableTaskChain(ctx, job.ID, func(ctx context.Context) (bool, error) {
			status, present, err := a.managementGroupFeatureStatus(ctx, job.CloudAccountID, feature)
			// On RBAC errors, DISABLED is enough only when the client will run the delete step.
			return !present || (!combined && status == core.StatusDisabled), err
		}); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				failures = append(failures, err)
			} else {
				failures = append(failures, fmt.Errorf("failed to wait for disable task %s for subscription %s: %s", job.ID, job.CloudAccountID, err))
			}
			continue
		}

		status, present, err := a.managementGroupFeatureStatus(ctx, job.CloudAccountID, feature)
		switch {
		case err != nil:
			failures = append(failures, err)
		case !present:
			// The combined job or another removal already deleted the feature.
		case combined:
			failures = append(failures, fmt.Errorf("task %s finished but feature %s remains on subscription %s", job.ID, feature, job.CloudAccountID))
		case status != core.StatusDisabled:
			failures = append(failures, fmt.Errorf("task %s finished but feature %s has status %q on subscription %s, expected DISABLED", job.ID, feature, status, job.CloudAccountID))
		default:
			toDelete = append(toDelete, job.CloudAccountID)
		}
	}
	if len(toDelete) != 0 {
		// A long disable job can outlive the 60-minute session. Get a fresh
		// one if the original is close to expiring before the delete call.
		if time.Since(sessionStarted) >= 55*time.Minute {
			sessionID, err = a.managementGroupSession(ctx, params.TenantDomain, params.AppID)
			if err != nil {
				return errors.Join(append(failures, err)...)
			}
		}
		failures = append(failures, deleteFeature(ctx, toDelete, sessionID))
	}
	return errors.Join(failures...)
}

// managementGroupFeatureStatus treats a missing cloud account or feature as
// absent. Any other read error must not be mistaken for a successful removal.
func (a API) managementGroupFeatureStatus(ctx context.Context, id uuid.UUID, feature core.Feature) (core.Status, bool, error) {
	account, err := a.SubscriptionByID(ctx, id)
	if errors.Is(err, graphql.ErrNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to verify %s on subscription %s: %s", feature, id, err)
	}
	f, present := account.Feature(feature)
	return f.Status, present, nil
}

func toManagementGroups(groups []gqlazure.ManagementGroup) []ManagementGroup {
	result := make([]ManagementGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, ManagementGroup{
			NativeID:     group.NativeID,
			Name:         group.Name,
			CustomerID:   group.CustomerManagementGroupID,
			IsAuthorized: group.IsAuthorized,
		})
	}

	return result
}
