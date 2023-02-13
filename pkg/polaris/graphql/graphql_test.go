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

package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testnet"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func TestErrorsWithNoError(t *testing.T) {
	buf := []byte(`{
	    "data": {
	        "awsNativeAccountConnection": {}
        }
	}`)

	var gqlErr GQLError
	if err := json.Unmarshal(buf, &gqlErr); err != nil {
		t.Fatal(err)
	}
	if gqlErr.isError() {
		t.Error("gqlErr should not represent an error")
	}
}

func TestGqlError(t *testing.T) {
	buf, err := os.ReadFile("testdata/error_graphql.json")
	if err != nil {
		t.Fatal(err)
	}

	var gqlErr GQLError
	if err := json.Unmarshal(buf, &gqlErr); err != nil {
		t.Fatal(err)
	}
	if !gqlErr.isError() {
		t.Error("gqlErr should represent an error")
	}
	expected := "INTERNAL: invalid status transition of feature CLOUDACCOUNTS from CONNECTED to CONNECTING"
	if msg := gqlErr.Error(); msg != expected {
		t.Fatalf("invalid error message: %v", msg)
	}
}

func TestRequestUnauthenticated(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/error_json_from_auth.json")
	if err != nil {
		t.Fatal(err)
	}

	client, lis := NewTestClient("john", "doe", log.DiscardLogger{})

	// Respond with status code 401 and additional details in the body.
	srv := testnet.ServeJSONWithStaticToken(lis, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(401)
		if err := tmpl.Execute(w, nil); err != nil {
			panic(err)
		}
	})
	defer srv.Shutdown(context.Background())

	_, err = client.Request(context.Background(), "me { name }", nil)
	if err == nil {
		t.Fatal("graphql request should fail")
	}
	if !strings.HasSuffix(err.Error(), "JWT validation failed: Missing or invalid credentials (code 16)") {
		t.Fatal(err)
	}
}

func TestRequestWithInternalServerErrorJSONBody(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/error_graphql.json")
	if err != nil {
		t.Fatal(err)
	}

	client, lis := NewTestClient("john", "doe", log.DiscardLogger{})

	// Respond with status code 500 and additional details in the JSON body.
	srv := testnet.ServeJSONWithStaticToken(lis, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(500)
		if err := tmpl.Execute(w, nil); err != nil {
			panic(err)
		}
	})
	defer srv.Shutdown(context.Background())

	_, err = client.Request(context.Background(), "me { name }", nil)
	if err == nil {
		t.Fatal("graphql request should fail")
	}
	if !strings.HasSuffix(err.Error(),
		"INTERNAL: invalid status transition of feature CLOUDACCOUNTS from CONNECTED to CONNECTING") {
		t.Fatal(err)
	}
}

func TestRequestWithInternalServerErrorNoBody(t *testing.T) {
	client, lis := NewTestClient("john", "doe", log.DiscardLogger{})

	// Respond with status code 500 and no additional details.
	srv := testnet.ServeWithStaticToken(lis, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(500)
	})
	defer srv.Shutdown(context.Background())

	_, err := client.Request(context.Background(), "me { name }", nil)
	if err == nil {
		t.Fatal("graphql request should fail")
	}
	if !strings.HasSuffix(err.Error(), "graphql response has no body (status code 500)") {
		t.Fatal(err)
	}
}

func TestRequestWithInternalServerErrorTextBody(t *testing.T) {
	client, lis := NewTestClient("john", "doe", log.DiscardLogger{})

	// Respond with status code 500 and additional details in the text body.
	srv := testnet.ServeWithStaticToken(lis, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(500)
		w.Write([]byte("database is corrupt"))
	})
	defer srv.Shutdown(context.Background())

	_, err := client.Request(context.Background(), "me { name }", nil)
	if err == nil {
		t.Fatal("graphql request should fail")
	}
	if !strings.HasSuffix(err.Error(),
		"graphql response has Content-Type text/plain (status code 500): \"database is corrupt\"") {
		t.Fatal(err)
	}
}

func TestExtractOperationName(t *testing.T) {
	tt := []struct {
		query      string
		expectedOp string
	}{
		// empty / invalid
		{"", ""},
		{"{ foo }", ""},
		{"{ foo() }", ""},
		{"{ foo(bar) }", ""},
		{"{ (", ""},
		{"{ query(", ""},
		{" query ({", ""},
		{" query {", ""},

		// valid
		{"mutation RubrikPolarisSDKRequest($input: String!) { foo(input: $input){} }", "RubrikPolarisSDKRequest"},
		{" query Foo { foo(input: $input){} }", "Foo"},
		{"query Bar () { foo(input: $input){} }", "Bar"},
		{"query	Buz	{ foo(input: $input){} }", "Buz"}, // with tabs
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.query, func(t *testing.T) {
			if op := operationName(tc.query); op != tc.expectedOp {
				t.Fatalf("%q returned %q, expected %q", tc.query, op, tc.expectedOp)
			}
		})
	}
}
