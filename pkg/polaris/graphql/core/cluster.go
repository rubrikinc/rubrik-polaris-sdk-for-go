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

package core

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NodeRegistrationConfig holds the node registration configuration.
type NodeRegistrationConfig struct {
	ID              string    `json:"id"`
	Capacity        string    `json:"capacity"`
	ClusterUUID     uuid.UUID `json:"clusterUuid"`
	IsEntitled      bool      `json:"isEntitled"`
	ManufactureTime string    `json:"manufactureTime"`
	Platform        string    `json:"platform"`
	Serial          string    `json:"serial"`
	SystemUUID      uuid.UUID `json:"systemUuid"`
	TeleportToken   string    `json:"teleportToken"`
}

// RegisterCluster registers the cluster with RSC. Note, even though
// managedByPolaris is set to true, depending on the RSC customer account
// entitlements, the cluster may be registered as Life of Device and not Hybrid.
func (a API) RegisterCluster(ctx context.Context, managedByPolaris bool, nodeConfigs []NodeRegistrationConfig, isOfflineRegistration bool) (string, string, error) {
	a.log.Print(log.Trace)

	query := registerClusterQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ManagedByPolaris      bool                     `json:"managedByPolaris"`
		NodeConfigs           []NodeRegistrationConfig `json:"nodeConfigs"`
		IsOfflineRegistration bool                     `json:"isOfflineRegistration"`
	}{ManagedByPolaris: managedByPolaris, NodeConfigs: nodeConfigs, IsOfflineRegistration: isOfflineRegistration})
	if err != nil {
		return "", "", graphql.RequestError(query, err)
	}
	graphql.LogResponse(a.log, query, buf)

	var payload struct {
		Data struct {
			Result struct {
				Token       string `json:"token"`
				PubKey      string `json:"pubKey"`
				ProductType string `json:"productType"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", "", graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Token, payload.Data.Result.ProductType, nil
}
