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

package exocompute

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// KMSSharingMode describes which automated KMS permissions a mapping uses.
type KMSSharingMode string

const (
	KMSSharingModeNone              KMSSharingMode = "KMS_SHARING_MODE_NONE"
	KMSSharingModeSharingOnly       KMSSharingMode = "KMS_SHARING_MODE_SHARING_ONLY"
	KMSSharingModeSharingAndGateway KMSSharingMode = "KMS_SHARING_MODE_SHARING_AND_GATEWAY"
	KMSSharingModeUnspecified       KMSSharingMode = "KMS_SHARING_MODE_UNSPECIFIED"
)

// KMSMappingProvisioningStatus describes the state of an AWS KMS mapping.
type KMSMappingProvisioningStatus string

const (
	KMSMappingProvisioningStatusConnected   KMSMappingProvisioningStatus = "KMS_MAPPING_PROVISIONING_STATUS_CONNECTED"
	KMSMappingProvisioningStatusConnecting  KMSMappingProvisioningStatus = "KMS_MAPPING_PROVISIONING_STATUS_CONNECTING"
	KMSMappingProvisioningStatusDisabling   KMSMappingProvisioningStatus = "KMS_MAPPING_PROVISIONING_STATUS_DISABLING"
	KMSMappingProvisioningStatusFailed      KMSMappingProvisioningStatus = "KMS_MAPPING_PROVISIONING_STATUS_FAILED"
	KMSMappingProvisioningStatusUnspecified KMSMappingProvisioningStatus = "KMS_MAPPING_PROVISIONING_STATUS_UNSPECIFIED"
)

// AWSKMSMappingStatus reports the sharing mode and provisioning state for an
// application cloud account. A CONNECTING mapping reports its requested mode
// before permissions are in effect.
type AWSKMSMappingStatus struct {
	AppCloudAccountID  uuid.UUID                    `json:"applicationCloudAccountId"`
	GatewayKMSKeyARN   string                       `json:"gatewayKmsKeyArn"`
	SharingMode        KMSSharingMode               `json:"kmsSharingMode"`
	ProvisioningStatus KMSMappingProvisioningStatus `json:"provisioningStatus"`
	StackARN           string                       `json:"stackArn"`
}

// AWSKMSMappingAction describes any AWS-side work required to complete a map
// or unmap. Apply the CloudFormation update or change the application
// account's IAM role policy, depending on the onboarding method.
type AWSKMSMappingAction struct {
	AppCloudAccountID uuid.UUID `json:"applicationCloudAccountId"`
	CloudFormationURL string    `json:"cloudFormationUrl"`
	ErrorMessage      string    `json:"errorMessage"`
	PendingApproval   bool      `json:"isPendingApproval"`
	ManualRolePolicy  string    `json:"manualRolePolicyJson"`
	TemplateURL       string    `json:"templateUrl"`
}

type awsKMSMappingReply struct {
	Success bool                  `json:"isSuccess"`
	Results []AWSKMSMappingAction `json:"results"`
}

// MapAWSCloudAccountWithKMSSharing maps an application account to an Exocompute
// host and requests automated KMS key sharing. It does not create a gateway key.
func MapAWSCloudAccountWithKMSSharing(ctx context.Context, gql *graphql.Client, appCloudAccountID, hostCloudAccountID uuid.UUID) (AWSKMSMappingAction, error) {
	query := mapAwsCloudAccountsWithKmsSharingQuery
	reply, err := requestAWSKMSMapping[awsKMSMappingReply](ctx, gql, query, struct {
		AppCloudAccountID  uuid.UUID `json:"applicationCloudAccountId"`
		HostCloudAccountID uuid.UUID `json:"exocomputeCloudAccountId"`
	}{appCloudAccountID, hostCloudAccountID})
	if err != nil {
		return AWSKMSMappingAction{}, err
	}
	// The backend omits ineligible applications rather than returning a
	// per-account error for them.
	if reply.Success && len(reply.Results) == 0 {
		return AWSKMSMappingAction{}, graphql.ResponseError(query, fmt.Errorf(
			"no KMS mapping result for application cloud account %s; verify that Cloud Native Protection or RDS Protection is enabled",
			appCloudAccountID,
		))
	}
	action, err := reply.action(appCloudAccountID)
	if err != nil {
		return AWSKMSMappingAction{}, graphql.ResponseError(query, err)
	}
	return action, nil
}

// UnmapAWSCloudAccountWithKMSSharing unmaps an application account and returns
// the AWS-side revoke action, if any. A pending revoke with no CloudFormation
// URL and an empty policy requires removing the policy and confirming the revoke.
func UnmapAWSCloudAccountWithKMSSharing(ctx context.Context, gql *graphql.Client, appCloudAccountID uuid.UUID) (AWSKMSMappingAction, error) {
	query := unmapAwsCloudAccountsWithKmsSharingQuery
	reply, err := requestAWSKMSMapping[awsKMSMappingReply](ctx, gql, query, struct {
		AppCloudAccountID uuid.UUID `json:"applicationCloudAccountId"`
	}{appCloudAccountID})
	if err != nil {
		return AWSKMSMappingAction{}, err
	}
	if reply.Success && len(reply.Results) == 0 {
		return AWSKMSMappingAction{AppCloudAccountID: appCloudAccountID}, nil
	}
	action, err := reply.action(appCloudAccountID)
	if err != nil {
		return AWSKMSMappingAction{}, graphql.ResponseError(query, err)
	}
	return action, nil
}

func (r awsKMSMappingReply) action(appCloudAccountID uuid.UUID) (AWSKMSMappingAction, error) {
	if len(r.Results) != 1 {
		return AWSKMSMappingAction{}, fmt.Errorf("expected one KMS mapping result, got %d", len(r.Results))
	}
	action := r.Results[0]
	if action.AppCloudAccountID != appCloudAccountID {
		return AWSKMSMappingAction{}, fmt.Errorf("KMS mapping result is for application cloud account %s, expected %s", action.AppCloudAccountID, appCloudAccountID)
	}
	if action.ErrorMessage != "" {
		return AWSKMSMappingAction{}, errors.New(action.ErrorMessage)
	}
	if !r.Success {
		return AWSKMSMappingAction{}, errors.New("KMS mapping operation did not succeed")
	}
	return action, nil
}

// GetAWSKMSMappingStatus returns the provisioning state for the application
// account, including mappings that are pending approval or being revoked.
func GetAWSKMSMappingStatus(ctx context.Context, gql *graphql.Client, appCloudAccountID uuid.UUID) (AWSKMSMappingStatus, error) {
	query := awsExocomputeKmsMappingStatusQuery
	reply, err := requestAWSKMSMapping[struct {
		Results []AWSKMSMappingStatus `json:"results"`
	}](ctx, gql, query, struct {
		AppCloudAccountID uuid.UUID `json:"applicationCloudAccountId"`
	}{appCloudAccountID})
	if err != nil {
		return AWSKMSMappingStatus{}, err
	}
	if len(reply.Results) != 1 || reply.Results[0].AppCloudAccountID != appCloudAccountID {
		return AWSKMSMappingStatus{}, graphql.ResponseError(query, fmt.Errorf("missing KMS mapping status for application cloud account %s", appCloudAccountID))
	}

	return reply.Results[0], nil
}

// UpdateManualKMSMapping confirms a sharing-only mapping or revoke for a
// manually onboarded AWS application account after the IAM change is applied.
func UpdateManualKMSMapping(ctx context.Context, gql *graphql.Client, appCloudAccountID uuid.UUID) error {
	query := updateManualKmsMappingQuery
	reply, err := requestAWSKMSMapping[struct {
		Results []struct {
			AppCloudAccountID uuid.UUID `json:"applicationCloudAccountId"`
			ErrorMessage      string    `json:"errorMessage"`
		} `json:"results"`
	}](ctx, gql, query, struct {
		AppCloudAccountID uuid.UUID `json:"applicationCloudAccountId"`
	}{appCloudAccountID})
	if err != nil {
		return err
	}
	if len(reply.Results) != 1 || reply.Results[0].AppCloudAccountID != appCloudAccountID {
		return graphql.ResponseError(query, fmt.Errorf("missing KMS confirmation result for application cloud account %s", appCloudAccountID))
	}
	if message := reply.Results[0].ErrorMessage; message != "" {
		return graphql.ResponseError(query, errors.New(message))
	}

	return nil
}

func requestAWSKMSMapping[R any](ctx context.Context, gql *graphql.Client, query string, params any) (R, error) {
	gql.Log().Print(log.Trace)

	var payload struct {
		Data struct {
			Result R `json:"result"`
		} `json:"data"`
	}
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return payload.Data.Result, graphql.RequestError(query, err)
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return payload.Data.Result, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}
