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

// MapAWSCloudAccountsParams holds the parameters for an AWS application cloud
// accounts map operation.
type MapAWSCloudAccountsParams struct {
	AppCloudAccountIDs []uuid.UUID `json:"cloudAccountIds"`
	HostCloudAccountID uuid.UUID   `json:"exocomputeCloudAccountId"`
}

func (p MapAWSCloudAccountsParams) MapQuery() (string, any, MapAWSCloudAccountsResult) {
	params := struct {
		MapAWSCloudAccountsParams
		CloudVendor string `json:"cloudVendor"`
	}{MapAWSCloudAccountsParams: p, CloudVendor: "AWS"}
	return mapCloudAccountExocomputeAccountQuery, params, MapAWSCloudAccountsResult{}
}

// MapAWSCloudAccountsResult holds the result of an AWS application cloud
// accounts map operation.
type MapAWSCloudAccountsResult struct {
	Success bool `json:"isSuccess"`
}

func (r MapAWSCloudAccountsResult) Validate() error {
	if !r.Success {
		return errors.New("failed to map application cloud accounts")
	}
	return nil
}

// UnmapAWSCloudAccountsParams holds the parameters for an AWS application cloud
// accounts unmap operation.
type UnmapAWSCloudAccountsParams struct {
	AppCloudAccountIDs []uuid.UUID `json:"cloudAccountIds"`
}

func (p UnmapAWSCloudAccountsParams) UnmapQuery() (string, any, UnmapAWSCloudAccountsResult) {
	params := struct {
		UnmapAWSCloudAccountsParams
		CloudVendor string `json:"cloudVendor"`
	}{UnmapAWSCloudAccountsParams: p, CloudVendor: "AWS"}
	return unmapCloudAccountExocomputeAccountQuery, params, UnmapAWSCloudAccountsResult{}
}

// UnmapAWSCloudAccountsResult holds the result of an AWS application cloud
// accounts unmap operation.
type UnmapAWSCloudAccountsResult struct {
	Success bool `json:"isSuccess"`
}

func (r UnmapAWSCloudAccountsResult) Validate() error {
	if !r.Success {
		return errors.New("failed to unmap application cloud accounts")
	}
	return nil
}
