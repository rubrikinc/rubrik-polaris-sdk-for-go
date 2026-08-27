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

package token

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"
	"time"

	internalerrors "github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/errors"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func TestServiceAccountSource(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var payload struct {
			GrantType    string `json:"grant_type"`
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Error(err)
			return
		}

		if payload.GrantType != "client_credentials" || payload.ClientID != "client-id" || payload.ClientSecret != "client-secret" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(struct {
			ClientID    string `json:"client_id"`
			AccessToken string `json:"access_token"`
		}{ClientID: payload.ClientID, AccessToken: dummyValidToken}); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	tok, err := serviceAccountTestSource(srv).token(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if tok.expired() {
		t.Fatal("invalid token, already expired")
	}

	_, err = serviceAccountTestSourceWithWrongSecret(srv).token(t.Context())
	if err == nil || !strings.HasSuffix(err.Error(), "token response has Content-Type text/plain; charset=utf-8 (status code 403): \"Forbidden\"") {
		t.Fatalf("invalid error: %s", err)
	}
}

func TestServiceAccountSourceWithBadCredentials(t *testing.T) {
	srv := httptest.NewServer(serviceAccountBadCredentialsHandler(t))
	defer srv.Close()

	_, err := serviceAccountTestSource(srv).token(t.Context())
	var jsonErr internalerrors.JSONError
	if err == nil || !errors.As(err, &jsonErr) {
		t.Fatalf("expected token request to fail with a JSONError: %s", err)
	}
	if !strings.HasSuffix(err.Error(), "UNAUTHENTICATED: incorrect client secret (code: 401, traceId: 8nZ8530f7NslEHow1KIBgQ==)") {
		t.Fatal(err)
	}
}

func TestServiceAccountSourceWithInternalServerErrorNoBody(t *testing.T) {
	srv := httptest.NewServer(internalServerErrorWithNoBodyHandler())
	defer srv.Close()

	_, err := serviceAccountTestSource(srv).token(t.Context())
	if err == nil {
		t.Fatal("expected token request to fail")
	}
	if !strings.HasSuffix(err.Error(), "token response has no body (status code 500)") {
		t.Fatal(err)
	}
}

func TestServiceAccountSourceWithInternalServerErrorTextBody(t *testing.T) {
	srv := httptest.NewServer(internalServerErrorWithTextBodyHandler())
	defer srv.Close()

	_, err := serviceAccountTestSource(srv).token(t.Context())
	if err == nil {
		t.Fatal("expected token request to fail")
	}
	if !strings.HasSuffix(err.Error(), " token response has Content-Type text/plain; charset=utf-8 (status code 500): \"Internal Server Error\"") {
		t.Fatal(err)
	}
}

func TestServiceAccountSourceWithCanceledContext(t *testing.T) {
	srv := httptest.NewServer(validTokenHandler(t))
	defer srv.Close()

	// Canceled context.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	account := serviceAccountTestSource(srv)
	_, err := account.token(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestServiceAccountSourceWithDeadlineExceededContext(t *testing.T) {
	srv := httptest.NewServer(validTokenHandler(t))
	defer srv.Close()

	// Deadline exceeded context.
	ctx, cancel := context.WithTimeout(t.Context(), 0*time.Second)
	defer cancel()

	account := serviceAccountTestSource(srv)
	_, err := account.token(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func serviceAccountTestSource(srv *httptest.Server) *ServiceAccountSource {
	return &ServiceAccountSource{
		log:          &log.DiscardLogger{},
		client:       srv.Client(),
		tokenURL:     srv.URL + "/api/client_token",
		clientID:     "client-id",
		clientSecret: "client-secret",
	}
}

func serviceAccountTestSourceWithWrongSecret(srv *httptest.Server) *ServiceAccountSource {
	return &ServiceAccountSource{
		log:          &log.DiscardLogger{},
		client:       srv.Client(),
		tokenURL:     srv.URL + "/api/client_token",
		clientID:     "client-id",
		clientSecret: "wrong-client-secret",
	}
}

func serviceAccountBadCredentialsHandler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("testdata/service_account_auth_error_response.json")
		if err != nil {
			t.Error(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := tmpl.Execute(w, nil); err != nil {
			t.Error(err)
		}
	})
}
