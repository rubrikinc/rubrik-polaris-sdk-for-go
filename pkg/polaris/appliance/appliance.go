// Copyright 2022 Rubrik, Inc.
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

// Package appliance encapsulates methods to help interact with Rubrik Appliance.
package appliance

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/token"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// TokenFromServiceAccount returns a token to access appliance
// REST APIs. It is issued on behalf of the given service account.
func TokenFromServiceAccount(clientID, clientSecret, accessTokenURI string, clusterID uuid.UUID, logger log.Logger) (string, error) {
	if !strings.HasSuffix(accessTokenURI, "/client_token") {
		return "", errors.New("invalid access token uri")
	}
	tokenURL := strings.TrimSuffix(accessTokenURI, "/client_token") + "/cdm_client_token"

	body, err := json.Marshal(struct {
		ClientID		string `json:"client_id"`
		ClientSecret	string `json:"client_secret"`
		ClusterUuid		uuid.UUID `json:"cluster_uuid"`
	}{ClientID: clientID, ClientSecret: clientSecret, ClusterUuid: clusterID})
	if err != nil {
		return "", fmt.Errorf("internal: failed to marshal token request body: %v", err)
	}

	resp, err := token.Request(http.DefaultClient, tokenURL, body, logger)
	if err != nil {
		return "", err
	}

	var payload struct {
		Session struct {
			AccessToken string `json:"token"`
		}  `json:"session"`
	}
	if err := json.Unmarshal(resp, &payload); err != nil {
		return "", fmt.Errorf("failed to unmarshal token response body: %v", err)
	}

	if payload.Session.AccessToken == "" {
		return "", errors.New("invalid token")
	}

	return payload.Session.AccessToken, nil
}
