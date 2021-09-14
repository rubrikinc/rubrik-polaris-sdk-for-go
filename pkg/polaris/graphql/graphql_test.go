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

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func TestErrorsWithNoError(t *testing.T) {
	buf := []byte(`{
	    "data": {
	        "awsNativeAccountConnection": {}
        }
	}`)

	var jsonErr jsonError
	if err := json.Unmarshal(buf, &jsonErr); err != nil {
		t.Fatal(err)
	}
	if jsonErr.isError() {
		t.Error("jsonErr should not represent an error")
	}

	var gqlErr gqlError
	if err := json.Unmarshal(buf, &gqlErr); err != nil {
		t.Fatal(err)
	}
	if gqlErr.isError() {
		t.Error("gqlErr should not represent an error")
	}
}

func TestJsonError(t *testing.T) {
	buf, err := os.ReadFile("testdata/error_json_from_auth.json")
	if err != nil {
		t.Fatal(err)
	}

	var jsonErr1 jsonError
	if err := json.Unmarshal(buf, &jsonErr1); err != nil {
		t.Fatal(err)
	}
	if !jsonErr1.isError() {
		t.Error("jsonErr should represent an error")
	}
	expected := "polaris: code 16: JWT validation failed: Missing or invalid credentials"
	if msg := jsonErr1.Error(); msg != expected {
		t.Errorf("invalid error message: %v", msg)
	}

	buf, err = os.ReadFile("testdata/error_json_from_polaris.json")
	if err != nil {
		t.Fatal(err)
	}

	var jsonErr2 jsonError
	if err := json.Unmarshal(buf, &jsonErr2); err != nil {
		t.Fatal(err)
	}
	if !jsonErr2.isError() {
		t.Error("jsonErr should represent an error")
	}
	expected = "polaris: code 401: UNAUTHENTICATED: wrong username or password"
	if msg := jsonErr2.Error(); msg != expected {
		t.Errorf("invalid error message: %v", msg)
	}
}

func TestGqlError(t *testing.T) {
	buf, err := os.ReadFile("testdata/error_graphql.json")
	if err != nil {
		t.Fatal(err)
	}

	var gqlErr gqlError
	if err := json.Unmarshal(buf, &gqlErr); err != nil {
		t.Fatal(err)
	}
	if !gqlErr.isError() {
		t.Error("gqlErr should represent an error")
	}
	expected := "polaris: INTERNAL: invalid status transition of feature CLOUDACCOUNTS from CONNECTED to CONNECTING"
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
	srv := TestServeJSONWithToken(lis, func(w http.ResponseWriter, req *http.Request) {
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
	if !strings.HasPrefix(err.Error(), "polaris: code 16: JWT validation failed") {
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
	srv := TestServeJSONWithToken(lis, func(w http.ResponseWriter, req *http.Request) {
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
	if !strings.HasPrefix(err.Error(), "polaris: INTERNAL: invalid status transition") {
		t.Fatal(err)
	}
}

func TestRequestWithInternalServerErrorNoBody(t *testing.T) {
	client, lis := NewTestClient("john", "doe", log.DiscardLogger{})

	// Respond with status code 500 and no additional details.
	srv := TestServeWithToken(lis, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(500)
	})
	defer srv.Shutdown(context.Background())

	_, err := client.Request(context.Background(), "me { name }", nil)
	if err == nil {
		t.Fatal("graphql request should fail")
	}
	if !strings.HasPrefix(err.Error(), "polaris: 500 Internal Server Error") {
		t.Fatal(err)
	}
}

func TestRequestWithInternalServerErrorTextBody(t *testing.T) {
	client, lis := NewTestClient("john", "doe", log.DiscardLogger{})

	// Respond with status code 500 and additional details in the text body.
	srv := TestServeWithToken(lis, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(500)
		w.Write([]byte("database is corrupt"))
	})
	defer srv.Shutdown(context.Background())

	_, err := client.Request(context.Background(), "me { name }", nil)
	if err == nil {
		t.Fatal("graphql request should fail")
	}
	if !strings.HasPrefix(err.Error(), "polaris: 500 Internal Server Error") {
		t.Fatal(err)
	}
}
