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

package token

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
)

// dummyExpiredToken is a JWT token that is already expired.
var dummyExpiredToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMj" +
	"M0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE2M" +
	"jI0OTExNTR9.y3TkH5_8Pv7Vde1I-ll2BJ29dX4tYKGIhrAA314VGa0"

// dummyValidToken is a JWT token that is still valid.
var dummyValidToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0" +
	"NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjQ3Nzg" +
	"zNzUzMDZ9.jAAX5cAp7UVLY6Kj1KS6UVPhxV2wtNNuYIUrXm_vGQ0"

func TestTokenExpired(t *testing.T) {
	tok := token{}
	if !tok.expired() {
		t.Fatal("empty token should be expired")
	}

	tok, err := fromJWT(dummyExpiredToken)
	if err != nil {
		t.Fatal(err)
	}
	if !tok.expired() {
		t.Error("token should be expired")
	}

	tok, err = fromJWT(dummyValidToken)
	if err != nil {
		t.Fatal(err)
	}
	if tok.expired() {
		t.Error("token should not be expired")
	}
}

func TestTokenSetAsHeader(t *testing.T) {
	tok, err := fromJWT(dummyExpiredToken)
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		Header: make(http.Header),
	}
	tok.setAsAuthHeader(req)

	if auth := req.Header.Get("Authorization"); auth != fmt.Sprintf("Bearer %s", dummyExpiredToken) {
		t.Errorf("invalid Authorization header: %s", auth)
	}
}

func TestTokenSource(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	// Respond with 200 and a valid token as long as the correct username and
	// password are received.
	srv := httptest.NewServer(handler.JSON(func(w http.ResponseWriter, req *http.Request) {
		var payload struct {
			Username string
			Password string
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			cancel(err)
			return
		}

		if payload.Username != "john" || payload.Password != "doe" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if err := json.NewEncoder(w).Encode(struct {
			AccessToken string `json:"access_token"`
		}{AccessToken: dummyValidToken}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	// Request token and verify that it's not expired.
	token, err := NewTestUserSource(srv, "john", "doe").token(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if token.expired() {
		t.Fatal("invalid token, already expired")
	}

	// Request token with invalid credentials.
	_, err = NewTestUserSource(srv, "jane", "doe").token(ctx)
	if err == nil || !strings.HasPrefix(err.Error(), "failed to acquire local user access token:") {
		t.Fatalf("invalid error: %v", err)
	}
}

func TestTokenSourceWithBadCredentials(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/auth_error_response.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	// Respond with status code 401 and additional details in the body.
	srv := httptest.NewServer(handler.JSON(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(401)
		if err := tmpl.Execute(w, nil); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	_, err = NewTestUserSource(srv, "john", "doe").token(ctx)
	if err == nil {
		t.Fatal("token request should fail")
	}
	if !strings.HasSuffix(err.Error(), "UNAUTHENTICATED: wrong username or password (code: 401, traceId: n2jJpBU8qkEy3k09s9JNkg==)") {
		t.Fatal(err)
	}
}

func TestTokenSourceWithInternalServerErrorNoBody(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	// Respond with status code 500 and no additional details.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	_, err := NewTestUserSource(srv, "john", "doe").token(ctx)
	if err == nil {
		t.Fatal("token request should fail")
	}
	if !strings.HasSuffix(err.Error(), "token response has no body (status code 500)") {
		t.Fatal(err)
	}
}

func TestTokenSourceWithInternalServerErrorTextBody(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	// Respond with status code 500 and no additional details.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(500)
		w.Write([]byte("user database is corrupt"))
	}))
	defer srv.Close()

	_, err := NewTestUserSource(srv, "john", "doe").token(ctx)
	if err == nil {
		t.Fatal("token request should fail")
	}
	if !strings.HasSuffix(err.Error(),
		"token response has Content-Type text/plain (status code 500): \"user database is corrupt\"") {
		t.Fatal(err)
	}
}
