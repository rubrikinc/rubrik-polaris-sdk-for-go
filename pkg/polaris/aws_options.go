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

package polaris

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// AwsConfigOption accepts AWS configuration options.
type AwsConfigOption interface {
	awsConfig(ctx context.Context, opts *options) error
}

// awsAccount returns the AWS account id and name. Note that if the AWS user
// does not have permissions for Organizations the account name will be empty.
func awsAccount(ctx context.Context, config aws.Config) (string, string, error) {
	stsClient := sts.NewFromConfig(config)
	id, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", "", err
	}

	// Organizations calls might fail due to missing permissions.
	orgClient := organizations.NewFromConfig(config)
	info, err := orgClient.DescribeAccount(ctx, &organizations.DescribeAccountInput{AccountId: id.Account})
	if err != nil {
		return *id.Account, "", nil
	}

	return *id.Account, *info.Account.Name, nil
}

type awsConfigOption struct {
	parse func(context.Context, *options) error
}

func (o *awsConfigOption) awsConfig(ctx context.Context, opts *options) error {
	return o.parse(ctx, opts)
}

func (o *awsConfigOption) id(ctx context.Context, opts *options) error {
	return o.parse(ctx, opts)
}

func (o *awsConfigOption) query(ctx context.Context, opts *options) error {
	return o.parse(ctx, opts)
}

// FromAwsConfig passes the specified AWS configuration as an option to a
// function accepting AwsConfigOption, IDOption or QueryOption as argument.
// When given multiple times to a variadic function the last configuration
// given will be used.
func FromAwsConfig(config aws.Config) *awsConfigOption {
	return &awsConfigOption{func(ctx context.Context, opts *options) error {
		id, name, err := awsAccount(ctx, config)
		if err != nil {
			return err
		}

		opts.awsID = id
		if name != "" {
			opts.name = name
		}

		opts.awsConfig = &config
		return nil
	}}
}

// FromAwsDefault passes the default AWS configuration as an option to a
// function accepting AwsConfigOption, IDOption or QueryOption as argument.
// When given multiple times to a variadic function the last configuration
// given will be used.
func FromAwsDefault() *awsConfigOption {
	return &awsConfigOption{func(ctx context.Context, opts *options) error {
		config, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return err
		}

		id, name, err := awsAccount(ctx, config)
		if err != nil {
			return err
		}

		opts.awsID = id
		if name != "" {
			opts.name = name
		}

		opts.awsConfig = &config
		return nil
	}}
}

// FromAwsProfile passes the AWS configuration identified by the given profile
// as an option to a function accepting AwsConfigOption, IDOption or
// QueryOption as argument. When given multiple times to a variadic function
// the last profile given will be used.
func FromAwsProfile(profile string) *awsConfigOption {
	return &awsConfigOption{func(ctx context.Context, opts *options) error {
		config, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
		if err != nil {
			return err
		}

		id, name, err := awsAccount(ctx, config)
		if err != nil {
			return err
		}

		opts.awsID = id
		if name != "" {
			opts.name = name
		}

		opts.awsConfig = &config
		opts.awsProfile = profile
		return nil
	}}
}

type queryAndIDOption struct {
	parse func(context.Context, *options) error
}

func (o *queryAndIDOption) id(ctx context.Context, opts *options) error {
	return o.parse(ctx, opts)
}

func (o *queryAndIDOption) query(ctx context.Context, opts *options) error {
	return o.parse(ctx, opts)
}

// WithAwsID passes the specified AWS id as an option to a function accepting
// IDOption or QueryOption as argument. When given multiple times to a variadic
// function only the first id will be used. Note that cloud service provider
// specific options that also specifies an id, directly or indirectly, takes
// priority.
func WithAwsID(id string) *queryAndIDOption {
	return &queryAndIDOption{func(ctx context.Context, opts *options) error {
		if len(id) != 12 {
			return errors.New("polaris: invalid length for aws id")
		}
		if opts.awsID == "" {
			opts.awsID = id
		}
		return nil
	}}
}
