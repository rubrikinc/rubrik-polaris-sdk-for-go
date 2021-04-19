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

package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"
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

// AwsCloudAccounts holds details about a cloud account and the features
// associated with that account.
type AwsCloudAccount struct {
	AwsCloudAccount struct {
		CloudType           string `json:"cloudType"`
		ID                  string `json:"id"`
		NativeID            string `json:"nativeId"`
		AccountName         string `json:"accountName"`
		Message             string `json:"message"`
		SeamlessFlowEnabled bool   `json:"seamlessFlowEnabled"`
	} `json:"awsCloudAccount"`
	FeatureDetails []struct {
		AwsRegions []string `json:"awsRegions"`
		Feature    string   `json:"feature"`
		RoleArn    string   `json:"roleArn"`
		StackArn   string   `json:"stackArn"`
		Status     string   `json:"status"`
	} `json:"featureDetails"`
}

// AwsNativeAccount holds details about a native account and the SLAs
// configured.
type AwsNativeAccount struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Regions       []string `json:"regions"`
	Status        string   `json:"status"`
	SLAAssignment string   `json:"slaAssignment"`

	ConfiguredSLADomain struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"configuredSlaDomain"`

	EffectiveSLADomain struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"effectiveSlaDomain"`
}

// AwsNativeAccounts returns the native account matching the specified filters.
func (c *Client) AwsNativeAccounts(ctx context.Context, protectionFeature AwsProtectionFeature, nameFilter string) ([]AwsNativeAccount, error) {
	c.log.Print(log.Trace, "graphql.Client.AwsNativeAccounts")

	accounts := make([]AwsNativeAccount, 0, 10)

	var endCursor string
	for {
		buf, err := c.Request(ctx, awsNativeAccountsQuery, struct {
			After             string `json:"after,omitempty"`
			ProtectionFeature string `json:"awsNativeProtectionFeature,omitempty"`
			NameFilter        string `json:"filter,omitempty"`
		}{After: endCursor, ProtectionFeature: string(protectionFeature), NameFilter: nameFilter})
		if err != nil {
			return nil, err
		}

		c.log.Printf(log.Debug, "AwsNativeAccounts(%q, %q): %s", protectionFeature, nameFilter, string(buf))

		var payload struct {
			Data struct {
				Query struct {
					Count int `json:"count"`
					Edges []struct {
						Node AwsNativeAccount `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"awsNativeAccounts"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, err
		}

		for _, account := range payload.Data.Query.Edges {
			accounts = append(accounts, account.Node)
		}

		if !payload.Data.Query.PageInfo.HasNextPage {
			break
		}
		endCursor = payload.Data.Query.PageInfo.EndCursor
	}

	return accounts, nil
}

// AwsCloudAccounts returns the cloud accounts matching the specified filters.
// The columnFilter can be used to search for AWS account ID, account name and
// role arn. Note that this call is locked to the cloud native
// protection feature.
func (c *Client) AwsCloudAccounts(ctx context.Context, columnFilter string) ([]AwsCloudAccount, error) {
	c.log.Print(log.Trace, "graphql.Client.AwsCloudAccounts")

	buf, err := c.Request(ctx, awsCloudAccountsQuery, struct {
		ColumnFilter string `json:"columnFilter,omitempty"`
	}{ColumnFilter: columnFilter})
	if err != nil {
		return nil, err
	}

	c.log.Printf(log.Debug, "AwsCloudAccounts(%q): %s", columnFilter, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Accounts []AwsCloudAccount `json:"awsCloudAccounts"`
			} `json:"awsCloudAccounts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	return payload.Data.Query.Accounts, nil
}

// AwsNativeProtectionAccountAdd adds the native AWS account. The cfm prefix
// for the return values are short form CloudFormation.
func (c *Client) AwsNativeProtectionAccountAdd(ctx context.Context, awsAccountID, awsAccountName string, awsRegions []string) (cfmName, cfmURL, cfmTemplateURL string, err error) {
	c.log.Print(log.Trace, "graphql.Client.AwsNativeProtectionAccountAdd")

	buf, err := c.Request(ctx, awsNativeProtectionAccountAddQuery, struct {
		AccountID string   `json:"accountId,omitempty"`
		Name      string   `json:"accountName,omitempty"`
		Regions   []string `json:"regions,omitempty"`
	}{AccountID: awsAccountID, Name: awsAccountName, Regions: awsRegions})
	if err != nil {
		return "", "", "", err
	}

	c.log.Printf(log.Debug, "AwsNativeProtectionAccountAdd(%q, %q, %q): %s", awsAccountID, awsAccountName, awsRegions, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Name        string `json:"cloudFormationName"`
				URL         string `json:"cloudFormationUrl"`
				TemplateURL string `json:"cloudFormationTemplateUrl"`
				Message     string `json:"errorMessage"`
			} `json:"awsNativeProtectionAccountAdd"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", "", "", err
	}
	if payload.Data.Query.Message != "" {
		return "", "", "", fmt.Errorf("polaris: %s", payload.Data.Query.Message)
	}
	if payload.Data.Query.Name == "" {
		return "", "", "", fmt.Errorf("polaris: invalid CloudFormation stack name: %q", payload.Data.Query.Name)
	}

	return payload.Data.Query.Name, payload.Data.Query.URL, payload.Data.Query.TemplateURL, nil
}

// AwsCloudAccountUpdateFeatureInitiate initiates a manual feature update.
// The cfm prefix for the return values are short form CloudFormation. Note
// that this call is locked to the cloud native protection feature.
func (c *Client) AwsCloudAccountUpdateFeatureInitiate(ctx context.Context, accountID string) (cfmURL, cfmTemplateURL string, err error) {
	c.log.Print(log.Trace, "graphql.Client.AwsCloudAccountUpdateFeatureInitiate")

	buf, err := c.Request(ctx, awsCloudAccountUpdateFeatureInitiateQuery, struct {
		AccountID string `json:"polarisAccountId,omitempty"`
	}{AccountID: accountID})
	if err != nil {
		return "", "", err
	}

	c.log.Printf(log.Debug, "AwsCloudAccountUpdateFeatureInitiate(%q, %q, %q): %s", accountID, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				URL         string `json:"cloudFormationUrl"`
				TemplateURL string `json:"templateUrl"`
			} `json:"awsCloudAccountUpdateFeatureInitiate"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", "", err
	}

	return payload.Data.Query.URL, payload.Data.Query.TemplateURL, nil
}

// AwsStartNativeAccountDisableJob deletes the native account. After being
// deleted the account will have the status disabled.
func (c *Client) AwsStartNativeAccountDisableJob(ctx context.Context, accountID string, protectionFeature AwsProtectionFeature, deleteSnapshots bool) (TaskChainUUID, error) {
	c.log.Print(log.Trace, "graphql.Client.AwsStartNativeAccountDisableJob")

	buf, err := c.Request(ctx, startAwsNativeAccountDisableJobQuery, struct {
		AccountID         string `json:"polarisAccountId,omitempty"`
		ProtectionFeature string `json:"awsNativeProtectionFeature"`
		DeleteSnapshots   bool   `json:"deleteNativeSnapshots"`
	}{AccountID: accountID, ProtectionFeature: string(protectionFeature), DeleteSnapshots: deleteSnapshots})
	if err != nil {
		return "", err
	}

	c.log.Printf(log.Debug, "AwsStartNativeAccountDisableJob(%q, %q, %t): %s", accountID, protectionFeature, deleteSnapshots, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Error string        `json:"error"`
				JobID TaskChainUUID `json:"jobId"`
			} `json:"startAwsNativeAccountDisableJob"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", err
	}
	if payload.Data.Query.Error != "" {
		return "", fmt.Errorf("polaris: %s", payload.Data.Query.Error)
	}

	return payload.Data.Query.JobID, nil
}

// AwsCloudAccountDeleteInitiate initiates the deletion of a cloud account.
// The cfm prefix for the return values are short form CloudFormation. Note
// that this call is locked to the cloud native protection feature.
func (c *Client) AwsCloudAccountDeleteInitiate(ctx context.Context, accountID string) (cfmURL string, err error) {
	c.log.Print(log.Trace, "graphql.Client.AwsCloudAccountDeleteInitiate")

	buf, err := c.Request(ctx, awsCloudAccountDeleteInitiateQuery, struct {
		AccountID string `json:"polarisAccountId,omitempty"`
	}{AccountID: accountID})
	if err != nil {
		return "", err
	}

	c.log.Printf(log.Debug, "AwsCloudAccountDeleteInitiate(%q): %s", accountID, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				URL string `json:"cloudFormationUrl"`
			} `json:"awsCloudAccountDeleteInitiate"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", err
	}

	return payload.Data.Query.URL, nil
}

// AwsCloudAccountDeleteProcess finalizes the deletion of a cloud account.
// The message returned by the GraphQL API call is converted into a Go error.
// Note that this call is locked to the cloud native protection feature.
func (c *Client) AwsCloudAccountDeleteProcess(ctx context.Context, accountID string) (err error) {
	c.log.Print(log.Trace, "graphql.Client.AwsCloudAccountDeleteProcess")

	buf, err := c.Request(ctx, awsCloudAccountDeleteProcessQuery, struct {
		AccountID string `json:"polarisAccountId,omitempty"`
	}{AccountID: accountID})
	if err != nil {
		return err
	}

	c.log.Printf(log.Debug, "AwsCloudAccountDeleteProcess(%q): %s", accountID, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Message string `json:"message"`
			} `json:"awsCloudAccountDeleteProcess"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return err
	}

	// On success the message starts with "successfully".
	if !strings.HasPrefix(strings.ToLower(payload.Data.Query.Message), "successfully") {
		return fmt.Errorf("polaris: %s", payload.Data.Query.Message)
	}

	return nil
}

// AwsCloudAccountSave updates the settings of the cloud account. The message
// returned by the GraphQL API call is converted into a Go error. At this time
// only the AWS regions can be updated.
func (c *Client) AwsCloudAccountSave(ctx context.Context, accountID string, awsRegions []string) (err error) {
	c.log.Print(log.Trace, "graphql.Client.AwsCloudAccountSave")

	buf, err := c.Request(ctx, awsCloudAccountSaveQuery, struct {
		AccountID  string   `json:"polarisAccountId,omitempty"`
		AwsRegions []string `json:"awsRegions,omitempty"`
	}{AccountID: accountID, AwsRegions: awsRegions})
	if err != nil {
		return err
	}

	c.log.Printf(log.Debug, "AwsCloudAccountSave(%q, %q): %s", accountID, awsRegions, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Message string `json:"message"`
			} `json:"awsCloudAccountSave"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return err
	}

	// On success the message starts with "successfully".
	if !strings.HasPrefix(strings.ToLower(payload.Data.Query.Message), "successfully") {
		return fmt.Errorf("polaris: %s", payload.Data.Query.Message)
	}

	return nil
}
