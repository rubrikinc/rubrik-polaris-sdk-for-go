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
	"testing"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/devops"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// TestValidateHostStorage covers the conditional rules coupling host/storage
// type to their dependent identifiers.
func TestValidateHostStorage(t *testing.T) {
	loc := uuid.New()
	acct := uuid.New()
	region := azure.RegionEastUS
	unknown := azure.RegionUnknown

	tests := []struct {
		name    string
		host    devops.HostType
		storage devops.StorageType
		locID   *uuid.UUID
		acctID  *uuid.UUID
		region  *azure.Region
		wantErr bool
	}{{
		name:    "BYOS without backup location",
		host:    devops.HostTypeRubrik,
		storage: devops.StorageTypeBYOS,
		region:  &region,
		wantErr: true,
	}, {
		name:    "customer host without exocompute account",
		host:    devops.HostTypeCustomer,
		storage: devops.StorageTypeRCV,
		wantErr: true,
	}, {
		name:    "rubrik host without region",
		host:    devops.HostTypeRubrik,
		storage: devops.StorageTypeRCV,
		wantErr: true,
	}, {
		name:    "rubrik host with unknown region",
		host:    devops.HostTypeRubrik,
		storage: devops.StorageTypeRCV,
		region:  &unknown,
		wantErr: true,
	}, {
		name:    "valid rubrik host RCV",
		host:    devops.HostTypeRubrik,
		storage: devops.StorageTypeRCV,
		region:  &region,
		wantErr: false,
	}, {
		name:    "valid customer host RCV",
		host:    devops.HostTypeCustomer,
		storage: devops.StorageTypeRCV,
		acctID:  &acct,
		wantErr: false,
	}, {
		name:    "valid BYOS customer host",
		host:    devops.HostTypeCustomer,
		storage: devops.StorageTypeBYOS,
		locID:   &loc,
		acctID:  &acct,
		wantErr: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHostStorage(tt.host, tt.storage, tt.locID, tt.acctID, tt.region)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHostStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
