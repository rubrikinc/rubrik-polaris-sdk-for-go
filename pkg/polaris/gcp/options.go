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

package gcp

import (
	"context"
)

type options struct {
	name    string
	orgName string
}

// OptionFunc gives the value passed to the function creating the OptionFunc
// to the specified options instance.
type OptionFunc func(ctx context.Context, opts *options) error

// Name returns an OptionFunc that gives the specified name to the options
// instance.
func Name(name string) OptionFunc {
	return func(ctx context.Context, opts *options) error {
		opts.name = name
		return nil
	}
}

// Organization returns an OptionFunc that gives the specified organization
// name to the options instance.
func Organization(name string) OptionFunc {
	return func(ctx context.Context, opts *options) error {
		opts.orgName = name
		return nil
	}
}
