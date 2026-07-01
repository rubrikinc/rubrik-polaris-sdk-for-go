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

package graphql

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

func TestErrorsWithNoError(t *testing.T) {
	buf := []byte(`{
	    "data": {
	        "awsNativeAccountConnection": {}
        }
	}`)

	var gqlErr GQLError
	if err := json.Unmarshal(buf, &gqlErr); err != nil {
		t.Fatal(err)
	}
	if gqlErr.isError() {
		t.Error("gqlErr should not represent an error")
	}
}

func TestGqlError(t *testing.T) {
	buf, err := os.ReadFile("testdata/graphql_error_response.json")
	if err != nil {
		t.Fatal(err)
	}

	var gqlErr GQLError
	if err := json.Unmarshal(buf, &gqlErr); err != nil {
		t.Fatal(err)
	}
	if !gqlErr.isError() {
		t.Error("gqlErr should represent an error")
	}
	expected := "INTERNAL: invalid status transition of feature CLOUDACCOUNTS from CONNECTED to CONNECTING " +
		"(code: 500, traceId: 9D7LJciYbUSaTTaLQuJcMA==)"
	if msg := gqlErr.Error(); msg != expected {
		t.Fatalf("invalid error message: %v", msg)
	}
}

func TestGQLErrorIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		code      int
		operation string
		want      bool
	}{{
		name:    "account flags error is retryable regardless of operation",
		message: "Error checking account flags to determine access. Please try again.",
		code:    http.StatusForbidden,
		want:    true,
	}, {
		name:    "transient transport code is retryable regardless of operation",
		message: "UNAVAILABLE: Connection closed while performing TLS negotiation",
		code:    http.StatusServiceUnavailable,
		want:    true,
	}, {
		name:      "bad gateway is retryable",
		message:   "some error",
		code:      http.StatusBadGateway,
		operation: "someQuery",
		want:      true,
	}, {
		name:      "gateway timeout is retryable",
		message:   "some error",
		code:      http.StatusGatewayTimeout,
		operation: "someQuery",
		want:      true,
	}, {
		name:      "too many requests is retryable",
		message:   "some error",
		code:      http.StatusTooManyRequests,
		operation: "someQuery",
		want:      true,
	}, {
		name:      "internal error is retryable for allowlisted operation",
		message:   "an internal error occurred while trying to communicate with AWS",
		code:      http.StatusInternalServerError,
		operation: "allVpcsByRegionFromAws",
		want:      true,
	}, {
		name:      "unprocessable code is not retryable for allowlisted operation",
		message:   "some error",
		code:      http.StatusUnprocessableEntity,
		operation: "allVpcsByRegionFromAws",
		want:      false,
	}, {
		name:      "internal error is not retryable for non-allowlisted operation",
		message:   "some error",
		code:      http.StatusInternalServerError,
		operation: "someOtherQuery",
		want:      false,
	}, {
		name:      "forbidden without account flags message is not retryable",
		message:   "Account does not have the appropriate features enabled to access the field.",
		code:      http.StatusForbidden,
		operation: "allVpcsByRegionFromAws",
		want:      false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gqlErr := gqlErrorWith(t, tt.message, tt.code)
			if got := gqlErr.isRetryable(tt.operation); got != tt.want {
				t.Errorf("isRetryable(%q) = %v, want %v", tt.operation, got, tt.want)
			}
		})
	}
}

func TestHTTPErrorIsRetryable(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{{
		name:       "service unavailable is retryable",
		statusCode: http.StatusServiceUnavailable,
		want:       true,
	}, {
		name:       "internal server error is not retryable",
		statusCode: http.StatusInternalServerError,
		want:       false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The operation is ignored for httpError.
			e := httpError{statusCode: tt.statusCode}
			if got := e.isRetryable("someQuery"); got != tt.want {
				t.Errorf("isRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// gqlErrorWith returns a GQLError carrying a single error with the given
// message and extension code.
func gqlErrorWith(t *testing.T, message string, code int) GQLError {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"errors": []map[string]any{
			{"message": message, "extensions": map[string]any{"code": code}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var gqlErr GQLError
	if err := json.Unmarshal(body, &gqlErr); err != nil {
		t.Fatal(err)
	}
	return gqlErr
}
