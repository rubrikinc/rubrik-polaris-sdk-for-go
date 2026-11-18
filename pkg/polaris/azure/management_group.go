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

// ManagementGroupHierarchy is the flattened hierarchy under a management group.
type ManagementGroupHierarchy struct {
	ManagementGroups []ManagementGroup
	Subscriptions    []ManagementGroupHierarchySubscription
}

// ManagementGroupHierarchyOptions are the options for listing a management
// group hierarchy.
type ManagementGroupHierarchyOptions struct {
	// ScopedManagementGroupIDs limits the hierarchy to the specified child
	// management groups.
	ScopedManagementGroupIDs []string
	// NoRecurse limits the hierarchy to the immediate children.
	NoRecurse bool
	// Features are the features to check subscription eligibility for.
	Features []core.Feature
}

// ManagementGroups returns the management groups visible to the customer app
// with the specified app ID. The secret stored in RSC for the app is used.
func (a API) ManagementGroups(ctx context.Context, tenantDomain string, appID uuid.UUID) ([]ManagementGroup, error) {
	a.log.Print(log.Trace)

	gql := gqlazure.Wrap(a.client)
	sessionID, err := gql.InitiateCloudAccountAddWithCustomerApp(ctx, gqlazure.PublicCloud, tenantDomain, appID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to initiate session: %w", err)
	}

	groups, err := gql.ListManagementGroups(ctx, gqlazure.PublicCloud, tenantDomain, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list management groups: %w", err)
	}

	return toManagementGroups(groups), nil
}

// ManagementGroupHierarchy returns the flattened hierarchy of management groups
// and subscriptions under the management group with the specified native ID.
// The secret stored in RSC for the app is used.
func (a API) ManagementGroupHierarchy(ctx context.Context, tenantDomain string, appID uuid.UUID, managementGroupID string,
	opts ManagementGroupHierarchyOptions) (ManagementGroupHierarchy, error) {
	a.log.Print(log.Trace)

	gql := gqlazure.Wrap(a.client)
	sessionID, err := gql.InitiateCloudAccountAddWithCustomerApp(ctx, gqlazure.PublicCloud, tenantDomain, appID, "")
	if err != nil {
		return ManagementGroupHierarchy{}, fmt.Errorf("failed to initiate session: %w", err)
	}

	return a.managementGroupHierarchy(ctx, sessionID, tenantDomain, managementGroupID, opts)
}

func (a API) managementGroupHierarchy(ctx context.Context, sessionID uuid.UUID, tenantDomain, managementGroupID string,
	opts ManagementGroupHierarchyOptions) (ManagementGroupHierarchy, error) {
	features := make([]string, 0, len(opts.Features))
	for _, feature := range opts.Features {
		features = append(features, feature.Name)
	}

	hierarchy, err := gqlazure.Wrap(a.client).ListManagementGroupHierarchy(ctx, gqlazure.ManagementGroupHierarchyRequest{
		Cloud:                    gqlazure.PublicCloud,
		TenantDomain:             tenantDomain,
		SessionID:                sessionID,
		ManagementGroupID:        managementGroupID,
		ScopedManagementGroupIDs: opts.ScopedManagementGroupIDs,
		ShouldRecurse:            !opts.NoRecurse,
		Features:                 features,
		Operation:                gqlazure.OperationAdd,
		AuthType:                 gqlazure.AuthTypeNonOAuth,
	})
	if err != nil {
		return ManagementGroupHierarchy{}, fmt.Errorf("failed to list management group hierarchy: %w", err)
	}

	result := ManagementGroupHierarchy{
		Subscriptions: make([]ManagementGroupHierarchySubscription, 0, len(hierarchy.Subscriptions)),
	}
	for _, group := range hierarchy.ManagementGroups {
		result.ManagementGroups = append(result.ManagementGroups, ManagementGroup{
			NativeID:     group.NativeID,
			Name:         group.Name,
			CustomerID:   group.CustomerManagementGroupID,
			IsAuthorized: group.IsAuthorized,
		})
	}
	for _, sub := range hierarchy.Subscriptions {
		result.Subscriptions = append(result.Subscriptions, ManagementGroupHierarchySubscription{
			NativeID:            sub.NativeID,
			Name:                sub.Name,
			IsEligible:          sub.IsEligible,
			IneligibilityReason: string(sub.IneligibilityReason),
		})
	}

	return result, nil
}

// ManagementGroupSubscription is a subscription, and the features to enable on
// it, to add through a management group.
type ManagementGroupSubscription struct {
	NativeID uuid.UUID
	Features []ManagementGroupFeature
}

// AddManagementGroupSubscriptions adds the subscriptions to RSC through the
// management group, without OAuth. The app's service principal must have been
// given the Contributor and User Access Administrator roles at the management
// group scope. The secret stored in RSC for the app is used.
//
// Each subscription must be eligible for all of its features. The function
// blocks until RSC has onboarded all the features, and returns an error if the
// onboarding fails or only partially succeeds.
func (a API) AddManagementGroupSubscriptions(ctx context.Context, tenantDomain string, appID uuid.UUID, managementGroupID string,
	regions []azure.Region, subscriptions []ManagementGroupSubscription) error {
	a.log.Print(log.Trace)

	if len(subscriptions) == 0 {
		return errors.New("at least one subscription is required")
	}

	gql := gqlazure.Wrap(a.client)
	sessionID, err := gql.InitiateCloudAccountAddWithCustomerApp(ctx, gqlazure.PublicCloud, tenantDomain, appID, "")
	if err != nil {
		return fmt.Errorf("failed to initiate session: %w", err)
	}

	// The management group name is required by RSC. An empty list means the
	// elevated permissions have not been granted to the app.
	groups, err := gql.ListManagementGroups(ctx, gqlazure.PublicCloud, tenantDomain, sessionID)
	if err != nil {
		return fmt.Errorf("failed to list management groups: %w", err)
	}
	groupIdx := slices.IndexFunc(groups, func(g gqlazure.ManagementGroup) bool {
		return strings.EqualFold(g.NativeID, managementGroupID)
	})
	if groupIdx < 0 {
		return fmt.Errorf("management group %q not visible to the app, grant the Contributor and User Access "+
			"Administrator roles to the app's service principal at the management group scope: %w",
			managementGroupID, graphql.ErrNotFound)
	}
	group := groups[groupIdx]

	// The hierarchy gives the subscription names and the eligibility.
	var features []core.Feature
	for _, sub := range subscriptions {
		for _, feature := range sub.Features {
			if !slices.ContainsFunc(features, feature.Feature.Equal) {
				features = append(features, feature.Feature)
			}
		}
	}
	hierarchy, err := a.managementGroupHierarchy(ctx, sessionID, tenantDomain, managementGroupID,
		ManagementGroupHierarchyOptions{Features: features})
	if err != nil {
		return err
	}

	input := make([]gqlazure.ManagementGroupSubscription, 0, len(subscriptions))
	for _, sub := range subscriptions {
		idx := slices.IndexFunc(hierarchy.Subscriptions, func(s ManagementGroupHierarchySubscription) bool {
			return s.NativeID == sub.NativeID
		})
		if idx < 0 {
			return fmt.Errorf("subscription %s not found in management group %q: %w", sub.NativeID, managementGroupID, graphql.ErrNotFound)
		}
		if found := hierarchy.Subscriptions[idx]; !found.IsEligible {
			return fmt.Errorf("subscription %s is not eligible for onboarding: %s", sub.NativeID, found.IneligibilityReason)
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

	taskChainID, err := gql.AddCloudAccountThroughManagementGroup(ctx, sessionID, tenantDomain, regions,
		gqlazure.ManagementGroupInput{NativeID: group.NativeID, Name: group.Name}, input)
	if err != nil {
		return fmt.Errorf("failed to add subscriptions: %w", err)
	}

	// The task chain can succeed even though some of the subscriptions failed
	// to onboard, so we also check that every feature has been connected.
	err = core.Wrap(a.client).WaitForTaskChainWithStatus(ctx, taskChainID, managementGroupPollInterval, func(ctx context.Context) (bool, error) {
		accounts, err := a.Subscriptions(ctx, "")
		if err != nil {
			return false, fmt.Errorf("failed to retrieve subscriptions: %w", err)
		}
		for _, sub := range subscriptions {
			idx := slices.IndexFunc(accounts, func(c CloudAccount) bool { return c.NativeID == sub.NativeID })
			if idx < 0 {
				return false, nil
			}
			for _, feature := range sub.Features {
				onboarded, ok := accounts[idx].Feature(feature.Feature)
				if !ok || onboarded.Status != core.StatusConnected {
					return false, nil
				}
			}
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait for task chain %s: %w", taskChainID, err)
	}

	return nil
}

// RemoveManagementGroupSubscriptionFeature disables the feature on the
// subscriptions with the specified RSC cloud account IDs, using the customer
// app so that RSC also removes the Azure role assignments. The app's service
// principal must still have the Contributor and User Access Administrator
// roles at the management group scope. The secret stored in RSC for the app is
// used.
//
// If deleteSnapshots is true, the snapshots of a protection feature are
// deleted, otherwise they are kept. The function blocks until the feature has
// been disabled on all the subscriptions.
func (a API) RemoveManagementGroupSubscriptionFeature(ctx context.Context, tenantDomain string, appID uuid.UUID,
	cloudAccountIDs []uuid.UUID, feature core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	if len(cloudAccountIDs) == 0 {
		return nil
	}

	gql := gqlazure.Wrap(a.client)
	sessionID, err := gql.InitiateCloudAccountAddWithCustomerApp(ctx, gqlazure.PublicCloud, tenantDomain, appID, "")
	if err != nil {
		return fmt.Errorf("failed to initiate session: %w", err)
	}

	jobIDs, err := gql.StartDisableCloudAccountsJobWithSession(ctx, cloudAccountIDs, feature, sessionID, deleteSnapshots)
	if err != nil {
		return fmt.Errorf("failed to disable feature %s: %w", feature, err)
	}

	for _, jobID := range jobIDs {
		err := core.Wrap(a.client).WaitForTaskChainWithStatus(ctx, jobID, managementGroupPollInterval, func(ctx context.Context) (bool, error) {
			for _, id := range cloudAccountIDs {
				account, err := a.SubscriptionByID(ctx, id)
				if errors.Is(err, graphql.ErrNotFound) {
					continue
				}
				if err != nil {
					return false, fmt.Errorf("failed to retrieve subscription %s: %w", id, err)
				}
				// A missing feature has already been removed.
				if f, ok := account.Feature(feature); ok && f.Status != core.StatusDisabled {
					return false, nil
				}
			}
			return true, nil
		})
		if err != nil {
			return fmt.Errorf("failed to wait for task chain %s: %w", jobID, err)
		}
	}

	return nil
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
