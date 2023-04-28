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

package errors

import (
	"encoding/json"
	"os"
	"testing"
)

func TestJSONError(t *testing.T) {
	buf, err := os.ReadFile("testdata/error_json_from_auth.json")
	if err != nil {
		t.Fatal(err)
	}

	var jsonErr1 JSONError
	if err := json.Unmarshal(buf, &jsonErr1); err != nil {
		t.Fatal(err)
	}
	if !jsonErr1.IsError() {
		t.Error("jsonErr1 should represent an error")
	}
	expected := "JWT validation failed: Missing or invalid credentials (code: 16)"
	if msg := jsonErr1.Error(); msg != expected {
		t.Errorf("invalid error message: %v", msg)
	}

	buf, err = os.ReadFile("testdata/error_json_from_polaris.json")
	if err != nil {
		t.Fatal(err)
	}

	var jsonErr2 JSONError
	if err := json.Unmarshal(buf, &jsonErr2); err != nil {
		t.Fatal(err)
	}
	if !jsonErr2.IsError() {
		t.Error("jsonErr2 should represent an error")
	}
	expected = "UNAUTHENTICATED: wrong username or password (code: 401, traceId: n2jJpBU8qkEy3k09s9JNkg==)"
	if msg := jsonErr2.Error(); msg != expected {
		t.Errorf("invalid error message: %v", msg)
	}
}

func TestJSONErrorsWithNoError(t *testing.T) {
	buf := []byte(`{
	    "data": {
	        "awsNativeAccountConnection": {}
        }
	}`)

	var jsonErr JSONError
	if err := json.Unmarshal(buf, &jsonErr); err != nil {
		t.Fatal(err)
	}
	if jsonErr.IsError() {
		t.Error("jsonErr should not represent an error")
	}
}
