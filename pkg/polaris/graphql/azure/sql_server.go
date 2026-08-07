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
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// SQLAuthenticationMechanism represents the mechanism used to authenticate
// against a SQL Server when setting up backups.
type SQLAuthenticationMechanism string

const (
	// SQLAuthenticationUnspecified means no authentication mechanism has been
	// specified.
	SQLAuthenticationUnspecified SQLAuthenticationMechanism = "AUTHENTICATION_MECHANISM_UNSPECIFIED"

	// SQLAuthentication authenticates using traditional SQL Server credentials.
	SQLAuthentication SQLAuthenticationMechanism = "SQL_AUTHENTICATION"

	// SQLAuthenticationEntraIDAuthCode authenticates using a Microsoft Entra ID
	// authorization code. Using it requires an OAuth session ID, which the SDK
	// does not support yet.
	SQLAuthenticationEntraIDAuthCode SQLAuthenticationMechanism = "AZURE_ACTIVE_DIRECTORY_AUTH_CODE"
)

// LoginCredentials holds the login and password of a database user.
type LoginCredentials struct {
	Login    string        `json:"login"`
	Password secret.String `json:"password"`
}

// AsyncJobID maps an RSC object to the job started for it.
type AsyncJobID struct {
	// JobID is the ID of the task chain started for the object.
	JobID uuid.UUID `json:"jobId"`
	// ObjectID is the RSC ID of the object the job was started for.
	ObjectID uuid.UUID `json:"rubrikObjectId"`
}

// AsyncJobError maps an RSC object to the reason no job was started for it.
type AsyncJobError struct {
	Error string `json:"error"`
	// ObjectID is the RSC ID of the object the job failed for.
	ObjectID uuid.UUID `json:"rubrikObjectId"`
}

// BatchAsyncJobStatus is the result of starting a batch of asynchronous jobs.
// Objects which passed pre-validation are listed in JobIDs, the rest are listed
// in Errors.
type BatchAsyncJobStatus struct {
	JobIDs []AsyncJobID    `json:"jobIds"`
	Errors []AsyncJobError `json:"errors"`
}

// SetupCloudNativeSQLServerBackup sets up backups for the specified SQL Server
// databases or servers. Server IDs are currently only supported for Azure SQL
// Managed Instance BAK backups.
//
// The admin credentials are the credentials of a database user with permission
// to create the user RSC uses to perform backups. They are only required when
// the authentication mechanism is SQLAuthentication.
//
// The returned job IDs are task chain IDs which can be passed to
// core.API.WaitForTaskChain to wait for the setup to finish.
func (a API) SetupCloudNativeSQLServerBackup(ctx context.Context, serverIDs, databaseIDs []uuid.UUID, adminCredentials LoginCredentials, authMechanism SQLAuthenticationMechanism) (BatchAsyncJobStatus, error) {
	a.log.Print(log.Trace)

	// The RSC API expects both ID fields to be present, so nil slices are
	// normalized to empty slices rather than being omitted.
	if serverIDs == nil {
		serverIDs = []uuid.UUID{}
	}
	if databaseIDs == nil {
		databaseIDs = []uuid.UUID{}
	}

	// The credentials are only required for SQLAuthentication, so omit them
	// entirely when none were given rather than sending an empty login and
	// password. omitempty only fires for a nil pointer, it does not apply to
	// struct values.
	credentials := &adminCredentials
	if adminCredentials == (LoginCredentials{}) {
		credentials = nil
	}

	query := setupCloudNativeSqlServerBackupQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ServerIDs        []uuid.UUID                `json:"serverIds"`
		DatabaseIDs      []uuid.UUID                `json:"databaseIds"`
		AdminCredentials *LoginCredentials          `json:"adminCredentials,omitempty"`
		AuthMechanism    SQLAuthenticationMechanism `json:"authMechanism,omitempty"`
	}{ServerIDs: serverIDs, DatabaseIDs: databaseIDs, AdminCredentials: credentials, AuthMechanism: authMechanism})
	if err != nil {
		return BatchAsyncJobStatus{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result BatchAsyncJobStatus `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return BatchAsyncJobStatus{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// ClearCloudNativeSQLServerBackupCredentials clears the backup credentials of
// the specified objects. The object IDs can refer to any level of the hierarchy
// the credentials were set on, e.g. subscriptions, resource groups, servers or
// databases. The workload type must match the object type the credentials apply
// to, e.g. hierarchy.WorkloadAzureSQLMIDB for Azure SQL Managed Instances.
//
// Returns the IDs of the objects whose credentials were cleared and the IDs of
// the objects whose credentials could not be cleared.
func (a API) ClearCloudNativeSQLServerBackupCredentials(ctx context.Context, objectIDs []uuid.UUID, workloadType hierarchy.Workload) (successIDs, failedIDs []uuid.UUID, err error) {
	a.log.Print(log.Trace)

	query := clearCloudNativeSqlServerBackupCredentialsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ObjectIDs    []uuid.UUID        `json:"objectIds"`
		WorkloadType hierarchy.Workload `json:"workloadType"`
	}{ObjectIDs: objectIDs, WorkloadType: workloadType})
	if err != nil {
		return nil, nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				SuccessObjectIDs []uuid.UUID `json:"successObjectIds"`
				FailedObjectIDs  []uuid.UUID `json:"failedObjectIds"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.SuccessObjectIDs, payload.Data.Result.FailedObjectIDs, nil
}
