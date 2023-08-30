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
	"os"
	"path"
	"strings"
)

type APIVersion string

const (
	V1       APIVersion = "v1"
	V2       APIVersion = "v2"
	Internal APIVersion = "internal"
)

// Client is used to make API calls to the CDM platform.
type Client struct {
	NodeIP   string
	Username string
	Password string
	Token    string
	client   *http.Client
}

// NewClientFromCredentials creates a new client from the provided Rubrik
// cluster credentials.
func NewClientFromCredentials(nodeIP, username, password string, allowInsecureTLS bool) (*Client, error) {
	if nodeIP == "" {
		return nil, errors.New("node ip required")
	}
	if username == "" {
		return nil, errors.New("username required")
	}
	if password == "" {
		return nil, errors.New("password required")
	}

	client := &http.Client{}
	if allowInsecureTLS {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &Client{
		NodeIP:   nodeIP,
		Username: username,
		Password: password,
		client:   client,
	}, nil
}

// NewClientFromToken creates a new client from the provided authentication
// token.
func NewClientFromToken(nodeIP, token string, allowInsecureTLS bool) (*Client, error) {
	if nodeIP == "" {
		return nil, errors.New("node ip required")
	}
	if token == "" {
		return nil, errors.New("authentication token required")
	}

	client := &http.Client{}
	if allowInsecureTLS {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &Client{
		NodeIP: nodeIP,
		Token:  token,
		client: client,
	}, nil
}

// NewClientFromEnv creates a new client from the current environment. The
// following environment variables are read:
//
//	rubrik_cdm_node_ip (or RUBRIK_CDM_NODE_IP)
//
//	rubrik_cdm_token (or RUBRIK_CDM_TOKEN)
//
//	rubrik_cdm_username (or RUBRIK_CDM_USERNAME)
//
//	rubrik_cdm_password (or RUBRIK_CDM_PASSWORD)
//
// If a token is found in the environment it will always take precedence over
// any other form of credentials.
func NewClientFromEnv(allowInsecureTLS bool) (*Client, error) {
	nodeIP := fromEnv("RUBRIK_CDM_NODE_IP")

	if token := fromEnv("RUBRIK_CDM_TOKEN"); token != "" {
		return NewClientFromToken(nodeIP, token, allowInsecureTLS)
	}

	username := fromEnv("RUBRIK_CDM_USERNAME")
	password := fromEnv("RUBRIK_CDM_PASSWORD")
	return NewClientFromCredentials(nodeIP, username, password, allowInsecureTLS)
}

func fromEnv(name string) string {
	if value := os.Getenv(strings.ToLower(name)); value != "" {
		return value
	}
	if value := os.Getenv(strings.ToUpper(name)); value != "" {
		return value
	}

	return ""
}

// Get sends a GET request to the provided Rubrik API endpoint and returns the
// response and status code.
func (c *Client) Get(ctx context.Context, version APIVersion, endpoint string) ([]byte, int, error) {
	req, err := c.request(ctx, http.MethodGet, version, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(ctx, req)
}

// Post sends a POST request to the provided Rubrik API endpoint and returns the
// response and status code.
func (c *Client) Post(ctx context.Context, version APIVersion, endpoint string, payload any) ([]byte, int, error) {
	req, err := c.request(ctx, http.MethodPost, version, endpoint, payload)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(ctx, req)
}

// Patch sends a PATCH request to the provided Rubrik API endpoint and returns
// the response and status code.
func (c *Client) Patch(ctx context.Context, version APIVersion, endpoint string, payload any) ([]byte, int, error) {
	req, err := c.request(ctx, http.MethodPatch, version, endpoint, payload)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(ctx, req)
}

// Delete sends a DELETE request to the provided Rubrik API endpoint and returns
// the response and status code.
func (c *Client) Delete(ctx context.Context, version APIVersion, endpoint string) ([]byte, int, error) {
	req, err := c.request(ctx, http.MethodDelete, version, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(ctx, req)
}

func (c *Client) request(ctx context.Context, method string, version APIVersion, endpoint string, payload any) (*http.Request, error) {
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
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("https://%s/api/%s%s", c.NodeIP, version, endpoint), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create requeset: %s", err)
	}
	if c.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	} else {
		req.SetBasicAuth(c.Username, c.Password)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Rubrik Polaris Go SDK")

	return req, nil
}

func (c *Client) doRequest(ctx context.Context, req *http.Request) ([]byte, int, error) {
	res, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to perform request: %s", err)
	}
	defer res.Body.Close()

	// Consume response.
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response body: %s", err)
	}

	return buf, res.StatusCode, nil
}
