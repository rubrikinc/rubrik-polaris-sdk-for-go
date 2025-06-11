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

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// errorMessage returns an error message from the specified response and HTTP
// status code.
func errorMessage(res []byte, code int) string {
	msg := fmt.Sprintf("%s (%d)", http.StatusText(code), code)

	var cdmErr struct {
		Type    string `json:"errorType"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(res, &cdmErr); err == nil {
		if cdmErr.Type != "" && cdmErr.Message != "" {
			return fmt.Sprintf("%s: %s: %s", msg, cdmErr.Type, cdmErr.Message)
		}
	}

	if res := string(res); res != "" {
		msg = fmt.Sprintf("%s: %s", msg, res)
	}

	return msg
}
