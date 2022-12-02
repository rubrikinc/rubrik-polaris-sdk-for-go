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

// Package graphql provides direct access to the Polaris GraphQL API. Can be
// used to execute both raw queries and prepared low level queries used by the
// high level part of the SDK.
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
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/errors"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testnet"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/token"
)

type gqlLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type gqlDetails struct {
	Message   string        `json:"message"`
	Path      []interface{} `json:"path"`
	Locations []gqlLocation `json:"locations"`
}

// gqlError is returned by Polaris in the body of a response as a JSON document
// when certain types of GraphQL errors occur.
type gqlError struct {
	Data   interface{}  `json:"data"`
	Errors []gqlDetails `json:"errors"`
}

// isError determines if the gqlError unmarshalled from a JSON document
// represents an error or not.
func (e gqlError) isError() bool {
	return len(e.Errors) > 0
}

func (e gqlError) Error() string {
	return e.Errors[0].Message
}

// Client is used to make GraphQL calls to the Polaris platform.
type Client struct {
	Version string
	app     string
	gqlURL  string
	client  *http.Client
	log     log.Logger
}

func NewClient(
	app,
	gqlURL string,
	tokenSource token.Source,
	transport http.RoundTripper,
	logger log.Logger,
	) *Client {
	return &Client{
		app:    app,
		gqlURL: gqlURL,
		client: &http.Client{
			Transport: token.NewRoundTripper(transport, tokenSource),
		},
		log: logger,
	}
}

// NewClientFromLocalUser returns a new Client with the specified configuration.
func NewClientFromLocalUser(app, apiURL, username, password string, logger log.Logger) *Client {
	tokenSource := token.NewLocalUserSource(http.DefaultClient, apiURL, username, password, logger)
	gqlURL := fmt.Sprintf("%s/graphql", apiURL)
	return NewClient(app, gqlURL, tokenSource, http.DefaultTransport, logger)
}

// NewClientFromServiceAccount returns a new Client with the specified
// configuration.
func NewClientFromServiceAccount(app, apiURL, accessTokenURI, clientID, clientSecret string, logger log.Logger) *Client {
	tokenSource := token.NewServiceAccountSource(http.DefaultClient, accessTokenURI, clientID, clientSecret, logger)
	gqlURL := fmt.Sprintf("%s/graphql", apiURL)
	return NewClient(app, gqlURL, tokenSource, http.DefaultTransport, logger)
}

func NewClientFromServiceAccountWithTransport(app, apiURL, accessTokenURI, clientID, clientSecret string, transport http.RoundTripper, logger log.Logger) *Client {
	tokenSource := token.NewServiceAccountSource(http.DefaultClient, accessTokenURI, clientID, clientSecret, logger)
	gqlURL := fmt.Sprintf("%s/graphql", apiURL)
	return NewClient(app, gqlURL, tokenSource, transport, logger)
}

// NewTestClient returns a new Client intended to be used by unit tests.
func NewTestClient(username, password string, logger log.Logger) (*Client, *testnet.TestListener) {
	testClient, listener := testnet.NewPipeNet()
	tokenSource := token.NewLocalUserSource(testClient, "http://test/api", username, password, logger)

	client := &Client{
		gqlURL: "http://test/api/graphql",
		client: &http.Client{
			Transport: token.NewRoundTripper(testClient.Transport, tokenSource),
		},
		log: logger,
	}

	return client, listener
}

// operationName tries to extract the operation name from a query
// e.g.:
//   "mutation RubrikPolarisSDKRequest($input: String!) { foo(input: $input){} }"
// returns:
//   "RubrikPolarisSDKRequest".
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

// Request posts the specified GraphQL query with the given variables to the
// Polaris platform. Returns the response JSON text as is.
func (c *Client) Request(ctx context.Context, query string, variables interface{}) ([]byte, error) {
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

	// Remote responded without a body. For status code 200 this means we are
	// missing the GraphQL response. For an error we have no additional details.
	if res.ContentLength == 0 {
		return nil, fmt.Errorf("graphql response has no body (status code %d)", res.StatusCode)
	}

	buf, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read graphql response body (status code %d): %v", res.StatusCode, err)
	}

	// Verify that the content type of the body is JSON. For status code 200
	// this mean we received something that isn't a GraphQL response. For an
	// error we have no additional JSON details.
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
	var jsonErr errors.JSONError
	if err := json.Unmarshal(buf, &jsonErr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal graphql response body as an error (status code %d): %v",
			res.StatusCode, err)
	}
	if jsonErr.IsError() {
		return nil, fmt.Errorf("graphql response body is an error (status code %d): %v", res.StatusCode, jsonErr)
	}

	var gqlErr gqlError
	if err := json.Unmarshal(buf, &gqlErr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal graphql response body as an error (status code %d): %v",
			res.StatusCode, err)
	}
	if gqlErr.isError() {
		return nil, fmt.Errorf("graphql response body is an error (status code %d): %v", res.StatusCode, gqlErr)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("graphql response has status code: %s", res.Status)
	}

	return buf, nil
}

// Log returns the logger used by the client.
func (c *Client) Log() log.Logger {
	return c.log
}
