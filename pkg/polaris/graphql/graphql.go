//go:generate go run queries_gen.go

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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ErrorResponse returned by Polaris when an error occurs in a GraphQL query.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ToError converts the ErrorResponse to a standard Go error or nil if the
// ErrorResponse isn't an error.
func (e ErrorResponse) ToError() error {
	if e.Code == 0 && e.Message == "" {
		return nil
	}

	return fmt.Errorf("polaris: remote responded with code %d: %q", e.Code, e.Message)
}

// Client is used to make GraphQL calls to the Polaris platform.
type Client struct {
	gqlURL     string
	httpClient *http.Client
}

// NewClient returns a new Client with the specified configuration.
func NewClient(apiURL, username, password string) *Client {
	src := tokenSource{
		tokenURL: fmt.Sprintf("%s/session", apiURL),
		username: username,
		password: password,
	}

	return &Client{
		gqlURL: fmt.Sprintf("%s/graphql", apiURL),
		httpClient: &http.Client{
			Transport: &transport{
				next: http.DefaultTransport,
				src:  src,
			},
		},
	}
}

// Post the specified query with the given variables to Polaris. Returns the
// response JSON text as is.
func (c *Client) Request(ctx context.Context, query string, variables interface{}) ([]byte, error) {
	buf, err := json.Marshal(struct {
		Query     string      `json:"query"`
		Variables interface{} `json:"variables,omitempty"`
	}{Query: query, Variables: variables})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.gqlURL, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return nil, fmt.Errorf("error: wrong content type for response: %s", contentType)
	}

	buf, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var errRes ErrorResponse
	if err := json.Unmarshal(buf, &errRes); err != nil {
		return nil, err
	}
	if err := errRes.ToError(); err != nil {
		return nil, err
	}

	return buf, nil
}
