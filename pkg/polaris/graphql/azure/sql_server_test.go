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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
)

const (
	testMIServerID = "11111111-1111-4111-8111-111111111111"
	testMIJobID    = "01900000-0000-7000-8000-000000000001"
)

// decodeVars decodes the GraphQL request variables into vars. It is called from
// the httptest handler goroutine, so failures cancel the context with the cause
// instead of failing the test directly. Returns false when the request could
// not be decoded, in which case the handler must return without responding.
func decodeVars(cancel context.CancelCauseFunc, req *http.Request, vars any) bool {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		cancel(err)
		return false
	}

	var decoded struct {
		Variables json.RawMessage `json:"variables"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		cancel(err)
		return false
	}
	if err := json.Unmarshal(decoded.Variables, vars); err != nil {
		cancel(err)
		return false
	}

	return true
}

// TestSetupCloudNativeSQLServerBackup verifies the variables sent for the SQL
// authentication flow and that a recorded response is unmarshalled correctly.
func TestSetupCloudNativeSQLServerBackup(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/setup_cloud_native_sql_server_backup_response.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	var gotVars struct {
		ServerIDs        []string `json:"serverIds"`
		DatabaseIDs      []string `json:"databaseIds"`
		AdminCredentials struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		} `json:"adminCredentials"`
		AuthMechanism string `json:"authMechanism"`
	}
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		if !decodeVars(cancel, req, &gotVars) {
			return
		}

		if err := tmpl.Execute(w, struct {
			JobID    string
			ServerID string
		}{JobID: testMIJobID, ServerID: testMIServerID}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	serverID := uuid.MustParse(testMIServerID)
	status, err := Wrap(graphql.NewTestClient(srv)).SetupCloudNativeSQLServerBackup(ctx,
		[]uuid.UUID{serverID}, nil, LoginCredentials{Login: "draper", Password: "s3cret"}, SQLAuthentication)
	if err != nil {
		t.Fatalf("SetupCloudNativeSQLServerBackup failed: %v", err)
	}

	if len(gotVars.ServerIDs) != 1 || gotVars.ServerIDs[0] != testMIServerID {
		t.Errorf("serverIds: got %v", gotVars.ServerIDs)
	}
	// RSC expects databaseIds to be present and empty for the instance-level
	// flow, not omitted. A nil slice and an empty slice both format as [], so
	// report the two cases distinctly.
	switch {
	case gotVars.DatabaseIDs == nil:
		t.Error("databaseIds: omitted from the request, want an empty array")
	case len(gotVars.DatabaseIDs) != 0:
		t.Errorf("databaseIds: got %v, want an empty array", gotVars.DatabaseIDs)
	}
	if gotVars.AdminCredentials.Login != "draper" {
		t.Errorf("adminCredentials.login: got %q", gotVars.AdminCredentials.Login)
	}
	// secret.String must still send the real value over the wire, it is only
	// redacted when logged.
	if gotVars.AdminCredentials.Password != "s3cret" {
		t.Errorf("adminCredentials.password was not sent verbatim")
	}
	if gotVars.AuthMechanism != "SQL_AUTHENTICATION" {
		t.Errorf("authMechanism: got %q", gotVars.AuthMechanism)
	}

	if len(status.Errors) != 0 {
		t.Errorf("errors: got %v, want none", status.Errors)
	}
	if len(status.JobIDs) != 1 {
		t.Fatalf("got %d job IDs, want 1", len(status.JobIDs))
	}
	if status.JobIDs[0].JobID != uuid.MustParse(testMIJobID) {
		t.Errorf("jobId: got %q", status.JobIDs[0].JobID)
	}
	if status.JobIDs[0].ObjectID != serverID {
		t.Errorf("rubrikObjectId: got %q", status.JobIDs[0].ObjectID)
	}
}

// TestSetupCloudNativeSQLServerBackupPreValidationError verifies that an object
// which fails pre-validation is returned in Errors with no job started.
func TestSetupCloudNativeSQLServerBackupPreValidationError(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/setup_cloud_native_sql_server_backup_pre_validation_error_response.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		if err := tmpl.Execute(w, struct {
			Error    string
			ServerID string
		}{Error: "invalid credentials", ServerID: testMIServerID}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	status, err := Wrap(graphql.NewTestClient(srv)).SetupCloudNativeSQLServerBackup(ctx,
		[]uuid.UUID{uuid.MustParse(testMIServerID)}, nil, LoginCredentials{Login: "draper", Password: "s3cret"},
		SQLAuthentication)
	if err != nil {
		t.Fatalf("SetupCloudNativeSQLServerBackup failed: %v", err)
	}

	if len(status.JobIDs) != 0 {
		t.Errorf("job IDs: got %v, want none", status.JobIDs)
	}
	if len(status.Errors) != 1 {
		t.Fatalf("got %d errors, want 1", len(status.Errors))
	}
	if status.Errors[0].Error != "invalid credentials" {
		t.Errorf("error: got %q", status.Errors[0].Error)
	}
	if status.Errors[0].ObjectID != uuid.MustParse(testMIServerID) {
		t.Errorf("rubrikObjectId: got %q", status.Errors[0].ObjectID)
	}
}

// TestClearCloudNativeSQLServerBackupCredentials verifies that the workload
// type is sent using its GraphQL enum name and that the reply is split into
// successful and failed object IDs.
func TestClearCloudNativeSQLServerBackupCredentials(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/clear_cloud_native_sql_server_backup_credentials_response.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	const failedID = "22222222-2222-4222-8222-222222222222"

	var gotVars struct {
		ObjectIDs    []string `json:"objectIds"`
		WorkloadType string   `json:"workloadType"`
	}
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		if !decodeVars(cancel, req, &gotVars) {
			return
		}

		if err := tmpl.Execute(w, struct {
			SuccessID string
			FailedID  string
		}{SuccessID: testMIServerID, FailedID: failedID}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	successIDs, failedIDs, err := Wrap(graphql.NewTestClient(srv)).ClearCloudNativeSQLServerBackupCredentials(ctx,
		[]uuid.UUID{uuid.MustParse(testMIServerID)}, hierarchy.WorkloadAzureSQLMIDB)
	if err != nil {
		t.Fatalf("ClearCloudNativeSQLServerBackupCredentials failed: %v", err)
	}

	if len(gotVars.ObjectIDs) != 1 || gotVars.ObjectIDs[0] != testMIServerID {
		t.Errorf("objectIds: got %v", gotVars.ObjectIDs)
	}
	// The GraphQL enum name, not the SDK string value AZURE_SQL_MANAGED_INSTANCE_DB.
	if gotVars.WorkloadType != "AzureSqlManagedInstanceDb" {
		t.Errorf("workloadType: got %q, want %q", gotVars.WorkloadType, "AzureSqlManagedInstanceDb")
	}

	if len(successIDs) != 1 || successIDs[0] != uuid.MustParse(testMIServerID) {
		t.Errorf("successObjectIds: got %v", successIDs)
	}
	if len(failedIDs) != 1 || failedIDs[0] != uuid.MustParse(failedID) {
		t.Errorf("failedObjectIds: got %v", failedIDs)
	}
}

// Credentials for the add credentials mutation tests. The password is typed as
// secret.String wherever it is carried, per the SDK convention for credentials.
const (
	testBackupLogin                  = "backup-user"
	testBackupPassword secret.String = "backup-secret"
)

// addCredentialsVars is the variable payload of the add credentials mutation.
// BackupCredentials is a pointer so the test can tell an omitted credentials
// field apart from one sent with zero values.
type addCredentialsVars struct {
	ObjectIDs         []string `json:"objectIds"`
	WorkloadType      string   `json:"workloadType"`
	ShouldUseAad      bool     `json:"shouldUseAad"`
	BackupCredentials *struct {
		Login    string        `json:"login"`
		Password secret.String `json:"password"`
	} `json:"backupCredentials"`
}

// TestAddCloudNativeSQLServerBackupCredentialsWithSQLLogin verifies that
// passing credentials sends them with shouldUseAad false, which is what a
// server only supporting SQL Server authentication requires.
func TestAddCloudNativeSQLServerBackupCredentialsWithSQLLogin(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/add_cloud_native_sql_server_backup_credentials_response.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	var gotVars addCredentialsVars
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		if !decodeVars(cancel, req, &gotVars) {
			return
		}
		if err := tmpl.Execute(w, struct{ SuccessID string }{SuccessID: testMIServerID}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	successIDs, failedIDs, err := Wrap(graphql.NewTestClient(srv)).AddCloudNativeSQLServerBackupCredentials(ctx,
		[]uuid.UUID{uuid.MustParse(testMIServerID)}, hierarchy.WorkloadAzureSQLMIDB,
		&LoginCredentials{Login: testBackupLogin, Password: testBackupPassword})
	if err != nil {
		t.Fatalf("AddCloudNativeSQLServerBackupCredentials failed: %v", err)
	}

	if len(gotVars.ObjectIDs) != 1 || gotVars.ObjectIDs[0] != testMIServerID {
		t.Errorf("objectIds: got %v", gotVars.ObjectIDs)
	}
	if gotVars.WorkloadType != "AzureSqlManagedInstanceDb" {
		t.Errorf("workloadType: got %q, want %q", gotVars.WorkloadType, "AzureSqlManagedInstanceDb")
	}
	if gotVars.ShouldUseAad {
		t.Error("shouldUseAad: got true, want false when SQL Server credentials are given")
	}
	if gotVars.BackupCredentials == nil {
		t.Fatal("backupCredentials: not sent, want the given credentials")
	}
	if gotVars.BackupCredentials.Login != testBackupLogin {
		t.Errorf("login: got %q", gotVars.BackupCredentials.Login)
	}
	// secret.String is a plain string alias, redacted only where a caller opts
	// in via secret.Redact, so the password must reach RSC verbatim. The value
	// itself is deliberately not printed on failure.
	if gotVars.BackupCredentials.Password != testBackupPassword {
		t.Error("password: got a value other than the credential passed in, want it to reach RSC verbatim")
	}

	if len(successIDs) != 1 || successIDs[0] != uuid.MustParse(testMIServerID) {
		t.Errorf("successObjectIds: got %v", successIDs)
	}
	if len(failedIDs) != 0 {
		t.Errorf("failedObjectIds: got %v, want none", failedIDs)
	}
}

// TestAddCloudNativeSQLServerBackupCredentialsUsingEntraID verifies that
// passing no credentials sends shouldUseAad true and omits the credentials
// field entirely, which is what a server supporting Entra ID requires. Sending
// an empty credentials object instead would be rejected.
func TestAddCloudNativeSQLServerBackupCredentialsUsingEntraID(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/add_cloud_native_sql_server_backup_credentials_response.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	// Decoded into a map rather than a struct, so that an omitted field can be
	// told apart from one sent as an explicit null. A nil pointer would decode
	// identically from both, which would make the assertion below vacuous.
	var gotVars map[string]any
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		if !decodeVars(cancel, req, &gotVars) {
			return
		}
		if err := tmpl.Execute(w, struct{ SuccessID string }{SuccessID: testMIServerID}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	_, _, err = Wrap(graphql.NewTestClient(srv)).AddCloudNativeSQLServerBackupCredentials(ctx,
		[]uuid.UUID{uuid.MustParse(testMIServerID)}, hierarchy.WorkloadAzureSQLMIDB, nil)
	if err != nil {
		t.Fatalf("AddCloudNativeSQLServerBackupCredentials failed: %v", err)
	}

	if shouldUseAad, _ := gotVars["shouldUseAad"].(bool); !shouldUseAad {
		t.Error("shouldUseAad: got false, want true when no credentials are given")
	}
	if credentials, ok := gotVars["backupCredentials"]; ok {
		t.Errorf("backupCredentials: got %v, want the field to be absent, not null", credentials)
	}
}

// TestSQLServerSetupScripts verifies that the per-server script details are
// unmarshalled, including the auth type which decides whether a server can use
// SQL Server credentials.
func TestSQLServerSetupScripts(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/sql_server_setup_scripts_bulk_response.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	const entraIDServerID = "33333333-3333-4333-8333-333333333333"

	var gotVars struct {
		ServerIDs []string `json:"serverIds"`
	}
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		if !decodeVars(cancel, req, &gotVars) {
			return
		}
		if err := tmpl.Execute(w, struct {
			SQLOnlyServerID string
			EntraIDServerID string
		}{SQLOnlyServerID: testMIServerID, EntraIDServerID: entraIDServerID}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	scripts, err := Wrap(graphql.NewTestClient(srv)).SQLServerSetupScripts(ctx,
		[]uuid.UUID{uuid.MustParse(testMIServerID), uuid.MustParse(entraIDServerID)})
	if err != nil {
		t.Fatalf("SQLServerSetupScripts failed: %v", err)
	}

	if len(gotVars.ServerIDs) != 2 {
		t.Errorf("serverIds: got %v, want both servers", gotVars.ServerIDs)
	}

	if len(scripts) != 2 {
		t.Fatalf("got %d scripts, want 2", len(scripts))
	}
	if scripts[0].ServerID != uuid.MustParse(testMIServerID) {
		t.Errorf("serverId: got %q", scripts[0].ServerID)
	}
	if scripts[0].AuthType != AzureSQLAuthTypeSQLOnly {
		t.Errorf("authType: got %q, want %q", scripts[0].AuthType, AzureSQLAuthTypeSQLOnly)
	}
	if scripts[0].Script == "" {
		t.Error("script: got empty, want the setup script body")
	}
	if scripts[1].AuthType != AzureSQLAuthTypeSQLAndEntraID {
		t.Errorf("authType: got %q, want %q", scripts[1].AuthType, AzureSQLAuthTypeSQLAndEntraID)
	}
}
