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

package azure

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NativeSubscription represents a Polaris native subscription.
type NativeSubscription struct {
	ID            uuid.UUID          `json:"id"`
	Name          string             `json:"name"`
	NativeID      uuid.UUID          `json:"nativeId"`
	Status        string             `json:"status"`
	SLAAssignment core.SLAAssignment `json:"slaAssignment"`
	Configured    core.SLADomain     `json:"configuredSlaDomain"`
	Effective     core.SLADomain     `json:"effectiveSlaDomain"`
}

// NativeSubscription returns the native subscription with the specified
// Polaris native subscription id.
func (a API) NativeSubscription(ctx context.Context, id uuid.UUID) (NativeSubscription, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.NativeSubscription")

	buf, err := a.GQL.Request(ctx, azureNativeSubscriptionQuery, struct {
		ID uuid.UUID `json:"fid"`
	}{ID: id})
	if err != nil {
		return NativeSubscription{}, err
	}

	a.GQL.Log().Printf(log.Debug, "azureNativeSubscription(%q): %s", id, string(buf))

	var payload struct {
		Data struct {
			Subscription NativeSubscription `json:"azureNativeSubscriptionConnection"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return NativeSubscription{}, err
	}

	return payload.Data.Subscription, nil
}

// NativeSubscriptions returns the native subscriptions matching the specified
// filter. The filter can be used to search for a substring in the subscription
// name.
func (a API) NativeSubscriptions(ctx context.Context, filter string) ([]NativeSubscription, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.NativeSubscriptions")

	var subscriptions []NativeSubscription
	var cursor string
	for {
		buf, err := a.GQL.Request(ctx, azureNativeSubscriptionConnectionQuery, struct {
			After  string `json:"after,omitempty"`
			Filter string `json:"filter"`
		}{After: cursor, Filter: filter})
		if err != nil {
			return nil, err
		}

		a.GQL.Log().Printf(log.Debug, "azureNativeSubscriptionConnection(%q): %s", filter, string(buf))

		var payload struct {
			Data struct {
				Query struct {
					Count int `json:"count"`
					Edges []struct {
						Node NativeSubscription `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"azureNativeSubscriptionConnection"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, err
		}
		for _, subscription := range payload.Data.Query.Edges {
			subscriptions = append(subscriptions, subscription.Node)
		}

		if !payload.Data.Query.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Query.PageInfo.EndCursor
	}

	return subscriptions, nil
}

// DeleteNativeSubscription starts a task chain job to disables the native
// subscription with the specified Polaris native subscription id. If
// deleteSnapshots is true the snapshots are deleted. Returns the Polaris task
// chain id.
func (a API) DeleteNativeSubscription(ctx context.Context, id uuid.UUID, deleteSnapshots bool) (uuid.UUID, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.DeleteNativeSubscription")

	buf, err := a.GQL.Request(ctx, deleteAzureNativeSubscriptionQuery, struct {
		ID              uuid.UUID `json:"subscriptionId"`
		DeleteSnapshots bool      `json:"shouldDeleteNativeSnapshots"`
	}{ID: id, DeleteSnapshots: deleteSnapshots})
	if err != nil {
		return uuid.Nil, err
	}

	a.GQL.Log().Printf(log.Debug, "deleteAzureNativeSubscription(%q, %t): %s", id, deleteSnapshots, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				TaskChainID uuid.UUID `json:"taskchainUuid"`
			} `json:"deleteAzureNativeSubscription"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, err
	}

	return payload.Data.Query.TaskChainID, nil
}
