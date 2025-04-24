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
)

// String is a type that can be redacted.
type String string

// Redact returns a copy of value where all String values of exported fields
// have been replaced with the redaction text. The redaction text defaults to
// "REDACTED".
func Redact[T any](value T, options ...Option) T {
	// Untyped nil value.
	switch any(value).(type) {
	case nil:
		return value
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Chan || v.Kind() == reflect.Func {
		return value
	}

	// Typed nil value.
	if (v.Kind() == reflect.Interface || v.Kind() == reflect.Map || v.Kind() == reflect.Pointer || v.Kind() == reflect.Slice) && v.IsNil() {
		return value
	}

	ro := &redactOptions{
		debugMode:  false,
		redactText: "REDACTED",
	}

	for _, o := range options {
		switch v := o.(type) {
		case DebugMode:
			ro.debugMode = bool(v)
		case RedactionText:
			ro.redactText = string(v)
		}
	}

	return proxyValue(v, ro).Interface().(T)
}
