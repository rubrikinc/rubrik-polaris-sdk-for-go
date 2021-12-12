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

// Package appliance enapsulates methods to help interact with Rubrik Appliance.
package appliance

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/token"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
)

// ApplianceTokenFromServiceAccount returns a token to access appliance
// REST APIs. It is issued on behalf of the given service account.
func ApplianceTokenFromServiceAccount(account *polaris.ServiceAccount, clusterID uuid.UUID, logger log.Logger) (string, error) {
	if !strings.HasSuffix(account.AccessTokenURI, "/client_token") {
		return "", errors.New("invalid access token uri")
	}
	tokenURL := strings.TrimSuffix(account.AccessTokenURI, "/client_token") + "/cdm_client_token"

	body, err := json.Marshal(struct {
		ClientID		string `json:"client_id"`
		ClientSecret	string `json:"client_secret"`
		ClusterUuid		uuid.UUID `json:"cluster_uuid"`
	}{ClientID: account.ClientID, ClientSecret: account.ClientSecret, ClusterUuid: clusterID})
	if err != nil {
		logger.Printf(log.Error, "failed to marshat token request body %v", err)	
		return "", errors.New("internal: failed to marshal token request body")
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
		logger.Printf(log.Error, "failed to unmarshal cluster token response body %v", err)
		return "", errors.New("failed to unmarshal token response body")
	}

	if payload.Session.AccessToken == "" {

		return "", errors.New("invalid token")
	}

	return payload.Session.AccessToken, nil
}
