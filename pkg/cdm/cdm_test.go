// Copyright 2025 Rubrik, Inc.
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

import "testing"

func TestErrorMessage(t *testing.T) {
	testCases := []struct {
		name string
		res  []byte
		code int
		msg  string
	}{{
		name: "EmptyResponse",
		res:  []byte(``),
		code: 404,
		msg:  "Not Found (404)",
	}, {
		name: "TextResponse",
		res:  []byte(`error message`),
		code: 500,
		msg:  "Internal Server Error (500): error message",
	}, {
		name: "EmptyJSON",
		res:  []byte(`{}`),
		code: 503,
		msg:  "Service Unavailable (503): {}",
	}, {
		name: "JSONResponse",
		res:  []byte(`{"name":"value"}`),
		code: 403,
		msg:  `Forbidden (403): {"name":"value"}`,
	}, {
		name: "PartialJSONErrorResponse",
		res:  []byte(`{"errorType":"errorType"}`),
		code: 500,
		msg:  `Internal Server Error (500): {"errorType":"errorType"}`,
	}, {
		name: "FullJSONErrorResponse",
		res:  []byte(`{"errorType":"errorType","message":"message"}`),
		code: 404,
		msg:  "Not Found (404): errorType: message",
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if msg := errorMessage(testCase.res, testCase.code); msg != testCase.msg {
				t.Fatalf("%s != %s", msg, testCase.msg)
			}
		})
	}
}
