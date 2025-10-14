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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NativeAccount represents an RSC native account.
type NativeAccount struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Regions []struct {
		Region        aws.RegionEnum `json:"region"`
		HasExocompute bool           `json:"isExocomputeConfigured"`
	} `json:"regionSpecs"`
	Status     string         `json:"status"`
	Assignment sla.Assignment `json:"slaAssignment"`
	Configured sla.DomainRef  `json:"configuredSlaDomain"`
	Effective  sla.DomainRef  `json:"effectiveSlaDomain"`
}

// NativeAccount returns the native account with the specified RSC native
// account id.
func (a API) NativeAccount(ctx context.Context, id uuid.UUID, feature ProtectionFeature) (NativeAccount, error) {
	a.log.Print(log.Trace)

	query := awsNativeAccountQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID      uuid.UUID         `json:"awsNativeAccountRubrikId"`
		Feature ProtectionFeature `json:"awsNativeProtectionFeature"`
	}{ID: id, Feature: feature})
	if err != nil {
		return NativeAccount{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Account NativeAccount `json:"awsNativeAccount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return NativeAccount{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Account, nil
}

// NativeAccounts returns the native accounts matching the specified filter.
// The filter can be used to search for a substring in account name.
func (a API) NativeAccounts(ctx context.Context, feature ProtectionFeature, filter string) ([]NativeAccount, error) {
	a.log.Print(log.Trace)

	var accounts []NativeAccount
	var cursor string
	for {
		query := awsNativeAccountsQuery
		buf, err := a.GQL.Request(ctx, query, struct {
			After   string            `json:"after,omitempty"`
			Feature ProtectionFeature `json:"awsNativeProtectionFeature"`
			Filter  string            `json:"filter"`
		}{After: cursor, Feature: feature, Filter: filter})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

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
			return nil, graphql.UnmarshalError(query, err)
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

// StartNativeAccountDisableJob starts a task chain job to disable the native
// account with the specified RSC cloud account id. Returns the RSC task chain
// id.
func (a API) StartNativeAccountDisableJob(ctx context.Context, id uuid.UUID, feature ProtectionFeature, deleteSnapshots bool) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	query := startAwsNativeAccountDisableJobQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID              uuid.UUID         `json:"awsAccountRubrikId"`
		Feature         ProtectionFeature `json:"awsNativeProtectionFeature"`
		DeleteSnapshots bool              `json:"shouldDeleteNativeSnapshots"`
	}{ID: id, Feature: feature, DeleteSnapshots: deleteSnapshots})
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Query struct {
				Error string    `json:"error"`
				JobID uuid.UUID `json:"jobId"`
			} `json:"startAwsNativeAccountDisableJob"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	if payload.Data.Query.Error != "" {
		return uuid.Nil, errors.New(payload.Data.Query.Error)
	}

	return payload.Data.Query.JobID, nil
}

// StartExocomputeDisableJob starts a task chain job to disable the Exocompute
// feature for the account with the specified RSC native account id. Returns the
// RSC task chain id.
func (a API) StartExocomputeDisableJob(ctx context.Context, nativeID uuid.UUID) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	query := startAwsExocomputeDisableJobQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID uuid.UUID `json:"cloudAccountId"`
	}{ID: nativeID})
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Error string    `json:"error"`
				JobID uuid.UUID `json:"jobId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	if payload.Data.Result.Error != "" {
		return uuid.Nil, errors.New(payload.Data.Result.Error)
	}

	return payload.Data.Result.JobID, nil
}
