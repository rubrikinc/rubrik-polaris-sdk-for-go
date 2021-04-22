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
	"fmt"
	"strings"
	"time"

	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// AwsProtectionFeature represents the protection features of an AWS cloud
// account.
type AwsProtectionFeature string

const (
	// AwsEC2 AWS EC2.
	AwsEC2 AwsProtectionFeature = "EC2"

	// AwsRDS AWS RDS.
	AwsRDS AwsProtectionFeature = "RDS"
)

// AwsAccountFeature AWS account features.
type AwsAccountFeature struct {
	Feature    string
	AwsRegions []string
	RoleArn    string
	StackArn   string
	Status     string
}

// AwsCloudAccount AWS cloud account.
type AwsCloudAccount struct {
	ID       string
	NativeID string
	Name     string
	Message  string
	Features []AwsAccountFeature
}

// awsStackExist returns true if a CloudFormation stack with the specified name
// exists, false otherwise.
func awsStackExist(ctx context.Context, config aws.Config, stackName string) (bool, error) {
	client := cloudformation.NewFromConfig(config)
	_, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{StackName: &stackName})
	if err == nil {
		return true, nil
	}

	doesNotExist := fmt.Sprintf("Stack with id %s does not exist", stackName)
	if strings.HasSuffix(err.Error(), doesNotExist) {
		return false, nil
	}

	return false, err
}

// awsWaitForStack blocks until the CloudFormation stack create/update/delete
// has completed. When the stack operation completes the final state of the
// operation is returned.
func awsWaitForStack(ctx context.Context, config aws.Config, stackName string) (types.StackStatus, error) {
	client := cloudformation.NewFromConfig(config)
	for {
		stacks, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
			StackName: &stackName,
		})
		if err != nil {
			return "", err
		}
		stack := stacks.Stacks[0]

		switch stack.StackStatus {
		case types.StackStatusCreateInProgress, types.StackStatusDeleteInProgress, types.StackStatusUpdateInProgress:
		default:
			return stack.StackStatus, nil
		}

		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

// AwsAccount returns cloud accounts the same way as AwsAccounts but expects
// to return only a single account, otherwise it returns an error.
func (c *Client) AwsAccount(ctx context.Context, queryOpt QueryOption) (AwsCloudAccount, error) {
	c.log.Print(log.Trace, "polaris.Client.AwsAccount")

	accounts, err := c.AwsAccounts(ctx, queryOpt)
	if err != nil {
		return AwsCloudAccount{}, err
	}
	if len(accounts) < 1 {
		return AwsCloudAccount{}, fmt.Errorf("polaris: account %w", ErrNotFound)
	}
	if len(accounts) > 1 {
		return AwsCloudAccount{}, fmt.Errorf("polaris: account %w", ErrNotUnique)
	}

	return accounts[0], nil
}

// AwsAccounts returns all cloud accounts with cloud native protection matching
// the given query option.
func (c *Client) AwsAccounts(ctx context.Context, queryOpt QueryOption) ([]AwsCloudAccount, error) {
	c.log.Print(log.Trace, "polaris.Client.AwsAccounts")

	opts := options{}
	if err := queryOpt.query(ctx, &opts); err != nil {
		return nil, err
	}

	filter := opts.awsID
	if filter == "" {
		filter = opts.name
	}
	gqlAccounts, err := c.gql.AwsCloudAccounts(ctx, filter)
	if err != nil {
		return nil, err
	}

	accounts := make([]AwsCloudAccount, 0, len(gqlAccounts))
	for _, gqlAccount := range gqlAccounts {
		features := make([]AwsAccountFeature, 0, len(gqlAccount.FeatureDetails))
		for _, gqlFeature := range gqlAccount.FeatureDetails {
			features = append(features, AwsAccountFeature{
				Feature:    gqlFeature.Feature,
				AwsRegions: fromPolarisRegionNames(gqlFeature.AwsRegions),
				RoleArn:    gqlFeature.RoleArn,
				StackArn:   gqlFeature.StackArn,
				Status:     gqlFeature.Status,
			})
		}

		accounts = append(accounts, AwsCloudAccount{
			ID:       gqlAccount.AwsCloudAccount.ID,
			NativeID: gqlAccount.AwsCloudAccount.NativeID,
			Name:     gqlAccount.AwsCloudAccount.AccountName,
			Message:  gqlAccount.AwsCloudAccount.Message,
			Features: features,
		})
	}

	return accounts, nil
}

// AwsAccountAdd adds the AWS account identified by the AwsConfigOption to
// Polaris. The optional AddOptions can be used to specify name and regions.
// If name isn't explicitly given AWS Organizations will be used to lookup the
// AWS account name. If that fails the name will be derived from the AWS account
// id and, if available, the profile name. If no regions are given the default
// region for the AWS configuration will be used.
func (c *Client) AwsAccountAdd(ctx context.Context, awsOpt AwsConfigOption, addOpts ...AddOption) error {
	c.log.Print(log.Trace, "polaris.Client.AwsAccountAdd")

	opts := options{}
	if awsOpt == nil {
		return errors.New("polaris: option not allowed to be nil")
	}
	if err := awsOpt.awsConfig(ctx, &opts); err != nil {
		return err
	}
	for _, opt := range addOpts {
		if err := opt.add(ctx, &opts); err != nil {
			return err
		}
	}
	if opts.awsConfig == nil {
		return errors.New("polaris: missing aws configuration")
	}
	if opts.name == "" {
		opts.name = opts.awsID
		if opts.awsProfile != "" {
			opts.name += " : " + opts.awsProfile
		}
	}
	if len(opts.regions) == 0 {
		opts.regions = append(opts.regions, opts.awsConfig.Region)
	}

	cfmName, _, cfmTmplURL, err := c.gql.AwsNativeProtectionAccountAdd(ctx, opts.awsID, opts.name, opts.regions)
	if err != nil {
		return err
	}

	exist, err := awsStackExist(ctx, *opts.awsConfig, cfmName)
	if err != nil {
		return err
	}

	// Create/Update the CloudFormation stack.
	client := cloudformation.NewFromConfig(*opts.awsConfig)
	if exist {
		c.log.Printf(log.Info, "Updating CloudFormation stack: %s", cfmName)

		stack, err := client.UpdateStack(ctx, &cloudformation.UpdateStackInput{
			StackName:    &cfmName,
			TemplateURL:  &cfmTmplURL,
			Capabilities: []types.Capability{types.CapabilityCapabilityIam},
		})
		if err != nil {
			return err
		}

		stackStatus, err := awsWaitForStack(ctx, *opts.awsConfig, *stack.StackId)
		if err != nil {
			return err
		}
		if stackStatus != types.StackStatusUpdateComplete {
			return fmt.Errorf("polaris: failed to update CloudFormation stack: %v", *stack.StackId)
		}
	} else {
		c.log.Printf(log.Info, "Creating CloudFormation stack: %s", cfmName)

		stack, err := client.CreateStack(ctx, &cloudformation.CreateStackInput{
			StackName:    &cfmName,
			TemplateURL:  &cfmTmplURL,
			Capabilities: []types.Capability{types.CapabilityCapabilityIam},
		})
		if err != nil {
			return err
		}

		stackStatus, err := awsWaitForStack(ctx, *opts.awsConfig, *stack.StackId)
		if err != nil {
			return err
		}
		if stackStatus != types.StackStatusCreateComplete {
			return fmt.Errorf("polaris: failed to create CloudFormation stack: %v", *stack.StackId)
		}
	}

	return nil
}

// AwsAccountSetRegions updates the AWS regions for the AWS account identified
// by the ID option.
func (c *Client) AwsAccountSetRegions(ctx context.Context, idOpts IDOption, regions ...string) error {
	c.log.Print(log.Trace, "polaris.Client.AwsAccountSetRegions")

	opts := options{}
	if idOpts == nil {
		return errors.New("polaris: option not allowed to be nil")
	}
	if err := idOpts.id(ctx, &opts); err != nil {
		return err
	}

	if opts.id == "" {
		if opts.awsID == "" {
			return errors.New("polaris: missing account id")
		}

		account, err := c.AwsAccount(ctx, WithAwsID(opts.awsID))
		if err != nil {
			return err
		}

		opts.id = account.ID
	}

	if len(regions) == 0 {
		return errors.New("polaris: missing regions")
	}

	if err := c.gql.AwsCloudAccountSave(ctx, opts.id, toPolarisRegionNames(regions...)); err != nil {
		return err
	}

	return nil
}

// AwsAccountRemove removes the AWS account identified by the AwsConfigOption
// from Polaris. If deleteSnapshots are true the snapshots are deleted otherwise
// they are kept.
func (c *Client) AwsAccountRemove(ctx context.Context, awsOpt AwsConfigOption, deleteSnapshots bool) error {
	c.log.Print(log.Trace, "polaris.Client.AwsAccountRemove")

	opts := options{}
	if awsOpt == nil {
		return errors.New("polaris: option not allowed to be nil")
	}
	if err := awsOpt.awsConfig(ctx, &opts); err != nil {
		return err
	}
	if opts.awsConfig == nil {
		return errors.New("polaris: missing aws configuration")
	}

	account, err := c.AwsAccount(ctx, WithAwsID(opts.awsID))
	if err != nil {
		return err
	}

	jobID, err := c.gql.AwsStartNativeAccountDisableJob(ctx, account.ID, graphql.AwsEC2, deleteSnapshots)
	if err != nil {
		return err
	}

	state, err := c.gql.WaitForTaskChain(ctx, jobID, 10*time.Second)
	if err != nil {
		return err
	}
	if state != graphql.TaskChainSucceeded {
		return fmt.Errorf("polaris: taskchain failed: jobID=%v, state=%v", jobID, state)
	}

	cfmURL, err := c.gql.AwsCloudAccountDeleteInitiate(ctx, account.ID)
	if err != nil {
		return err
	}

	c.log.Printf(log.Debug, "cfmURL: %s", cfmURL)

	// Check if there are more features than cloud native protection and get
	// the stack Arn/ID.
	var stackArn string
	var protRDS bool
	for _, feature := range account.Features {
		switch feature.Feature {
		case "CLOUD_NATIVE_PROTECTION":
			stackArn = feature.StackArn
		case "RDS_PROTECTION":
			protRDS = true
		}
	}

	// Remove/Update CloudFormation stack.
	client := cloudformation.NewFromConfig(*opts.awsConfig)
	if protRDS {
		c.log.Printf(log.Info, "Updating CloudFormation stack: %s", stackArn)

		usePrevious := true
		_, err := client.UpdateStack(ctx, &cloudformation.UpdateStackInput{
			StackName:           &stackArn,
			UsePreviousTemplate: &usePrevious,
			Capabilities:        []types.Capability{types.CapabilityCapabilityIam},
		})
		if err != nil {
			return err
		}

		stackStatus, err := awsWaitForStack(ctx, *opts.awsConfig, stackArn)
		if err != nil {
			return err
		}
		if stackStatus != types.StackStatusUpdateComplete {
			return fmt.Errorf("polaris: failed to update CloudFormation stack: %v", stackArn)
		}
	} else {
		c.log.Printf(log.Info, "Deleting CloudFormation stack: %s", stackArn)

		_, err = client.DeleteStack(ctx, &cloudformation.DeleteStackInput{
			StackName: &stackArn,
		})
		if err != nil {
			return err
		}

		stackStatus, err := awsWaitForStack(ctx, *opts.awsConfig, stackArn)
		if err != nil {
			return err
		}
		if stackStatus != types.StackStatusDeleteComplete {
			return fmt.Errorf("polaris: failed to delete CloudFormation stack: %v", stackArn)
		}
	}

	if err := c.gql.AwsCloudAccountDeleteProcess(ctx, account.ID); err != nil {
		return err
	}

	return nil
}
