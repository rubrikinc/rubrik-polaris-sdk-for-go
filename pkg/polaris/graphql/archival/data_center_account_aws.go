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

import (
	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/secret"
)

// AWSCloudAccount represents an AWS data center cloud account.
// Note, these cloud accounts are separate from the cloud native cloud accounts.
type AWSCloudAccount struct {
	ID               uuid.UUID `json:"cloudAccountId"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	CloudProvider    string    `json:"cloudProvider"`
	ConnectionStatus string    `json:"connectionStatus"`
	AccessKey        string    `json:"accessKey"`
}

func (AWSCloudAccount) ListQuery(filters []ListAccountFilter) (string, any) {
	extraFilters := []ListAccountFilter{{
		Field: "ACCOUNT_PROVIDER_TYPE",
		Text:  "CLOUD_ACCOUNT_AWS",
	}}
	return allCloudAccountsQuery, struct {
		Filters []ListAccountFilter `json:"filters"`
	}{Filters: append(filters, extraFilters...)}
}

func (r AWSCloudAccount) Validate() bool {
	return r.AccessKey != ""
}

// CreateAWSCloudAccountParams holds the parameters for an AWS data center
// cloud account create operation.
type CreateAWSCloudAccountParams struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	AccessKey   secret.String `json:"accessKey"`
	SecretKey   secret.String `json:"secretKey"`
}

// CreateAWSCloudAccountResult holds the result of an AWS data center cloud
// account create operation.
type CreateAWSCloudAccountResult struct {
	ID string `json:"cloudAccountId"`
}

func (CreateAWSCloudAccountResult) CreateQuery(createParams CreateAWSCloudAccountParams) (string, any) {
	return createAwsAccountQuery, createParams
}

func (r CreateAWSCloudAccountResult) Validate() (uuid.UUID, error) {
	return uuid.Parse(r.ID)
}

// UpdateAWSCloudAccountParams holds the parameters for an AWS data center
// cloud account update operation.
type UpdateAWSCloudAccountParams CreateAWSCloudAccountParams

// UpdateAWSCloudAccountResult holds the result of an AWS data center cloud
// account update operation.
type UpdateAWSCloudAccountResult CreateAWSCloudAccountResult

func (UpdateAWSCloudAccountResult) UpdateQuery(cloudAccountID uuid.UUID, updateParams UpdateAWSCloudAccountParams) (string, any) {
	return updateAwsAccountQuery, struct {
		ID uuid.UUID `json:"id"`
		UpdateAWSCloudAccountParams
	}{ID: cloudAccountID, UpdateAWSCloudAccountParams: updateParams}
}

// DeleteAWSCloudAccountResult holds the result of an AWS data center cloud
// account delete operation.
type DeleteAWSCloudAccountResult struct{}

func (DeleteAWSCloudAccountResult) DeleteQuery(id uuid.UUID) (string, any) {
	return deleteAwsDataCenterKeyBasedCloudAccountQuery, struct {
		ID uuid.UUID `json:"cloudAccountId"`
	}{ID: id}
}
