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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CloudAccount represents an RSC Cloud Account for AWS.
type CloudAccount struct {
	Cloud               Cloud     `json:"cloudType"`
	ID                  uuid.UUID `json:"id"`
	NativeID            string    `json:"nativeId"`
	Name                string    `json:"accountName"`
	Message             string    `json:"message"`
	SeamlessFlowEnabled bool      `json:"seamlessFlowEnabled"`
}

// Feature represents an RSC Cloud Account feature for AWS, e.g Cloud Native
// Protection.
type Feature struct {
	Name     core.Feature `json:"feature"`
	Regions  []Region     `json:"awsRegions"`
	RoleArn  string       `json:"roleArn"`
	StackArn string       `json:"stackArn"`
	Status   core.Status  `json:"status"`
}

// FeatureVersion maps an RSC Cloud Account feature to a version number.
type FeatureVersion struct {
	Name    string `json:"feature"`
	Version int    `json:"version"`
}

// CloudAccountWithFeatures hold details about a cloud account and the features
// associated with that account.
type CloudAccountWithFeatures struct {
	Account  CloudAccount `json:"awsCloudAccount"`
	Features []Feature    `json:"featureDetails"`
}

// CloudAccountWithFeatures returns the cloud account with the specified
// Polaris cloud account id.
func (a API) CloudAccountWithFeatures(ctx context.Context, id uuid.UUID, feature core.Feature) (CloudAccountWithFeatures, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, awsCloudAccountWithFeaturesQuery, struct {
		ID       uuid.UUID      `json:"cloudAccountId"`
		Features []core.Feature `json:"features"`
	}{ID: id, Features: []core.Feature{feature}})
	if err != nil {
		return CloudAccountWithFeatures{}, fmt.Errorf("failed to request awsCloudAccountWithFeatures: %w", err)
	}
	a.log.Printf(log.Debug, "awsCloudAccountWithFeatures(%q, %q): %s", id, feature, string(buf))

	var payload struct {
		Data struct {
			Result CloudAccountWithFeatures `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return CloudAccountWithFeatures{}, fmt.Errorf("failed to unmarshal awsCloudAccountWithFeatures: %v", err)
	}

	return payload.Data.Result, nil
}

// CloudAccountsWithFeatures returns the cloud accounts matching the specified
// filter. The filter can be used to search for AWS account id, account name
// and role arn.
func (a API) CloudAccountsWithFeatures(ctx context.Context, feature core.Feature, filter string) ([]CloudAccountWithFeatures, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, allAwsCloudAccountsWithFeaturesQuery, struct {
		Feature core.Feature `json:"feature"`
		Filter  string       `json:"columnSearchFilter"`
	}{Filter: filter, Feature: feature})
	if err != nil {
		return nil, fmt.Errorf("failed to request allAwsCloudAccountsWithFeatures: %w", err)
	}
	a.log.Printf(log.Debug, "allAwsCloudAccountsWithFeatures(%q, %q): %s", filter, feature, string(buf))

	var payload struct {
		Data struct {
			Result []CloudAccountWithFeatures `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allAwsCloudAccountsWithFeatures: %v", err)
	}

	return payload.Data.Result, nil
}

// CloudAccountInitiate holds information about the CloudFormation stack
// that needs to be created in AWS to give permission to Polaris for managing
// the account being added. It also holds feature version information.
type CloudAccountInitiate struct {
	CloudFormationURL string           `json:"cloudFormationUrl"`
	ExternalID        string           `json:"externalId"` // Deprecated: no replacement.
	FeatureVersions   []FeatureVersion `json:"featureVersions"`
	StackName         string           `json:"stackName"`
	TemplateURL       string           `json:"templateUrl"`
}

// ValidateAndCreateCloudAccount begins the process of adding the specified AWS
// account to RSC. The returned CloudAccountInitiate value must be passed on to
// FinalizeCloudAccountProtection which is the next step in the process of
// adding an AWS account to RSC.
func (a API) ValidateAndCreateCloudAccount(ctx context.Context, id, name string, feature core.Feature) (CloudAccountInitiate, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, validateAndCreateAwsCloudAccountQuery, struct {
		ID       string         `json:"nativeId"`
		Name     string         `json:"accountName"`
		Features []core.Feature `json:"features"`
	}{ID: id, Name: name, Features: []core.Feature{feature}})
	if err != nil {
		return CloudAccountInitiate{}, fmt.Errorf("failed to request validateAndCreateAwsCloudAccount: %w", err)
	}
	a.log.Printf(log.Debug, "validateAndCreateAwsCloudAccount(%q, %q, %q): %s", id, name, feature, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				InitiateResponse CloudAccountInitiate `json:"initiateResponse"`
				ValidateResponse struct {
					InvalidAwsAccounts []struct {
						AccountName string `json:"accountName"`
						NativeID    string `json:"nativeId"`
						Message     string `json:"message"`
					} `json:"invalidAwsAccounts"`
					InvalidAwsAdminAccount struct {
						AccountName string `json:"accountName"`
						NativeID    string `json:"nativeId"`
						Message     string `json:"message"`
					} `json:"invalidAwsAdminAccount"`
				} `json:"validateResponse"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return CloudAccountInitiate{}, fmt.Errorf("failed to unmarshal validateAndCreateAwsCloudAccount: %v", err)
	}
	if msg := payload.Data.Result.ValidateResponse.InvalidAwsAdminAccount.Message; msg != "" {
		return CloudAccountInitiate{}, fmt.Errorf("invalid admin account: %v", msg)
	}
	if accounts := payload.Data.Result.ValidateResponse.InvalidAwsAccounts; len(accounts) != 0 {
		return CloudAccountInitiate{}, fmt.Errorf("invalid account: %v", accounts[0].Message)
	}

	return payload.Data.Result.InitiateResponse, nil
}

// FinalizeCloudAccountProtection finalizes the process of the adding the
// specified AWS account to RSC. The message returned by the GraphQL API is
// converted into a Go error. After this function a CloudFormation stack must
// be created using the information returned by ValidateAndCreateCloudAccount.
func (a API) FinalizeCloudAccountProtection(ctx context.Context, id, name string, feature core.Feature, regions []Region, init CloudAccountInitiate) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, finalizeAwsCloudAccountProtectionQuery, struct {
		ID             string           `json:"nativeId"`
		Name           string           `json:"accountName"`
		Regions        []Region         `json:"awsRegions,omitempty"`
		ExternalID     string           `json:"externalId"`
		FeatureVersion []FeatureVersion `json:"featureVersion"`
		Feature        core.Feature     `json:"feature"`
		StackName      string           `json:"stackName"`
	}{ID: id, Name: name, Regions: regions, ExternalID: init.ExternalID, FeatureVersion: init.FeatureVersions, Feature: feature, StackName: init.StackName})
	if err != nil {
		return fmt.Errorf("failed to request finalizeAwsCloudAccountProtection: %w", err)
	}
	a.log.Printf(log.Debug, "finalizeAwsCloudAccountProtection(%q, %q, %q, %q, %v, %q, %q): %s", id, name, regions, init.ExternalID,
		init.FeatureVersions, feature, init.StackName, string(buf))

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
		return fmt.Errorf("failed to unmarshal finalizeAwsCloudAccountProtection: %v", err)
	}

	// On success the message starts with "successfully".
	if !strings.HasPrefix(strings.ToLower(payload.Data.Query.Message), "successfully") {
		return errors.New(payload.Data.Query.Message)
	}

	return nil
}

// PrepareCloudAccountDeletion prepares the deletion of the cloud account
// identified by the specified RSC cloud account id.
// FinalizeCloudAccountDeletion is the next step in the process.
func (a API) PrepareCloudAccountDeletion(ctx context.Context, id uuid.UUID, feature core.Feature) (cloudFormationURL string, err error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, prepareAwsCloudAccountDeletionQuery, struct {
		ID      uuid.UUID    `json:"cloudAccountId"`
		Feature core.Feature `json:"feature"`
	}{ID: id, Feature: feature})
	if err != nil {
		return "", fmt.Errorf("failed to request prepareAwsCloudAccountDeletion: %w", err)
	}
	a.log.Printf(log.Debug, "prepareAwsCloudAccountDeletion(%q, %q): %s", id, feature, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				URL string `json:"cloudFormationUrl"`
			} `json:"prepareAwsCloudAccountDeletion"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", fmt.Errorf("failed to unmarshal prepareAwsCloudAccountDeletion: %v", err)
	}
	if payload.Data.Query.URL == "" {
		return "", errors.New("CloudFormation url is empty")
	}

	return payload.Data.Query.URL, nil
}

// FinalizeCloudAccountDeletion finalizes the deletion of the cloud account
// identified by the specified RSC cloud account id. The message returned by the
// GraphQL API call is converted into a Go error.
func (a API) FinalizeCloudAccountDeletion(ctx context.Context, id uuid.UUID, feature core.Feature) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, finalizeAwsCloudAccountDeletionQuery, struct {
		ID      uuid.UUID    `json:"cloudAccountId"`
		Feature core.Feature `json:"feature"`
	}{ID: id, Feature: feature})
	if err != nil {
		return fmt.Errorf("failed to request finalizeAwsCloudAccountDeletion: %w", err)
	}
	a.log.Printf(log.Debug, "finalizeAwsCloudAccountDeletion(%q, %q): %s", id, feature, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Message string `json:"message"`
			} `json:"finalizeAwsCloudAccountDeletion"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal finalizeAwsCloudAccountDeletion: %v", err)
	}

	// On success the message starts with "successfully".
	if !strings.HasPrefix(strings.ToLower(payload.Data.Query.Message), "successfully") {
		return errors.New(payload.Data.Query.Message)
	}

	return nil
}

// UpdateCloudAccountFeature updates the settings of the cloud account. The
// message returned by the GraphQL API call is converted into a Go error. At
// this time only the regions can be updated.
func (a API) UpdateCloudAccountFeature(ctx context.Context, action core.CloudAccountAction, id uuid.UUID, feature core.Feature, regions []Region) error {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, updateAwsCloudAccountQuery, struct {
		Action  core.CloudAccountAction `json:"action"`
		ID      uuid.UUID               `json:"cloudAccountId"`
		Regions []Region                `json:"awsRegions"`
		Feature core.Feature            `json:"feature"`
	}{Action: action, ID: id, Regions: regions, Feature: feature})
	if err != nil {
		return fmt.Errorf("failed to request updateAwsCloudAccount: %w", err)
	}
	a.log.Printf(log.Debug, "updateAwsCloudAccount(%q, %q, %q, %q): %s", action, id, regions, feature, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Message string `json:"message"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal updateAwsCloudAccount: %v", err)
	}

	// On success the message starts with "successfully".
	if !strings.HasPrefix(strings.ToLower(payload.Data.Result.Message), "successfully") {
		return errors.New(payload.Data.Result.Message)
	}

	return nil
}

// VPC represents an AWS VPC together with AWS subnets and AWS security groups.
type VPC struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Subnets []struct {
		ID               string `json:"id"`
		Name             string `json:"name"`
		AvailabilityZone string `json:"availabilityZone"`
	} `json:"subnets"`
	SecurityGroups []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"securityGroups"`
}

// AllVpcsByRegion returns all VPCs including their subnets for the specified
// RSC cloud account id.
func (a API) AllVpcsByRegion(ctx context.Context, id uuid.UUID, regions Region) ([]VPC, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, allVpcsByRegionFromAwsQuery, struct {
		ID     uuid.UUID `json:"awsAccountRubrikId"`
		Region Region    `json:"region"`
	}{ID: id, Region: regions})
	if err != nil {
		return nil, fmt.Errorf("failed to request allVpcsByRegionFromAws: %w", err)
	}
	a.log.Printf(log.Debug, "allVpcsByRegionFromAws(%q, %q): %s", id, regions, string(buf))

	var payload struct {
		Data struct {
			VPCs []VPC `json:"allVpcsByRegionFromAws"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allVpcsByRegionFromAws: %v", err)
	}

	return payload.Data.VPCs, nil
}

// PrepareFeatureUpdateForAwsCloudAccount returns a CloudFormation URL and a
// template URL from RSC which can be used to update the CloudFormation stack.
func (a API) PrepareFeatureUpdateForAwsCloudAccount(ctx context.Context, id uuid.UUID, features []core.Feature) (cfmURL string, tmplURL string, err error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, prepareFeatureUpdateForAwsCloudAccountQuery, struct {
		ID       uuid.UUID      `json:"cloudAccountId"`
		Features []core.Feature `json:"features"`
	}{ID: id, Features: features})
	if err != nil {
		return "", "", fmt.Errorf("failed to request prepareFeatureUpdateForAwsCloudAccount: %w", err)
	}
	a.log.Printf(log.Debug, "prepareFeatureUpdateForAwsCloudAccount(%q, %v): %s", id, features, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				CloudFormationURL string `json:"cloudFormationUrl"`
				TemplateURL       string `json:"templateUrl"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal prepareFeatureUpdateForAwsCloudAccount: %v", err)
	}

	return payload.Data.Result.CloudFormationURL, payload.Data.Result.TemplateURL, nil
}
