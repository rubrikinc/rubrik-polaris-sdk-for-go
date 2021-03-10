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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// SLAAssignment -
type SLAAssignment string

// SLADomain -
type SLADomain struct {
	ID   string
	Name string
}

// AwsProtectionFeature -
type AwsProtectionFeature string

const (
	// AwsEC2 -
	AwsEC2 AwsProtectionFeature = "EC2"

	// AwsRDS -
	AwsRDS AwsProtectionFeature = "RDS"
)

type AccountFeature struct {
	Feature    string
	AwsRegions []string
	RoleArn    string
	StackArn   string
	Status     string
}

// AwsCloudAccount -
type AwsCloudAccount struct {
	ID       string
	NativeID string
	Name     string
	Message  string
	Features []AccountFeature
}

// AwsAccount -
type AwsAccount struct {
	ID         string
	Regions    []string
	Status     string
	Name       string
	Assignment SLAAssignment
	Configured SLADomain
	Effective  SLADomain
}

// awsAccountID returns the AWS account id and account name. Note that if the
// AWS user does not have permissions for Organizations the account name will
// be empty.
func (c *Client) awsAccountID(ctx context.Context, config aws.Config) (string, string, error) {
	c.log.Print(Trace, "Client.awsAccountID")

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
		c.log.Print(Debug, "Failed to access Organizations")
		return *id.Account, "", nil
	}

	return *id.Account, *info.Account.Name, nil
}

// awsCFMCreateStackWaitFor blocks until the CloudFormation stack create/delete
// has completed. When the stack create/delete completes the final state of the
// operation is returned.
func (c *Client) awsCFMCreateStackWaitFor(ctx context.Context, config aws.Config, stackName string) (types.StackStatus, error) {
	c.log.Print(Trace, "Client.awsCFMCreateStackWaitFor")

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

// ListAwsAccounts lists the AWS accounts added to Polaris. The filter
// parameter can be used to filter the result on a substring of the account
// name.
func (c *Client) ListAwsAccounts(ctx context.Context, feature AwsProtectionFeature, filter string) ([]AwsAccount, error) {
	c.log.Print(Trace, "Client.ListAwsAccounts")

	accounts := make([]AwsAccount, 0, 5)

	var endCursor string
	for {
		buf, err := c.gql.Request(ctx, graphql.AwsAccountsQuery, struct {
			After   string `json:"after,omitempty"`
			Feature string `json:"awsNativeProtectionFeature,omitempty"`
			Filter  string `json:"filter,omitempty"`
		}{After: endCursor, Feature: string(feature), Filter: filter})
		if err != nil {
			return nil, err
		}

		var payload struct {
			Data struct {
				Query struct {
					Count int `json:"count"`
					Edges []struct {
						Node graphql.AwsNativeAccount `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"awsNativeAccountConnection"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, err
		}

		for _, edge := range payload.Data.Query.Edges {
			configured := SLADomain{
				ID:   edge.Node.ConfiguredSLADomain.ID,
				Name: edge.Node.ConfiguredSLADomain.Name,
			}
			effective := SLADomain{
				ID:   edge.Node.EffectiveSLADomain.ID,
				Name: edge.Node.EffectiveSLADomain.Name,
			}
			accounts = append(accounts, AwsAccount{
				ID:         edge.Node.ID,
				Regions:    edge.Node.Regions,
				Status:     edge.Node.Status,
				Name:       edge.Node.Name,
				Assignment: SLAAssignment(edge.Node.SLAAssignment),
				Configured: configured,
				Effective:  effective,
			})
		}

		if !payload.Data.Query.PageInfo.HasNextPage {
			break
		}
		endCursor = payload.Data.Query.PageInfo.EndCursor
	}

	return accounts, nil
}

// AwsAccountsDetail -
func (c *Client) AwsAccountsDetail(ctx context.Context, filter string) ([]AwsCloudAccount, error) {
	c.log.Print(Trace, "Client.AwsAccountsDetail")

	accounts := make([]AwsCloudAccount, 0, 5)

	buf, err := c.gql.Request(ctx, graphql.AwsAccountsDetailQuery, struct {
		Filter string `json:"filter,omitempty"`
	}{Filter: filter})
	if err != nil {
		return nil, err
	}

	var payload struct {
		Data struct {
			Query struct {
				Accounts []graphql.AwsCloudAccount `json:"awsCloudAccounts"`
			} `json:"awsCloudAccounts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	for _, internal := range payload.Data.Query.Accounts {
		account := AwsCloudAccount{
			ID:       internal.AwsCloudAccount.ID,
			NativeID: internal.AwsCloudAccount.NativeID,
			Name:     internal.AwsCloudAccount.AccountName,
			Message:  internal.AwsCloudAccount.Message,
		}
		for _, feature := range internal.FeatureDetails {
			account.Features = append(account.Features, AccountFeature{
				Feature:    feature.Feature,
				AwsRegions: feature.AwsRegions,
				RoleArn:    feature.RoleArn,
				StackArn:   feature.StackArn,
				Status:     feature.Status,
			})
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// AddAwsAccount -
func (c *Client) AddAwsAccount(ctx context.Context, config aws.Config) error {
	c.log.Print(Trace, "Client.AddAwsAccount")

	// Lookup AWS account id and name
	awsAccountID, awsAccountName, err := c.awsAccountID(ctx, config)
	if err != nil {
		return err
	}
	if awsAccountName == "" {
		awsAccountName = "Trinity-TPM-DevOps"
	}

	buf, err := c.gql.Request(ctx, graphql.AwsAccountsAddQuery, struct {
		AccountID string   `json:"account_id,omitempty"`
		Name      string   `json:"account_name,omitempty"`
		Regions   []string `json:"regions,omitempty"`
	}{AccountID: awsAccountID, Name: awsAccountName, Regions: []string{config.Region}})
	if err != nil {
		return err
	}

	fmt.Println(string(buf))

	var payload struct {
		Data struct {
			Query struct {
				CloudFormationName        string `json:"cloudFormationName"`
				CloudFormationTemplateURL string `json:"cloudFormationTemplateUrl"`
				CloudFormationURL         string `json:"cloudFormationUrl"`
				ErrorMessage              string `json:"errorMessage"`
			} `json:"awsNativeProtectionAccountAdd"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return err
	}
	if payload.Data.Query.ErrorMessage != "" {
		return fmt.Errorf("polaris: %s", payload.Data.Query.ErrorMessage)
	}
	if payload.Data.Query.CloudFormationName == "" {
		return fmt.Errorf("polaris: invalid CloudFormation stack name: %q", payload.Data.Query.CloudFormationName)
	}

	c.log.Printf(Debug, "CloudFormation stack name: %q", payload.Data.Query.CloudFormationName)

	// Create CloudFormation stack.
	client3 := cloudformation.NewFromConfig(config)
	stack, err := client3.CreateStack(ctx, &cloudformation.CreateStackInput{
		StackName:    &payload.Data.Query.CloudFormationName,
		TemplateURL:  &payload.Data.Query.CloudFormationTemplateURL,
		Capabilities: []types.Capability{types.CapabilityCapabilityIam},
	})
	if err != nil {
		return err
	}

	stackStatus, err := c.awsCFMCreateStackWaitFor(ctx, config, *stack.StackId)
	if err != nil {
		return err
	}
	if stackStatus != types.StackStatusCreateComplete {
		return fmt.Errorf("polaris: failed to create CloudFormation stack: %v", *stack.StackId)
	}

	return nil
}

// UpdateAwsAccount -
func (c *Client) UpdateAwsAccount(ctx context.Context, config aws.Config) error {
	c.log.Print(Trace, "Client.UpdateAwsAccount")

	// Lookup AWS account id
	awsAccountID, _, err := c.awsAccountID(ctx, config)
	if err != nil {
		return err
	}

	// Lookup Polaris cloud account from AWS account id
	accounts, err := c.AwsAccountsDetail(ctx, awsAccountID)
	if err != nil {
		return err
	}
	if len(accounts) > 1 {
		return errors.New("account id refers to multiple accounts")
	}
	account := accounts[0]

	// Iterate over the cloud native protection features.
	for _, feature := range account.Features {
		if feature.Feature != "CLOUD_NATIVE_PROTECTION" {
			continue
		}
		c.log.Printf(Debug, "Account: %s/%s - %s\n", account.Name, account.ID, feature.Status)

		if feature.Status != "MISSING_PERMISSIONS" {
			continue
		}
		buf, err := c.gql.Request(ctx, graphql.AwsAccountsUpdateInitiateQuery, struct {
			AccountID string `json:"polaris_account_id,omitempty"`
			Feature   string `json:"aws_native_protection_feature"`
		}{AccountID: account.ID, Feature: feature.Feature})
		if err != nil {
			return err
		}

		// TODO: setup CloudFormation stack?
		var payload struct {
			Data struct {
				Query struct {
					CloudFormationURL string `json:"cloudFormationUrl"`
					TemplateURL       string `json:"templateUrl"`
				} `json:"awsCloudAccountUpdateFeatureInitiate"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return err
		}

		c.log.Print(Info, "Account updated")
	}

	return nil
}

// DeleteAwsAccount -
func (c *Client) DeleteAwsAccount(ctx context.Context, config aws.Config) error {
	c.log.Print(Trace, "Client.DeleteAwsAccount")

	// Lookup AWS account id.
	awsAccountID, _, err := c.awsAccountID(ctx, config)
	if err != nil {
		return err
	}

	// Lookup Polaris cloud account from AWS account id.
	accounts, err := c.AwsAccountsDetail(ctx, awsAccountID)
	if err != nil {
		return err
	}
	if len(accounts) < 1 {
		return fmt.Errorf("polaris: no account matching account id found")
	}
	if len(accounts) > 1 {
		return fmt.Errorf("polaris: account id refers to multiple accounts")
	}
	account := accounts[0]

	// Disable account.
	buf, err := c.gql.Request(ctx, graphql.AwsAccountsDisableQuery, struct {
		AccountID string `json:"polaris_account_id,omitempty"`
	}{AccountID: account.ID})
	if err != nil {
		return err
	}

	var payload1 struct {
		Data struct {
			Query struct {
				TaskChainID string `json:"taskchainUuid"`
			} `json:"deleteAwsNativeAccount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload1); err != nil {
		return err
	}

	status, err := c.taskChainWaitFor(ctx, taskChainID(payload1.Data.Query.TaskChainID))
	if err != nil {
		return err
	}
	if status != taskChainSucceeded {
		return fmt.Errorf("polaris: taskchain failed: taskChainUUID=%v, taskStatus=%v", payload1.Data.Query.TaskChainID, status)
	}

	// Initiate delete.
	buf, err = c.gql.Request(ctx, graphql.AwsAccountsDeleteInitiateQuery, struct {
		AccountID string `json:"polaris_account_id,omitempty"`
	}{AccountID: account.ID})
	if err != nil {
		return err
	}

	var payload2 struct {
		Data struct {
			Query struct {
				CloudFormationURL string `json:"cloudFormationUrl"`
			} `json:"awsCloudAccountDeleteInitiate"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload2); err != nil {
		return err
	}

	// Remove CloudFormation stack.
	for _, feature := range account.Features {
		if feature.Feature != "CLOUD_NATIVE_PROTECTION" {
			continue
		}

		stackName := feature.StackArn
		if strings.Count(stackName, "/") < 2 || len(stackName) < 2 {
			return fmt.Errorf("polaris: invalid stack arn: stackArn=%v", feature.StackArn)
		}
		stackName = stackName[strings.Index(stackName, "/")+1 : strings.LastIndex(stackName, "/")]

		// TODO: handle multiple regions?
		client := cloudformation.NewFromConfig(config)
		stacks, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{StackName: &stackName})
		if err != nil {
			return err
		}
		stack := stacks.Stacks[0]

		_, err = client.DeleteStack(ctx, &cloudformation.DeleteStackInput{StackName: stack.StackId})
		if err != nil {
			return err
		}

		stackStatus, err := c.awsCFMCreateStackWaitFor(ctx, config, *stack.StackId)
		if err != nil {
			return err
		}
		if stackStatus != types.StackStatusDeleteComplete {
			return fmt.Errorf("polaris: failed to delete CloudFormation stack: %v", stackName)
		}
	}

	// Process delete.
	buf, err = c.gql.Request(ctx, graphql.AwsAccountsDeleteCommitQuery, struct {
		AccountID string `json:"polaris_account_id,omitempty"`
	}{AccountID: account.ID})
	if err != nil {
		return err
	}

	var payload3 struct {
		Data struct {
			Query struct {
				Message string `json:"message"`
			} `json:"awsCloudAccountDeleteProcess"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload3); err != nil {
		return err
	}

	fmt.Println(payload3.Data.Query.Message)

	return nil
}
