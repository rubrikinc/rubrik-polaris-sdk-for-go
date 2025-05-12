//go:generate go run ../queries_gen.go azure

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

// Package azure provides a low-level interface to the Azure GraphQL queries
// provided by the Polaris platform.
package azure

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Cloud represents the Azure cloud type.
type Cloud string

const (
	ChinaCloud  Cloud = "AZURECHINACLOUD"
	PublicCloud Cloud = "AZUREPUBLICCLOUD"
)

// API wraps around GraphQL clients to give them the RSC Azure API.
type API struct {
	Version string // Deprecated: use GQL.DeploymentVersion
	GQL     *graphql.Client
	log     log.Logger
}

// Wrap the GraphQL client in the Azure API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

// SetCloudAccountCustomerAppCredentials sets the credentials for the customer
// application for the specified tenant domain. If shouldReplace is true and the
// app already has a service principal, it will be replaced. If the tenant
// domain is empty, set it for all the tenants of the customer.
func (a API) SetCloudAccountCustomerAppCredentials(ctx context.Context, cloud Cloud, appID, appTenantID uuid.UUID, appName, appTenantDomain, appSecretKey string, shouldReplace bool) error {
	a.log.Print(log.Trace)

	query := setAzureCloudAccountCustomerAppCredentialsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Cloud         Cloud         `json:"azureCloudType"`
		ID            uuid.UUID     `json:"appId"`
		Name          string        `json:"appName"`
		SecretKey     secret.String `json:"appSecretKey"`
		TenantID      uuid.UUID     `json:"appTenantId"`
		TenantDomain  string        `json:"tenantDomainName"`
		ShouldReplace bool          `json:"shouldReplace"`
	}{Cloud: cloud, ID: appID, Name: appName, TenantID: appTenantID, TenantDomain: appTenantDomain, SecretKey: secret.String(appSecretKey), ShouldReplace: shouldReplace})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result {
		return graphql.ResponseError(query, errors.New("set app credentials failed"))
	}

	return nil
}
