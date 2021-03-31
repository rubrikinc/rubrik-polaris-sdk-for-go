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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// token holds a textual access token and the time when it expires.
type token struct {
	token  string
	expiry time.Time
}

// expired returns true if the token has expired.
func (t token) expired() bool {
	return t.expiry.Before(time.Now())
}

// setAsAuthHeader adds an Authorization header with a bearer token to the
// specified request.
func (t token) setAsAuthHeader(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t.token))
}

// tokenSource holds all the information needed to obtain a token and the way
// to obtain it.
type tokenSource struct {
	client   *http.Client
	tokenURL string
	username string
	password string
}

// newTokenSource returns a new token source that uses the http.DefaultClient
// to obtain tokens.
func newTokenSource(apiURL, username, password string) tokenSource {
	return tokenSource{
		client:   http.DefaultClient,
		tokenURL: fmt.Sprintf("%s/session", apiURL),
		username: username,
		password: password,
	}
}

// newTestTokenSource returns a new token source that uses a pipe network
// connection to obtain tokens. Intended to be used by unit tests.
func newTestTokenSource(username, password string) (tokenSource, *TestListener) {
	dialer, listener := newPipeNet()

	// Copy the default transport and install the test dialer.
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		panic("http.DefaultTransport is not of type *http.Transport")
	}
	transport = transport.Clone()
	transport.DialContext = dialer

	src := tokenSource{
		client: &http.Client{
			Transport: transport,
		},
		tokenURL: "http://test/api/session",
		username: username,
		password: password,
	}

	return src, listener
}

// token returns a new token from the token source.
func (src *tokenSource) token() (token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare the token request body.
	buf, err := json.Marshal(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{Username: src.username, Password: src.password})
	if err != nil {
		return token{}, err
	}

	// Request an access token from the remote token endpoint.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, src.tokenURL, bytes.NewReader(buf))
	if err != nil {
		return token{}, err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	res, err := src.client.Do(req)
	if err != nil {
		return token{}, err
	}
	defer res.Body.Close()

	// Remote responded without a body. For status code 200 this means we are
	// missing the token. For an error we have no additional details.
	if res.ContentLength == 0 {
		if res.StatusCode != 200 {
			return token{}, fmt.Errorf("polaris: %s", res.Status)
		} else {
			return token{}, errors.New("polaris: no body")
		}
	}

	// Verify that the content type of the body is JSON.
	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return token{}, fmt.Errorf("polaris: wrong content-type: %s", contentType)
	}

	buf, err = io.ReadAll(res.Body)
	if err != nil {
		return token{}, err
	}

	// Remote responded with a JSON document. Try to parse it as an error
	// message.
	var jsonErr jsonError
	if err := json.Unmarshal(buf, &jsonErr); err != nil {
		return token{}, err
	}
	if jsonErr.isError() {
		return token{}, jsonErr
	}
	if res.StatusCode != 200 {
		return token{}, fmt.Errorf("polaris: %s", res.Status)
	}

	// Try to parse the JSON document as an access token.
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return token{}, err
	}
	if payload.AccessToken == "" {
		return token{}, errors.New("polaris: invalid token")
	}

	// Note that token expiration is hardcoded to 2 hours. We should inspect
	// the returned JWT token and determine this from the iat and exp fields.
	token := token{
		token:  payload.AccessToken,
		expiry: time.Now().Add(2 * time.Hour),
	}
	return token, nil
}

// cloneRequest does a shallow copy of the request and a deep copy of the
// request's headers.
func cloneRequest(req *http.Request) *http.Request {
	clone := &http.Request{}
	*clone = *req
	clone.Header = req.Header.Clone()
	return clone
}

// tokenTransport decorates an existing transport and injects an Authorization
// header with a valid access token. The token is automatically refreshed when
// it expires.
type tokenTransport struct {
	mutex sync.Mutex
	next  http.RoundTripper
	src   tokenSource
	token token
}

// RoundTrip handles a single HTTP request. Note that a RoundTripper must be
// safe for concurrent use by multiple goroutines.
func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
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
		t.token, err = t.src.token()
		if err != nil {
			t.mutex.Unlock()
			return nil, err
		}
	}
	t.token.setAsAuthHeader(authReq)
	t.mutex.Unlock()

	// At this point the next RoundTripper is responsible for closing the
	// request body.
	closeBody = false
	return t.next.RoundTrip(authReq)
}
