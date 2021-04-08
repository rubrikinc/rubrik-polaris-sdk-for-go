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
	"net/http"
	"testing"
	"text/template"
	"time"
)

func TestTokenExpired(t *testing.T) {
	tok := token{
		token:  "token",
		expiry: time.Now().Add(-1 * time.Minute),
	}
	if !tok.expired() {
		t.Error("token should be expired")
	}

	tok = token{
		token:  "token",
		expiry: time.Now().Add(1 * time.Minute),
	}
	if tok.expired() {
		t.Error("token should not be expired")
	}
}

func TestTokenSetAsHeader(t *testing.T) {
	tok := token{
		token:  "token",
		expiry: time.Now(),
	}

	req := &http.Request{
		Header: make(http.Header),
	}
	tok.setAsAuthHeader(req)

	if auth := req.Header.Get("Authorization"); auth != "Bearer token" {
		t.Errorf("invalid Authorization header: %s", auth)
	}
}

func TestTokenSource(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/session.json")
	if err != nil {
		t.Fatal(err)
	}

	src, lis := newTestTokenSource("john", "doe")

	// Respond with status code 200 and a token created by concatenating the
	// username and password.
	srv := serveJSON(lis, func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var payload struct {
			Username string
			Password string
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if err := tmpl.Execute(w, payload); err != nil {
			panic(err)
		}
	})
	defer srv.Shutdown(context.Background())

	token, err := src.token()
	if err != nil {
		t.Fatal(err)
	}
	if token.token != "john:doe" {
		t.Fatalf("invalid token: %v", token.token)
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

	src, lis := newTestTokenSource("john", "doe")

	// Respond with status code 401 and additional details in the body.
	srv := serveJSON(lis, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(401)
		if err := tmpl.Execute(w, nil); err != nil {
			panic(err)
		}
	})
	defer srv.Shutdown(context.Background())

	if _, err := src.token(); err == nil {
		t.Fatal("token request should fail")
	}
}

func TestTokenSourceWithInternalServerErrorNoBody(t *testing.T) {
	src, lis := newTestTokenSource("john", "doe")

	// Respond with status code 500 and no additional details.
	srv := serve(lis, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(500)
	})
	defer srv.Shutdown(context.Background())

	if _, err := src.token(); err == nil {
		t.Fatal("token request should fail")
	}
}
