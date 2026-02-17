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
	"errors"
	"fmt"

	graphqlaws "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type account struct {
	cloud    graphqlaws.Cloud
	NativeID string
	name     string
	config   *aws.Config
}

// AccountFunc returns an account initialized from the values passed to the
// function creating the AccountFunc.
type AccountFunc func(ctx context.Context) (account, error)

// Config returns an AccountFunc that initializes the account with values from
// the specified AWS configuration and values from the AWS cloud.
func Config(config aws.Config) AccountFunc {
	return func(ctx context.Context) (account, error) {
		id, name, err := awsAccountInfo(ctx, config)
		if err != nil {
			return account{}, fmt.Errorf("failed to access AWS account: %v", err)
		}

		if name == "" {
			name = id
		}

		return account{NativeID: id, name: name, config: &config}, nil
	}
}

// Default returns an AccountFunc that initializes the account with values from
// the default profile (~/.aws/credentials and ~/.aws/config) and the AWS cloud.
// Credentials and region from the profile can be overridden by environment
// variables.
func Default() AccountFunc {
	return ProfileWithRegionAndRole("default", "", "")
}

// DefaultWithRegion returns an AccountFunc that initializes the account with
// values from the default profile (~/.aws/credentials and ~/.aws/config) and
// the AWS cloud. Credentials and region from the profile can be overridden by
// environment variables.
func DefaultWithRegion(region string) AccountFunc {
	return ProfileWithRegionAndRole("default", region, "")
}

// DefaultWithRole returns an AccountFunc that initializes the account with
// values from the default profile (~/.aws/credentials and ~/.aws/config) and
// the AWS cloud. After the account has been initialized it assumes the role
// specified by the role ARN. Credentials and region from the profile can be
// overridden by environment variables.
func DefaultWithRole(roleARN string) AccountFunc {
	return ProfileWithRegionAndRole("default", "", roleARN)
}

// DefaultWithRegionAndRole returns an AccountFunc that initializes the account
// with values from the default profile (~/.aws/credentials and ~/.aws/config)
// and the AWS cloud. After the account has been initialized it assumes the role
// specified by the role ARN. Credentials and region from the profile can be
// overridden by environment variables.
func DefaultWithRegionAndRole(region, roleARN string) AccountFunc {
	return ProfileWithRegionAndRole("default", region, roleARN)
}

// Profile returns an AccountFunc that initializes the account with values from
// the named profile (~/.aws/credentials and ~/.aws/config) and the AWS cloud.
// If the profile specified is "default", credentials and region from the
// profile can be overridden by environment variables.
func Profile(profile string) AccountFunc {
	return ProfileWithRegionAndRole(profile, "", "")
}

// ProfileWithAccountID returns an AccountFunc that initializes the account
// with the specified account ID and values from the named profile
// (~/.aws/credentials and ~/.aws/config). If the profile specified is
// "default", credentials and region from the profile can be overridden by
// environment variables.
func ProfileWithAccountID(profile, accountID string) AccountFunc {
	return func(ctx context.Context) (account, error) {
		return accountFromProfile(ctx, profile, accountID, "", "")
	}
}

// ProfileWithRegion returns an AccountFunc that initializes the account with
// values from the named profile (~/.aws/credentials and ~/.aws/config) and the
// AWS cloud. If the profile specified is "default", credentials and region from
// the profile can be overridden by environment variables.
func ProfileWithRegion(profile, region string) AccountFunc {
	return ProfileWithRegionAndRole(profile, region, "")
}

// ProfileWithRole returns an AccountFunc that initializes the account with
// values from the named profile (~/.aws/credentials and ~/.aws/config) and the
// AWS cloud. After the account has been initialized it assumes the role
// specified by the role ARN. If the profile specified is "default", credentials
// and region from the profile can be overridden by environment variables.
func ProfileWithRole(profile string, roleArn string) AccountFunc {
	return ProfileWithRegionAndRole(profile, "", roleArn)
}

// ProfileWithRegionAndRole returns an AccountFunc that initializes the account
// with values from the named profile (~/.aws/credentials and ~/.aws/config) and
// the AWS cloud. After the account has been initialized it assumes the role
// specified by the role ARN. If the profile specified is "default", credentials
// and region from the profile can be overridden by environment variables.
func ProfileWithRegionAndRole(profile, region, roleARN string) AccountFunc {
	return func(ctx context.Context) (account, error) {
		return accountFromProfile(ctx, profile, "", region, roleARN)
	}
}

// accountFromProfile returns an account initialized from the specified profile,
// the cloud and passed in values.
func accountFromProfile(ctx context.Context, profile, accountID, region, roleARN string) (account, error) {
	// When profileToLoad is the empty string environment variables can be
	// used to override the credentials loaded by LoadDefaultConfig.
	profileToLoad := profile
	if profileToLoad == "default" {
		profileToLoad = ""
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profileToLoad),
		config.WithRegion(region))
	if err != nil {
		return account{}, fmt.Errorf("failed to load profile %q: %v", profile, err)
	}
	if cfg.Region == "" {
		return account{}, errors.New("missing AWS region, used for AWS CloudFormation stack operations")
	}

	if roleARN != "" {
		stsClient := sts.NewFromConfig(cfg)
		cfg.Credentials = aws.NewCredentialsCache(stscreds.NewAssumeRoleProvider(stsClient, roleARN))
	}

	// If the accountID is empty, we use the profile to look up the account ID
	// using the STS service.
	var name string
	if accountID == "" {
		accountID, name, err = awsAccountInfo(ctx, cfg)
		if err != nil {
			return account{}, fmt.Errorf("failed to access AWS account: %v", err)
		}
	}
	if name == "" {
		name = accountID + " : " + profile
	}

	return account{cloud: graphqlaws.CloudStandard, NativeID: accountID, name: name, config: &cfg}, nil
}

// awsAccount returns the account id and name. Note that if the AWS user does
// not have permissions for Organizations the account name will be empty.
func awsAccountInfo(ctx context.Context, config aws.Config) (string, string, error) {
	stsClient := sts.NewFromConfig(config)
	callerID, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", "", fmt.Errorf("failed to get AWS identity from STS: %v", err)
	}

	// Organizations call might fail due to missing permissions.
	orgClient := organizations.NewFromConfig(config)
	info, err := orgClient.DescribeAccount(ctx, &organizations.DescribeAccountInput{AccountId: callerID.Account})
	if err != nil {
		return *callerID.Account, "", nil
	}

	return *callerID.Account, *info.Account.Name, nil
}

// Account returns an AccountFunc that initializes the account with specified
// cloud type and AWS account id.
func Account(cloud, awsAccountID string) AccountFunc {
	return AccountWithName(cloud, awsAccountID, awsAccountID)
}

// AccountWithName returns an AccountFunc that initializes the account with
// specified cloud type, AWS account id and account name.
func AccountWithName(cloud, awsAccountID, name string) AccountFunc {
	return func(ctx context.Context) (account, error) {
		c, err := graphqlaws.ParseCloud(cloud)
		if err != nil {
			return account{}, fmt.Errorf("failed to parse cloud: %s", err)
		}

		if !verifyAccountID(awsAccountID) {
			return account{}, fmt.Errorf("invalid AWS account id")
		}

		return account{cloud: c, NativeID: awsAccountID, name: name}, nil
	}
}
