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
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Client is used to make API calls to the CDM platform.
type Client struct {
	*client
	nodeIP string
	Log    log.Logger
}

// NewClient creates a new client without any cluster credentials.
func NewClient(nodeIP string, allowInsecureTLS bool) *Client {
	return NewClientWithLogger(nodeIP, allowInsecureTLS, log.DiscardLogger{})
}

// NewClientWithLogger creates a new client without any cluster credentials.
// The client logs to the provided logger.
func NewClientWithLogger(nodeIP string, allowInsecureTLS bool, logger log.Logger) *Client {
	return &Client{
		client: newClientWithLogger(allowInsecureTLS, logger),
		nodeIP: nodeIP,
		Log:    logger,
	}
}

// NewClientFromCredentials creates a new client from the provided Rubrik
// cluster credentials.
func NewClientFromCredentials(nodeIP, username, password string, allowInsecureTLS bool) (*Client, error) {
	if nodeIP == "" {
		return nil, errors.New("node ip required")
	}

	client, err := newClientFromCredentials(username, password, allowInsecureTLS)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		nodeIP: nodeIP,
		Log:    log.DiscardLogger{},
	}, nil
}

// NewClientFromToken creates a new client from the provided authentication
// token.
func NewClientFromToken(nodeIP, token string, allowInsecureTLS bool) (*Client, error) {
	if nodeIP == "" {
		return nil, errors.New("node ip required")
	}

	client, err := newClientFromToken(token, allowInsecureTLS)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		nodeIP: nodeIP,
		Log:    log.DiscardLogger{},
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
	c.Log.Print(log.Trace)

	req, err := c.request(ctx, http.MethodGet, c.nodeIP, version, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(req)
}

// Post sends a POST request to the provided Rubrik API endpoint and returns the
// response and status code.
func (c *Client) Post(ctx context.Context, version APIVersion, endpoint string, payload any) ([]byte, int, error) {
	c.Log.Print(log.Trace)

	req, err := c.request(ctx, http.MethodPost, c.nodeIP, version, endpoint, payload)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(req)
}

// Put sends a PUT request to the provided Rubrik API endpoint and returns the
// response and status code.
func (c *Client) Put(ctx context.Context, version APIVersion, endpoint string, payload any) ([]byte, int, error) {
	c.Log.Print(log.Trace)

	req, err := c.request(ctx, http.MethodPut, c.nodeIP, version, endpoint, payload)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(req)
}

// Patch sends a PATCH request to the provided Rubrik API endpoint and returns
// the response and status code.
func (c *Client) Patch(ctx context.Context, version APIVersion, endpoint string, payload any) ([]byte, int, error) {
	c.Log.Print(log.Trace)

	req, err := c.request(ctx, http.MethodPatch, c.nodeIP, version, endpoint, payload)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(req)
}

// Delete sends a DELETE request to the provided Rubrik API endpoint and returns
// the response and status code.
func (c *Client) Delete(ctx context.Context, version APIVersion, endpoint string) ([]byte, int, error) {
	c.Log.Print(log.Trace)

	req, err := c.request(ctx, http.MethodDelete, c.nodeIP, version, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(req)
}
