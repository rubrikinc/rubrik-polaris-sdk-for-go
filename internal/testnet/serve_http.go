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

package testnet

import (
	"errors"
	"net"
	"net/http"
)

// Serve serves the handler function over HTTP by accepting incoming connections
// on the specified listener. Intended to be used with a pipenet in unit tests.
func Serve(lis net.Listener, handler HandlerFunc) CancelFunc {
	return serve(lis, serveHTTP, handler)
}

// ServeJSON serves the handler function using HTTP by accepting incoming
// connections on the specified listener. The response content-type is set to
// application/json. Intended to be used with a pipenet in unit tests.
func ServeJSON(lis net.Listener, handler HandlerFunc) CancelFunc {
	return Serve(lis, func(w http.ResponseWriter, req *http.Request) error {
		w.Header().Set("Content-Type", "application/json")
		return handler(w, req)
	})
}

// ServeWithStaticToken serves the handler function and tokens using HTTP by
// accepting incoming connections on specified listener. Intended to be used
// with a pipenet in unit tests.
func ServeWithStaticToken(lis net.Listener, handler HandlerFunc) CancelFunc {
	return serve(lis, serveHTTP, func(w http.ResponseWriter, req *http.Request) error {
		switch req.URL.Path {
		case "/api/session":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjQ3NzgzNzUzMDZ9.jAAX5cAp7UVLY6Kj1KS6UVPhxV2wtNNuYIUrXm_vGQ0",
				"is_eula_accepted": true
			}`))
			return nil
		case "/api/graphql":
			return handler(w, req)
		default:
			return errors.New("invalid url path")
		}
	})
}

// ServeJSONWithStaticToken serves the handler function and tokens using HTTP by
// accepting incoming connections on specified listener. The response
// content-type is set to application/json. Intended to be used with a pipenet
// in unit tests.
func ServeJSONWithStaticToken(lis net.Listener, handler HandlerFunc) CancelFunc {
	return ServeWithStaticToken(lis, func(w http.ResponseWriter, req *http.Request) error {
		w.Header().Set("Content-Type", "application/json")
		return handler(w, req)
	})
}

// serveHTTP starts an HTTP server
func serveHTTP(lis net.Listener, server *http.Server) error {
	if err := server.Serve(lis); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
