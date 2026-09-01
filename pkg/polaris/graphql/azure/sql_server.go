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

// AddCloudNativeSQLServerBackupCredentials registers the credentials RSC uses
// to back up the specified objects, after a setup script has been run against
// them. The object IDs can refer to any level of the hierarchy the credentials
// apply to, e.g. subscriptions, resource groups, servers or databases. The
// workload type must match the object type, e.g. hierarchy.WorkloadAzureSQLMIDB
// for Azure SQL Managed Instances.
//
// Pass credentials to authenticate as a SQL Server user, which requires that
// the setup script created that user. Pass nil to authenticate using Microsoft
// Entra ID instead, which is only possible for a server supporting Entra ID
// authentication, see SQLServerSetupScripts for how to determine that.
//
// Unlike SetupCloudNativeSQLServerBackup this is not asynchronous: the
// credentials are recorded by the request itself, with no task chain to wait
// for. Returns the IDs of the objects the credentials were added for and the
// IDs of the objects they could not be added for.
func (a API) AddCloudNativeSQLServerBackupCredentials(ctx context.Context, objectIDs []uuid.UUID, workloadType hierarchy.Workload, credentials *LoginCredentials) (successIDs, failedIDs []uuid.UUID, err error) {
	a.log.Print(log.Trace)

	query := addCloudNativeSqlServerBackupCredentialsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ObjectIDs         []uuid.UUID        `json:"objectIds"`
		WorkloadType      hierarchy.Workload `json:"workloadType"`
		BackupCredentials *LoginCredentials  `json:"backupCredentials,omitempty"`
		ShouldUseAad      bool               `json:"shouldUseAad"`
	}{
		ObjectIDs:         objectIDs,
		WorkloadType:      workloadType,
		BackupCredentials: credentials,
		ShouldUseAad:      credentials == nil,
	})
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

// AzureSQLAuthenticationType represents the authentication mechanisms a SQL
// Server supports.
type AzureSQLAuthenticationType string

const (
	// AzureSQLAuthTypeUnspecified means the authentication type is unknown.
	AzureSQLAuthTypeUnspecified AzureSQLAuthenticationType = "AUTH_TYPE_UNSPECIFIED"

	// AzureSQLAuthTypeEntraIDOnly means the server only supports Microsoft
	// Entra ID authentication. RSC spells Entra ID by its former name, AAD.
	AzureSQLAuthTypeEntraIDOnly AzureSQLAuthenticationType = "AAD_ONLY"

	// AzureSQLAuthTypeSQLAndEntraID means the server supports both SQL Server
	// and Microsoft Entra ID authentication.
	AzureSQLAuthTypeSQLAndEntraID AzureSQLAuthenticationType = "SQL_AUTH_AND_AAD"

	// AzureSQLAuthTypeSQLOnly means the server only supports SQL Server
	// authentication.
	AzureSQLAuthTypeSQLOnly AzureSQLAuthenticationType = "SQL_AUTH_ONLY"
)

// SQLServerSetupScript is the setup script for a single SQL Server, and the
// authentication mechanisms that server supports.
type SQLServerSetupScript struct {
	// ServerID is the RSC ID of the server the script is for.
	ServerID uuid.UUID `json:"serverId"`
	// AuthType is the authentication mechanisms the server supports. It
	// determines whether backup credentials can be registered as a SQL Server
	// user or must use Microsoft Entra ID.
	AuthType AzureSQLAuthenticationType `json:"authType"`
	// Script is the T-SQL setup script to run against the server. It creates
	// the objects RSC needs to perform backups. The script is a template: the
	// user RSC authenticates as is created by a procedure call at the end of
	// the script, which must be supplied with the intended login and password
	// before the script is run.
	Script string `json:"script"`
}

// SQLServerSetupScripts returns the setup script for each of the specified SQL
// Servers, to be run against the server before its backup credentials are
// registered with AddCloudNativeSQLServerBackupCredentials.
//
// The scripts are generated by RSC and differ per server, both by server and by
// the authentication mechanisms it supports. They contain no secret: the only
// password in a script is generated by the script itself when it runs.
func (a API) SQLServerSetupScripts(ctx context.Context, serverIDs []uuid.UUID) ([]SQLServerSetupScript, error) {
	a.log.Print(log.Trace)

	query := sqlServerSetupScriptsBulkQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ServerIDs []uuid.UUID `json:"serverIds"`
	}{ServerIDs: serverIDs})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				ScriptDetails []SQLServerSetupScript `json:"scriptDetails"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.ScriptDetails, nil
}
