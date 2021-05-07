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
	"errors"
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

// AwsCloudAccountSelector holds details about a cloud account and the features
// associated with that account.
type AwsCloudAccountSelector struct {
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

// AwsFeatureVersion maps a feature to a version number.
type AwsFeatureVersion struct {
	Feature string `json:"feature"`
	Version int    `json:"version"`
}

// AwsCloudAccountInitiate holds information about the CloudFormation stack
// that needs to be created in AWS to give permission to Polaris for managing
// the account being added. It also holds feature version information.
type AwsCloudAccountInitiate struct {
	CloudFormationURL string              `json:"cloudFormationUrl"`
	ExternalID        string              `json:"externalId"`
	FeatureVersions   []AwsFeatureVersion `json:"featureVersionList"`
	StackName         string              `json:"stackName"`
	TemplateURL       string              `json:"templateUrl"`
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

// AwsCloudAccount returns the cloud accounts with the specified Polaris cloud
// account UUID. Note that this call is locked to the cloud native protection
// feature.
func (c *Client) AwsCloudAccount(ctx context.Context, cloudAccountUUID string) (AwsCloudAccountSelector, error) {
	c.log.Print(log.Trace, "graphql.Client.AwsCloudAccount")

	buf, err := c.Request(ctx, awsCloudAccountSelectorQuery, struct {
		CloudAccountUUID string `json:"cloud_account_uuid"`
	}{CloudAccountUUID: cloudAccountUUID})
	if err != nil {
		return AwsCloudAccountSelector{}, err
	}

	c.log.Printf(log.Debug, "AwsCloudAccount(%q): %s", cloudAccountUUID, string(buf))

	var payload struct {
		Data struct {
			Account AwsCloudAccountSelector `json:"awsCloudAccountSelector"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AwsCloudAccountSelector{}, err
	}

	return payload.Data.Account, nil
}

// AwsCloudAccounts returns the cloud accounts matching the specified filters.
// The columnFilter can be used to search for AWS account ID, account name and
// role arn. Note that this call is locked to the cloud native protection
// feature.
func (c *Client) AwsCloudAccounts(ctx context.Context, columnFilter string) ([]AwsCloudAccountSelector, error) {
	c.log.Print(log.Trace, "graphql.Client.AwsCloudAccounts")

	buf, err := c.Request(ctx, awsAllCloudAccountsQuery, struct {
		ColumnFilter string `json:"column_filter,omitempty"`
	}{ColumnFilter: columnFilter})
	if err != nil {
		return nil, err
	}

	c.log.Printf(log.Debug, "AwsCloudAccounts(%q): %s", columnFilter, string(buf))

	var payload struct {
		Data struct {
			Accounts []AwsCloudAccountSelector `json:"allAwsCloudAccounts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	return payload.Data.Accounts, nil
}

// AwsValidateAndCreateCloudAccount begins the process of adding the specified
// AWS account to Polaris. The returned AwsCloudAccountInitiate value must be
// passed on to AwsFinalizeCloudAccountProtection which is the next step in the
// process.
func (c *Client) AwsValidateAndCreateCloudAccount(ctx context.Context, accountName, awsAccountID string) (AwsCloudAccountInitiate, error) {
	c.log.Print(log.Trace, "graphql.Client.AwsValidateAndCreateCloudAccount")

	buf, err := c.Request(ctx, awsValidateAndCreateCloudAccountQuery, struct {
		AccountName  string `json:"account_name"`
		AwsAccountID string `json:"aws_account_id"`
	}{AccountName: accountName, AwsAccountID: awsAccountID})
	if err != nil {
		return AwsCloudAccountInitiate{}, err
	}

	c.log.Printf(log.Debug, "AwsValidateAndCreateCloudAccount(%q, %q): %s", accountName, awsAccountID, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				InitiateResponse AwsCloudAccountInitiate `json:"initiateResponse"`
				ValidateResponse struct {
					InvalidAwsAccounts []struct {
						AccountName string `json:"accountName"`
						NativeID    string `json:"nativeId"`
						Message     string `json:"message"`
					}
					InvalidAwsAdminAccount struct {
						AccountName string `json:"accountName"`
						NativeID    string `json:"nativeId"`
						Message     string `json:"message"`
					}
				} `json:"validateReponse"`
			} `json:"validateAndCreateAwsCloudAccount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AwsCloudAccountInitiate{}, err
	}
	if len(payload.Data.Query.ValidateResponse.InvalidAwsAccounts) != 0 {
		return AwsCloudAccountInitiate{}, errors.New("polaris: invalid aws accounts")
	}

	return payload.Data.Query.InitiateResponse, nil
}

// AwsFinalizeCloudAccountProtection finalizes the process of the adding the
// specified AWS account to Polaris. After this function a CloudFormation stack
// must be created using the information returned by
// AwsValidateAndCreateCloudAccount.
func (c *Client) AwsFinalizeCloudAccountProtection(ctx context.Context, accountName, awsAccountID string, awsRegions []string, accountInit AwsCloudAccountInitiate) error {
	c.log.Print(log.Trace, "graphql.Client.AwsFinalizeCloudAccountProtection")

	buf, err := c.Request(ctx, awsFinalizeCloudAccountProtectionQuery, struct {
		AccountName     string              `json:"account_name"`
		AwsAccountID    string              `json:"aws_account_id"`
		AwsRegions      []string            `json:"aws_regions,omitempty"`
		ExternalID      string              `json:"external_id"`
		FeatureVersions []AwsFeatureVersion `json:"feature_versions"`
		StackName       string              `json:"stack_name"`
	}{
		AccountName:     accountName,
		AwsAccountID:    awsAccountID,
		AwsRegions:      awsRegions,
		ExternalID:      accountInit.ExternalID,
		FeatureVersions: accountInit.FeatureVersions,
		StackName:       accountInit.StackName,
	})
	if err != nil {
		return err
	}

	c.log.Printf(log.Debug, "AwsFinalizeCloudAccountProtection(%q, %q, %q, %q, %q, %q): %s", accountName, awsAccountID, awsRegions,
		accountInit.ExternalID, accountInit.FeatureVersions, accountInit.StackName, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				AwsChildAccounts []struct {
					AccountName string `json:"accountName"`
					NativeId    string `json:"nativeId"`
					Message     string `json:"message"`
				}
				Message string `json:"message"`
			} `json:"finalizeAwsCloudAccountProtection"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return err
	}

	return nil
}

// AwsStartNativeAccountDisableJob deletes the native account. After being
// deleted the account will have the status disabled.
func (c *Client) AwsStartNativeAccountDisableJob(ctx context.Context, accountID string, protectionFeature AwsProtectionFeature, deleteSnapshots bool) (TaskChainUUID, error) {
	c.log.Print(log.Trace, "graphql.Client.AwsStartNativeAccountDisableJob")

	buf, err := c.Request(ctx, awsStartNativeAccountDisableJobQuery, struct {
		AccountID         string `json:"polarisAccountId"`
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

// AwsPrepareCloudAccountDeletion prepares the deletion of a cloud account.
// The cfm prefix for the return values are short form CloudFormation. Note
// that this call is locked to the cloud native protection feature.
func (c *Client) AwsPrepareCloudAccountDeletion(ctx context.Context, cloudAccountUUID string) (cfmURL string, err error) {
	c.log.Print(log.Trace, "graphql.Client.AwsPrepareCloudAccountDeletion")

	buf, err := c.Request(ctx, awsPrepareCloudAccountDeletionQuery, struct {
		CloudAccountUUID string `json:"cloud_account_uuid,omitempty"`
	}{CloudAccountUUID: cloudAccountUUID})
	if err != nil {
		return "", err
	}

	c.log.Printf(log.Debug, "AwsPrepareCloudAccountDeletion(%q): %s", cloudAccountUUID, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				URL string `json:"cloudFormationUrl"`
			} `json:"prepareAwsCloudAccountDeletion"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", err
	}

	return payload.Data.Query.URL, nil
}

// AwsFinalizeCloudAccountDeletion finalizes the deletion of a cloud account.
// The message returned by the GraphQL API call is converted into a Go error.
// Note that this call is locked to the cloud native protection feature.
func (c *Client) AwsFinalizeCloudAccountDeletion(ctx context.Context, cloudAccountUUID string) (err error) {
	c.log.Print(log.Trace, "graphql.Client.AwsFinalizeCloudAccountDeletion")

	buf, err := c.Request(ctx, awsFinalizeCloudAccountDeletionQuery, struct {
		CloudAccountUUID string `json:"cloud_account_uuid"`
	}{CloudAccountUUID: cloudAccountUUID})
	if err != nil {
		return err
	}

	c.log.Printf(log.Debug, "AwsFinalizeCloudAccountDeletion(%q): %s", cloudAccountUUID, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Message string `json:"message"`
			} `json:"finalizeAwsCloudAccountDeletion"`
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

// AwsUpdateCloudAccount updates the settings of the cloud account. The message
// returned by the GraphQL API call is converted into a Go error. At this time
// only the regions can be updated.
func (c *Client) AwsUpdateCloudAccount(ctx context.Context, cloudAccountUUID string, awsRegions []string) (err error) {
	c.log.Print(log.Trace, "graphql.Client.AwsUpdateCloudAccount")

	buf, err := c.Request(ctx, awsUpdateCloudAccountQuery, struct {
		CloudAccountUUID string   `json:"cloud_account_uuid"`
		AwsRegions       []string `json:"aws_regions"`
	}{CloudAccountUUID: cloudAccountUUID, AwsRegions: awsRegions})
	if err != nil {
		return err
	}

	c.log.Printf(log.Debug, "AwsCloudAccountSave(%q, %q): %s", cloudAccountUUID, awsRegions, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Message string `json:"message"`
			} `json:"updateAwsCloudAccount"`
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
