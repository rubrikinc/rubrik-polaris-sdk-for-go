// Copyright 2024 Rubrik, Inc.
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

package archival

import "github.com/google/uuid"

// AzureCloudAccount represents an Azure data center cloud account.
// Note, these cloud accounts are separate from the cloud native cloud accounts.
type AzureCloudAccount struct {
	ID               uuid.UUID `json:"cloudAccountId"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	CloudProvider    string    `json:"cloudProvider"`
	ConnectionStatus string    `json:"connectionStatus"`
	SubscriptionID   uuid.UUID `json:"subscriptionId"`
	TenantID         uuid.UUID `json:"tenantId"`
}

func (AzureCloudAccount) ListQuery(filters []ListAccountFilter) (string, any) {
	extraFilters := []ListAccountFilter{{
		Field: "ACCOUNT_PROVIDER_TYPE",
		Text:  "CLOUD_ACCOUNT_AZURE",
	}}
	return allCloudAccountsQuery, struct {
		Filters []ListAccountFilter `json:"filters"`
	}{Filters: append(filters, extraFilters...)}
}

func (r AzureCloudAccount) Validate() bool {
	return r.SubscriptionID != uuid.Nil
}

// CreateAzureCloudAccountParams holds the parameters for an Azure data center
// cloud account create operation.
type CreateAzureCloudAccountParams struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	SubscriptionID string `json:"subscriptionId"`
}

// CreateAzureCloudAccountResult holds the result of an Azure data center cloud
// account create operation.
type CreateAzureCloudAccountResult struct {
	ID string `json:"cloudAccountId"`
}

func (CreateAzureCloudAccountResult) CreateQuery(createParams CreateAzureCloudAccountParams) (string, any) {
	return createAzureAccountQuery, createParams
}

func (r CreateAzureCloudAccountResult) Validate() (uuid.UUID, error) {
	return uuid.Parse(r.ID)
}

// UpdateAzureCloudAccountParams holds the parameters for an Azure data center
// cloud account update operation.
type UpdateAzureCloudAccountParams CreateAzureCloudAccountParams

// UpdateAzureCloudAccountResult holds the result of an Azure data center cloud
// account update operation.
type UpdateAzureCloudAccountResult CreateAzureCloudAccountResult

func (UpdateAzureCloudAccountResult) UpdateQuery(cloudAccountID uuid.UUID, updateParams UpdateAzureCloudAccountParams) (string, any) {
	return updateAzureAccountQuery, struct {
		ID uuid.UUID `json:"id"`
		UpdateAzureCloudAccountParams
	}{ID: cloudAccountID, UpdateAzureCloudAccountParams: updateParams}
}

func (r UpdateAzureCloudAccountResult) Validate() (uuid.UUID, error) {
	return uuid.Parse(r.ID)
}

// DeleteAzureCloudAccountResult holds the result of an Azure data center cloud
// account delete operation.
type DeleteAzureCloudAccountResult struct{}

func (DeleteAzureCloudAccountResult) DeleteQuery(id uuid.UUID) (string, any) {
	return deleteAzureDataCenterCloudAccountQuery, struct {
		ID uuid.UUID `json:"cloudAccountId"`
	}{ID: id}
}
