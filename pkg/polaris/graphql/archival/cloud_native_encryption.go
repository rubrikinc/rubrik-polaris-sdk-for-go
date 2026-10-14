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

package archival

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/kms"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// RekeyStatus represents the RSC GraphQL ArchivalGroupRekeyStatus enum type.
// It is the rekey status of an archival group, aggregated across its child
// archival locations.
type RekeyStatus string

const (
	RekeyStatusUnspecified RekeyStatus = "ARCHIVAL_GROUP_REKEY_STATUS_UNSPECIFIED"
	RekeyStatusInProgress  RekeyStatus = "ARCHIVAL_GROUP_REKEY_IN_PROGRESS"
	RekeyStatusSuccess     RekeyStatus = "ARCHIVAL_GROUP_REKEY_SUCCESS"
)

// RekeyStatusValues returns all archival group rekey statuses.
func RekeyStatusValues() []RekeyStatus {
	return slices.Clone([]RekeyStatus{
		RekeyStatusUnspecified,
		RekeyStatusInProgress,
		RekeyStatusSuccess,
	})
}

// EncryptionJobStatus represents the RSC GraphQL
// CnpArchivalLocationEncryptionJobStatus enum type. It is the status of a
// single key management operation for an archival location.
type EncryptionJobStatus string

const (
	EncryptionJobUnspecified EncryptionJobStatus = "CNP_ARCHIVAL_LOCATION_ENCRYPTION_JOB_STATUS_UNSPECIFIED"
	EncryptionJobQueued      EncryptionJobStatus = "CNP_ARCHIVAL_LOCATION_ENCRYPTION_JOB_STATUS_QUEUED"
	EncryptionJobInProgress  EncryptionJobStatus = "CNP_ARCHIVAL_LOCATION_ENCRYPTION_JOB_STATUS_IN_PROGRESS"
	EncryptionJobSucceeded   EncryptionJobStatus = "CNP_ARCHIVAL_LOCATION_ENCRYPTION_JOB_STATUS_SUCCEEDED"
	EncryptionJobFailed      EncryptionJobStatus = "CNP_ARCHIVAL_LOCATION_ENCRYPTION_JOB_STATUS_FAILED"
)

// EncryptionJobStatusValues returns all encryption job statuses.
func EncryptionJobStatusValues() []EncryptionJobStatus {
	return slices.Clone([]EncryptionJobStatus{
		EncryptionJobUnspecified,
		EncryptionJobQueued,
		EncryptionJobInProgress,
		EncryptionJobSucceeded,
		EncryptionJobFailed,
	})
}

// EncryptionJob holds the status of a single key management operation for
// an archival location. A status of unspecified means the operation has not run
// for the location.
type EncryptionJob struct {
	Status           EncryptionJobStatus `json:"status"`
	LastSuccessfulAt *time.Time          `json:"lastSuccessfulAt"`
	ErrorMessage     string              `json:"errorMessage"`
}

// ArchivalLocationEncryptionDetails holds the encryption details of one
// child archival location in a cloud native archival group. The location is
// the group's template location excluded.
type ArchivalLocationEncryptionDetails struct {
	ID             string `json:"archivalLocationId"`
	Region         string `json:"region"`
	EncryptionJobs struct {
		LastMasterKeyRekey *EncryptionJob `json:"lastMasterKeyRekeyStatus"`
	} `json:"encryptionJobsSummary"`
	KMSKey *struct {
		KeyManagerID string   `json:"keyVaultId"`
		Type         kms.Type `json:"kmsType"`
		Key          *struct {
			GCPKMSKey *struct {
				KeyName    string `json:"keyName"`
				KeyVersion string `json:"keyVersion"`
			} `json:"gcpKmsKey"`
		} `json:"key"`
	} `json:"kmsKey"`
}

// ArchivalGroupEncryptionDetails holds the encryption details of a cloud
// native archival group and its child archival locations with an active key
// hierarchy.
type ArchivalGroupEncryptionDetails struct {
	ID          uuid.UUID                           `json:"archivalGroupId"`
	Name        string                              `json:"name"`
	RekeyStatus RekeyStatus                         `json:"rekeyStatus"`
	Locations   []ArchivalLocationEncryptionDetails `json:"archivalLocations"`
}

// ListArchivalGroupEncryptionDetails returns the encryption details of the
// cloud native archival groups with a name starting with the specified prefix.
// All groups are returned when the prefix is empty.
func ListArchivalGroupEncryptionDetails(ctx context.Context, gql *graphql.Client, namePrefix string) ([]ArchivalGroupEncryptionDetails, error) {
	gql.Log().Print(log.Trace)

	var cursor string
	var groups []ArchivalGroupEncryptionDetails
	for {
		query := cnpArchivalGroupsEncryptionDetailsQuery
		buf, err := gql.Request(ctx, query, struct {
			First            int    `json:"first"`
			After            string `json:"after,omitempty"`
			NameSearchPrefix string `json:"nameSearchPrefix,omitempty"`
		}{First: 100, After: cursor, NameSearchPrefix: namePrefix})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []ArchivalGroupEncryptionDetails `json:"nodes"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, graphql.UnmarshalError(query, err)
		}

		groups = append(groups, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return groups, nil
}

// RekeyArchivalGroupMasterKeyParams holds the parameters for a cloud native
// archival group master key rekey operation.
type RekeyArchivalGroupMasterKeyParams struct {
	// KeyManagerID is the ID of the key manager holding the new key.
	KeyManagerID uuid.UUID

	// GCPKMSKey is the new GCP Cloud KMS key. Both the key name and the key
	// version are required.
	GCPKMSKey GCPKMSKey
}

// RekeyArchivalGroupMasterKey rekeys the master key of the cloud native
// archival group with the specified ID, moving the group's template and all of
// its active archival locations to the new key. The rekey is asynchronous, the
// rekey status of the group and its locations can be read with
// ListArchivalGroupEncryptionDetails.
func RekeyArchivalGroupMasterKey(ctx context.Context, gql *graphql.Client, groupID uuid.UUID, params RekeyArchivalGroupMasterKeyParams) error {
	gql.Log().Print(log.Trace)

	query := rekeyCnpArchivalGroupMasterKeyQuery
	buf, err := gql.Request(ctx, query, struct {
		ArchivalGroupID uuid.UUID `json:"archivalGroupId"`
		KeyVaultID      uuid.UUID `json:"keyVaultId"`
		KMSType         kms.Type  `json:"kmsType"`
		GCPKMSKey       struct {
			KeyName    string `json:"keyName"`
			KeyVersion string `json:"keyVersion"`
		} `json:"gcpKmsKey"`
	}{
		ArchivalGroupID: groupID,
		KeyVaultID:      params.KeyManagerID,
		KMSType:         kms.TypeGCPKMS,
		GCPKMSKey: struct {
			KeyName    string `json:"keyName"`
			KeyVersion string `json:"keyVersion"`
		}{KeyName: params.GCPKMSKey.KeyName, KeyVersion: params.GCPKMSKey.KeyVersion},
	})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result json.RawMessage `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}

// RekeyArchivalGroupRootKey rekeys the root key of the cloud native archival
// group with the specified ID. The master key is unchanged, only the root key
// wrapping the data encryption keys is replaced. The rekey is asynchronous.
func RekeyArchivalGroupRootKey(ctx context.Context, gql *graphql.Client, groupID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := rekeyCnpArchivalGroupRootKeyQuery
	buf, err := gql.Request(ctx, query, struct {
		ArchivalGroupID uuid.UUID `json:"archivalGroupId"`
	}{ArchivalGroupID: groupID})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result json.RawMessage `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}
