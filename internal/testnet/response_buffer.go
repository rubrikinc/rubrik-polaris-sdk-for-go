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

package testnet

import (
	"bytes"
	"net/http"
)

// responseBuffer is a http.ResponseWriter that buffers the response, so that a
// different response can be sent if an error occurs while generating the
// response.
type responseBuffer struct {
	header     http.Header
	statusCode int
	buffer     *bytes.Buffer
}

func newResponseBuffer(w http.ResponseWriter) *responseBuffer {
	return &responseBuffer{
		header:     w.Header().Clone(),
		statusCode: 200,
		buffer:     &bytes.Buffer{},
	}
}

func (rb *responseBuffer) Header() http.Header {
	return rb.header
}

func (rb *responseBuffer) Write(data []byte) (int, error) {
	return rb.buffer.Write(data)
}

func (rb *responseBuffer) WriteHeader(statusCode int) {
	rb.statusCode = statusCode
}

func (rb *responseBuffer) copyTo(w http.ResponseWriter) {
	h := w.Header()
	for key, val := range rb.header {
		h[key] = val
	}

	w.WriteHeader(rb.statusCode)
	w.Write(rb.buffer.Bytes())
}
