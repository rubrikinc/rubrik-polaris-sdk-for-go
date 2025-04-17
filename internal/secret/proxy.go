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
	"log"
	"reflect"
)

var stringType = reflect.TypeOf(String(""))

type redactOptions struct {
	debugMode  bool
	redactText string
}

func proxyArray(value reflect.Value, options *redactOptions) reflect.Value {
	if options.debugMode {
		log.Printf("[DEBUG] Proxy array %s", value.Type())
	}

	pv := reflect.New(value.Type()).Elem()
	for i := 0; i < value.Len(); i++ {
		pv.Index(i).Set(proxyValue(value.Index(i), options))
	}

	return pv
}

func proxyInterface(value reflect.Value, options *redactOptions) reflect.Value {
	if options.debugMode {
		log.Printf("[DEBUG] Proxy interface %s", value.Type())
	}

	pv := reflect.New(value.Type()).Elem()
	pv.Set(proxyValue(value.Elem(), options))

	return pv
}

func proxyMap(value reflect.Value, options *redactOptions) reflect.Value {
	if options.debugMode {
		log.Printf("[DEBUG] Proxy map %s", value.Type())
	}

	pv := reflect.MakeMapWithSize(value.Type(), value.Len())
	for i := value.MapRange(); i.Next(); {
		pv.SetMapIndex(proxyValue(i.Key(), options), proxyValue(i.Value(), options))
	}

	return pv
}

func proxyPointer(value reflect.Value, options *redactOptions) reflect.Value {
	if options.debugMode {
		log.Printf("[DEBUG] Proxy pointer %s", value.Type())
	}

	pv := reflect.New(value.Type()).Elem()
	pv.Set(proxyValue(value.Elem(), options).Addr())

	return pv
}

func proxySlice(value reflect.Value, options *redactOptions) reflect.Value {
	if options.debugMode {
		log.Printf("[DEBUG] Proxy slice %s", value.Type())
	}

	pv := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
	for i := 0; i < value.Len(); i++ {
		pv.Index(i).Set(proxyValue(value.Index(i), options))
	}

	return pv
}

func proxyStruct(value reflect.Value, options *redactOptions) reflect.Value {
	if options.debugMode {
		log.Printf("[DEBUG] Proxy struct %s", value.Type())
	}

	pv := reflect.New(value.Type()).Elem()
	for _, f := range reflect.VisibleFields(value.Type()) {
		if f.IsExported() {
			pv.FieldByIndex(f.Index).Set(proxyValue(value.FieldByIndex(f.Index), options))
		}
	}

	return pv
}

func proxyValue(value reflect.Value, options *redactOptions) reflect.Value {
	switch value.Kind() {
	case reflect.Array:
		return proxyArray(value, options)
	case reflect.Interface:
		return proxyInterface(value, options)
	case reflect.Map:
		return proxyMap(value, options)
	case reflect.Ptr:
		return proxyPointer(value, options)
	case reflect.Slice:
		return proxySlice(value, options)
	case reflect.Struct:
		return proxyStruct(value, options)
	default:
		if options.debugMode {
			log.Printf("[DEBUG] Proxy value %s", value.Type())
		}

		if t := value.Type(); t == stringType {
			value = reflect.New(t).Elem()
			value.SetString(options.redactText)
		}

		return value
	}
}
