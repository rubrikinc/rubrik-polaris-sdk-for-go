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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// UserSource holds all the information needed to obtain a token for a user
// account.
type UserSource struct {
	log      log.Logger
	client   *http.Client
	tokenURL string
	username string
	password string
}

// Deprecated: use UserSource
type LocalUserSource = UserSource

// NewUserSource returns a new token source that uses the specified client to
// obtain tokens.
func NewUserSource(client *http.Client, apiURL, username, password string) *UserSource {
	return NewUserSourceWithLogger(client, apiURL, username, password, &log.DiscardLogger{})
}

// NewUserSourceWithLogger returns a new token source that uses the specified
// client to obtain tokens.
func NewUserSourceWithLogger(client *http.Client, tokenURL, username, password string, logger log.Logger) *UserSource {
	return &UserSource{
		log:      logger,
		client:   client,
		tokenURL: tokenURL,
		username: username,
		password: password,
	}
}

// NewTestUserSource returns a new test token source that uses a client matching
// the specified test server to obtain tokens. Intended to be used in unit
// tests.
func NewTestUserSource(testServer *httptest.Server, username, password string) *UserSource {
	return &UserSource{
		log:      &log.DiscardLogger{},
		client:   testServer.Client(),
		tokenURL: testServer.URL + "/api/session",
		username: username,
		password: password,
	}
}

// Deprecated: use NewUserSourceWithLogger.
func NewLocalUserSource(client *http.Client, apiURL, username, password string, logger log.Logger) *LocalUserSource {
	return NewUserSourceWithLogger(client, apiURL, username, password, logger)
}

// token returns a new token from the user token source.
func (src *UserSource) token(ctx context.Context) (token, error) {
	// Prepare the token request body.
	body, err := json.Marshal(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{Username: src.username, Password: src.password})
	if err != nil {
		return token{}, fmt.Errorf("failed to marshal token request body: %v", err)
	}

	resp, err := RequestWithContext(ctx, src.client, src.tokenURL, body, src.log)
	if err != nil {
		return token{}, fmt.Errorf("failed to acquire local user access token: %v", err)
	}

	// Try to parse the JSON document as an access token.
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(resp, &payload); err != nil {
		return token{}, fmt.Errorf("failed to unmarshal token response body: %v", err)
	}
	if payload.AccessToken == "" {
		return token{}, errors.New("invalid token")
	}

	return fromJWT(payload.AccessToken)
}
