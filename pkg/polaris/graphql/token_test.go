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
	"strings"
	"testing"
	"text/template"
)

func TestTokenExpired(t *testing.T) {
	tok := token{}
	if !tok.expired() {
		t.Fatal("empty token should be expired")
	}

	tok, err := fromJWT("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE2MjI0OTExNTR9.y3TkH5_8Pv7Vde1I-ll2BJ29dX4tYKGIhrAA314VGa0")
	if err != nil {
		t.Fatal(err)
	}
	if !tok.expired() {
		t.Error("token should be expired")
	}

	tok, err = fromJWT("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjQ3NzgzNzUzMDZ9.jAAX5cAp7UVLY6Kj1KS6UVPhxV2wtNNuYIUrXm_vGQ0")
	if err != nil {
		t.Fatal(err)
	}
	if tok.expired() {
		t.Error("token should not be expired")
	}
}

func TestTokenSetAsHeader(t *testing.T) {
	tok, err := fromJWT("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE2MjI0OTExNTR9.y3TkH5_8Pv7Vde1I-ll2BJ29dX4tYKGIhrAA314VGa0")
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		Header: make(http.Header),
	}
	tok.setAsAuthHeader(req)

	if auth := req.Header.Get("Authorization"); auth != "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE2MjI0OTExNTR9.y3TkH5_8Pv7Vde1I-ll2BJ29dX4tYKGIhrAA314VGa0" {
		t.Errorf("invalid Authorization header: %s", auth)
	}
}

func TestTokenSource(t *testing.T) {
	src, lis := newLocalUserTestSource("john", "doe")

	// Respond with 200 and a valid token as long as the correct username and
	// password are received.
	srv := serveJSON(lis, func(w http.ResponseWriter, req *http.Request) {
		var payload struct {
			Username string
			Password string
		}
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if payload.Username != "john" || payload.Password != "doe" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		json.NewEncoder(w).Encode(struct {
			AccessToken string `json:"access_token"`
		}{AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjQ3NzgzNzUzMDZ9.jAAX5cAp7UVLY6Kj1KS6UVPhxV2wtNNuYIUrXm_vGQ0"})
	})
	defer srv.Shutdown(context.Background())

	// Request token and verify that it's not expired.
	token, err := src.token()
	if err != nil {
		t.Fatal(err)
	}
	if token.expired() {
		t.Fatal("invalid token, already expired")
	}
}

func TestTokenSourceWithBadCredentials(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/error_json_from_polaris.json")
	if err != nil {
		t.Fatal(err)
	}

	src, lis := newLocalUserTestSource("john", "doe")

	// Respond with status code 401 and additional details in the body.
	srv := serveJSON(lis, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(401)
		if err := tmpl.Execute(w, nil); err != nil {
			panic(err)
		}
	})
	defer srv.Shutdown(context.Background())

	_, err = src.token()
	if err == nil {
		t.Fatal("token request should fail")
	}
	if err.Error() != "polaris: code 401: UNAUTHENTICATED: wrong username or password" {
		t.Fatal(err)
	}
}

func TestTokenSourceWithInternalServerErrorNoBody(t *testing.T) {
	src, lis := newLocalUserTestSource("john", "doe")

	// Respond with status code 500 and no additional details.
	srv := serve(lis, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(500)
	})
	defer srv.Shutdown(context.Background())

	_, err := src.token()
	if err == nil {
		t.Fatal("token request should fail")
	}
	if !strings.HasPrefix(err.Error(), "polaris: 500 Internal Server Error") {
		t.Fatal(err)
	}
}

func TestTokenSourceWithInternalServerErrorTextBody(t *testing.T) {
	src, lis := newLocalUserTestSource("john", "doe")

	// Respond with status code 500 and no additional details.
	srv := serve(lis, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(500)
		w.Write([]byte("user database is corrupt"))
	})
	defer srv.Shutdown(context.Background())

	_, err := src.token()
	if err == nil {
		t.Fatal("token request should fail")
	}
	if !strings.HasPrefix(err.Error(), "polaris: 500 Internal Server Error") {
		t.Fatal(err)
	}
}
