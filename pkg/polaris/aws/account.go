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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type account struct {
	id     string
	name   string
	config aws.Config
}

// AccountFunc returns an account initialized from the values passed to the
// function creating the AccountFunc.
type AccountFunc func(ctx context.Context) (account, error)

// awsAccountID returns the account id.
func awsAccountID(ctx context.Context, config aws.Config) (string, error) {
	stsClient := sts.NewFromConfig(config)
	id, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}

	return *id.Account, nil
}

// awsAccount returns the account id and name. Note that if the AWS user does
// not have permissions for Organizations the account name will be empty.
func awsAccount(ctx context.Context, config aws.Config) (string, string, error) {
	id, err := awsAccountID(ctx, config)
	if err != nil {
		return "", "", err
	}

	// Organizations calls might fail due to missing permissions.
	orgClient := organizations.NewFromConfig(config)
	info, err := orgClient.DescribeAccount(ctx, &organizations.DescribeAccountInput{AccountId: &id})
	if err != nil {
		return id, "", nil
	}

	return id, *info.Account.Name, nil
}

// Config returns an AccountFunc that initializes the account with values from
// the specified config and the cloud.
func Config(config aws.Config) AccountFunc {
	return func(ctx context.Context) (account, error) {
		id, name, err := awsAccount(ctx, config)
		if err != nil {
			return account{}, err
		}

		if name == "" {
			name = id
		}

		return account{id: id, name: name, config: config}, nil
	}
}

// Default returns an AccountFunc that initializes the account with values from
// the default credentials, the default region and the cloud.
func Default() AccountFunc {
	return func(ctx context.Context) (account, error) {
		config, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return account{}, err
		}

		id, name, err := awsAccount(ctx, config)
		if err != nil {
			return account{}, err
		}

		if name == "" {
			name = id + " : default"
		}

		return account{id: id, name: name, config: config}, nil
	}
}

// Profile returns an AccountFunc that initializes the account with values from
// the specified profile, the default region and the cloud.
func Profile(profile string) AccountFunc {
	return ProfileAndRegion(profile, "")
}

// Profile returns an AccountFunc that initializes the account with values from
// the specified profile, the given region and the cloud.
func ProfileAndRegion(profile, region string) AccountFunc {
	return func(ctx context.Context) (account, error) {
		config, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile),
			config.WithRegion(region))
		if err != nil {
			return account{}, err
		}

		id, name, err := awsAccount(ctx, config)
		if err != nil {
			return account{}, err
		}

		// Derive name from AWS id and profile.
		if name == "" {
			name = id + " : " + profile
		}

		return account{id: id, name: name, config: config}, nil
	}
}
