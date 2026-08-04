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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
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
// The login and password are the credentials of a SQL Server user with
// permission to create the user RSC uses to perform backups. They are sent to
// RSC but never stored by the SDK. Microsoft Entra ID authentication is not
// supported yet.
//
// A pollInterval of zero uses a default interval. Note that the credentials are
// only validated against the managed instance once the task chain runs, so
// invalid credentials surface as a failed task chain rather than as an error
// from the initial request.
func (a API) SetupSQLManagedInstanceBackup(ctx context.Context, serverIDs []uuid.UUID, login string, password secret.String, pollInterval time.Duration) error {
	a.log.Print(log.Trace)

	if len(serverIDs) == 0 {
		return errors.New("at least one server ID is required")
	}
	if pollInterval <= 0 {
		pollInterval = defaultSQLServerBackupSetupPollInterval
	}

	status, err := gqlazure.Wrap(a.client).SetupCloudNativeSQLServerBackup(ctx, serverIDs, nil,
		gqlazure.LoginCredentials{Login: login, Password: password}, gqlazure.SQLAuthentication)
	if err != nil {
		return fmt.Errorf("failed to set up SQL Managed Instance backup: %s", err)
	}

	// Objects which fail pre-validation never get a task chain, so they must be
	// reported separately from the task chain results below.
	for _, e := range status.Errors {
		return fmt.Errorf("failed to set up backup for SQL Managed Instance server %s: %s", e.ObjectID, e.Error)
	}
	if len(status.JobIDs) == 0 {
		return errors.New("failed to set up SQL Managed Instance backup: no jobs were started")
	}

	coreAPI := core.Wrap(a.client)
	for _, job := range status.JobIDs {
		state, err := coreAPI.WaitForTaskChain(ctx, job.JobID, pollInterval)
		if err != nil {
			return fmt.Errorf("failed to wait for backup setup of SQL Managed Instance server %s: %s", job.ObjectID, err)
		}
		if state != core.TaskChainSucceeded {
			return fmt.Errorf("backup setup of SQL Managed Instance server %s ended in state %s, verify that the "+
				"credentials are valid for the managed instance", job.ObjectID, state)
		}
	}

	return nil
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
