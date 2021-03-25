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
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// errorResponse returned by Polaris when an error occurs in a GraphQL query.
type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ToError converts the ErrorResponse to a standard Go error or nil if the
// ErrorResponse isn't an error.
func (e errorResponse) ToError() error {
	if e.Code == 0 && e.Message == "" {
		return nil
	}

	return fmt.Errorf("polaris: remote responded with code %d: %q", e.Code, e.Message)
}

type errorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type errorDetail struct {
	Message   string          `json:"message"`
	Path      []string        `json:"path"`
	Locations []errorLocation `json:"locations"`
}

// errorResponseWithDetails returned by Polaris when an error occurs in a
// GraphQL query.
type errorResponseWithDetails struct {
	Data   interface{}   `json:"data"`
	Errors []errorDetail `json:"errors"`
}

// ToError converts the errorResponseWithDetails to a standard Go error or nil
// if the errorResponseWithDetails isn't an error.
func (e errorResponseWithDetails) ToError() error {
	if len(e.Errors) == 0 {
		return nil
	}

	return fmt.Errorf("polaris: remote responded with: %q", e.Errors[0].Message)
}

// Client is used to make GraphQL calls to the Polaris platform.
type Client struct {
	gqlURL     string
	httpClient *http.Client
	log        log.Logger
}

// NewClient returns a new Client with the specified configuration.
func NewClient(apiURL, username, password string, logger log.Logger) *Client {
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
		log: logger,
	}
}

// Request posts the specified GraphQL query with the given variables to the
// Polaris platform. Returns the response JSON text as is.
func (c *Client) Request(ctx context.Context, query string, variables interface{}) ([]byte, error) {
	c.log.Print(log.Trace, "graphql.Request")

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

	// Check if the response matches one of the two error responses.
	var errRes errorResponse
	if err := json.Unmarshal(buf, &errRes); err != nil {
		return nil, err
	}
	if err := errRes.ToError(); err != nil {
		return nil, err
	}

	var errResWithDetails errorResponseWithDetails
	if err := json.Unmarshal(buf, &errResWithDetails); err != nil {
		return nil, err
	}
	if err := errResWithDetails.ToError(); err != nil {
		return nil, err
	}

	return buf, nil
}

// TaskChain is a collection of sequential tasks that all must complete for the
// task chain to be considered complete.
type TaskChain struct {
	ID            int64          `json:"id"`
	TaskChainUUID TaskChainUUID  `json:"taskchainUuid"`
	State         TaskChainState `json:"state"`
	ProgressedAt  time.Time      `json:"progressedAt"`
}

// TaskChainState represents the state of a Polaris task chain.
type TaskChainState string

const (
	TaskChainInvalid   TaskChainState = ""
	TaskChainCanceled  TaskChainState = "CANCELED"
	TaskChainCanceling TaskChainState = "CANCELING"
	TaskChainFailed    TaskChainState = "FAILED"
	TaskChainReady     TaskChainState = "READY"
	TaskChainRunning   TaskChainState = "RUNNING"
	TaskChainSucceeded TaskChainState = "SUCCEEDED"
	TaskChainUndoing   TaskChainState = "UNDOING"
)

// TaskChainUUID represents the identity of a Polaris task chain.
type TaskChainUUID string

// KorgTaskChainStatus returns the task chain for the specified task chain
// UUID.
func (c *Client) KorgTaskChainStatus(ctx context.Context, taskChainID TaskChainUUID) (TaskChain, error) {
	c.log.Print(log.Trace, "graphql.Client.KorgTaskChainStatus")

	buf, err := c.Request(ctx, coreTaskchainStatusQuery, struct {
		TaskChainID string `json:"taskchainId,omitempty"`
	}{TaskChainID: string(taskChainID)})
	if err != nil {
		return TaskChain{}, err
	}

	c.log.Printf(log.Debug, "KorgTaskChainStatus(%q): %s", string(taskChainID), string(buf))

	var payload struct {
		Data struct {
			Query struct {
				TaskChain TaskChain `json:"taskchain"`
			} `json:"getKorgTaskchainStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return TaskChain{}, err
	}

	return payload.Data.Query.TaskChain, nil
}

// WaitForTaskChain blocks until the Polaris task chain with the specified task
// chain ID has completed. When the task chain completes the final state of the
// task chain is returned.
func (c *Client) WaitForTaskChain(ctx context.Context, taskChainID TaskChainUUID) (TaskChainState, error) {
	c.log.Print(log.Trace, "graphql.Client.WaitForTaskChain")

	for {
		taskChain, err := c.KorgTaskChainStatus(ctx, taskChainID)
		if err != nil {
			return TaskChainInvalid, err
		}

		if taskChain.State == TaskChainSucceeded || taskChain.State == TaskChainCanceled || taskChain.State == TaskChainFailed {
			return taskChain.State, nil
		}

		c.log.Print(log.Debug, "Waiting for Polaris task chain")

		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			return TaskChainInvalid, ctx.Err()
		}
	}
}
