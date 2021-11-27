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
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// serviceAccountSource holds all the information needed to obtain a token
// for a service account.
type serviceAccountSource struct {
	logger       log.Logger
	client       *http.Client
	tokenURL     string
	clientID     string
	clientSecret string
}

// newServiceAccountSource returns a new token source that uses the
// http.DefaultClient to obtain tokens.
func newServiceAccountSource(accessTokenURL, clientID, clientSecret string, logger log.Logger) *serviceAccountSource {
	return &serviceAccountSource{
		logger:       logger,
		client:       http.DefaultClient,
		tokenURL:     accessTokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// token returns a new token from the service account token source.
func (src *serviceAccountSource) token() (Token, error) {
	// Prepare the token request body.
	body, err := json.Marshal(struct {
		GrantType    string `json:"grant_type"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}{GrantType: "client_credentials", ClientID: src.clientID, ClientSecret: src.clientSecret})
	if err != nil {
		return Token{}, fmt.Errorf("failed to marshal token request body: %v", err)
	}

	resp, err := fetchTokenWithRetries(src.client, src.tokenURL, body, src.logger)
	if err != nil {
		return Token{}, err
	}

	// Try to parse the JSON document as an access token. Verify that the
	// response has the same client id as the request.
	var payload struct {
		ClientID    string `json:"client_id"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(resp, &payload); err != nil {
		return Token{}, fmt.Errorf("failed to unmarshal token response body: %v", err)
	}
	if payload.ClientID != src.clientID {
		return Token{}, errors.New("invalid client id")
	}
	if payload.AccessToken == "" {
		return Token{}, errors.New("invalid token")
	}

	return fromJWT(payload.AccessToken)
}

func (src *serviceAccountSource) applianceToken(applianceUuid string) (Token, error) {
	body, err := json.Marshal(struct {
		ClientID		string `json:"client_id"`
		ClientSecret	string `json:"client_secret"`
		ClusterUuid		string `json:"cluster_uuid"`
	}{ClientID: src.clientID, ClientSecret: src.clientSecret, ClusterUuid: applianceUuid})
	if err != nil {
		return Token{}, fmt.Errorf("failed to marshal token request body: %v", err)
	}

	// Extract the API URL from the token access URI.
	i := strings.LastIndex(src.tokenURL, "/")
	if i < 0 {
		return Token{}, errors.New("invalid access token uri")
	}
	baseApiURL := src.tokenURL[:i]

	applianceTokenURL := baseApiURL + "/cdm_client_token"
	resp, err := fetchTokenWithRetries(src.client, applianceTokenURL, body, src.logger)
	if err != nil {
		return Token{}, err
	}

	var payload struct {
		Session struct {
			AccessToken string `json:"token"`
		}  `json:"session"`
	}
	if err := json.Unmarshal(resp, &payload); err != nil {
		return Token{}, fmt.Errorf("failed to unmarshal token response body: %v", err)
	}

	if payload.Session.AccessToken == "" {
		return Token{}, errors.New("invalid token")
	}

	return fromJWT(payload.Session.AccessToken)
}

func fetchTokenWithRetries(
	client *http.Client,
	tokenUrl string,
	body []byte,
	logger log.Logger,
) ([]byte, error) {
	for attempt := 1; ; attempt++ {
		logger.Printf(log.Debug, "Acquire access token (attempt: %d)", attempt)
		resp, err := requestToken(client, tokenUrl, body)
		if err == nil {
			return resp, nil
		}
		if !errors.Is(err, errTokenRequestTimeout) || attempt == tokenRequestAttempts {
			return []byte{}, fmt.Errorf("failed to acquire service account access token: %v", err)
		}
	}
}
