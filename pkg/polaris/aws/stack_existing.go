// Copyright 2026 Rubrik, Inc.
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
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
)

type cloudFormationStackClient interface {
	DescribeStacks(context.Context, *cloudformation.DescribeStacksInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	GetTemplateSummary(context.Context, *cloudformation.GetTemplateSummaryInput, ...func(*cloudformation.Options)) (*cloudformation.GetTemplateSummaryOutput, error)
	UpdateStack(context.Context, *cloudformation.UpdateStackInput, ...func(*cloudformation.Options)) (*cloudformation.UpdateStackOutput, error)
}

// CloudFormationStack is a verified, existing single-account AWS stack.
type CloudFormationStack struct {
	client cloudFormationStackClient
	id     string
}

// OpenCloudFormationStack verifies that the specified AWS profile and/or IAM
// role can access the exact stack ARN in its owning account and region. It
// waits for any in-progress stack operation to finish. It never creates a stack.
func OpenCloudFormationStack(ctx context.Context, stackARN, profile, roleARN string) (*CloudFormationStack, error) {
	stack, err := arn.Parse(stackARN)
	parts := strings.Split(stack.Resource, "/")
	if err != nil || stack.Service != "cloudformation" || stack.Region == "" ||
		!verifyAccountID(stack.AccountID) || len(parts) != 3 || parts[0] != "stack" || parts[1] == "" || parts[2] == "" {
		return nil, errors.New("expected the ARN of an existing CloudFormation stack")
	}
	if profile == "" && roleARN == "" {
		return nil, errors.New("an explicit AWS profile or IAM role is required")
	}
	if profile == "" {
		profile = "default"
	}
	account, err := ProfileWithRegionAndRole(profile, stack.Region, roleARN)(ctx)
	if err != nil {
		return nil, err
	}
	if account.NativeID != stack.AccountID {
		return nil, fmt.Errorf("AWS credentials are for account %s, but the CloudFormation stack is in account %s", account.NativeID, stack.AccountID)
	}
	existing := &CloudFormationStack{client: cloudformation.NewFromConfig(*account.config), id: stackARN}
	if _, err := existing.waitForUpdateableStack(ctx); err != nil {
		return nil, err
	}
	return existing, nil
}

// Update applies the given S3 template to the existing stack, preserving the
// values of parameters already set on the stack, and waits for the update to
// finish. A stack that no longer exists is not recreated.
func (s *CloudFormationStack) Update(ctx context.Context, templateURL string) error {
	u, err := url.Parse(templateURL)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.Fragment != "" || u.Path == "" {
		return errors.New("invalid CloudFormation S3 template URL")
	}
	host := u.Hostname()
	s3 := strings.TrimSuffix(strings.TrimSuffix(host, ".amazonaws.com.cn"), ".amazonaws.com")
	if s3 == host || !(s3 == "s3" || strings.HasPrefix(s3, "s3.") || strings.HasPrefix(s3, "s3-") ||
		strings.HasSuffix(s3, ".s3") || strings.Contains(s3, ".s3.") || strings.Contains(s3, ".s3-")) {
		return errors.New("CloudFormation template must be hosted on Amazon S3")
	}

	stack, err := s.waitForUpdateableStack(ctx)
	if err != nil {
		return err
	}

	var parameters []types.Parameter
	if len(stack.Parameters) > 0 {
		summary, err := s.client.GetTemplateSummary(ctx, &cloudformation.GetTemplateSummaryInput{TemplateURL: &templateURL})
		if err != nil {
			return s.templateError(ctx, "inspect CloudFormation template", err)
		}
		existing := make(map[string]bool, len(stack.Parameters))
		for _, param := range stack.Parameters {
			existing[aws.ToString(param.ParameterKey)] = true
		}
		for _, param := range summary.Parameters {
			if existing[aws.ToString(param.ParameterKey)] {
				parameters = append(parameters, types.Parameter{ParameterKey: param.ParameterKey, UsePreviousValue: aws.Bool(true)})
			}
		}
	}
	reply, err := s.client.UpdateStack(ctx, &cloudformation.UpdateStackInput{
		StackName:    &s.id,
		TemplateURL:  &templateURL,
		Parameters:   parameters,
		Capabilities: []types.Capability{types.CapabilityCapabilityNamedIam},
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) && apiError.ErrorCode() == "ValidationError" &&
			apiError.ErrorMessage() == "No updates are to be performed." {
			return nil
		}
		return s.templateError(ctx, "update CloudFormation stack", err)
	}
	if aws.ToString(reply.StackId) != s.id {
		return errors.New("CloudFormation update returned an unexpected stack ID")
	}

	stack, err = s.wait(ctx)
	if err != nil {
		return err
	}
	if stack.StackStatus != types.StackStatusUpdateComplete {
		return fmt.Errorf("CloudFormation stack %s update finished with status %s", s.id, stack.StackStatus)
	}
	return nil
}

func (s *CloudFormationStack) waitForUpdateableStack(ctx context.Context) (types.Stack, error) {
	stack, err := s.wait(ctx)
	if err != nil {
		return types.Stack{}, err
	}
	switch stack.StackStatus {
	case types.StackStatusCreateComplete, types.StackStatusUpdateComplete,
		types.StackStatusUpdateRollbackComplete, types.StackStatusImportComplete:
		return stack, nil
	default:
		return types.Stack{}, fmt.Errorf("CloudFormation stack %s cannot be updated in status %s", s.id, stack.StackStatus)
	}
}

func (s *CloudFormationStack) wait(ctx context.Context) (types.Stack, error) {
	for {
		reply, err := s.client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{StackName: &s.id})
		if err != nil {
			return types.Stack{}, fmt.Errorf("describe CloudFormation stack %s: %w", s.id, err)
		}
		if len(reply.Stacks) != 1 || aws.ToString(reply.Stacks[0].StackId) != s.id {
			return types.Stack{}, fmt.Errorf("could not verify the exact CloudFormation stack %s", s.id)
		}
		stack := reply.Stacks[0]
		if !strings.HasSuffix(string(stack.StackStatus), "_IN_PROGRESS") {
			return stack, nil
		}
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			return types.Stack{}, ctx.Err()
		}
	}
}

// AWS errors from operations that accept a template URL may echo it. Preserve
// the AWS error code, but never include the signed URL in an error.
func (s *CloudFormationStack) templateError(ctx context.Context, operation string, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	var apiError smithy.APIError
	if errors.As(err, &apiError) {
		return fmt.Errorf("%s for %s: AWS returned %s (see CloudFormation for details)", operation, s.id, apiError.ErrorCode())
	}
	return fmt.Errorf("%s for %s: AWS request failed (see CloudFormation for details)", operation, s.id)
}
