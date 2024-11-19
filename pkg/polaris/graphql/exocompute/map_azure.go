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

package exocompute

import (
	"errors"

	"github.com/google/uuid"
)

// MapAzureCloudAccountsParams holds the parameters for an Azure application
// cloud accounts map operation.
type MapAzureCloudAccountsParams struct {
	AppCloudAccountIDs []uuid.UUID `json:"cloudAccountIds"`
	HostCloudAccountID uuid.UUID   `json:"exocomputeCloudAccountId"`
}

func (p MapAzureCloudAccountsParams) MapQuery() (string, any, MapAzureCloudAccountsResult) {
	return mapAzureCloudAccountExocomputeSubscriptionQuery, p, MapAzureCloudAccountsResult{}
}

// MapAzureCloudAccountsResult holds the result of an Azure application cloud
// accounts map operation.
type MapAzureCloudAccountsResult struct {
	Success bool `json:"isSuccess"`
}

func (r MapAzureCloudAccountsResult) Validate() error {
	if !r.Success {
		return errors.New("failed to map application cloud accounts")
	}
	return nil
}

// UnmapAzureCloudAccountsParams holds the parameters for an Azure application
// cloud accounts unmap operation.
type UnmapAzureCloudAccountsParams struct {
	AppCloudAccountIDs []uuid.UUID `json:"cloudAccountIds"`
}

func (p UnmapAzureCloudAccountsParams) UnmapQuery() (string, any, UnmapAzureCloudAccountsResult) {
	return unmapAzureCloudAccountExocomputeSubscriptionQuery, p, UnmapAzureCloudAccountsResult{}
}

// UnmapAzureCloudAccountsResult holds the result of an Azure application cloud
// accounts unmap operation.
type UnmapAzureCloudAccountsResult struct {
	Success bool `json:"isSuccess"`
}

func (r UnmapAzureCloudAccountsResult) Validate() error {
	if !r.Success {
		return errors.New("failed to unmap application cloud accounts")
	}

	return nil
}
