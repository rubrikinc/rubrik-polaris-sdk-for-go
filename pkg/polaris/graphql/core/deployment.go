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

package core

import (
	"cmp"
	"context"
	"encoding/json"
	"slices"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Deprecated: use GQL.DeploymentVersion.
func (a API) DeploymentVersion(ctx context.Context) (string, error) {
	a.log.Print(log.Trace)

	query := deploymentVersionQuery
	buf, err := a.GQL.Request(ctx, query, struct{}{})
	if err != nil {
		return "", graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			DeploymentVersion string `json:"deploymentVersion"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", graphql.UnmarshalError(query, err)
	}

	return payload.Data.DeploymentVersion, nil
}

// DeploymentIPAddresses returns the deployment IP addresses.
func (a API) DeploymentIPAddresses(ctx context.Context) ([]string, error) {
	a.log.Print(log.Trace)

	query := allDeploymentIpAddressesQuery
	buf, err := a.GQL.Request(ctx, query, struct{}{})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			DeploymentIPAddresses []string `json:"allDeploymentIpAddresses"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.DeploymentIPAddresses, nil
}

// EnabledFeaturesForAccount returns all features enable for the RSC account.
func (a API) EnabledFeaturesForAccount(ctx context.Context) ([]Feature, error) {
	a.log.Print(log.Trace)

	query := allEnabledFeaturesForAccountQuery
	buf, err := a.GQL.Request(ctx, query, struct{}{})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Features []string `json:"features"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	var features []Feature
	for _, feature := range payload.Data.Result.Features {
		features = append(features, Feature{Name: feature})
	}
	slices.SortFunc(features, func(i, j Feature) int {
		return cmp.Compare(i.Name, j.Name)
	})

	return features, nil
}
