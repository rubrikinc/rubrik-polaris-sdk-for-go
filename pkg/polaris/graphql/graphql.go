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

// Package graphql provides direct access to the RSC GraphQL API. Can be used
// to execute both raw queries and prepared low-level queries used by the
// high-level part of the SDK.
//
// The graphql package tries to stay as close as possible to the GraphQL API:
//
// - Get as a prefix is dropped.
//
// - Names are turned into CamelCase.
//
// - CSP acronyms (e.g. aws) that are part of names are removed.
//
// - Query parameters with values in a well-defined range are turned into
// types.
package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	internalerrors "github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/errors"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/token"
)

// Client is used to make GraphQL calls to the Polaris platform.
type Client struct {
	Version string // Deprecated: use DeploymentVersion.
	gqlURL  string
	client  *http.Client
	log     log.Logger
}

// NewClient returns a new Client for the specified API URL.
func NewClient(apiURL string, tokenSource token.Source) *Client {
	return NewClientWithLogger(apiURL, tokenSource, &log.DiscardLogger{})
}

// NewClientWithLogger returns a new Client for the specified API URL, logging
// to the given logger.
func NewClientWithLogger(apiURL string, tokenSource token.Source, logger log.Logger) *Client {
	logger.Printf(log.Info, "Polaris API URL: %s", apiURL)

	client := &Client{
		gqlURL: apiURL + "/graphql",
		client: &http.Client{
			Transport: token.NewRoundTripper(http.DefaultTransport, tokenSource),
		},
		log: logger,
	}

	return client
}

// Deprecated: use NewClientWithLogger.
func NewClientFromLocalUser(app, apiURL, username, password string, logger log.Logger) *Client {
	tokenSource := token.NewUserSourceWithLogger(http.DefaultClient, apiURL, username, password, logger)

	return NewClient(apiURL, tokenSource)
}

// Deprecated: use NewClientWithLogger.
func NewClientFromServiceAccount(app, apiURL, accessTokenURI, clientID, clientSecret string, logger log.Logger) *Client {
	tokenSource := token.NewServiceAccountSourceWithLogger(http.DefaultClient, accessTokenURI, clientID, clientSecret, logger)

	return NewClient(apiURL, tokenSource)
}

// NewTestClient returns a new Client intended to be used by unit tests.
func NewTestClient(testServer *httptest.Server) *Client {
	tokenSource := token.NewTestUserSource(testServer, "username", "password")
	return NewTestClientWithTokenSource(testServer, tokenSource)
}

// NewTestClientWithTokenSource returns a new Client with a custom token source
// intended to be used by unit tests.
func NewTestClientWithTokenSource(testServer *httptest.Server, tokenSource token.Source) *Client {
	return &Client{
		gqlURL: testServer.URL + "/api/graphql",
		client: &http.Client{
			Transport: token.NewRoundTripper(testServer.Client().Transport, tokenSource),
		},
		log: &log.DiscardLogger{},
	}
}

// DeploymentVersion returns the deployed version of RSC.
func (c *Client) DeploymentVersion(ctx context.Context) (Version, error) {
	c.log.Print(log.Trace)

	buf, err := c.Request(ctx, "query SdkGolangDeploymentVersion { deploymentVersion }", struct{}{})
	if err != nil {
		return "", fmt.Errorf("failed to request deploymentVersion: %w", err)
	}
	c.log.Printf(log.Debug, "deploymentVersion(): %s", string(buf))

	var payload struct {
		Data struct {
			DeploymentVersion string `json:"deploymentVersion"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", fmt.Errorf("failed to unmarshal deploymentVersion: %v", err)
	}

	return Version(payload.Data.DeploymentVersion), nil
}

// Log returns the logger.
func (c *Client) Log() log.Logger {
	return c.log
}

// SetLogger sets the logger to use.
func (c *Client) SetLogger(logger log.Logger) {
	c.log = logger
}

const requestRetryAttempts = 10

// Request posts the specified GraphQL query/mutation with the given variables
// to the Polaris platform. Returns the response JSON text as is. If the request
// fails due to temporary error, it will be retried automatically.
func (c *Client) Request(ctx context.Context, query string, variables interface{}) ([]byte, error) {
	c.log.Print(log.Trace)

	// Log variables before calling the query/mutation.
	buf, err := json.Marshal(variables)
	if err != nil {
		buf = []byte(fmt.Sprintf("marshaling of variables failed: %s", err))
	}
	c.log.Printf(log.Debug, "%s params: %s", QueryName(query), string(buf))

	return c.RequestWithoutLogging(ctx, query, variables)
}

// RequestWithoutLogging posts the specified GraphQL query/mutation with the
// given variables to the Polaris platform. Returns the response JSON text as
// is. The variables are not logged before the request is made. Certain
// temporary errors will be retried.
func (c *Client) RequestWithoutLogging(ctx context.Context, query string, variables interface{}) ([]byte, error) {
	c.log.Print(log.Trace)

	retryAttempt := 0
	for {
		buf, err := c.RequestWithoutRetry(ctx, query, variables)

		var gqlErr GQLError
		if errors.As(err, &gqlErr) && gqlErr.isTemporary() {
			if retryAttempt++; retryAttempt > requestRetryAttempts {
				return nil, fmt.Errorf("request failed after %d retries: %w", retryAttempt-1, err)
			}

			c.log.Printf(log.Debug, "Endpoint temporarily unavailable (retry attempt: %d/%d): %s", retryAttempt,
				requestRetryAttempts, err)
			select {
			case <-time.After(10 * time.Second):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		return buf, err
	}
}

// RequestWithoutRetry posts the specified GraphQL query/mutation with the given
// variables to the Polaris platform. Returns the response JSON text as is.
func (c *Client) RequestWithoutRetry(ctx context.Context, query string, variables interface{}) ([]byte, error) {
	c.log.Print(log.Trace)

	// Extract operation name from query to pass in the body of the request for
	// metrics.
	operation := operationName(query)

	// Prepare the query request body.
	buf, err := json.Marshal(struct {
		Query     string      `json:"query"`
		Variables interface{} `json:"variables,omitempty"`
		Operation string      `json:"operationName,omitempty"`
	}{Query: query, Variables: variables, Operation: operation})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal graphql request body: %v", err)
	}

	// Send the query to the remote API endpoint.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.gqlURL, bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("failed to create graphql request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request graphql field: %v", err)
	}
	defer res.Body.Close()

	// Remote responded without a body. For status code 200, this means we
	// are missing the GraphQL response. For an error, we have no additional
	// details.
	if res.ContentLength == 0 {
		return nil, fmt.Errorf("graphql response has no body (status code %d)", res.StatusCode)
	}

	buf, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read graphql response body (status code %d): %v", res.StatusCode, err)
	}

	// Verify that the content type of the body is JSON. For status code 200,
	// this means we received something that isn't a GraphQL response. For an
	// error, we have no additional JSON details.
	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		snippet := string(buf)
		if len(snippet) > 512 {
			snippet = snippet[:512]
		}
		return nil, fmt.Errorf("graphql response has Content-Type %s (status code %d): %q",
			contentType, res.StatusCode, snippet)
	}

	// Remote responded with a JSON document. Try to parse it as both known
	// error message formats.
	var jsonErr internalerrors.JSONError
	if err := json.Unmarshal(buf, &jsonErr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal graphql response body as an error (status code %d): %v",
			res.StatusCode, err)
	}
	if jsonErr.IsError() {
		return nil, fmt.Errorf("graphql response body is an error (status code %d): %w", res.StatusCode, jsonErr)
	}

	var gqlErr GQLError
	if err := json.Unmarshal(buf, &gqlErr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal graphql response body as an error (status code %d): %v",
			res.StatusCode, err)
	}
	if gqlErr.isError() {
		return nil, fmt.Errorf("graphql response body is an error (status code %d): %w", res.StatusCode, gqlErr)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("graphql response has status code: %s", res.Status)
	}

	return buf, nil
}

// LogResponse logs the response from a GraphQL query/mutation.
func LogResponse(logger log.Logger, query string, response []byte) {
	logger.Printf(log.Debug, "%s response: %s", query, string(response))
}

// RequestError returns a standard formatted error detailing the failure when
// calling a query/mutation.
func RequestError(query string, err error) error {
	return fmt.Errorf("failed to request %s: %w", QueryName(query), err)
}

// UnmarshalError returns a standard formatted error detailing the failure to
// unmarshal the response from a GraphQL query/mutation.
func UnmarshalError(query string, err error) error {
	return fmt.Errorf("failed to unmarshal %s: %s", QueryName(query), err)
}

// ResponseError returns a standard formatted error detailing the error received
// in response to a query/mutation.
func ResponseError(query string, err error) error {
	return fmt.Errorf("%s response is an error: %s", QueryName(query), err)
}

// operationName tries to extract the operation name from a query
// e.g.:
//
//	"mutation RubrikPolarisSDKRequest($input: String!) { foo(input: $input){} }"
//
// returns:
//
//	"RubrikPolarisSDKRequest".
//
// TODO: do we need to improve this since it currently is just a best effort extraction?
func operationName(query string) string {
	// Trim leading white spaces
	query = strings.TrimSpace(query)

	// Find index of first GQL whitespace (either space or tab)
	i := strings.Index(query, " ")
	if i2 := strings.Index(query, "\t"); i2 != -1 && i2 < i {
		i = i2
	}

	j := strings.Index(query, "{")
	// Check if it is a query shorthand (or invalid query)
	if i == -1 || j == -1 || j < i {
		return ""
	}

	// Do not include variable definitions
	if k := strings.Index(query[i:j], "("); k != -1 {
		return strings.TrimSpace(query[i : i+k])
	}

	return strings.TrimSpace(query[i:j])
}
