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

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// FeatureFlagName represents a feature flag affecting the workflows supported
// by the Go SDK.
type FeatureFlagName string

const (
	FeatureFlagAzureSQLDBCopyBackup     FeatureFlagName = "CNP_AZURE_SQL_DB_COPY_BACKUP"
	FeatureFlagGCPDisableDeleteCombined FeatureFlagName = "CNP_GCP_DISABLE_DELETE_COMBINED"
)

// FeatureFlag holds the name and state of a single RSC feature flag.
type FeatureFlag struct {
	Name    string
	Enabled bool
}

// FeatureFlags returns all the RSC feature flags.
func (a API) FeatureFlags(ctx context.Context) ([]FeatureFlag, error) {
	a.log.Print(log.Trace)

	query := featureFlagAllQuery
	buf, err := a.GQL.Request(ctx, query, struct{}{})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Flags []struct {
					Name    string `json:"name"`
					Variant string `json:"variant"`
				} `json:"flags"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	flags := make([]FeatureFlag, 0, len(payload.Data.Result.Flags))
	for _, flag := range payload.Data.Result.Flags {
		flags = append(flags, FeatureFlag{Name: flag.Name, Enabled: flag.Variant == "true"})
	}

	return flags, nil
}

// FeatureFlag returns a specific RSC feature flag.
func (a API) FeatureFlag(ctx context.Context, name FeatureFlagName) (FeatureFlag, error) {
	a.log.Print(log.Trace)

	flag, err := a.featureFlag(ctx, name)
	if err != nil {
		return FeatureFlag{}, err
	}
	if flag.Variant == "true" || flag.Variant == "false" {
		return FeatureFlag{Name: flag.Name, Enabled: flag.Variant == "true"}, nil
	}

	unifiedFlag, err := a.singleUnifiedFeatureFlag(ctx, name)
	if err != nil {
		return FeatureFlag{}, err
	}

	return FeatureFlag{Name: unifiedFlag.Name, Enabled: unifiedFlag.Variant == "true"}, nil
}

type internalFlag struct {
	Name    string `json:"name"`
	Variant string `json:"variant"`
}

// featureFlag returns the value of a non-unified feature flag.
func (a API) featureFlag(ctx context.Context, name FeatureFlagName) (internalFlag, error) {
	a.log.Print(log.Trace)

	query := featureFlagQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		FlagName FeatureFlagName `json:"flagName"`
	}{FlagName: name})
	if err != nil {
		return internalFlag{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Flag internalFlag `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return internalFlag{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Flag, nil
}

// singleUnifiedFeatureFlag returns the value of a unified feature flag.
func (a API) singleUnifiedFeatureFlag(ctx context.Context, name FeatureFlagName) (internalFlag, error) {
	a.log.Print(log.Trace)

	query := singleUnifiedFeatureFlagQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		FlagName FeatureFlagName `json:"flagName"`
	}{FlagName: name})
	if err != nil {
		return internalFlag{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Flag internalFlag `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return internalFlag{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Flag, nil
}
