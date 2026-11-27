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
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
	"github.com/google/go-cmp/cmp"
)

const testCFStackARN = "arn:aws:cloudformation:us-west-2:123456789012:stack/example/stack-id"

type fakeExistingCloudFormation struct {
	stacks        []types.Stack
	parameters    []types.ParameterDeclaration
	getError      error
	updateError   error
	describeInput []*cloudformation.DescribeStacksInput
	updateInput   *cloudformation.UpdateStackInput
}

func (f *fakeExistingCloudFormation) DescribeStacks(_ context.Context, in *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	f.describeInput = append(f.describeInput, in)
	if len(f.stacks) == 0 {
		return &cloudformation.DescribeStacksOutput{}, nil
	}
	stack := f.stacks[0]
	if len(f.stacks) > 1 {
		f.stacks = f.stacks[1:]
	}
	return &cloudformation.DescribeStacksOutput{Stacks: []types.Stack{stack}}, nil
}

func (f *fakeExistingCloudFormation) GetTemplateSummary(_ context.Context, _ *cloudformation.GetTemplateSummaryInput, _ ...func(*cloudformation.Options)) (*cloudformation.GetTemplateSummaryOutput, error) {
	return &cloudformation.GetTemplateSummaryOutput{Parameters: f.parameters}, f.getError
}

func (f *fakeExistingCloudFormation) UpdateStack(_ context.Context, in *cloudformation.UpdateStackInput, _ ...func(*cloudformation.Options)) (*cloudformation.UpdateStackOutput, error) {
	f.updateInput = in
	return &cloudformation.UpdateStackOutput{StackId: aws.String(testCFStackARN)}, f.updateError
}

func TestOpenExistingCloudFormationStackValidation(t *testing.T) {
	if _, err := OpenCloudFormationStack(t.Context(), testCFStackARN, "", ""); err == nil || !strings.Contains(err.Error(), "explicit AWS") {
		t.Fatalf("expected an error for implicit credentials, got %v", err)
	}
	stackSet := "arn:aws:cloudformation:us-west-2:123456789012:stackset/example:stack-id"
	if _, err := OpenCloudFormationStack(t.Context(), stackSet, "example", ""); err == nil || !strings.Contains(err.Error(), "existing CloudFormation stack") {
		t.Fatalf("expected an error for StackSet ARN, got %v", err)
	}
}

func TestUpdateExistingCloudFormationStack(t *testing.T) {
	const templateURL = "https://bucket.s3.us-west-2.amazonaws.com/template?X-Amz-Signature=example"
	initial := types.Stack{
		StackId: aws.String(testCFStackARN), StackStatus: types.StackStatusCreateComplete,
		Parameters: []types.Parameter{{ParameterKey: aws.String("Keep")}, {ParameterKey: aws.String("Removed")}},
	}
	for _, tc := range []struct {
		name        string
		url         string
		stackID     string
		startStatus types.StackStatus
		finalStatus types.StackStatus
		updateError error
		getError    error
		wantErr     string
		wantUpdate  bool
	}{
		{name: "UpdatesExactStack", finalStatus: types.StackStatusUpdateComplete, wantUpdate: true},
		{name: "NoChanges", updateError: &smithy.GenericAPIError{Code: "ValidationError", Message: "No updates are to be performed."}, wantUpdate: true},
		{name: "Rollback", finalStatus: types.StackStatusUpdateRollbackComplete, wantErr: "UPDATE_ROLLBACK_COMPLETE", wantUpdate: true},
		{name: "MissingStack", stackID: "different-stack", wantErr: "could not verify the exact"},
		{name: "DeletedStack", startStatus: types.StackStatusDeleteComplete, wantErr: "cannot be updated in status DELETE_COMPLETE"},
		{name: "NotS3", url: "https://example.com/template", wantErr: "Amazon S3"},
		{name: "SummaryErrorRedactsURL", getError: &smithy.GenericAPIError{Code: "ValidationError", Message: templateURL}, wantErr: "AWS returned ValidationError"},
		{name: "UpdateErrorRedactsURL", updateError: &smithy.GenericAPIError{Code: "AccessDenied", Message: templateURL}, wantErr: "AWS returned AccessDenied", wantUpdate: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start := initial
			if tc.stackID != "" {
				start.StackId = &tc.stackID
			}
			if tc.startStatus != "" {
				start.StackStatus = tc.startStatus
			}
			f := &fakeExistingCloudFormation{
				stacks:      []types.Stack{start, {StackId: aws.String(testCFStackARN), StackStatus: tc.finalStatus}},
				parameters:  []types.ParameterDeclaration{{ParameterKey: aws.String("Keep")}, {ParameterKey: aws.String("New")}},
				updateError: tc.updateError, getError: tc.getError,
			}
			s := &CloudFormationStack{client: f, id: testCFStackARN}
			template := templateURL
			if tc.url != "" {
				template = tc.url
			}
			err := s.Update(t.Context(), template)
			if tc.wantErr == "" && err != nil {
				t.Fatal(err)
			}
			if tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
			if err != nil && strings.Contains(err.Error(), "X-Amz-Signature") {
				t.Fatal("error revealed signed URL")
			}
			for _, in := range f.describeInput {
				if aws.ToString(in.StackName) != testCFStackARN {
					t.Fatal("DescribeStacks was not scoped to the exact stack ARN")
				}
			}
			if (f.updateInput != nil) != tc.wantUpdate {
				t.Fatalf("unexpected UpdateStack call: %t", f.updateInput != nil)
			}
			if tc.wantUpdate {
				want := &cloudformation.UpdateStackInput{
					StackName: aws.String(testCFStackARN), TemplateURL: aws.String(templateURL),
					Parameters:   []types.Parameter{{ParameterKey: aws.String("Keep"), UsePreviousValue: aws.Bool(true)}},
					Capabilities: []types.Capability{types.CapabilityCapabilityNamedIam},
				}
				if diff := cmp.Diff(want, f.updateInput); diff != "" {
					t.Fatalf("UpdateStack request mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestUpdateExistingCloudFormationStackWaitsBeforeUpdating(t *testing.T) {
	f := &fakeExistingCloudFormation{stacks: []types.Stack{{
		StackId: aws.String(testCFStackARN), StackStatus: types.StackStatusUpdateInProgress,
	}}}
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	err := (&CloudFormationStack{client: f, id: testCFStackARN}).Update(ctx, "https://bucket.s3.amazonaws.com/template")
	if !errors.Is(err, context.DeadlineExceeded) || f.updateInput != nil {
		t.Fatalf("expected timeout before submitting a concurrent update, got %v", err)
	}
}
