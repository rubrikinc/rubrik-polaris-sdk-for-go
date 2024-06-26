// Copyright 2023 Rubrik, Inc.
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

package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// SetPrivateContainerRegistryDetails sets the private container registry
// details.
func (a API) SetPrivateContainerRegistryDetails(ctx context.Context, id uuid.UUID, url string, nativeID string) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, setPrivateContainerRegistryDetailsQuery, struct {
		ID       uuid.UUID `json:"exocomputeAccountId"`
		URL      string    `json:"registryUrl"`
		NativeID string    `json:"awsNativeId,omitempty"`
	}{ID: id, URL: url, NativeID: nativeID})
	if err != nil {
		return fmt.Errorf("failed to request setPrivateContainerRegistryDetails: %w", err)
	}
	a.log.Printf(log.Debug, "setPrivateContainerRegistryDetails(%q, %q, %q): %s", id, url, nativeID, string(buf))

	return nil
}

// PrivateContainerRegistry returns the private container registry details.
func (a API) PrivateContainerRegistry(ctx context.Context, id uuid.UUID) (nativeID, URL string, err error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, privateContainerRegistryQuery, struct {
		ID uuid.UUID `json:"exocomputeCloudAccountId"`
	}{ID: id})
	if err != nil {
		return "", "", fmt.Errorf("failed to request privateContainerRegistry: %w", err)
	}
	a.log.Printf(log.Debug, "privateContainerRegistry(%q): %s", id, string(buf))
	var payload struct {
		Data struct {
			Result struct {
				PCRDetails struct {
					ImagePullDetails struct {
						NativeID string `json:"awsNativeId,omitempty"`
					} `json:"imagePullDetails"`
					URL string `json:"registryUrl"`
				} `json:"pcrDetails"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal privateContainerRegistry: %s", err)
	}

	return payload.Data.Result.PCRDetails.ImagePullDetails.NativeID, payload.Data.Result.PCRDetails.URL, nil
}
