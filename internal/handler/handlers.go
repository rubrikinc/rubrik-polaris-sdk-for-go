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

package handler

import (
	"encoding/json"
	"net/http"
)

// dummyToken is a dummy JWT token for testing purposes.
var dummyToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3O" +
	"DkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjQ3NzgzNzUz" +
	"MDZ9.jAAX5cAp7UVLY6Kj1KS6UVPhxV2wtNNuYIUrXm_vGQ0"

// Token handles the /api/session endpoint. Other paths are forwarded to the
// specified handler.
func Token(handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/session" {
			handler(w, req)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(struct {
			AccessToken    string `json:"access_token"`
			IsEulaAccepted bool   `json:"is_eula_accepted"`
		}{AccessToken: dummyToken, IsEulaAccepted: true}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

// GraphQL handles the /api/graphql endpoint. Other paths returns a status code
// 404 error.
func GraphQL(handler http.HandlerFunc) http.Handler {
	return Token(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/graphql" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		handler(w, req)
	})
}

// JSON sets the Content-Type header to application/json before calling the
// specified handler.
func JSON(handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler(w, req)
	})
}
