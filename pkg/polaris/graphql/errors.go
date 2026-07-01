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
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

var (
	// ErrNotFound signals that the specified entity could not be found.
	ErrNotFound = errors.New("not found")
)

// retryableGQLCodes maps a GraphQL operation name, as returned by QueryName, to
// the RSC error codes that are transient for that operation and safe to retry.
// Only idempotent read operations belong here, since a retry re-issues the
// request. Unlike the transient HTTP statuses, these codes, such as a catch-all
// 500, are not inherently retryable and must be scoped per operation.
var retryableGQLCodes = map[string][]int{
	"allVpcsByRegionFromAws": {
		http.StatusInternalServerError,
	},
}

// httpError represents an HTTP-level error with a status code.
type httpError struct {
	statusCode int
	msg        string
}

func (e httpError) Error() string {
	return e.msg
}

// isRetryable returns true if the HTTP status code indicates a transient
// condition that may resolve on retry. The operation is ignored because
// transport-level status codes are retryable regardless of the query.
func (e httpError) isRetryable(string) bool {
	return isTransientHTTPStatus(e.statusCode)
}

// GQLError is returned by RSC in the body of a response as a JSON document when
// certain types of GraphQL errors occur.
type GQLError struct {
	Data   interface{} `json:"data"`
	Errors []struct {
		Message   string        `json:"message"`
		Path      []interface{} `json:"path"`
		Locations []struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"locations"`
		Extensions struct {
			Code  int `json:"code"`
			Trace struct {
				Operation string `json:"operation"`
				TraceID   string `json:"traceId"`
				SpanID    string `json:"spanId"`
			} `json:"trace"`
		} `json:"extensions"`
	} `json:"errors"`
}

// isError determines if the gqlError unmarshalled from a JSON document
// represents an error or not.
func (e GQLError) isError() bool {
	return len(e.Errors) > 0
}

// isRetryable returns true if the error represents a transient condition that
// may resolve when the given operation is retried.
func (e GQLError) isRetryable(operation string) bool {
	// Intrinsically-transient message, independent of the operation. Kept as a
	// string match because it is a 403 whose message, unlike the code, is not
	// masked in production and is the only signal that distinguishes it from a
	// legitimate access denial carrying the same 403 code.
	for _, err := range e.Errors {
		if strings.HasPrefix(err.Message, "Error checking account flags to determine access. Please try again.") {
			return true
		}
	}

	// Transient transport codes, such as a gRPC UNAVAILABLE mapped to 503,
	// reported in the GraphQL response body rather than as an HTTP-level error.
	// Independent of the operation and preserved in production, unlike the
	// error message.
	if isTransientHTTPStatus(e.Code()) {
		return true
	}

	// Codes that are transient only for specific idempotent operations.
	if codes, ok := retryableGQLCodes[operation]; ok && slices.Contains(codes, e.Code()) {
		return true
	}

	return false
}

// Code returns the status code of the first error with a non-zero extension
// status code. If no error has a non-zero status code, 0 is returned.
func (e GQLError) Code() int {
	for _, err := range e.Errors {
		if err.Extensions.Code != 0 {
			return err.Extensions.Code
		}
	}

	return 0
}

func (e GQLError) Error() string {
	if len(e.Errors) > 0 {
		err := e.Errors[0]
		return fmt.Sprintf("%s (code: %d, traceId: %s)",
			err.Message, err.Extensions.Code, err.Extensions.Trace.TraceID)
	}

	return "Unknown GraphQL error"
}

// isTransientHTTPStatus returns true if the HTTP status code indicates a
// transient condition that may resolve on retry, independent of the operation:
// 502 Bad Gateway, 503 Service Unavailable, 504 Gateway Timeout and 429 Too
// Many Requests.
func isTransientHTTPStatus(code int) bool {
	switch code {
	case http.StatusBadGateway, http.StatusServiceUnavailable,
		http.StatusGatewayTimeout, http.StatusTooManyRequests:
		return true
	default:
		return false
	}
}
