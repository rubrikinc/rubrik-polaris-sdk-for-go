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

func TestUserSourceToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var payload struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Error(err)
			return
		}

		if payload.Username != "username" || payload.Password != "password" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(struct {
			AccessToken string `json:"access_token"`
		}{AccessToken: dummyValidToken}); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	tok, err := userTestSource(srv).token(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if tok.expired() {
		t.Fatal("invalid token, already expired")
	}

	_, err = userTestSourceWithWrongPassword(srv).token(t.Context())
	if err == nil || !strings.HasSuffix(err.Error(), "token response has Content-Type text/plain; charset=utf-8 (status code 403): \"Forbidden\"") {
		t.Fatalf("invalid error: %s", err)
	}
}

func TestUserSourceWithBadCredentials(t *testing.T) {
	srv := httptest.NewServer(userBadCredentialsHandler(t))
	defer srv.Close()

	_, err := userTestSource(srv).token(t.Context())
	var jsonErr internalerrors.JSONError
	if err == nil || !errors.As(err, &jsonErr) {
		t.Fatalf("expected token request to fail with a JSONError: %s", err)
	}
	if !strings.HasSuffix(err.Error(), "UNAUTHENTICATED: wrong username or password (code: 401, traceId: n2jJpBU8qkEy3k09s9JNkg==)") {
		t.Fatal(err)
	}
}

func TestUserSourceWithInternalServerErrorNoBody(t *testing.T) {
	srv := httptest.NewServer(internalServerErrorWithNoBodyHandler())
	defer srv.Close()

	_, err := userTestSource(srv).token(t.Context())
	if err == nil {
		t.Fatal("expected token request to fail")
	}
	if !strings.HasSuffix(err.Error(), "token response has no body (status code 500)") {
		t.Fatal(err)
	}
}

func TestUserSourceWithInternalServerErrorTextBody(t *testing.T) {
	srv := httptest.NewServer(internalServerErrorWithTextBodyHandler())
	defer srv.Close()

	_, err := userTestSource(srv).token(t.Context())
	if err == nil {
		t.Fatal("expected token request to fail")
	}
	if !strings.HasSuffix(err.Error(), " token response has Content-Type text/plain; charset=utf-8 (status code 500): \"Internal Server Error\"") {
		t.Fatal(err)
	}
}

func TestUserSourceWithCanceledContext(t *testing.T) {
	srv := httptest.NewServer(validTokenHandler(t))
	defer srv.Close()

	// Canceled context.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := userTestSource(srv).token(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestUserSourceWithDeadlineExceededContext(t *testing.T) {
	srv := httptest.NewServer(validTokenHandler(t))
	defer srv.Close()

	// Deadline exceeded context
	ctx, cancel := context.WithTimeout(t.Context(), 0*time.Second)
	defer cancel()

	if _, err := userTestSource(srv).token(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func userTestSource(srv *httptest.Server) *UserSource {
	return &UserSource{
		log:      &log.DiscardLogger{},
		client:   srv.Client(),
		tokenURL: srv.URL + "/api/session",
		username: "username",
		password: "password",
	}
}

func userTestSourceWithWrongPassword(srv *httptest.Server) *UserSource {
	return &UserSource{
		log:      &log.DiscardLogger{},
		client:   srv.Client(),
		tokenURL: srv.URL + "/api/session",
		username: "username",
		password: "wrong-password",
	}
}

func userBadCredentialsHandler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("testdata/user_auth_error_response.json")
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
