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
	"fmt"
	"net/http"
	"sync"
)

// RoundTripper decorates an existing RoundTripper and injects an Authorization
// header with a valid access token. The token is automatically refreshed when
// it expires.
type RoundTripper struct {
	mutex sync.Mutex
	next  http.RoundTripper
	src   Source
	token token
}

// NewRoundTripper returns a new token RoundTripper decorating the specified
// http.RoundTripper.
func NewRoundTripper(next http.RoundTripper, tokenSource Source) *RoundTripper {
	return &RoundTripper{next: next, src: tokenSource}
}

// cloneRequest does a shallow copy of the request and a deep copy of the
// request's headers.
func cloneRequest(req *http.Request) *http.Request {
	clone := &http.Request{}
	*clone = *req
	clone.Header = req.Header.Clone()
	return clone
}

// RoundTrip handles a single HTTP request. Note that a RoundTripper must be
// safe for concurrent use by multiple goroutines.
func (t *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	closeBody := true
	if req.Body != nil {
		defer func() {
			if closeBody {
				req.Body.Close()
			}
		}()
	}

	// Clone request and add the authorization token.
	authReq := cloneRequest(req)
	t.mutex.Lock()
	if t.token.expired() {
		var err error
		t.token, err = t.src.token(req.Context())
		if err != nil {
			t.mutex.Unlock()
			return nil, fmt.Errorf("failed to refresh access token: %v", err)
		}
	}
	t.token.setAsAuthHeader(authReq)
	t.mutex.Unlock()

	// At this point the next RoundTripper is responsible for closing the
	// request body.
	closeBody = false
	return t.next.RoundTrip(authReq)
}
