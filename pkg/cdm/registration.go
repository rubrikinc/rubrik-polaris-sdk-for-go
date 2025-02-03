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

package cdm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NodeDetails holds the details of a node's configuration.
type NodeDetails struct {
	BoardSerial       string    `json:"boardSerial"`
	SystemUUID        uuid.UUID `json:"systemUUID"`
	NodeID            string    `json:"nodeId"`
	ManufacturingTime string    `json:"manufacturingTime"`
	TeleportToken     string    `json:"teleportToken"`
	ClusterUUID       uuid.UUID `json:"clusterUuid"`
	PlatformType      string    `json:"platformType"`
	CapacityInBytes   string    `json:"capacityInBytes"`
	IsEntitled        bool      `json:"isEntitled"`
	Version           string    `json:"version"`
}

// ToNodeRegistrationConfig converts the node details to a node registration.
func (nd NodeDetails) ToNodeRegistrationConfig() core.NodeRegistrationConfig {
	return core.NodeRegistrationConfig{
		ID:              nd.NodeID,
		Capacity:        nd.CapacityInBytes,
		ClusterUUID:     nd.ClusterUUID,
		IsEntitled:      nd.IsEntitled,
		ManufactureTime: nd.ManufacturingTime,
		Platform:        nd.PlatformType,
		Serial:          nd.BoardSerial,
		SystemUUID:      nd.SystemUUID,
		TeleportToken:   nd.TeleportToken,
	}
}

// OfflineEntitle returns the node details for offline entitlement.
func (c *Client) OfflineEntitle(ctx context.Context) ([]NodeDetails, error) {
	c.Log.Print(log.Trace)

	endpoint := "/cluster/me/offline_entitle"
	res, code, err := c.Get(ctx, V1, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed GET request %q: %s", endpoint, err)
	}
	if code != 200 {
		return nil, fmt.Errorf("failed GET request %q: %s", endpoint, errorMessage(res, code))
	}

	var payload struct {
		Data []NodeDetails `json:"data"`
	}
	if err := json.Unmarshal(res, &payload); err != nil {
		return nil, err
	}

	return payload.Data, nil
}

// SetRegisteredMode sets the registered mode for the cluster using the auth
// token. Returns the mode set as a string.
func (c *Client) SetRegisteredMode(ctx context.Context, authToken string) (string, error) {
	c.Log.Print(log.Trace)

	endpoint := "/cluster/me/registered_mode"
	res, code, err := c.Put(ctx, Internal, endpoint, struct {
		AuthToken string `json:"authToken"`
	}{AuthToken: authToken})
	if err != nil {
		return "", fmt.Errorf("failed PUT request %q: %s", endpoint, err)
	}
	if code != 200 {
		return "", fmt.Errorf("failed PUT request %q: %s", endpoint, errorMessage(res, code))
	}

	var payload struct {
		RegisteredMode struct {
			Result string `json:"result"`
		}
	}
	if err := json.Unmarshal(res, &payload); err != nil {
		return "", err
	}

	return payload.RegisteredMode.Result, nil
}
