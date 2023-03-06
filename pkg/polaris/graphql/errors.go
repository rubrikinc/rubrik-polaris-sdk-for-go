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

import "errors"

var (
	// ErrNotFound signals that the specified entity could not be found.
	ErrNotFound = errors.New("not found")
)

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

func (e GQLError) Error() string {
	if len(e.Errors) > 0 {
		return e.Errors[0].Message
	}

	return "Unknown GraphQL error"
}
