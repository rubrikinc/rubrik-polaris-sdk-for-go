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

// Package graphql contains code to interact directly with the Polaris GraphQL
// API. Can be used to execute both raw queries and prepared low level queries
// used by the high level part of the SDK.
//
// The graphql package tries to stay as close as possible to the GraphQL API:
//
// - Get as a prefix is dropped.
//
// - Names are turned into CamelCase.
//
// - CSP ackronyms (e.g. aws) that are part of names are turned into prefixes.
//
// - Query parameters with values in a well defined range are turned into
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
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

var (
	// ErrAlreadyEnabled signals that the specified feature has  already been
	// enabled.
	ErrAlreadyEnabled = errors.New("already enabled")

	// ErrNotFound signals that the specified entity could not be found.
	ErrNotFound = errors.New("not found")

	// ErrNotUnique signals that a request did not result in a unique entity.
	ErrNotUnique = errors.New("not unique")
)

// jsonError is returned by Polaris in the body of a response as a JSON
// document when certain types of errors occur.
type jsonError struct {
	Code    int    `json:"code"`
	URI     string `json:"uri"`
	Message string `json:"message"`
}

// isError determines if the jsonError unmarshallad from a JSON document
// represents an error or not.
func (e jsonError) isError() bool {
	return e.Code != 0 || e.Message != ""
}

func (e jsonError) Error() string {
	return fmt.Sprintf("polaris: code %d: %s", e.Code, e.Message)
}

type gqlLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type gqlDetails struct {
	Message   string        `json:"message"`
	Path      []string      `json:"path"`
	Locations []gqlLocation `json:"locations"`
}

// gqlError is returned by Polaris in the body of a response as a JSON document
// when certain types of GraphQL errors occur.
type gqlError struct {
	Data   interface{}  `json:"data"`
	Errors []gqlDetails `json:"errors"`
}

// isError determines if the gqlError unmarshallad from a JSON document
// represents an error or not.
func (e gqlError) isError() bool {
	return len(e.Errors) > 0
}

func (e gqlError) Error() string {
	return fmt.Sprintf("polaris: %s", e.Errors[0].Message)
}

// Client is used to make GraphQL calls to the Polaris platform.
type Client struct {
	Version string
	app     string
	gqlURL  string
	client  *http.Client
	log     log.Logger
}

// NewClientFromLocalUser returns a new Client with the specified configuration.
func NewClientFromLocalUser(ctx context.Context, app, apiURL, username, password string, logger log.Logger) *Client {
	return &Client{
		app:    app,
		gqlURL: fmt.Sprintf("%s/graphql", apiURL),
		client: &http.Client{
			Transport: &tokenTransport{
				next: http.DefaultTransport,
				src:  newLocalUserSource(apiURL, username, password),
			},
		},
		log: logger,
	}
}

// NewClientFromServiceAccount returns a new Client with the specified configuration.
func NewClientFromServiceAccount(ctx context.Context, app, apiURL, accessTokenURI, clientID, clientSecret string, logger log.Logger) *Client {
	return &Client{
		app:    app,
		gqlURL: fmt.Sprintf("%s/graphql", apiURL),
		client: &http.Client{
			Transport: &tokenTransport{
				next: http.DefaultTransport,
				src:  newServiceAccountSource(accessTokenURI, clientID, clientSecret),
			},
		},
		log: logger,
	}
}

// NewTestClient - Intended to be used by unit tests.
func NewTestClient(username, password string, logger log.Logger) (*Client, *TestListener) {
	src, lis := newLocalUserTestSource(username, password)

	client := &Client{
		gqlURL: "http://test/api/graphql",
		client: &http.Client{
			Transport: &tokenTransport{
				next: src.client.Transport,
				src:  src,
			},
		},
		log: logger,
	}

	return client, lis
}

// Request posts the specified GraphQL query with the given variables to the
// Polaris platform. Returns the response JSON text as is.
func (c *Client) Request(ctx context.Context, query string, variables interface{}) ([]byte, error) {
	c.log.Print(log.Trace, "polaris/graphql.Request")

	// Extract operation name from query to pass in the body of the request for
	// metrics.
	var operation string
	i := strings.Index(query, " ")
	j := strings.Index(query, "(")
	if i != -1 && j != -1 {
		operation = query[i+1 : j]
	}

	// Prepare the query request body.
	buf, err := json.Marshal(struct {
		Query     string      `json:"query"`
		Variables interface{} `json:"variables,omitempty"`
		Operation string      `json:"operationName,omitempty"`
	}{Query: query, Variables: variables, Operation: operation})
	if err != nil {
		return nil, err
	}

	// Send the query to the remote API endpoint.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.gqlURL, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Remote responded without a body. For status code 200 this means we are
	// missing the GraphQL response. For an error we have no additional details.
	if res.ContentLength == 0 {
		if res.StatusCode == 200 {
			return nil, errors.New("polaris: no body")
		}
		return nil, fmt.Errorf("polaris: %s", res.Status)
	}

	buf, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Verify that the content type of the body is JSON. For status code 200
	// this mean we received something that isn't a GraphQL response. For an
	// error we have no additional JSON details.
	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		if res.StatusCode == 200 {
			return nil, fmt.Errorf("polaris: wrong content-type: %s", contentType)
		}
		return nil, fmt.Errorf("polaris: %s - %s", res.Status, string(buf))
	}

	// Remote responded with a JSON document. Try to parse it as both known
	// error message formats.
	var jsonErr jsonError
	if err := json.Unmarshal(buf, &jsonErr); err != nil {
		return nil, err
	}
	if jsonErr.isError() {
		return nil, jsonErr
	}

	var gqlErr gqlError
	if err := json.Unmarshal(buf, &gqlErr); err != nil {
		return nil, err
	}
	if gqlErr.isError() {
		return nil, gqlErr
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("polaris: %s", res.Status)
	}

	return buf, nil
}

// Log returns the logger used by the client.
func (c *Client) Log() log.Logger {
	return c.log
}
