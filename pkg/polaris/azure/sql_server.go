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

package azure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// defaultSQLServerBackupSetupPollInterval is how often the backup setup task
// chains are polled when no interval is specified.
const defaultSQLServerBackupSetupPollInterval = 10 * time.Second

// SetupSQLManagedInstanceBackup sets up backups for the specified Azure SQL
// Managed Instance servers using SQL Server authentication, and blocks until
// every task chain started has finished.
//
// The credentials are those of a SQL Server user with permission to create the
// user RSC uses to perform backups. They are sent to RSC but never stored by
// the SDK. Microsoft Entra ID authentication is not supported yet.
//
// A pollInterval of zero uses a default interval. Note that the credentials are
// only validated against the managed instance once the task chain runs, so
// invalid credentials surface as a failed task chain rather than as an error
// from the initial request.
//
// Servers are set up independently of each other, so a failure for one server
// does not stop the others. All failures are collected and returned together as
// a joined error.
func (a API) SetupSQLManagedInstanceBackup(ctx context.Context, serverIDs []uuid.UUID, credentials gqlazure.LoginCredentials, pollInterval time.Duration) error {
	a.log.Print(log.Trace)

	if len(serverIDs) == 0 {
		return errors.New("at least one server ID is required")
	}
	if pollInterval <= 0 {
		pollInterval = defaultSQLServerBackupSetupPollInterval
	}

	status, err := gqlazure.Wrap(a.client).SetupCloudNativeSQLServerBackup(ctx, serverIDs, nil, credentials,
		gqlazure.SQLAuthentication)
	if err != nil {
		return fmt.Errorf("failed to set up SQL Managed Instance backup: %s", err)
	}

	// Objects which fail pre-validation never get a task chain, so they are
	// reported separately from the task chain results below. Note that they are
	// only collected here, not returned yet: task chains may already be running
	// for the other servers and those must still be waited for.
	var errs []error
	acknowledged := make(map[uuid.UUID]struct{}, len(status.JobIDs)+len(status.Errors))
	for _, e := range status.Errors {
		acknowledged[e.ObjectID] = struct{}{}
		errs = append(errs, fmt.Errorf("failed to set up backup for SQL Managed Instance server %s: %s",
			e.ObjectID, e.Error))
	}
	for _, job := range status.JobIDs {
		acknowledged[job.ObjectID] = struct{}{}
	}

	// RSC acknowledges each requested server with either a task chain or a
	// pre-validation error. A server in neither list was silently dropped, so
	// report it instead of returning success for a server never set up. This
	// also covers RSC returning nothing at all, in which case every requested
	// server is reported.
	for _, serverID := range serverIDs {
		if _, ok := acknowledged[serverID]; !ok {
			errs = append(errs, fmt.Errorf("no job was started for SQL Managed Instance server %s", serverID))
		}
	}

	// Wait for every task chain, recording rather than returning failures, so
	// that a failure for one server does not leave the remaining task chains
	// running unwaited.
	coreAPI := core.Wrap(a.client)
	for _, job := range status.JobIDs {
		// Once the context is done, waiting for the remaining task chains only
		// produces one duplicate failure per server, so stop and report the
		// cause once. Wrapped with %w, unlike the errors below, so that callers
		// can tell cancellation apart from a genuine setup failure.
		if err := ctx.Err(); err != nil {
			errs = append(errs, fmt.Errorf("interrupted while waiting for SQL Managed Instance backup setup: %w", err))
			break
		}

		state, err := coreAPI.WaitForTaskChain(ctx, job.JobID, pollInterval)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("failed to wait for backup setup of SQL Managed Instance server %s: %s",
				job.ObjectID, err))
		case state != core.TaskChainSucceeded:
			errs = append(errs, fmt.Errorf("backup setup of SQL Managed Instance server %s ended in state %s, "+
				"verify that the credentials are valid for the managed instance", job.ObjectID, state))
		}
	}

	return errors.Join(errs...)
}

// ClearSQLManagedInstanceBackupCredentials clears the backup credentials of the
// specified Azure SQL Managed Instance servers.
func (a API) ClearSQLManagedInstanceBackupCredentials(ctx context.Context, serverIDs []uuid.UUID) error {
	a.log.Print(log.Trace)

	if len(serverIDs) == 0 {
		return errors.New("at least one server ID is required")
	}

	_, failedIDs, err := gqlazure.Wrap(a.client).ClearCloudNativeSQLServerBackupCredentials(ctx, serverIDs,
		hierarchy.WorkloadAzureSQLMIDB)
	if err != nil {
		return fmt.Errorf("failed to clear SQL Managed Instance backup credentials: %s", err)
	}
	if len(failedIDs) > 0 {
		return fmt.Errorf("failed to clear backup credentials for SQL Managed Instance servers: %v", failedIDs)
	}

	return nil
}
