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

package aws

import (
	"context"
	"fmt"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
)

type options struct {
	name    string
	regions []aws.Region
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

// Region returns an OptionFunc that gives the specified region to the options
// instance.
func Region(region string) OptionFunc {
	return func(ctx context.Context, opts *options) error {
		r, err := aws.ParseRegion(region)
		if err != nil {
			return fmt.Errorf("failed to parse region: %v", err)
		}

		opts.regions = append(opts.regions, r)
		return nil
	}
}

// Regions returns an OptionFunc that gives the specified regions to the
// options instance.
func Regions(regions ...string) OptionFunc {
	return func(ctx context.Context, opts *options) error {
		set := make(map[aws.Region]struct{}, len(regions)+len(opts.regions))
		for _, region := range opts.regions {
			set[region] = struct{}{}
		}

		for _, r := range regions {
			region, err := aws.ParseRegion(r)
			if err != nil {
				return fmt.Errorf("failed to parse region: %v", err)
			}

			if _, ok := set[region]; !ok {
				opts.regions = append(opts.regions, region)
				set[region] = struct{}{}
			}
		}

		return nil
	}
}
