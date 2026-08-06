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

package azure

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	azure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NativeSubscriptionByID returns the native subscription with the specified RSC
// object ID.
func (a API) NativeSubscriptionByID(ctx context.Context, objectID uuid.UUID) (azure.NativeSubscription, error) {
	a.log.Print(log.Trace)

	natives, err := a.NativeSubscriptionsByFilter(ctx, azure.NativeSubscriptionFilters{})
	if err != nil {
		return azure.NativeSubscription{}, err
	}
	for _, native := range natives {
		if native.ID == objectID {
			return native, nil
		}
	}

	return azure.NativeSubscription{}, fmt.Errorf("native subscription %q %w", objectID, graphql.ErrNotFound)
}

// NativeSubscriptionByCloudAccountID returns the native subscription with the
// specified RSC cloud account ID.
func (a API) NativeSubscriptionByCloudAccountID(ctx context.Context, cloudAccountID uuid.UUID) (azure.NativeSubscription, error) {
	a.log.Print(log.Trace)

	natives, err := a.NativeSubscriptionsByFilter(ctx, azure.NativeSubscriptionFilters{})
	if err != nil {
		return azure.NativeSubscription{}, err
	}
	for _, native := range natives {
		if native.CloudAccountID == cloudAccountID {
			return native, nil
		}
	}

	return azure.NativeSubscription{}, fmt.Errorf("native subscription with cloud account id %q %w", cloudAccountID, graphql.ErrNotFound)
}

// Deprecated: use NativeSubscriptionsByFilter instead.
func (a API) NativeSubscriptions(ctx context.Context, filter string) ([]azure.NativeSubscription, error) {
	a.log.Print(log.Trace)

	return azure.Wrap(a.client).NativeSubscriptions(ctx, filter)
}

// NativeSubscriptionsByFilter returns all native subscriptions matching the
// specified filters. See azure.NativeSubscriptionFilters for the available
// filter criteria; a zero value returns all native subscriptions. Results are
// sorted by name in ascending order.
func (a API) NativeSubscriptionsByFilter(ctx context.Context, filters azure.NativeSubscriptionFilters) ([]azure.NativeSubscription, error) {
	a.log.Print(log.Trace)

	nativeSubs, err := azure.Wrap(a.client).NativeSubscriptionsByFilter(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list native subscriptions: %w", err)
	}

	// If an RSC subscription refresh runs at the same time as the subscriptions
	// are read, the result might contain duplicates, remove those.
	slices.SortFunc(nativeSubs, func(lhs, rhs azure.NativeSubscription) int {
		return cmp.Compare(lhs.ID.String(), rhs.ID.String())
	})
	nativeSubs = slices.CompactFunc(nativeSubs, func(lhs, rhs azure.NativeSubscription) bool {
		return lhs.ID == rhs.ID
	})

	slices.SortFunc(nativeSubs, func(lhs, rhs azure.NativeSubscription) int {
		return cmp.Compare(lhs.Name, rhs.Name)
	})

	return nativeSubs, nil
}

// WaitForNativeSubscription blocks until a native subscription for the
// specified RSC cloud account ID becomes available, polling every 30 seconds.
// It returns nil once the native subscription is found, or an error if the
// context is cancelled, the timeout is exceeded, or the lookup fails for any
// reason other than the subscription not yet existing.
func (a API) WaitForNativeSubscription(ctx context.Context, cloudAccountID uuid.UUID) error {
	a.log.Print(log.Trace)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	a.log.Printf(log.Debug, "Waiting for native subscription with cloud account ID %q to become available", cloudAccountID)
	for {
		_, err := a.NativeSubscriptionByCloudAccountID(ctx, cloudAccountID)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.Canceled):
			return fmt.Errorf("%w while waiting for native subscription with cloud account id %q to become available",
				context.Canceled, cloudAccountID)
		case errors.Is(err, context.DeadlineExceeded):
			return fmt.Errorf("%w while waiting for native subscription with cloud account id %q to become available",
				context.DeadlineExceeded, cloudAccountID)
		case !errors.Is(err, graphql.ErrNotFound):
			return err
		}

		a.log.Printf(log.Debug, "Native subscription with cloud account ID %q still not available", cloudAccountID)

		select {
		case <-ctx.Done():
			return fmt.Errorf("%w while waiting for native subscription with cloud account id %q to become available",
				ctx.Err(), cloudAccountID)
		case <-time.After(30 * time.Second):
		}
	}
}

// Deprecated: use NativeResourceGroupsByFilter instead.
func (a API) NativeResourceGroups(ctx context.Context, subscriptionIDs []uuid.UUID, nameSubstring string) ([]azure.NativeResourceGroup, error) {
	a.log.Print(log.Trace)

	if len(subscriptionIDs) == 0 {
		subs, err := a.NativeSubscriptions(ctx, "")
		if err != nil {
			return nil, fmt.Errorf("failed to list native subscriptions: %s", err)
		}
		if len(subs) == 0 {
			return nil, nil
		}
		subscriptionIDs = make([]uuid.UUID, 0, len(subs))
		for _, s := range subs {
			subscriptionIDs = append(subscriptionIDs, s.ID)
		}
	}
	return azure.Wrap(a.client).NativeResourceGroups(ctx, subscriptionIDs, nameSubstring)
}

// NativeResourceGroupsByFilter returns all Azure resource groups matching the
// specified filters. See azure.ResourceGroupFilters for the available filter
// criteria; a zero value returns all resource groups. Results are sorted by
// subscription name, then by resource group name.
func (a API) NativeResourceGroupsByFilter(ctx context.Context, filters azure.ResourceGroupFilters) ([]azure.NativeResourceGroup, error) {
	a.log.Print(log.Trace)

	groups, err := azure.Wrap(a.client).NativeResourceGroupsByFilter(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list native resource groups: %w", err)
	}

	// If an RSC subscription refresh run at the same time as the resource
	// groups are read, the result might contain duplicates, remove those.
	slices.SortFunc(groups, func(lhs, rhs azure.NativeResourceGroup) int {
		return cmp.Compare(lhs.ID, rhs.ID)
	})
	groups = slices.CompactFunc(groups, func(lhs, rhs azure.NativeResourceGroup) bool {
		return lhs.ID == rhs.ID
	})

	slices.SortFunc(groups, func(lhs, rhs azure.NativeResourceGroup) int {
		return cmp.Or(
			cmp.Compare(lhs.Subscription.Name, rhs.Subscription.Name),
			cmp.Compare(lhs.Name, rhs.Name),
		)
	})

	return groups, nil
}
