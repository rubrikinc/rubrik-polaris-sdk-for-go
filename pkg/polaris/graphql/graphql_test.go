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
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Serve serves the handler function over HTTP by accepting incoming connections
// on the specified listener.
func serve(lis net.Listener, handler http.HandlerFunc) *http.Server {
	server := &http.Server{Handler: handler}
	go server.Serve(lis)
	return server
}

// serveJSON serves the handler function using HTTP by accepting incoming
// connections on the specified listener. The response content-type is set to
// application/json.
func serveJSON(lis net.Listener, handler http.HandlerFunc) *http.Server {
	return serve(lis, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler(w, req)
	})
}

// serveWithToken serves the handler function and tokens using HTTP by accepting
// incoming connections on specified listener. The response content-type is set
// to application/json.
func serveJSONWithToken(lis net.Listener, handler http.HandlerFunc) *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/api/session", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"access_token": "fake:test:token",
			"is_eula_accepted": true
		}`))
	})
	mux.HandleFunc("/api/graphql", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler(w, req)
	})

	server := &http.Server{Handler: mux}
	go server.Serve(lis)
	return server
}

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
		t.Error("expected jsonErr to not represent an error")
	}

	var gqlErr gqlError
	if err := json.Unmarshal(buf, &gqlErr); err != nil {
		t.Fatal(err)
	}
	if gqlErr.isError() {
		t.Error("expected gqlErr to not represent an error")
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
		t.Error("expected jsonErr to represent an error")
	}
	expected := "polaris: code 16: JWT validation failed: Missing or invalid credentials"
	if msg := jsonErr1.Error(); msg != expected {
		t.Errorf("invalid error message returned by Error, msg=%v", msg)
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
		t.Error("expected jsonErr to represent an error")
	}
	expected = "polaris: code 401: UNAUTHENTICATED: wrong username or password"
	if msg := jsonErr2.Error(); msg != expected {
		t.Errorf("invalid error message returned by Error, msg=%v", msg)
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
		t.Error("expected gqlErr to represent an error")
	}
	expected := "polaris: INTERNAL: invalid status transition of feature CLOUDACCOUNTS from CONNECTED to CONNECTING"
	if msg := gqlErr.Error(); msg != expected {
		t.Fatalf("invalid error message returned by Error, msg=%v", msg)
	}
}

func TestRequestUnauthenticated(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/error_json_from_auth.json")
	if err != nil {
		t.Fatal(err)
	}

	client, lis := NewTestClient("john", "doe", log.DiscardLogger{})

	// Respond with status code 401 and additional details in the body.
	srv := serveJSONWithToken(lis, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(401)
		if err := tmpl.Execute(w, nil); err != nil {
			panic(err)
		}
	})
	defer srv.Shutdown(context.Background())

	_, err = client.Request(context.Background(), "me { name }", nil)
	if err == nil {
		t.Fatal("expected the graphql request to fail")
	}
	if !strings.HasPrefix(err.Error(), "polaris: code 16: JWT validation failed") {
		t.Fatal(err)
	}
}

func TestRequestWithInternalServerError(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/error_graphql.json")
	if err != nil {
		t.Fatal(err)
	}

	client, lis := NewTestClient("john", "doe", log.DiscardLogger{})

	// Respond with status code 500 and additional details in the body.
	srv := serveJSONWithToken(lis, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(500)
		if err := tmpl.Execute(w, nil); err != nil {
			panic(err)
		}
	})
	defer srv.Shutdown(context.Background())

	_, err = client.Request(context.Background(), "me { name }", nil)
	if err == nil {
		t.Fatal("expected the graphql request to fail")
	}
	if !strings.HasPrefix(err.Error(), "polaris: INTERNAL: invalid status transition") {
		t.Fatal(err)
	}
}

func TestRequestWithInternalServerErrorNoBody(t *testing.T) {
	client, lis := NewTestClient("john", "doe", log.DiscardLogger{})

	// Respond with status code 500 and no additional details.
	srv := serve(lis, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(500)
	})
	defer srv.Shutdown(context.Background())

	_, err := client.Request(context.Background(), "me { name }", nil)
	if err == nil {
		t.Fatal("expected the graphql request to fail")
	}
	if !strings.HasSuffix(err.Error(), "polaris: 500 Internal Server Error") {
		t.Fatal(err)
	}
}

func TestKorgTaskChainStatus(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/korgtaskchainstatus.json")
	if err != nil {
		t.Fatal(err)
	}

	client, lis := NewTestClient("john", "doe", log.DiscardLogger{})

	// Respond with status code 200 and a valid body.
	srv := serveJSONWithToken(lis, func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var payload struct {
			Query     string `json:"query"`
			Variables struct {
				TaskChainID string `json:"taskchainId,omitempty"`
			} `json:"variables,omitempty"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			t.Fatal(err)
		}

		chain := struct {
			ChainState string
			ChainUUID  string
		}{ChainState: "RUNNING", ChainUUID: payload.Variables.TaskChainID}
		if err := tmpl.Execute(w, chain); err != nil {
			panic(err)
		}
	})
	defer srv.Shutdown(context.Background())

	chainUUID := TaskChainUUID("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	taskChain, err := client.KorgTaskChainStatus(context.Background(), chainUUID)
	if err != nil {
		t.Fatal(err)
	}
	if id := taskChain.ID; id != 12761540 {
		t.Errorf("invalid task chain id, id=%v", id)
	}
	if uuid := taskChain.TaskChainUUID; uuid != chainUUID {
		t.Errorf("invalid task chain uuid, uuid=%v", uuid)
	}
	if state := taskChain.State; state != TaskChainRunning {
		t.Errorf("invalid task chain state, state=%v", state)
	}
}
