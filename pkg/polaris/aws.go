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
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"
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

func awsFromPolarisRegionNames(polarisNames []string) []string {
	names := make([]string, 0, len(polarisNames))
	for _, name := range polarisNames {
		names = append(names, strings.ReplaceAll(strings.ToLower(name), "_", "-"))
	}

	return names
}

func awsToPolarisRegionNames(names []string) []string {
	polarisNames := make([]string, 0, len(names))
	for _, name := range names {
		polarisNames = append(polarisNames, strings.ReplaceAll(strings.ToUpper(name), "-", "_"))
	}

	return polarisNames
}

// awsAccountID returns the AWS account id and account name. Note that if the
// AWS user does not have permissions for Organizations the account name will
// be empty.
func (c *Client) awsAccountID(ctx context.Context, config aws.Config) (string, string, error) {
	c.log.Print(log.Trace, "polaris.Client.awsAccountID")

	stsClient := sts.NewFromConfig(config)
	id, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", "", err
	}

	// Organizations calls might fail due to missing permissions, in that case
	// we skip the account name.
	orgClient := organizations.NewFromConfig(config)
	info, err := orgClient.DescribeAccount(ctx, &organizations.DescribeAccountInput{AccountId: id.Account})
	if err != nil {
		c.log.Print(log.Info, "Failed to obtain account name from AWS Organizations")
		return *id.Account, "", nil
	}

	return *id.Account, *info.Account.Name, nil
}

// awsStackExist returns true if a CloudFormation stack with the specified name
// exists, false otherwise.
func (c *Client) awsStackExist(ctx context.Context, config aws.Config, stackName string) (bool, error) {
	c.log.Print(log.Trace, "polaris.Client.awsStackExist")

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
func (c *Client) awsWaitForStack(ctx context.Context, config aws.Config, stackName string) (types.StackStatus, error) {
	c.log.Print(log.Trace, "polaris.Client.awsWaitForStack")

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

		c.log.Print(log.Debug, "Waiting for AWS stack")

		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

// AwsAccounts returns all cloud accounts with cloud native protection under
// the current Polaris account.
func (c *Client) AwsAccounts(ctx context.Context) ([]AwsCloudAccount, error) {
	gqlAccounts, err := c.gql.AwsCloudAccounts(ctx, "")
	if err != nil {
		return nil, err
	}

	accounts := make([]AwsCloudAccount, 0, len(gqlAccounts))
	for _, gqlAccount := range gqlAccounts {
		features := make([]AwsAccountFeature, 0, len(gqlAccount.FeatureDetails))
		for _, gqlFeature := range gqlAccount.FeatureDetails {
			features = append(features, AwsAccountFeature{
				Feature:    gqlFeature.Feature,
				AwsRegions: awsFromPolarisRegionNames(gqlFeature.AwsRegions),
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

// AwsAccounts returns the cloud account with cloud native protection for the
// specified AWS account ID.
func (c *Client) AwsAccountFromID(ctx context.Context, awsAccountID string) (AwsCloudAccount, error) {
	c.log.Print(log.Trace, "polaris.Client.AwsAccountFromID")

	gqlAccounts, err := c.gql.AwsCloudAccounts(ctx, awsAccountID)
	if err != nil {
		return AwsCloudAccount{}, err
	}
	if len(gqlAccounts) < 1 {
		return AwsCloudAccount{}, ErrAccountNotFound
	}
	if len(gqlAccounts) > 1 {
		return AwsCloudAccount{}, errors.New("polaris: aws account id refers to multiple cloud accounts")
	}
	gqlAccount := gqlAccounts[0]

	features := make([]AwsAccountFeature, 0, len(gqlAccount.FeatureDetails))
	for _, gqlFeature := range gqlAccount.FeatureDetails {
		features = append(features, AwsAccountFeature{
			Feature:    gqlFeature.Feature,
			AwsRegions: awsFromPolarisRegionNames(gqlFeature.AwsRegions),
			RoleArn:    gqlFeature.RoleArn,
			StackArn:   gqlFeature.StackArn,
			Status:     gqlFeature.Status,
		})
	}

	account := AwsCloudAccount{
		ID:       gqlAccount.AwsCloudAccount.ID,
		NativeID: gqlAccount.AwsCloudAccount.NativeID,
		Name:     gqlAccount.AwsCloudAccount.AccountName,
		Message:  gqlAccount.AwsCloudAccount.Message,
		Features: features,
	}

	return account, nil
}

// AwsAccounts returns the cloud account with cloud native protection for the
// specified AWS config.
func (c *Client) AwsAccountFromConfig(ctx context.Context, config aws.Config) (AwsCloudAccount, error) {
	c.log.Print(log.Trace, "polaris.Client.AwsAccountDetails")

	// Lookup AWS account id
	awsAccountID, _, err := c.awsAccountID(ctx, config)
	if err != nil {
		return AwsCloudAccount{}, err
	}

	return c.AwsAccountFromID(ctx, awsAccountID)
}

// AwsAccountAdd adds the AWS account referred to by the given AWS config.
// The altName parameter specifies an alternative name for the account in case
// the AWS Organizations lookup of the account name fails to due to missing
// permissions.
func (c *Client) AwsAccountAdd(ctx context.Context, config aws.Config, altName string, awsRegions []string) error {
	c.log.Print(log.Trace, "polaris.Client.AwsAccountAdd")

	// Lookup AWS account id and name.
	awsAccountID, awsAccountName, err := c.awsAccountID(ctx, config)
	if err != nil {
		return err
	}
	if awsAccountName == "" {
		awsAccountName = altName
	}

	cfmName, _, cfmTemplateURL, err := c.gql.AwsNativeProtectionAccountAdd(ctx, awsAccountID, awsAccountName, awsRegions)
	if err != nil {
		return err
	}

	exist, err := c.awsStackExist(ctx, config, cfmName)
	if err != nil {
		return err
	}

	// Create/Update the CloudFormation stack.
	client := cloudformation.NewFromConfig(config)
	if exist {
		c.log.Printf(log.Info, "Updating CloudFormation stack: %s", cfmName)

		stack, err := client.UpdateStack(ctx, &cloudformation.UpdateStackInput{
			StackName:    &cfmName,
			TemplateURL:  &cfmTemplateURL,
			Capabilities: []types.Capability{types.CapabilityCapabilityIam},
		})
		if err != nil {
			return err
		}

		stackStatus, err := c.awsWaitForStack(ctx, config, *stack.StackId)
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
			TemplateURL:  &cfmTemplateURL,
			Capabilities: []types.Capability{types.CapabilityCapabilityIam},
		})
		if err != nil {
			return err
		}

		stackStatus, err := c.awsWaitForStack(ctx, config, *stack.StackId)
		if err != nil {
			return err
		}
		if stackStatus != types.StackStatusCreateComplete {
			return fmt.Errorf("polaris: failed to create CloudFormation stack: %v", *stack.StackId)
		}
	}

	return nil
}

// AwsAccountSetRegions updates the AWS regions for the AWS account with
// specified account ID.
func (c *Client) AwsAccountSetRegions(ctx context.Context, awsAccountID string, awsRegions []string) error {
	c.log.Print(log.Trace, "polaris.Client.AwsAccountAddRegions")

	account, err := c.AwsAccountFromID(ctx, awsAccountID)
	if err != nil {
		return err
	}
	if n := len(account.Features); n != 1 {
		return fmt.Errorf("polaris: invalid number of features: %d", n)
	}

	regions := awsToPolarisRegionNames(awsRegions)
	if err := c.gql.AwsCloudAccountSave(ctx, account.ID, regions); err != nil {
		return err
	}

	return nil
}

// AwsAccountRemove removes the AWS account with specified account ID from the
// Polaris account. If awsAccountID is empty the account ID of the AWS config
// will be used instead.
func (c *Client) AwsAccountRemove(ctx context.Context, config aws.Config, awsAccountID string) error {
	c.log.Print(log.Trace, "polaris.Client.AwsAccountDelete")

	// Lookup Polaris cloud account from AWS account id
	var account AwsCloudAccount
	var err error
	if awsAccountID != "" {
		account, err = c.AwsAccountFromID(ctx, awsAccountID)
	} else {
		account, err = c.AwsAccountFromConfig(ctx, config)
	}
	if err != nil {
		return err
	}

	taskChainID, err := c.gql.AwsDeleteNativeAccount(ctx, account.ID, graphql.AwsEC2, false)
	if err != nil {
		return err
	}

	state, err := c.gql.WaitForTaskChain(ctx, taskChainID, 10*time.Second)
	if err != nil {
		return err
	}
	if state != graphql.TaskChainSucceeded {
		return fmt.Errorf("polaris: taskchain failed: taskChainUUID=%v, state=%v", taskChainID, state)
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
	client := cloudformation.NewFromConfig(config)
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

		stackStatus, err := c.awsWaitForStack(ctx, config, stackArn)
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

		stackStatus, err := c.awsWaitForStack(ctx, config, stackArn)
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
