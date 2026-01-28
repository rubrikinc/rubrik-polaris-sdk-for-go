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
	pa := proxyArray(reflect.ValueOf(v), options)
	if !reflect.DeepEqual(pa.Interface(), v) {
		t.Fatalf("invalid proxy array value: %v", pa)
	}
}

func TestProxyArrayWithRedaction(t *testing.T) {
	v := [2]String{"str1", "str2"}
	pa := proxyArray(reflect.ValueOf(v), options)
	if !reflect.DeepEqual(pa.Interface(), [2]String{"R", "R"}) {
		t.Fatalf("invalid proxy array value: %v", pa)
	}
}

// TestProxyInterface uses proxyStruct to verify that a nil any field is
// handled correctly. ValueOf a nil interface is not a valid value.
func TestProxyInterface(t *testing.T) {
	var v struct {
		A any
	}
	ps := proxyStruct(reflect.ValueOf(v), options)
	if !reflect.DeepEqual(ps.Interface(), v) {
		t.Fatalf("proxy interface should be nil: %v", ps)
	}

	v = struct {
		A any
	}{A: "str"}
	ps = proxyStruct(reflect.ValueOf(v), options)
	if !reflect.DeepEqual(ps.Interface(), v) {
		t.Fatalf("invalid proxy interface value: %v", ps)
	}
}

// TestProxyInterface uses proxyStruct to verify that a nil any field is
// handled correctly. ValueOf a nil interface is not a valid value.
func TestProxyInterfaceWithRedaction(t *testing.T) {
	v := struct {
		A any
	}{A: String("str")}
	ps := proxyStruct(reflect.ValueOf(v), options)
	if !reflect.DeepEqual(ps.Interface(), struct {
		A any
	}{A: String("R")}) {
		t.Fatalf("invalid proxy interface value: %v", ps)
	}
}

func TestProxyMap(t *testing.T) {
	var v1 map[int]string
	pm := proxyMap(reflect.ValueOf(v1), options)
	if !reflect.DeepEqual(pm.Interface(), v1) {
		t.Fatalf("proxy map should be nil: %v", pm)
	}

	v1 = map[int]string{1: "str1", 2: "str2"}
	pm = proxyMap(reflect.ValueOf(v1), options)
	if !reflect.DeepEqual(pm.Interface(), v1) {
		t.Fatalf("invalid proxy map value: %v", pm)
	}

	var v2 map[string]int
	pm = proxyMap(reflect.ValueOf(v2), options)
	if !reflect.DeepEqual(pm.Interface(), v2) {
		t.Fatalf("proxy map should be nil: %v", pm)
	}

	v2 = map[string]int{"str1": 1, "str2": 2}
	pm = proxyMap(reflect.ValueOf(v2), options)
	if !reflect.DeepEqual(pm.Interface(), v2) {
		t.Fatalf("invalid proxy map value: %v", pm)
	}
}

func TestProxyMapWithRedaction(t *testing.T) {
	v1 := map[int]String{1: "str1", 2: "str2"}
	pm := proxyMap(reflect.ValueOf(v1), options)
	if !reflect.DeepEqual(pm.Interface(), map[int]String{1: "R", 2: "R"}) {
		t.Fatalf("invalid proxy map value: %v", pm)
	}

	v2 := map[String]int{"str1": 1, "str2": 2}
	pm = proxyMap(reflect.ValueOf(v2), options)
	if r := pm.Interface().(map[String]int); r["R"] != 1 && r["R"] != 2 {
		t.Fatalf("invalid proxy map value: %v", pm)
	}
}

func TestProxyPointer(t *testing.T) {
	var v1 *string
	pp := proxyPointer(reflect.ValueOf(v1), options)
	if r := pp.Interface().(*string); r != v1 {
		t.Fatalf("proxy pointer should be nil: %v", pp)
	}

	v2 := "str"
	pp = proxyPointer(reflect.ValueOf(&v2), options)
	if r := pp.Interface().(*string); *r != v2 {
		t.Fatalf("invalid proxy pointer value: %v", pp)
	}
}

func TestProxyPointerWithRedaction(t *testing.T) {
	var v1 *String
	pp := proxyPointer(reflect.ValueOf(v1), options)
	if r := pp.Interface().(*String); r != v1 {
		t.Fatalf("proxy pointer should be nil: %v", pp)
	}

	v2 := String("str")
	pp = proxyPointer(reflect.ValueOf(&v2), options)
	if r := pp.Interface().(*String); *r == v2 {
		t.Fatalf("invalid proxy pointer value: %v", pp)
	}
}

func TestProxySlice(t *testing.T) {
	var v []string
	ps := proxySlice(reflect.ValueOf(v), options)
	if !reflect.DeepEqual(ps.Interface(), v) {
		t.Fatalf("proxy slice should be nil: %v", ps)
	}

	v = []string{"str1", "str2"}
	ps = proxySlice(reflect.ValueOf(v), options)
	if !reflect.DeepEqual(ps.Interface(), v) {
		t.Fatalf("invalid proxy slice value: %v", ps)
	}
}

func TestProxySliceWithRedaction(t *testing.T) {
	var v []String
	ps := proxySlice(reflect.ValueOf(v), options)
	if !reflect.DeepEqual(ps.Interface(), v) {
		t.Fatalf("proxy slice should be nil: %v", ps)
	}

	v = []String{"str1", "str2"}
	ps = proxySlice(reflect.ValueOf(v), options)
	if !reflect.DeepEqual(ps.Interface(), []String{"R", "R"}) {
		t.Fatalf("invalid proxy slice value: %v", ps)
	}
}

func TestProxyStruct(t *testing.T) {
	v := struct {
		S1 string
		S2 string
	}{S1: "str1", S2: "str2"}
	ps := proxyStruct(reflect.ValueOf(v), options)
	if !reflect.DeepEqual(ps.Interface(), v) {
		t.Fatalf("invalid proxy struct value: %v", ps)
	}
}

func TestProxyStructWithRedaction(t *testing.T) {
	v := struct {
		F1 String
		F2 string
	}{F1: "str1", F2: "str2"}
	ps := proxyStruct(reflect.ValueOf(v), options)
	if !reflect.DeepEqual(ps.Interface(), struct {
		F1 String
		F2 string
	}{F1: "R", F2: "str2"}) {
		t.Fatalf("invalid proxy struct value: %v", ps)
	}
}

func TestProxyValue(t *testing.T) {
	v1 := "str"
	pv := proxyValue(reflect.ValueOf(v1), options)
	if pv.Interface().(string) != v1 {
		t.Fatalf("invalid proxy value: %v", pv)
	}

	v2 := 1
	pv = proxyValue(reflect.ValueOf(v2), options)
	if pv.Interface().(int) != v2 {
		t.Fatalf("invalid proxy value: %v", pv)
	}

	v3 := 2.0
	pv = proxyValue(reflect.ValueOf(v3), options)
	if pv.Interface().(float64) != v3 {
		t.Fatalf("invalid proxy value: %v", pv)
	}
}

func TestProxyValueWithRedaction(t *testing.T) {
	v := String("str")
	pv := proxyValue(reflect.ValueOf(v), options)
	if pv.Interface().(String) != "R" {
		t.Fatalf("invalid proxy value: %v", pv)
	}
}
