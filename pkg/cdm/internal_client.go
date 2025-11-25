// Copyright 2023 Rubrik, Inc.
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

package cdm

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

type APIVersion string

const (
	V1       APIVersion = "v1"
	V2       APIVersion = "v2"
	Internal APIVersion = "internal"
)

type authorization func(*http.Request)

type client struct {
	auth   authorization
	client *http.Client
	log    log.Logger
}

func newClientWithLogger(allowInsecureTLS bool, logger log.Logger) *client {
	httpClient := &http.Client{}
	if allowInsecureTLS {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = transport
	}

	return &client{
		client: httpClient,
		log:    logger,
	}
}

func newClientFromCredentialsWithLogger(username, password string, allowInsecureTLS bool, logger log.Logger) (*client, error) {
	if username == "" {
		return nil, errors.New("username required")
	}
	if password == "" {
		return nil, errors.New("password required")
	}

	httpClient := &http.Client{}
	if allowInsecureTLS {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = transport
	}

	return &client{
		auth: func(req *http.Request) {
			req.SetBasicAuth(username, password)
		},
		client: httpClient,
		log:    logger,
	}, nil
}

func newClientFromToken(token string, allowInsecureTLS bool) (*client, error) {
	return newClientFromTokenWithLogger(token, allowInsecureTLS, log.DiscardLogger{})
}

func newClientFromTokenWithLogger(token string, allowInsecureTLS bool, logger log.Logger) (*client, error) {
	if token == "" {
		return nil, errors.New("authentication token required")
	}

	httpClient := &http.Client{}
	if allowInsecureTLS {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = transport
	}

	return &client{
		auth: func(req *http.Request) {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
		},
		client: httpClient,
		log:    logger,
	}, nil
}

func (c *client) request(ctx context.Context, method, nodeIP string, version APIVersion, endpoint string, payload any) (*http.Request, error) {
	c.log.Print(log.Trace)

	if version != V1 && version != V2 && version != Internal {
		return nil, fmt.Errorf("invalid API version: %s", version)
	}

	// Clean endpoint path.
	if !path.IsAbs(endpoint) {
		endpoint = "/" + endpoint
	}
	endpoint = path.Clean(endpoint)

	// Encode payload.
	body := &bytes.Buffer{}
	if payload != nil {
		if err := json.NewEncoder(body).Encode(payload); err != nil {
			return nil, err
		}
	}

	// Create request.
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("https://%s/api/%s%s", nodeIP, version, endpoint), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create requeset: %s", err)
	}
	if c.auth != nil {
		c.auth(req)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Rubrik Polaris Go SDK")

	return req, nil
}

func (c *client) doRequest(req *http.Request) ([]byte, int, error) {
	c.log.Print(log.Trace)

	res, reqErr := c.client.Do(req)
	defer func() {
		var status string
		if res != nil {
			status = res.Status
		}
		if reqErr != nil {
			status = strings.TrimSpace(fmt.Sprintf("%s %s", status, reqErr))
		}
		c.log.Printf(log.Debug, "request %s %s - %s", req.Method, req.URL, status)
	}()
	if reqErr != nil {
		return nil, 0, fmt.Errorf("failed to perform request: %s", reqErr)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	// Consume response.
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response body: %s", err)
	}

	return buf, res.StatusCode, nil
}
