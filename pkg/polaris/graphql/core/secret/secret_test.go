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

package secret

import (
	"reflect"
	"testing"
)

func TestRedact(t *testing.T) {
	v1 := struct {
		F1 string
		F2 []int
	}{F1: "str", F2: []int{1, 2}}
	r1 := Redact(v1)

	if !reflect.DeepEqual(r1, v1) {
		t.Fatalf("invalid result: %v", r1)
	}

	v2 := struct {
		F1 String
		F2 []int
	}{F1: "str", F2: []int{1, 2}}
	r2 := Redact(v2)

	if !reflect.DeepEqual(r2, struct {
		F1 String
		F2 []int
	}{F1: "REDACTED", F2: []int{1, 2}}) {
		t.Fatalf("invalid result: %v", r2)
	}
}

func TestRedactWithAny(t *testing.T) {
	v1 := struct {
		F any
	}{F: "str"}
	r1 := Redact(v1)

	if !reflect.DeepEqual(r1, v1) {
		t.Fatalf("invalid result: %v", r1)
	}

	v2 := struct {
		F any
	}{F: String("str")}
	r2 := Redact(v2)

	if !reflect.DeepEqual(r2, struct {
		F any
	}{F: String("REDACTED")}) {
		t.Fatalf("invalid result: %v", r2)
	}
}
