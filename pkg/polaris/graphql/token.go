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

	"github.com/golang-jwt/jwt/v4"
)

const (
	// per request timeout
	tokenRequestTimeout = 15 * time.Second

	// number of attempts before failing timed-out token requests
	tokenRequestAttempts = 3
)

var errTokenRequestTimeout = errors.New("token request timeout")

type token struct {
	jwtToken *jwt.Token
}

// expired returns true if the token has expired or if the token has no
// expiration time associated with it.
func (t token) expired() bool {
	if t.jwtToken == nil {
		return true
	}

	claims, ok := t.jwtToken.Claims.(jwt.MapClaims)
	if ok {
		// Compare the expiry to 1 minute into the future to avoid the token
		// expiring in transit or because clocks being skewed.
		now := time.Now().Add(1 * time.Minute)
		return !claims.VerifyExpiresAt(now.Unix(), true)
	}

	return true
}

// setAsAuthHeader adds an Authorization header with a bearer token to the
// specified request.
func (t token) setAsAuthHeader(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t.jwtToken.Raw))
}

// fromJWT returns a new token from the JWT token text. Note that the token
// signature is not verified.
func fromJWT(text string) (token, error) {
	p := jwt.Parser{}
	jwtToken, _, err := p.ParseUnverified(text, jwt.MapClaims{})
	if err != nil {
		return token{}, err
	}

	return token{jwtToken: jwtToken}, nil
}

// tokenSource is used to obtain access tokens from a remote source.
type tokenSource interface {
	token() (token, error)
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

// requestToken tries to acquire a token using provided parameters. It returns
// errTokenRequestTimeout if the request exceeds tokenRequestTimeout.
func requestToken(client *http.Client, tokenURL string, requestBody []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), tokenRequestTimeout)
	defer cancel()

	// Request an access token from the remote token endpoint.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, errTokenRequestTimeout
		}
		return nil, err
	}
	defer res.Body.Close()
	// Remote responded without a body. For status code 200 this means we are
	// missing the token. For an error we have no additional details.
	if res.ContentLength == 0 {
		if res.StatusCode == 200 {
			return nil, errors.New("polaris: no body")
		}
		return nil, fmt.Errorf("polaris: %s", res.Status)
	}

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, errTokenRequestTimeout
		}
		return nil, err
	}

	// Verify that the content type of the body is JSON. For status code 200
	// this mean we received something that isn't JSON. For an error we have no
	// additional JSON details.
	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		if res.StatusCode == 200 {
			return nil, fmt.Errorf("polaris: wrong content-type: %s", contentType)
		}
		return nil, fmt.Errorf("polaris: %s - %s", res.Status, respBody)
	}

	// Remote responded with a JSON document. Try to parse it as an error
	// message.
	var jsonErr jsonError
	if err := json.Unmarshal(respBody, &jsonErr); err != nil {
		return nil, err
	}
	if jsonErr.isError() {
		return nil, jsonErr
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("polaris: %s", res.Status)
	}

	return respBody, nil
}
