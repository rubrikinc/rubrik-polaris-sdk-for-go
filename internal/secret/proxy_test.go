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

var options = &redactOptions{
	redactText: "R",
}

func TestProxyArray(t *testing.T) {
	v := [2]string{"str1", "str2"}
	r := proxyArray(reflect.ValueOf(v), options)

	if !reflect.DeepEqual(r.Interface(), v) {
		t.Fatalf("invalid result: %v", r)
	}
}

func TestProxyArrayWithRedaction(t *testing.T) {
	v := [2]String{"str1", "str2"}
	r := proxyArray(reflect.ValueOf(v), options)

	if !reflect.DeepEqual(r.Interface(), [2]String{"R", "R"}) {
		t.Fatalf("invalid result: %v", r.Interface())
	}
}

func TestProxyInterface(t *testing.T) {
	v := struct {
		F any
	}{F: "str"}
	r := proxyStruct(reflect.ValueOf(v), options)

	if !reflect.DeepEqual(r.Interface(), v) {
		t.Fatalf("invalid result: %v", r)
	}
}

func TestProxyWithRedaction(t *testing.T) {
	v := struct {
		F any
	}{F: String("str")}
	r := proxyStruct(reflect.ValueOf(v), options)

	if !reflect.DeepEqual(r.Interface(), struct {
		F any
	}{F: String("R")}) {
		t.Fatalf("invalid result: %v", r)
	}
}

func TestProxyMap(t *testing.T) {
	v1 := map[int]string{1: "str1", 2: "str2"}
	r1 := proxyMap(reflect.ValueOf(v1), options)

	if !reflect.DeepEqual(r1.Interface(), v1) {
		t.Fatalf("invalid result: %v", r1)
	}

	v2 := map[string]int{"str1": 1, "str2": 2}
	r2 := proxyMap(reflect.ValueOf(v2), options)

	if !reflect.DeepEqual(r2.Interface(), v2) {
		t.Fatalf("invalid result: %v", r2)
	}
}

func TestProxyMapWithRedaction(t *testing.T) {
	v1 := map[int]String{1: "str1", 2: "str2"}
	r1 := proxyMap(reflect.ValueOf(v1), options)

	if !reflect.DeepEqual(r1.Interface(), map[int]String{1: "R", 2: "R"}) {
		t.Fatalf("invalid result: %v", r1.Interface())
	}

	v2 := map[String]int{"str1": 1, "str2": 2}
	r2 := proxyMap(reflect.ValueOf(v2), options)

	m := r2.Interface().(map[String]int)
	if m["R"] != 1 && m["R"] != 2 {
		t.Fatalf("invalid result: %v", r2.Interface())
	}
}

func TestProxyPointer(t *testing.T) {
	v := "str"
	r := proxyPointer(reflect.ValueOf(&v), options)

	if p := r.Interface().(*string); *p != v {
		t.Fatalf("invalid result: %v", *p)
	}
}

func TestProxyPointerWithRedaction(t *testing.T) {
	v := String("str")
	r := proxyPointer(reflect.ValueOf(&v), options)

	if p := r.Interface().(*String); *p == v {
		t.Fatalf("invalid result: %v", *p)
	}
}

func TestProxySlice(t *testing.T) {
	v := []string{"str1", "str2"}
	r := proxySlice(reflect.ValueOf(v), options)

	if !reflect.DeepEqual(r.Interface(), v) {
		t.Fatalf("invalid result: %v", r)
	}
}

func TestProxySliceWithRedaction(t *testing.T) {
	v := []String{"str1", "str2"}
	r := proxySlice(reflect.ValueOf(v), options)

	if !reflect.DeepEqual(r.Interface(), []String{"R", "R"}) {
		t.Fatalf("invalid result: %v", r.Interface())
	}
}

func TestProxyStruct(t *testing.T) {
	v := struct {
		S1 string
		S2 string
	}{S1: "str1", S2: "str2"}
	r := proxyStruct(reflect.ValueOf(v), options)

	if !reflect.DeepEqual(r.Interface(), v) {
		t.Fatalf("invalid result: %v", r)
	}
}

func TestProxyStructWithRedaction(t *testing.T) {
	v := struct {
		F1 String
		F2 string
	}{F1: "str1", F2: "str2"}
	r := proxyStruct(reflect.ValueOf(v), options)

	if !reflect.DeepEqual(r.Interface(), struct {
		F1 String
		F2 string
	}{F1: "R", F2: "str2"}) {
		t.Fatalf("invalid result: %v", r.Interface())
	}
}

func TestProxyValue(t *testing.T) {
	v1 := "str"
	r1 := proxyValue(reflect.ValueOf(v1), options)

	if r1.Interface().(string) != v1 {
		t.Fatalf("invalid result: %v", r1)
	}

	v2 := 1
	r2 := proxyValue(reflect.ValueOf(v2), options)

	if r2.Interface().(int) != v2 {
		t.Fatalf("invalid result: %v", r2)
	}

	v3 := 2.0
	r3 := proxyValue(reflect.ValueOf(v3), options)

	if r3.Interface().(float64) != v3 {
		t.Fatalf("invalid result: %v", r3)
	}
}

func TestProxyValueWithRedaction(t *testing.T) {
	v := String("str")
	r := proxyValue(reflect.ValueOf(v), options)

	if r.Interface().(String) != "R" {
		t.Fatalf("invalid result: %v", r)
	}
}
