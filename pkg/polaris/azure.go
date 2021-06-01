// Copyright 2021 Rubrik, Inc.
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

package polaris

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

// hasPrefix returns true if at least one of the strings has at least one of
// the prefixes.
func hasPrefix(ztrings, prefixes []string) bool {
	for _, z := range ztrings {
		for _, p := range prefixes {
			if strings.HasPrefix(z, p) {
				return true
			}
		}
	}

	return false
}

// AzureFeature Azure feature.
type AzureFeature struct {
	Name    graphql.CloudAccountFeature
	Status  graphql.CloudAccountStatus
	Regions []graphql.AzureRegion
}

// AzureSubscription Azure subscription. Note that ID is the Polaris
// subscription id and NativeID the actual Azure subscription id.
type AzureSubscription struct {
	ID           uuid.UUID
	NativeID     uuid.UUID
	Name         string
	TenantDomain string
	Feature      AzureFeature
}

// toAzureSubscription converts a graphql AzureSubscription to a polaris
// AzureSubscription.
func toAzureSubscription(tenantDomain string, subscription graphql.AzureSubscription) (AzureSubscription, error) {
	subID, err := uuid.Parse(subscription.ID)
	if err != nil {
		return AzureSubscription{}, err
	}

	subNativeID, err := uuid.Parse(subscription.NativeID)
	if err != nil {
		return AzureSubscription{}, err
	}

	sub := AzureSubscription{
		ID:           subID,
		NativeID:     subNativeID,
		Name:         subscription.Name,
		TenantDomain: tenantDomain,
		Feature: AzureFeature{
			Name:    subscription.FeatureDetail.Feature,
			Status:  subscription.FeatureDetail.Status,
			Regions: subscription.FeatureDetail.Regions,
		},
	}

	return sub, nil
}

// AzureSubscriptionID -
type AzureSubscriptionID interface {
	subscriptionID() (uuid.UUID, error)
}

type azureSubscriptionID struct {
	parse func() (uuid.UUID, error)
}

func (id *azureSubscriptionID) subscriptionID() (uuid.UUID, error) {
	return id.parse()
}

// WithAzureSubscriptionID -
func WithAzureSubscriptionID(subscriptionID string) *azureSubscriptionID {
	return &azureSubscriptionID{func() (uuid.UUID, error) {
		id, err := uuid.Parse(subscriptionID)
		if err != nil {
			return uuid.Nil, err
		}

		return id, nil
	}}
}

type polarisSubscriptionID struct {
	parse func() (uuid.UUID, error)
}

func (id *polarisSubscriptionID) subscriptionID() (uuid.UUID, error) {
	return id.parse()
}

// WithPolarisSubscriptionID -
func WithPolarisSubscriptionID(subscriptionID string) *polarisSubscriptionID {
	return &polarisSubscriptionID{func() (uuid.UUID, error) {
		id, err := uuid.Parse(subscriptionID)
		if err != nil {
			return uuid.Nil, err
		}

		return id, nil
	}}
}

// AzureSubscriptions returns the Azure subscription identified by the
// specified subscription id.
func (c *Client) AzureSubscription(ctx context.Context, id AzureSubscriptionID) (AzureSubscription, error) {
	subID, err := id.subscriptionID()
	if err != nil {
		return AzureSubscription{}, err
	}

	tenants, err := c.gql.AzureCloudAccountTenants(ctx, graphql.CloudNativeProtection, true)
	if err != nil {
		return AzureSubscription{}, err
	}

	var equals func(uuid.UUID, graphql.AzureSubscription) bool
	switch id.(type) {
	case *azureSubscriptionID:
		equals = func(id uuid.UUID, sub graphql.AzureSubscription) bool {
			return id.String() == sub.NativeID
		}
	case *polarisSubscriptionID:
		equals = func(id uuid.UUID, sub graphql.AzureSubscription) bool {
			return id.String() == sub.ID
		}
	default:
		return AzureSubscription{}, errors.New("polaris: invalid subscription id")
	}

	for _, tenant := range tenants {
		for _, subscription := range tenant.Subscriptions {
			if equals(subID, subscription) {
				return toAzureSubscription(tenant.DomainName, subscription)
			}
		}
	}

	return AzureSubscription{}, fmt.Errorf("polaris: subscription %w", ErrNotFound)
}

// AzureSubscriptionQuery -
type AzureSubscriptionQuery interface {
	subscription(query *azureQueryBuilder)
}

type azureQueryBuilder struct {
	status map[graphql.CloudAccountStatus]struct{}
	prefix []string
}

type azureSubscriptionQuery struct {
	build func(query *azureQueryBuilder)
}

func (status *azureSubscriptionQuery) subscription(builder *azureQueryBuilder) {
	status.build(builder)
}

// WithStatus returns an Azure subscription query for a specific status.
func WithStatus(status graphql.CloudAccountStatus) *azureSubscriptionQuery {
	return &azureSubscriptionQuery{func(query *azureQueryBuilder) {
		if query.status == nil {
			query.status = make(map[graphql.CloudAccountStatus]struct{})
		}

		query.status[status] = struct{}{}
	}}
}

// WithPrefix returns an Azure subscription query for a specific prefix.
// The prefix is matched against the Azure subscription id, the Polaris
// subscription id and the subscription name.
func WithPrefix(prefix string) *azureSubscriptionQuery {
	return &azureSubscriptionQuery{func(query *azureQueryBuilder) {
		query.prefix = append(query.prefix, prefix)
	}}
}

// AzureSubscriptions returns a collection of Azure subscriptions matching the
// specified query. WithPrefix can be used to search for a prefix on the Azure
// subscription id, the Polaris subscription id or the subscription name.
// If both WithPrefix and WithStatus is given only subscription having both
// conditions are included in the collection. If multiple WithPrefix are given
// subscriptions have any of the prefixes are included in the collection.
func (c *Client) AzureSubscriptions(ctx context.Context, query ...AzureSubscriptionQuery) ([]AzureSubscription, error) {
	builder := azureQueryBuilder{}
	for _, q := range query {
		q.subscription(&builder)
	}

	tenants, err := c.gql.AzureCloudAccountTenants(ctx, graphql.CloudNativeProtection, true)
	if err != nil {
		return nil, err
	}

	subscriptions := make([]AzureSubscription, 0, 10)
	for _, tenant := range tenants {
		for _, sub := range tenant.Subscriptions {
			_, ok := builder.status[sub.FeatureDetail.Status]
			if len(builder.status) > 0 && !ok {
				continue
			}
			if len(builder.prefix) > 0 && !hasPrefix([]string{sub.ID, sub.NativeID, sub.Name}, builder.prefix) {
				continue
			}

			subscription, err := toAzureSubscription(tenant.DomainName, sub)
			if err != nil {
				return nil, err
			}
			subscriptions = append(subscriptions, subscription)
		}
	}

	return subscriptions, nil
}

// AzureSubscriptionIn -
type AzureSubscriptionIn struct {
	Cloud        graphql.AzureCloud
	ID           uuid.UUID
	Name         string
	TenantDomain string
	Regions      []graphql.AzureRegion
}

// AzureSubscriptionOut -
type AzureSubscriptionOut struct {
	TenantID uuid.UUID
	ID       uuid.UUID
	NativeID uuid.UUID
}

// AzureSubscriptionAdd -
func (c *Client) AzureSubscriptionAdd(ctx context.Context, subscription AzureSubscriptionIn) error {
	subsIn := []graphql.AzureSubscriptionIn{{
		ID:   subscription.ID.String(),
		Name: subscription.Name,
	}}

	permConf, err := c.gql.AzureCloudAccountPermissionConfig(ctx)
	if err != nil {
		return err
	}

	_, status, err := c.gql.AzureCloudAccountAddWithoutOAuth(ctx, subscription.Cloud, subscription.TenantDomain,
		subscription.Regions, graphql.CloudNativeProtection, subsIn, permConf.PermissionVersion)
	if len(status) != 1 {
		return errors.New("polaris: expected a single response")
	}
	if err != nil {
		return fmt.Errorf("polaris: %s", status[0].Error)
	}

	return err
}

// AzureSubscriptionSetRegions -
func (c *Client) AzureSubscriptionSetRegions(ctx context.Context, id AzureSubscriptionID, regions ...graphql.AzureRegion) error {
	subscription, err := c.AzureSubscription(ctx, id)
	if err != nil {
		return err
	}

	regionMap := make(map[graphql.AzureRegion]struct{})
	for _, region := range regions {
		regionMap[region] = struct{}{}
	}

	regionRemove := make([]graphql.AzureRegion, 0, 10)
	for _, region := range subscription.Feature.Regions {
		if _, ok := regionMap[region]; ok {
			delete(regionMap, region)
		} else {
			regionRemove = append(regionRemove, region)
		}
	}

	regionAdd := make([]graphql.AzureRegion, 0, 10)
	for region, _ := range regionMap {
		regionAdd = append(regionAdd, region)
	}

	sub := []graphql.AzureSubscriptionIn2{{
		ID:   subscription.ID.String(),
		Name: subscription.Name,
	}}

	status, err := c.gql.AzureCloudAccountUpdate(ctx, graphql.CloudNativeProtection, regionAdd, regionRemove, sub)
	if len(status) != 1 {
		return errors.New("polaris: expected a single response")
	}
	if err != nil {
		return err
	}

	return nil
}

// AzureSubscriptionSetName -
func (c *Client) AzureSubscriptionSetName(ctx context.Context, id AzureSubscriptionID, name string) error {
	subscription, err := c.AzureSubscription(ctx, id)
	if err != nil {
		return err
	}

	sub := []graphql.AzureSubscriptionIn2{{
		ID:   subscription.ID.String(),
		Name: name,
	}}

	status, err := c.gql.AzureCloudAccountUpdate(ctx, graphql.CloudNativeProtection, nil, nil, sub)
	if len(status) != 1 {
		return errors.New("polaris: expected a single response")
	}
	if err != nil {
		return err
	}

	return nil
}

// AzureSubscriptionRemove removes the Azure subscription identified by the
// specified subscription id.
func (c *Client) AzureSubscriptionRemove(ctx context.Context, id AzureSubscriptionID, deleteSnapshots bool) error {
	subscription, err := c.AzureSubscription(ctx, id)
	if err != nil {
		return err
	}

	// Lookup the Polaris Native ID from the Polaris subscription name and
	// the Azure subscription ID. The Polaris Native ID is needed to delete
	// the Polaris Native Account subscription.
	nativeSubs, err := c.gql.AzureNativeSubscriptionConnection(ctx, subscription.Name)
	if err != nil {
		return err
	}
	var nativeID string
	for _, nativeSub := range nativeSubs {
		if nativeSub.NativeID == subscription.NativeID.String() {
			nativeID = nativeSub.ID
			break
		}
	}
	if nativeID == "" {
		return errors.New("polaris: polaris native id not found")
	}

	jobID, err := c.gql.AzureDeleteNativeSubscription(ctx, nativeID, deleteSnapshots)
	if err != nil {
		return err
	}

	state, err := c.gql.WaitForTaskChain(ctx, jobID, 10*time.Second)
	if err != nil {
		return err
	}
	if state != graphql.TaskChainSucceeded {
		return fmt.Errorf("polaris: taskchain failed: jobID=%v, state=%v", jobID, state)
	}

	subscriptions := []string{
		subscription.ID.String(),
	}
	status, err := c.gql.AzureCloudAccountDeleteWithoutOAuth(ctx, subscriptions, graphql.CloudNativeProtection)
	if len(status) != 1 {
		return errors.New("polaris: expected a single response")
	}
	if err != nil {
		return fmt.Errorf("polaris: %s", status[0].Error)
	}

	return nil
}

// AzureServicePrincipal Azure service principal used by Polaris to access one
// or more Azure subscriptions.
type AzureServicePrincipal struct {
	Cloud        graphql.AzureCloud
	AppID        uuid.UUID
	AppName      string
	AppSecret    string
	TenantID     uuid.UUID
	TenantDomain string
}

// AzureServicePrincipalSet sets the service princiapl to use by subscriptions
// in the same tenant domain.
func (c *Client) AzureServicePrincipalSet(ctx context.Context, principal AzureServicePrincipal) error {
	return c.gql.AzureSetCustomerAppCredentials(ctx, principal.Cloud, principal.AppID.String(),
		principal.AppName, principal.TenantID.String(), principal.TenantDomain, principal.AppSecret)
}
