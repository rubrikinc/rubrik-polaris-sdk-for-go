// Copyright 2026 Rubrik, Inc.
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

package core

import "time"

// AsyncRequestStatus is the status of an asynchronous request submitted to a
// Rubrik cluster.
type AsyncRequestStatus struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Progress  float64           `json:"progress"`
	NodeID    string            `json:"nodeId"`
	StartTime time.Time         `json:"startTime"`
	EndTime   time.Time         `json:"endTime"`
	Links     []Link            `json:"links"`
	Error     *RequestErrorInfo `json:"error"`
}

// Link is a REST link to a resource on a Rubrik cluster.
type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

// RequestErrorInfo holds the error message of a failed asynchronous request.
type RequestErrorInfo struct {
	Message string `json:"message"`
}
