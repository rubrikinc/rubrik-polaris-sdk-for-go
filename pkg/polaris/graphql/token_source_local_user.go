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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// localUserSource holds all the information needed to obtain a token for a
// local user account.
type localUserSource struct {
	logger   log.Logger
	client   *http.Client
	tokenURL string
	username string
	password string
}

// newLocalUserSource returns a new token source that uses the http.DefaultClient
// to obtain tokens.
func newLocalUserSource(apiURL, username, password string, logger log.Logger) *localUserSource {
	return &localUserSource{
		logger:   logger,
		client:   http.DefaultClient,
		tokenURL: fmt.Sprintf("%s/session", apiURL),
		username: username,
		password: password,
	}
}

// newLocalUserTestSource returns a new token source that uses a pipe network
// connection to obtain tokens. Intended to be used by unit tests.
func newLocalUserTestSource(username, password string, logger log.Logger) (*localUserSource, *TestListener) {
	dialer, listener := newPipeNet()

	// Copy the default transport and install the test dialer.
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		panic("http.DefaultTransport is not of type *http.Transport")
	}
	transport = transport.Clone()
	transport.DialContext = dialer

	src := &localUserSource{
		logger: logger,
		client: &http.Client{
			Transport: transport,
		},
		tokenURL: "http://test/api/session",
		username: username,
		password: password,
	}

	return src, listener
}

// token returns a new token from the local user token source.
func (src *localUserSource) token() (token, error) {
	// Prepare the token request body.
	body, err := json.Marshal(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{Username: src.username, Password: src.password})
	if err != nil {
		return token{}, err
	}

	var resp []byte
	for attempt := 1; ; attempt++ {
		src.logger.Printf(log.Debug, "Acquire access token (attempt: %d)", attempt)
		resp, err = requestToken(src.client, src.tokenURL, body)
		if err == nil {
			break
		}
		if !errors.Is(err, errTokenRequestTimeout) || attempt == tokenRequestAttempts {
			return token{}, err
		}
	}

	// Try to parse the JSON document as an access token.
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(resp, &payload); err != nil {
		return token{}, err
	}
	if payload.AccessToken == "" {
		return token{}, errors.New("polaris: invalid token")
	}

	return fromJWT(payload.AccessToken)
}
