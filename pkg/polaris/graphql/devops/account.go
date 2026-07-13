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

package devops

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AzureOnboardingScript holds the onboarding scripts returned by
// GenerateAzureOnboardingScript. Both scripts are base64-encoded exactly as returned
// by the RSC API; decoding is left to the caller.
type AzureOnboardingScript struct {
	BashScript       string `json:"bashScript"`
	PowershellScript string `json:"powershellScript"`
}

// GenerateAzureOnboardingScriptParams holds the parameters for
// GenerateAzureOnboardingScript. TenantDomain carries the Azure AD tenant domain.
// OrganizationNativeIDs is omitted from the request when empty.
type GenerateAzureOnboardingScriptParams struct {
	TenantDomain          string         `json:"tenantId"`
	Cloud                 gqlazure.Cloud `json:"cloudType"`
	Features              []core.Feature `json:"featuresWithPermissionsGroups"`
	OrganizationNativeIDs []string       `json:"organizationNativeIds,omitempty"`
}

// GenerateAzureOnboardingScript generates the Azure DevOps onboarding scripts for the
// given tenant and features. The returned scripts are base64-encoded.
func GenerateAzureOnboardingScript(ctx context.Context, gql *graphql.Client, params GenerateAzureOnboardingScriptParams) (AzureOnboardingScript, error) {
	gql.Log().Print(log.Trace)

	query := generateOnboardingScriptQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return AzureOnboardingScript{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result AzureOnboardingScript `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AzureOnboardingScript{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// AddAzureCloudAccountParams holds the parameters for AddAzureCloudAccountWithoutOauth.
// TenantDomain carries the Azure AD tenant domain. The nullable pointer fields
// are omitted from the request when nil.
type AddAzureCloudAccountParams struct {
	OrganizationNativeIDs    []string       `json:"organizationNativeIds,omitempty"`
	TenantDomain             string         `json:"tenantId"`
	Cloud                    gqlazure.Cloud `json:"cloudType"`
	Features                 []core.Feature `json:"featuresWithPermissionsGroups"`
	HostType                 HostType       `json:"hostType,omitempty"`
	StorageType              StorageType    `json:"storageType,omitempty"`
	BackupLocationID         *uuid.UUID     `json:"backupLocationId,omitempty"`
	ExocomputeCloudAccountID *uuid.UUID     `json:"exocomputeCloudAccountId,omitempty"`
	ExocomputeRegion         *azure.Region  `json:"exocomputeRegion,omitempty"`
}

// AddAzureCloudAccountWithoutOauth onboards one or more Azure DevOps organizations
// using a customer-supplied application (non-OAuth). The mutation returns Void,
// so the caller must confirm the result by querying the organizations.
func AddAzureCloudAccountWithoutOauth(ctx context.Context, gql *graphql.Client, params AddAzureCloudAccountParams) error {
	gql.Log().Print(log.Trace)

	query := addCloudAccountWithoutOauthQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct{} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}

// UpdateAzureCloudAccountParams holds the parameters for UpdateAzureCloudAccount. The
// organization is keyed by OrganizationID. The nullable pointer fields are
// omitted from the request when nil.
type UpdateAzureCloudAccountParams struct {
	OrganizationID           uuid.UUID     `json:"organizationId"`
	BackupLocationID         *uuid.UUID    `json:"backupLocationId,omitempty"`
	BackupRegion             *azure.Region `json:"backupRegion,omitempty"`
	ExocomputeCloudAccountID *uuid.UUID    `json:"exocomputeCloudAccountId,omitempty"`
	HostType                 HostType      `json:"hostType,omitempty"`
	StorageType              StorageType   `json:"storageType,omitempty"`
	ExocomputeRegion         *azure.Region `json:"exocomputeRegion,omitempty"`
}

// UpdateAzureCloudAccount updates the backup location/region, exocompute account/
// region and host/storage type of an onboarded Azure DevOps organization. The
// mutation returns Void.
func UpdateAzureCloudAccount(ctx context.Context, gql *graphql.Client, params UpdateAzureCloudAccountParams) error {
	gql.Log().Print(log.Trace)

	query := updateCloudAccountQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct{} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}

// DeleteAzureCloudAccountWithoutOauth removes an onboarded Azure DevOps organization,
// keyed by organizationID. If deleteSnapshots is true, the organization's
// snapshots are also deleted. The mutation returns Void.
func DeleteAzureCloudAccountWithoutOauth(ctx context.Context, gql *graphql.Client, organizationID uuid.UUID, deleteSnapshots bool) error {
	gql.Log().Print(log.Trace)

	query := deleteCloudAccountWithoutOauthQuery
	buf, err := gql.Request(ctx, query, struct {
		OrganizationID  uuid.UUID `json:"organizationId"`
		DeleteSnapshots bool      `json:"deleteSnapshots"`
	}{OrganizationID: organizationID, DeleteSnapshots: deleteSnapshots})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct{} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}
