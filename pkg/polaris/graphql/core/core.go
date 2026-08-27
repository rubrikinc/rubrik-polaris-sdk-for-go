//go:generate go run ../queries_gen.go core

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

// Package core provides a low-level interface to core GraphQL queries provided
// by the RSC platform. E.g., task chains and enum definitions.
package core

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API wraps around GraphQL clients to give them the RSC Core API.
type API struct {
	Version string // Deprecated: use GQL.DeploymentVersion
	GQL     *graphql.Client
	log     log.Logger
}

// Wrap the GraphQL client in the Core API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

type CloudVendor string

const (
	CloudVendorUnknown CloudVendor = "VENDOR_UNKNOWN"
	CloudVendorAWS     CloudVendor = "AWS"
	CloudVendorAzure   CloudVendor = "AZURE"
	CloudVendorGCP     CloudVendor = "GCP"
	CloudVendorOCI     CloudVendor = "OCI"
)

// CloudAccountAction represents a Polaris cloud account action.
type CloudAccountAction string

const (
	Create              CloudAccountAction = "CREATE"
	Delete              CloudAccountAction = "DELETE"
	UpdateChildAccounts CloudAccountAction = "UPDATE_CHILD_ACCOUNTS"
	UpdatePermissions   CloudAccountAction = "UPDATE_PERMISSIONS"
	UpdateRegions       CloudAccountAction = "UPDATE_REGIONS"
)

// Status represents an RSC cloud account status.
type Status string

const (
	StatusConnected          Status = "CONNECTED"
	StatusConnecting         Status = "CONNECTING"
	StatusDisabled           Status = "DISABLED"
	StatusDisconnected       Status = "DISCONNECTED"
	StatusMissingPermissions Status = "MISSING_PERMISSIONS"
)

// FormatStatus returns the Status as a string using lower-case and with hyphen
// as a separator.
func FormatStatus(status Status) string {
	return strings.ReplaceAll(strings.ToLower(string(status)), "_", "-")
}

// FormatTimestamp converts a time.Time to RFC3339 format with milliseconds and
// Z suffix.
func FormatTimestamp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

// ValuesByEnum returns the enum values for the specified enum name in the RSC GraphQL API.
func (a API) ValuesByEnum(ctx context.Context, enumName string) ([]string, error) {
	a.log.Print(log.Trace)

	query := enumValuesQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		EnumName string `json:"enumName"`
	}{EnumName: enumName})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				EnumValues []struct {
					Name string `json:"name"`
				} `json:"enumValues"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	var enumValues []string
	for _, v := range payload.Data.Result.EnumValues {
		enumValues = append(enumValues, v.Name)
	}
	return enumValues, nil
}
