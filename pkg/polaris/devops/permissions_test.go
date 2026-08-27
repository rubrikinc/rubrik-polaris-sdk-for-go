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

package devops

import "testing"

// TestSortPermissionJSON verifies that sortPermissionJSON produces a canonical
// document: object keys sorted, array elements ordered by their marshaled bytes,
// and identical content yielding an identical result regardless of input order.
func TestSortPermissionJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{{
		name:  "sorts keys and elements",
		input: `[{"b":"2","a":"1"},{"a":"1","b":"0"}]`,
		want:  `[{"a":"1","b":"0"},{"a":"1","b":"2"}]`,
	}, {
		name:  "different input order yields same result",
		input: `[{"a":"1","b":"2"},{"b":"0","a":"1"}]`,
		want:  `[{"a":"1","b":"0"},{"a":"1","b":"2"}]`,
	}, {
		name:  "empty string passes through",
		input: "",
		want:  "",
	}, {
		name:  "null document",
		input: "null",
		want:  "null",
	}, {
		name:    "invalid JSON",
		input:   "{",
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sortPermissionJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("sortPermissionJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("sortPermissionJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}
