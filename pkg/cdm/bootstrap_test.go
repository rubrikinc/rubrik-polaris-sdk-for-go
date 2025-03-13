// Copyright 2025 Rubrik, Inc.
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

package cdm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
)

// TestIsBootstrapped verifies that the IsBootstrapped function parses and
// returns the correct value when the cluster returns a valid response.
func TestIsBootstrapped(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := json.NewEncoder(w).Encode(struct {
			Value bool `json:"value"`
		}{Value: true}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	ok, err := WrapBootstrap(testClient(srv)).IsBootstrapped(ctx, 1*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("invalid bootstrap status: false")
	}
}

// TestIsBootstrappedContextTimeout verifies that the IsBootstrapped function
// cancels its request and returns an error when the context is canceled.
func TestIsBootstrappedContextTimeout(t *testing.T) {
	ctx := context.Background()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer srv.Close()

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer timeoutCancel()

	_, err := WrapBootstrap(testClient(srv)).IsBootstrapped(timeoutCtx, 1*time.Second, 100*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("invalid error type: %T", err)
	}
}

// TestIsBootstrappedErrorInvalid verifies that the IsBootstrapped function
// returns an error when the cluster returns an invalid response.
func TestIsBootstrappedErrorInvalid(t *testing.T) {
	ctx := context.Background()

	// Return something which isn't valid JSON.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
	defer srv.Close()

	_, err := WrapBootstrap(testClient(srv)).IsBootstrapped(ctx, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "failed to unmarshal bootstrap status:") {
		t.Fatalf("invalid error: %v", err)
	}
}

// TestIsBootstrappedErrorRecover verifies that the IsBootstrapped function is
// able to recover from errors within the timeout duration.
func TestIsBootstrappedErrorRecover(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	count := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if count == 5 {
			if err := json.NewEncoder(w).Encode(struct {
				Value bool `json:"value"`
			}{Value: false}); err != nil {
				cancel(err)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		count++
	}))
	defer srv.Close()

	ok, err := WrapBootstrap(testClient(srv)).IsBootstrapped(ctx, 1*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("invalid bootstrap status: true")
	}
}

// TestIsBootstrappedErrorTimeout verifies that the IsBootstrapped function
// returns an error if the error is not recovered within the timeout duration.
func TestIsBootstrappedErrorTimeout(t *testing.T) {
	ctx := context.Background()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := WrapBootstrap(testClient(srv)).IsBootstrapped(ctx, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "timeout waiting for bootstrap status:") {
		t.Fatalf("invalid error: %v", err)
	}
}

// TestBootstrap verifies that the BootstrapCluster function returns the values
// returned by the cluster in response to a bootstrap request.
func TestBootstrap(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/internal/node_management/is_bootstrapped":
			if err := json.NewEncoder(w).Encode(struct {
				Value bool `json:"value"`
			}{Value: false}); err != nil {
				cancel(err)
			}
		case "/api/internal/cluster/me/bootstrap":
			w.WriteHeader(http.StatusAccepted)
			if err := json.NewEncoder(w).Encode(struct {
				ID int `json:"id"`
			}{ID: 1}); err != nil {
				cancel(err)
			}
		default:
			cancel(errors.New("invalid url path"))
		}
	}))
	defer srv.Close()

	id, err := WrapBootstrap(testClient(srv)).BootstrapCluster(ctx, ClusterConfig{}, 1*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Fatalf("invalid cluster id: %d", id)
	}
}

// TestBootstrapErrorIsBootstrapped verifies that the BootstrapCluster function
// returns the correct error when the IsBootstrapped function fails.
func TestBootstrapErrorIsBootstrapped(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/internal/node_management/is_bootstrapped":
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		default:
			cancel(errors.New("invalid url path"))
		}
	}))
	defer srv.Close()

	_, err := WrapBootstrap(testClient(srv)).BootstrapCluster(ctx, ClusterConfig{}, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "failed to check cluster bootstrap status:") {
		t.Fatalf("invalid error: %v", err)
	}
}

// TestBootstrapErrorUnprocessableEntity verifies that the BootstrapCluster
// function returns the correct error when the bootstrap request returns an
// unprocessable entity error.
func TestBootstrapErrorUnprocessableEntity(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/internal/node_management/is_bootstrapped":
			if err := json.NewEncoder(w).Encode(struct {
				Value bool `json:"value"`
			}{Value: false}); err != nil {
				cancel(err)
			}
		case "/api/internal/cluster/me/bootstrap":
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		default:
			cancel(errors.New("invalid url path"))
		}
	}))
	defer srv.Close()

	_, err := WrapBootstrap(testClient(srv)).BootstrapCluster(ctx, ClusterConfig{}, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "failed POST request") {
		t.Fatalf("invalid error: %v", err)
	}
}

// TestBootstrapErrorUnprocessableEntity verifies that the BootstrapCluster
// function returns the correct error when the bootstrap request returns an
// unprocessable entity error with a JSON body.
func TestBootstrapErrorUnprocessableEntityWithStatus(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/internal/node_management/is_bootstrapped":
			if err := json.NewEncoder(w).Encode(struct {
				Value bool `json:"value"`
			}{Value: false}); err != nil {
				cancel(err)
			}
		case "/api/internal/cluster/me/bootstrap":
			w.WriteHeader(http.StatusUnprocessableEntity)
			if err := json.NewEncoder(w).Encode(struct {
				ID     int    `json:"id"`
				Status string `json:"status"`
			}{ID: 0, Status: "invalid configuration"}); err != nil {
				cancel(err)
			}
		default:
			cancel(errors.New("invalid url path"))
		}
	}))
	defer srv.Close()

	_, err := WrapBootstrap(testClient(srv)).BootstrapCluster(ctx, ClusterConfig{}, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "failed POST request") && !strings.HasSuffix(err.Error(), "invalid configuration") {
		t.Fatalf("invalid error: %v", err)
	}
}

// TestBootstrapWithBootstrappedCluster verifies that the BootstrapCluster
// function returns an error when the cluster is already bootstrapped.
func TestBootstrapWithBootstrappedCluster(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/internal/node_management/is_bootstrapped":
			if err := json.NewEncoder(w).Encode(struct {
				Value bool `json:"value"`
			}{Value: true}); err != nil {
				cancel(err)
			}
		default:
			cancel(errors.New("invalid url path"))
		}
	}))
	defer srv.Close()

	_, err := WrapBootstrap(testClient(srv)).BootstrapCluster(ctx, ClusterConfig{}, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "cluster is already bootstrapped") {
		t.Fatalf("invalid error: %v", err)
	}
}

// TestWaitForBootstrap verifies that the WaitForBootstrap function parses and
// returns the correct value when the cluster returns a valid response to the
// bootstrap GET request.
func TestWaitForBootstrap(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/internal/cluster/me/bootstrap":
			if err := json.NewEncoder(w).Encode(struct {
				Status  string `json:"status"`
				Message string `json:"message"`
			}{Status: "", Message: ""}); err != nil {
				cancel(err)
			}
		default:
			cancel(errors.New("invalid url path"))
		}
	}))
	defer srv.Close()

	err := WrapBootstrap(testClient(srv)).WaitForBootstrap(ctx, 1, 1*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
}

// TestWaitForBootstrapContextTimeout verifies that the WaitForBootstrap
// function cancels its request and returns an error when the context is
// canceled.
func TestWaitForBootstrapContextTimeout(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/internal/cluster/me/bootstrap":
			if err := json.NewEncoder(w).Encode(struct {
				Status  string `json:"status"`
				Message string `json:"message"`
			}{Status: "IN_PROGRESS", Message: "bootstrap in progress"}); err != nil {
				cancel(err)
			}
		default:
			cancel(errors.New("invalid url path"))
		}
	}))
	defer srv.Close()

	ctx, ctxCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer ctxCancel()

	err := WrapBootstrap(testClient(srv)).WaitForBootstrap(ctx, 1, 1*time.Second, 100*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("invalid error type: %T", err)
	}
}

// TestWaitForBootstrapErrorRecover verifies that the WaitForBootstrap function
// is able to recover from errors within the timeout duration.
func TestWaitForBootstrapErrorRecover(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	count := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/internal/cluster/me/bootstrap":
			if count == 5 {
				if err := json.NewEncoder(w).Encode(struct {
					Status  string `json:"status"`
					Message string `json:"message"`
				}{Status: "", Message: ""}); err != nil {
					cancel(err)
				}
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			count++
		default:
			cancel(errors.New("invalid url path"))
		}
	}))
	defer srv.Close()

	err := WrapBootstrap(testClient(srv)).WaitForBootstrap(ctx, 1, 1*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
}

// TestWaitForBootstrapErrorTimeout verifies that the WaitForBootstrap function
// returns an error if the error is not recovered within the timeout duration.
func TestWaitForBootstrapErrorTimeout(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/internal/cluster/me/bootstrap":
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			cancel(errors.New("invalid url path"))
		}
	}))
	defer srv.Close()

	err := WrapBootstrap(testClient(srv)).WaitForBootstrap(ctx, 1, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "timeout waiting for bootstrap:") {
		t.Fatalf("invalid error: %v", err)
	}
}
