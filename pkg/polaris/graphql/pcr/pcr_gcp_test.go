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

package pcr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
)

func TestSetGCPRegistry(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	accountID := uuid.New()
	srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Variables map[string]json.RawMessage `json:"variables"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			cancel(err)
			return
		}

		want := map[string]string{
			"cloudType":           `"GCP"`,
			"exocomputeAccountId": `"` + accountID.String() + `"`,
			"registryUrl":         `"us-docker.pkg.dev/my-project/my-repo"`,
			"pcrGcpImagePullDetails": `{"serviceAccountEmail":"pull@my-project.iam.gserviceaccount.com",` +
				`"serviceAccountOwnershipToken":"id-token"}`,
		}
		for name, value := range want {
			if got := string(body.Variables[name]); got != value {
				t.Errorf("invalid variable %s: %s", name, got)
			}
		}
		for _, name := range []string{"pcrAwsImagePullDetails", "pcrAzureImagePullDetails"} {
			if _, ok := body.Variables[name]; ok {
				t.Errorf("unexpected variable %s", name)
			}
		}

		if _, err := w.Write([]byte(`{"data": {"result": null}}`)); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	err := SetRegistry(ctx, graphql.NewTestClient(srv), SetGCPRegistryParams{
		CloudAccountID:               accountID,
		RegistryURL:                  "us-docker.pkg.dev/my-project/my-repo",
		ServiceAccountEmail:          "pull@my-project.iam.gserviceaccount.com",
		ServiceAccountOwnershipToken: secret.String("id-token"),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetGCPRegistry(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	tmpl, err := template.ParseFiles("testdata/get_gcp_private_container_registry_response.tmpl")
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		if err := tmpl.Execute(w, struct {
			BundleVersion       string
			RegistryURL         string
			ServiceAccountEmail string
		}{
			BundleVersion:       "2026-10-01",
			RegistryURL:         "us-docker.pkg.dev/my-project/my-repo",
			ServiceAccountEmail: "pull@my-project.iam.gserviceaccount.com",
		}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	registry, err := GetRegistry(ctx, graphql.NewTestClient(srv), GetGCPRegistryParams{CloudAccountID: uuid.New()})
	if err != nil {
		t.Fatal(err)
	}
	if registry.LatestApprovedBundleVersion != "2026-10-01" {
		t.Errorf("invalid bundle version: %s", registry.LatestApprovedBundleVersion)
	}
	if url := registry.PCRDetails.RegistryURL; url != "us-docker.pkg.dev/my-project/my-repo" {
		t.Errorf("invalid registry url: %s", url)
	}
	if email := registry.PCRDetails.ImagePullDetails.ServiceAccountEmail; email != "pull@my-project.iam.gserviceaccount.com" {
		t.Errorf("invalid service account email: %s", email)
	}
}
