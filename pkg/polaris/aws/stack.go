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
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

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
		case types.StackStatusCreateInProgress,
			types.StackStatusDeleteInProgress,
			types.StackStatusRollbackInProgress,
			types.StackStatusUpdateInProgress,
			types.StackStatusUpdateCompleteCleanupInProgress,
			types.StackStatusUpdateRollbackInProgress,
			types.StackStatusUpdateRollbackCompleteCleanupInProgress,
			types.StackStatusReviewInProgress,
			types.StackStatusImportInProgress,
			types.StackStatusImportRollbackInProgress:
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

// awsUpdateStack creates the stack if it doesn't exist, otherwise it's
// updated.
func awsUpdateStack(ctx context.Context, config aws.Config, stackName, templateURL string) error {
	client := cloudformation.NewFromConfig(config)

	exist, err := awsStackExist(ctx, config, stackName)
	if err != nil {
		return err
	}

	if exist {
		stack, err := client.UpdateStack(ctx, &cloudformation.UpdateStackInput{
			StackName:    &stackName,
			TemplateURL:  &templateURL,
			Capabilities: []types.Capability{types.CapabilityCapabilityIam},
		})
		if err != nil {
			return err
		}

		stackStatus, err := awsWaitForStack(ctx, config, *stack.StackId)
		if err != nil {
			return err
		}
		if stackStatus != types.StackStatusUpdateComplete {
			return fmt.Errorf("polaris: failed to update CloudFormation stack: %v, status: %v", *stack.StackId, stackStatus)
		}
	} else {
		stack, err := client.CreateStack(ctx, &cloudformation.CreateStackInput{
			StackName:    &stackName,
			TemplateURL:  &templateURL,
			Capabilities: []types.Capability{types.CapabilityCapabilityIam},
		})
		if err != nil {
			return err
		}

		stackStatus, err := awsWaitForStack(ctx, config, *stack.StackId)
		if err != nil {
			return err
		}
		if stackStatus != types.StackStatusCreateComplete {
			return fmt.Errorf("polaris: failed to create CloudFormation stack: %v, status: %v", *stack.StackId, stackStatus)
		}
	}

	return nil
}

// awsDeleteStack deletes the stack.
func awsDeleteStack(ctx context.Context, config aws.Config, stackName string) error {
	client := cloudformation.NewFromConfig(config)

	_, err := client.DeleteStack(ctx, &cloudformation.DeleteStackInput{StackName: &stackName})
	if err != nil {
		return err
	}

	stackStatus, err := awsWaitForStack(ctx, config, stackName)
	if err != nil {
		return err
	}
	if stackStatus != types.StackStatusDeleteComplete {
		return fmt.Errorf("polaris: failed to delete CloudFormation stack: %v", stackName)
	}

	return nil
}
