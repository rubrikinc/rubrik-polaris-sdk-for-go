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

package cluster

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// SelfServeRollingUpgrade returns whether self-serve rolling upgrade is
// enabled for the account.
func SelfServeRollingUpgrade(ctx context.Context, gql *graphql.Client) (bool, error) {
	gql.Log().Print(log.Trace)

	query := selfServeRollingUpgradeQuery
	buf, err := gql.Request(ctx, query, struct{}{})
	if err != nil {
		return false, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Enabled bool `json:"enabled"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return false, graphql.UnmarshalError(query, err)
	}
	return payload.Data.Result.Enabled, nil
}

// SetSelfServeRollingUpgrade sets the account-level self-serve rolling upgrade
// setting to the specified value.
func SetSelfServeRollingUpgrade(ctx context.Context, gql *graphql.Client, enabled bool) error {
	gql.Log().Print(log.Trace)

	query := setSelfServeRollingUpgradeQuery
	buf, err := gql.Request(ctx, query, struct {
		Enabled bool `json:"enabled"`
	}{Enabled: enabled})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Enabled bool `json:"enabled"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if payload.Data.Result.Enabled != enabled {
		return graphql.ResponseError(query, fmt.Errorf("self-serve rolling upgrade not set: requested %t, got %t", enabled, payload.Data.Result.Enabled))
	}
	return nil
}
