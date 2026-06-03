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
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	azure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NativeSubscriptionByID returns the native subscription with the specified RSC
// native cloud account ID.
func (a API) NativeSubscriptionByID(ctx context.Context, nativeCloudAccountID uuid.UUID) (azure.NativeSubscription, error) {
	a.log.Print(log.Trace)

	natives, err := a.NativeSubscriptions(ctx, "")
	if err != nil {
		return azure.NativeSubscription{}, fmt.Errorf("failed to list native subscriptions: %s", err)
	}
	for _, native := range natives {
		if native.ID == nativeCloudAccountID {
			return native, nil
		}
	}

	return azure.NativeSubscription{}, fmt.Errorf("native subscription %q %w", nativeCloudAccountID, graphql.ErrNotFound)
}

// NativeSubscriptionByCloudAccountID returns the native subscription with the
// specified RSC cloud account ID.
func (a API) NativeSubscriptionByCloudAccountID(ctx context.Context, cloudAccountID uuid.UUID) (azure.NativeSubscription, error) {
	a.log.Print(log.Trace)

	natives, err := a.NativeSubscriptions(ctx, "")
	if err != nil {
		return azure.NativeSubscription{}, fmt.Errorf("failed to list native subscriptions: %s", err)
	}
	for _, native := range natives {
		if native.CloudAccountID == cloudAccountID {
			return native, nil
		}
	}

	return azure.NativeSubscription{}, fmt.Errorf("native subscription for subscription %q %w", cloudAccountID, graphql.ErrNotFound)
}

// NativeSubscriptions returns all native subscriptions matching the specified
// filter. The filter can be used to search for a substring in the subscription
// name.
func (a API) NativeSubscriptions(ctx context.Context, filter string) ([]azure.NativeSubscription, error) {
	return azure.Wrap(a.client).NativeSubscriptions(ctx, filter)
}

// NativeResourceGroups returns Azure resource groups visible to RSC,
// optionally filtered to the given subscription IDs and to those whose name
// contains nameSubstring. A nil/empty subscription slice returns resource
// groups across all managed subscriptions; the RSC API requires a non-empty
// subscription filter, so in that case the list of managed subscriptions is
// fetched and used as the filter. Passing an empty nameSubstring disables the
// name filter.
func (a API) NativeResourceGroups(ctx context.Context, subscriptionIDs []uuid.UUID, nameSubstring string) ([]azure.NativeResourceGroup, error) {
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
