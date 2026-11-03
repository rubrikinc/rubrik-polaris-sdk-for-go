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
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// rekeyPollInterval is the interval between checks of the rekey status when
// waiting for a rekey to complete.
const rekeyPollInterval = 10 * time.Second

// RekeyMasterKey rekeys the master key of the cloud native archival location
// with the specified target mapping ID, moving it to the key given in the
// parameters. Both RCV and customer hosted cloud native archival locations can
// be rekeyed. The function returns once the rekey has completed. If the rekey
// fails for any of the group's archival locations, an error holding the error
// message reported by RSC is returned.
func (a API) RekeyMasterKey(ctx context.Context, targetMappingID uuid.UUID, params archival.RekeyArchivalGroupMasterKeyParams) error {
	a.log.Print(log.Trace)

	if err := archival.RekeyArchivalGroupMasterKey(ctx, a.client, targetMappingID, params); err != nil {
		return fmt.Errorf("failed to rekey master key: %s", err)
	}

	if err := a.waitForRekey(ctx, targetMappingID); err != nil {
		return fmt.Errorf("failed to wait for rekey: %s", err)
	}

	return nil
}

// ArchivalGroupEncryptionDetails returns the encryption details of the cloud
// native archival groups with a name containing the specified name filter. All
// groups are returned when the name filter is empty. RSC only lists the groups
// encrypted with a key from a key manager.
func (a API) ArchivalGroupEncryptionDetails(ctx context.Context, nameFilter string) ([]archival.ArchivalGroupEncryptionDetails, error) {
	a.log.Print(log.Trace)

	groups, err := archival.ListArchivalGroupEncryptionDetails(ctx, a.client, nameFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption details: %s", err)
	}

	return groups, nil
}

// ArchivalGroupEncryptionDetailsByID returns the encryption details of the
// cloud native archival group with the specified ID. RSC only lists the groups
// encrypted with a key from a key manager, so graphql.ErrNotFound is returned
// both when the group doesn't exist and when it has no such key. The details
// can only be filtered by name, so all groups are listed.
func (a API) ArchivalGroupEncryptionDetailsByID(ctx context.Context, id uuid.UUID) (archival.ArchivalGroupEncryptionDetails, error) {
	a.log.Print(log.Trace)

	groups, err := a.ArchivalGroupEncryptionDetails(ctx, "")
	if err != nil {
		return archival.ArchivalGroupEncryptionDetails{}, err
	}

	for _, group := range groups {
		if group.ID == id {
			return group, nil
		}
	}

	return archival.ArchivalGroupEncryptionDetails{}, fmt.Errorf("encryption details for %q %w", id, graphql.ErrNotFound)
}

// waitForRekey waits for the master key rekey of the archival group with the
// specified ID to complete. A failed rekey of any of the group's archival
// locations is returned as an error. RSC rekeys the group's own archival
// location when the rekey is requested and only the child archival locations
// are rekeyed in the background, so a group without child archival locations
// has completed its rekey as soon as the request returns, and its rekey status
// is never reported as successful.
func (a API) waitForRekey(ctx context.Context, groupID uuid.UUID) error {
	a.log.Print(log.Trace)

	for {
		group, err := a.ArchivalGroupEncryptionDetailsByID(ctx, groupID)
		if err != nil {
			return err
		}

		if len(group.Locations) == 0 {
			return nil
		}

		// A failed location keeps the group in progress, so failures must be
		// checked before the group status.
		for _, location := range group.Locations {
			job := location.EncryptionJobs.LastMasterKeyRekey
			if job != nil && job.Status == archival.EncryptionJobFailed {
				return fmt.Errorf("rekey of archival location %q failed: %s", location.ID, job.ErrorMessage)
			}
		}
		if group.RekeyStatus == archival.RekeyStatusSuccess {
			return nil
		}

		select {
		case <-time.After(rekeyPollInterval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
