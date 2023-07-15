//go:generate go run ../queries_gen.go cdmk8s

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

// Package cdmk8s provides a low level interface to the CDM K8s GraphQL queries
// provided by the Polaris platform.
package cdmk8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API wraps around GraphQL clients to give them the RSC CDM K8s API.
type API struct {
	GQL *graphql.Client
	log log.Logger
}

// Wrap the RSC client in the CDM K8s API.
func Wrap(client *polaris.Client) API {
	return API{GQL: client.GQL, log: client.GQL.Log()}
}

// DeleteK8sResourceSet deletes the K8s resource set corresponding to the provided fid.
func (a API) DeleteK8sResourceSet(ctx context.Context, fid string, preserveSnapshots bool) (bool, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(
		ctx,
		deleteK8sResourcesetQuery,
		struct {
			Id                string `json:"id"`
			PreserveSnapshots bool   `json:"preserveSnapshots"`
		}{
			Id:                fid,
			PreserveSnapshots: preserveSnapshots,
		},
	)

	if err != nil {
		return false, fmt.Errorf("failed to request deleteK8sResourceset: %w", err)
	}
	a.log.Printf(log.Debug, "deleteK8sResourceset(%q, %q): %s", fid, preserveSnapshots, string(buf))

	var payload struct {
		Data struct {
			ResponseSuccess struct {
				Success bool `json:"success"`
			} `json:"deleteK8sResourceset"`
		} `json:"data"`
	}

	if err := json.Unmarshal(buf, &payload); err != nil {
		return false, fmt.Errorf("failed to unmarshal deleteK8sResourceset response: %v", err)
	}
	if !payload.Data.ResponseSuccess.Success {
		return false, fmt.Errorf("failed to delete k8s resource set with fid %q", fid)
	}

	return true, nil
}
