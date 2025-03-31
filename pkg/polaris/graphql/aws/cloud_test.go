// Copyright 2021 Rubrik, Inc.
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

package aws

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

func TestValidateAndCreateAWSCloudAccountWithDuplicate(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/validate_and_create_aws_cloud_account_response.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	// Respond with an error indicating that the account has already been added.
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}

		var payload struct {
			Query     string `json:"query"`
			Variables struct {
				ID   string `json:"nativeId"`
				Name string `json:"accountName"`
			} `json:"variables"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}

		if err := tmpl.Execute(w, struct {
			ID   string
			Name string
		}{ID: payload.Variables.ID, Name: payload.Variables.Name}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	_, err = Wrap(graphql.NewTestClient(srv)).ValidateAndCreateCloudAccount(ctx,
		"123456789012", "123456789012 : default", []core.Feature{core.FeatureCloudNativeProtection})
	if err == nil {
		t.Fatal("expected ValidateAndCreateCloudAccount to fail")
	}
	if msg := err.Error(); !strings.HasPrefix(msg, "invalid account: You do not need to add 123456789012") {
		t.Errorf("invalid error: %v", err)
	}
}
