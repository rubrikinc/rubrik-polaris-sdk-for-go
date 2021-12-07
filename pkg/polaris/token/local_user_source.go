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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// LocalUserSource holds all the information needed to obtain a token for a
// local user account.
type LocalUserSource struct {
	logger   log.Logger
	client   *http.Client
	tokenURL string
	username string
	password string
}

// NewLocalUserSource returns a new token source that uses the specified client
// to obtain tokens.
func NewLocalUserSource(client *http.Client, apiURL, username, password string, logger log.Logger) *LocalUserSource {
	return &LocalUserSource{
		logger:   logger,
		client:   client,
		tokenURL: fmt.Sprintf("%s/session", apiURL),
		username: username,
		password: password,
	}
}

// token returns a new token from the local user token source.
func (src *LocalUserSource) token() (token, error) {
	// Prepare the token request body.
	body, err := json.Marshal(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{Username: src.username, Password: src.password})
	if err != nil {
		return token{}, fmt.Errorf("failed to marshal token request body: %v", err)
	}

	resp, err := Request(src.client, src.tokenURL, body, src.logger)
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
