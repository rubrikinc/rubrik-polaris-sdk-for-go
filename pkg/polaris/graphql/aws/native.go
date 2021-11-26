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

package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NativeAccount represents a Polaris native account.
type NativeAccount struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Regions []struct {
		Region        Region `json:"region"`
		HasExocompute bool   `json:"isExocomputeConfigured"`
	} `json:"regionSpecs"`
	Status     string             `json:"status"`
	Assignment core.SLAAssignment `json:"slaAssignment"`
	Configured core.SLADomain     `json:"configuredSlaDomain"`
	Effective  core.SLADomain     `json:"effectiveSlaDomain"`
}

// NativeAccount returns the native account with the specified Polaris native
// account id.
func (a API) NativeAccount(ctx context.Context, id uuid.UUID, feature ProtectionFeature) (NativeAccount, error) {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, awsNativeAccountQuery, struct {
		ID      uuid.UUID         `json:"awsNativeAccountRubrikId"`
		Feature ProtectionFeature `json:"awsNativeProtectionFeature"`
	}{ID: id, Feature: feature})
	if err != nil {
		return NativeAccount{}, fmt.Errorf("failed to request NativeAccount: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "awsNativeAccount(%q, %q): %s", id, feature, string(buf))

	var payload struct {
		Data struct {
			Account NativeAccount `json:"awsNativeAccount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return NativeAccount{}, fmt.Errorf("failed to unmarshal NativeAccount: %v", err)
	}

	return payload.Data.Account, nil
}

// NativeAccounts returns the native accounts matching the specified filter.
// The filter can be used to search for a substring in account name.
func (a API) NativeAccounts(ctx context.Context, feature ProtectionFeature, filter string) ([]NativeAccount, error) {
	a.GQL.Log().Print(log.Trace)

	// nameSubstringFilter: NameSubstringFilter
	var accounts []NativeAccount
	var cursor string
	for {
		buf, err := a.GQL.Request(ctx, awsNativeAccountsQuery, struct {
			After   string            `json:"after,omitempty"`
			Feature ProtectionFeature `json:"awsNativeProtectionFeature"`
			Filter  string            `json:"filter"`
		}{After: cursor, Feature: feature, Filter: filter})
		if err != nil {
			return nil, fmt.Errorf("failed to request NativeAccounts: %v", err)
		}

		a.GQL.Log().Printf(log.Debug, "awsNativeAccounts(%q, %q, %q): %s", cursor, feature, filter, string(buf))

		var payload struct {
			Data struct {
				Query struct {
					Count int `json:"count"`
					Edges []struct {
						Node NativeAccount `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"awsNativeAccounts"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal NativeAccounts: %v", err)
		}
		for _, account := range payload.Data.Query.Edges {
			accounts = append(accounts, account.Node)
		}

		if !payload.Data.Query.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Query.PageInfo.EndCursor
	}

	return accounts, nil
}

// StartNativeAccountDisableJob starts a task chain job to disables the native
// account with the specified Polaris native account id. Returns the Polaris
// task chain id.
func (a API) StartNativeAccountDisableJob(ctx context.Context, id uuid.UUID, feature ProtectionFeature, deleteSnapshots bool) (uuid.UUID, error) {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, startAwsNativeAccountDisableJobQuery, struct {
		ID              uuid.UUID         `json:"awsAccountRubrikId"`
		Feature         ProtectionFeature `json:"awsNativeProtectionFeature"`
		DeleteSnapshots bool              `json:"shouldDeleteNativeSnapshots"`
	}{ID: id, Feature: feature, DeleteSnapshots: deleteSnapshots})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to request StartNativeAccountDisableJob: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "startAwsNativeAccountDisableJob(%q, %q, %t): %s", id, feature, deleteSnapshots, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Error string    `json:"error"`
				JobID uuid.UUID `json:"jobId"`
			} `json:"startAwsNativeAccountDisableJob"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal StartNativeAccountDisableJob: %v", err)
	}
	if payload.Data.Query.Error != "" {
		return uuid.Nil, errors.New(payload.Data.Query.Error)
	}

	return payload.Data.Query.JobID, nil
}
