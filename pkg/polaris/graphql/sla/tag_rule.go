// Copyright 2024 Rubrik, Inc.
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

package sla

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CloudNativeTagObjectType represents the valid object type values when
// creating a cloud native tag rule.
type CloudNativeTagObjectType string

const (
	TagObjectAWSEBSVolume                  CloudNativeTagObjectType = "AWS_EBS_VOLUME"
	TagObjectAWSEC2Instance                CloudNativeTagObjectType = "AWS_EC2_INSTANCE"
	TagObjectAWSDynamoDBTable              CloudNativeTagObjectType = "AWS_DYNAMODB_TABLE"
	TagObjectAWSRDSInstance                CloudNativeTagObjectType = "AWS_RDS_INSTANCE"
	TagObjectAWSS3Bucket                   CloudNativeTagObjectType = "AWS_S3_BUCKET"
	TagObjectAzureManagedDisk              CloudNativeTagObjectType = "AZURE_MANAGED_DISK"
	TagObjectAzureSQLDatabaseDB            CloudNativeTagObjectType = "AZURE_SQL_DATABASE_DB"
	TagObjectAzureSQLDatabaseServer        CloudNativeTagObjectType = "AZURE_SQL_DATABASE_SERVER"
	TagObjectAzureSQLManagedInstanceServer CloudNativeTagObjectType = "AZURE_SQL_MANAGED_INSTANCE_SERVER"
	TagObjectAzureStorageAccount           CloudNativeTagObjectType = "AZURE_STORAGE_ACCOUNT"
	TagObjectAzureVirtualMachine           CloudNativeTagObjectType = "AZURE_VIRTUAL_MACHINE"
)

// AllCloudNativeTagObjectTypes returns all cloud native tag object types.
func AllCloudNativeTagObjectTypes() []CloudNativeTagObjectType {
	return []CloudNativeTagObjectType{
		TagObjectAWSEBSVolume,
		TagObjectAWSEC2Instance,
		TagObjectAWSRDSInstance,
		TagObjectAWSDynamoDBTable,
		TagObjectAWSS3Bucket,
		TagObjectAzureManagedDisk,
		TagObjectAzureSQLDatabaseDB,
		TagObjectAzureSQLDatabaseServer,
		TagObjectAzureSQLManagedInstanceServer,
		TagObjectAzureStorageAccount,
		TagObjectAzureVirtualMachine,
	}
}

// AllCloudNativeTagObjectTypesAsStrings returns all cloud native tag object
// types as a slice of strings.
func AllCloudNativeTagObjectTypesAsStrings() []string {
	return []string{
		string(TagObjectAWSEBSVolume),
		string(TagObjectAWSEC2Instance),
		string(TagObjectAWSRDSInstance),
		string(TagObjectAWSDynamoDBTable),
		string(TagObjectAWSS3Bucket),
		string(TagObjectAzureManagedDisk),
		string(TagObjectAzureSQLDatabaseDB),
		string(TagObjectAzureSQLDatabaseServer),
		string(TagObjectAzureSQLManagedInstanceServer),
		string(TagObjectAzureStorageAccount),
		string(TagObjectAzureVirtualMachine),
	}
}

var managedObjectTypeMap = map[ManagedObjectType]CloudNativeTagObjectType{
	AWSNativeEBSVolume:            TagObjectAWSEBSVolume,
	AWSNativeEC2Instance:          TagObjectAWSEC2Instance,
	AWSNativeDynamoDBTable:        TagObjectAWSDynamoDBTable,
	AWSNativeRDSInstance:          TagObjectAWSRDSInstance,
	AWSNativeS3Bucket:             TagObjectAWSS3Bucket,
	AzureManagedDisk:              TagObjectAzureManagedDisk,
	AzureSQLDatabaseDB:            TagObjectAzureSQLDatabaseDB,
	AzureSQLDatabaseServer:        TagObjectAzureSQLDatabaseServer,
	AzureSQLManagedInstanceServer: TagObjectAzureSQLManagedInstanceServer,
	AzureStorageAccount:           TagObjectAzureStorageAccount,
	AzureVirtualMachine:           TagObjectAzureVirtualMachine,
}

// FromManagedObjectType returns the corresponding CloudNativeTagObjectType for
// the given ManagedObjectType.
func FromManagedObjectType(objectType ManagedObjectType) (CloudNativeTagObjectType, error) {
	if tagObjectType, ok := managedObjectTypeMap[objectType]; ok {
		return tagObjectType, nil
	}

	return "", fmt.Errorf("unsupported managed object type: %s", objectType)
}

// ManagedObjectType represents the object type of managed objects.
type ManagedObjectType string

const (
	AWSNativeAccount                   ManagedObjectType = "AWS_NATIVE_ACCOUNT"
	AWSNativeEBSVolume                 ManagedObjectType = "AWS_NATIVE_EBS_VOLUME"
	AWSNativeEC2Instance               ManagedObjectType = "AWS_NATIVE_EC2_INSTANCE"
	AWSNativeRDSInstance               ManagedObjectType = "AWS_NATIVE_RDS_INSTANCE"
	AWSNativeDynamoDBTable             ManagedObjectType = "AWS_NATIVE_DYNAMODB_TABLE"
	AWSNativeS3Bucket                  ManagedObjectType = "AWS_NATIVE_S3_BUCKET"
	AzureManagedDisk                   ManagedObjectType = "AZURE_MANAGED_DISK"
	AzureResourceGroup                 ManagedObjectType = "AZURE_RESOURCE_GROUP"
	AzureResourceGroupForVMHierarchy   ManagedObjectType = "AZURE_RESOURCE_GROUP_FOR_VM_HIERARCHY"
	AzureResourceGroupFprDiskHierarchy ManagedObjectType = "AZURE_RESOURCE_GROUP_FOR_DISK_HIERARCHY"
	AzureSQLDatabaseDB                 ManagedObjectType = "AZURE_SQL_DATABASE_DB"
	AzureSQLDatabaseServer             ManagedObjectType = "AZURE_SQL_DATABASE_SERVER"
	AzureSQLManagedInstanceDB          ManagedObjectType = "AZURE_SQL_MANAGED_INSTANCE_DB"
	AzureSQLManagedInstanceServer      ManagedObjectType = "AZURE_SQL_MANAGED_INSTANCE_SERVER"
	AzureStorageAccount                ManagedObjectType = "AZURE_STORAGE_ACCOUNT"
	AzureSubscription                  ManagedObjectType = "AZURE_SUBSCRIPTION"
	AzureUnmanagedDisk                 ManagedObjectType = "AZURE_UNMANAGED_DISK"
	AzureVirtualMachine                ManagedObjectType = "AZURE_VIRTUAL_MACHINE"
	CloudNativeTagRule                 ManagedObjectType = "CLOUD_NATIVE_TAG_RULE"
	GCPNativeDisk                      ManagedObjectType = "GCP_NATIVE_DISK"
	GCPNativeGCEInstance               ManagedObjectType = "GCP_NATIVE_GCE_INSTANCE"
	GCPNativeProject                   ManagedObjectType = "GCP_NATIVE_PROJECT"
)

// TagRule represents an RSC tag rule. Note, the ID field of the EffectiveSLA
// field is either a UUID or one of the strings: doNotProtect, noAssignment.
type TagRule struct {
	ID                uuid.UUID         `json:"id"`
	Name              string            `json:"name"`
	ObjectType        ManagedObjectType `json:"objectType"`
	Tag               Tag               `json:"tag"`
	AllACloudAccounts bool              `json:"applyToAllCloudAccounts"`
	CloudAccounts     []struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	} `json:"cloudNativeAccounts"`
	EffectiveDomain struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"effectiveSla"`
}

// Tag represents the tag of an RSC tag rule.
type Tag struct {
	Key       string `json:"tagKey"`
	Value     string `json:"tagValue"`
	AllValues bool   `json:"matchAllValues"`
}

// TagRuleFilter holds the filter for a tag rules list operation.
type TagRuleFilter struct {
	Field  string   `json:"field"`
	Values []string `json:"texts"`
}

// ListTagRules returns all RSC tag rules of the specified object type matching
// the specified tag rule filters.
func ListTagRules(ctx context.Context, gql *graphql.Client, objectType CloudNativeTagObjectType, filters []TagRuleFilter) ([]TagRule, error) {
	gql.Log().Print(log.Trace)

	// Skip retries when listing tag rules since some object types can result
	// in an error classified as a temporary error, even though it will never
	// succeed.
	query := cloudNativeTagRulesQuery
	buf, err := gql.RequestWithoutRetry(ctx, query, struct {
		ObjectType CloudNativeTagObjectType `json:"objectType"`
		Filters    []TagRuleFilter          `json:"filters,omitempty"`
	}{ObjectType: objectType, Filters: filters})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				TagRules []TagRule `json:"tagRules"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.TagRules, nil
}

// CreateTagRuleParams holds the parameters for a tag rule create operation.
type CreateTagRuleParams struct {
	Name             string                   `json:"tagRuleName"`
	ObjectType       CloudNativeTagObjectType `json:"objectType"`
	Tag              Tag                      `json:"tag"`
	CloudAccounts    *TagRuleCloudAccounts    `json:"cloudNativeAccountIds,omitempty"`
	AllCloudAccounts bool                     `json:"applyToAllCloudAccounts,omitempty"`
}

// TagRuleCloudAccounts holds the cloud accounts for a tag rule. Note, the IDs
// are Native Cloud Account IDs and not regular Cloud Account IDs.
type TagRuleCloudAccounts struct {
	AWSAccountIDs        []uuid.UUID `json:"awsNativeAccountIds,omitempty"`
	AzureSubscriptionIDs []uuid.UUID `json:"azureNativeSubscriptionIds,omitempty"`
	GCPProjectIDs        []uuid.UUID `json:"gcpNativeProjectIds,omitempty"`
}

// CreateTagRule creates a new tag rule. Returns the ID of the new tag rule.
func CreateTagRule(ctx context.Context, gql *graphql.Client, params CreateTagRuleParams) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	query := createCloudNativeTagRuleQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				ID string `json:"tagRuleId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	id, err := uuid.Parse(payload.Data.Result.ID)
	if err != nil {
		return uuid.Nil, graphql.ResponseError(query, err)
	}

	return id, nil
}

// UpdateTagRuleParams holds the parameters for a tag rule update operation.
type UpdateTagRuleParams struct {
	Name             string                `json:"tagRuleName"`
	CloudAccounts    *TagRuleCloudAccounts `json:"cloudNativeAccountIds,omitempty"`
	AllCloudAccounts bool                  `json:"applyToAllCloudAccounts,omitempty"`
}

// UpdateTagRule updates the tag rule with the specified ID.
func UpdateTagRule(ctx context.Context, gql *graphql.Client, tagRuleID uuid.UUID, params UpdateTagRuleParams) error {
	gql.Log().Print(log.Trace)

	query := updateCloudNativeTagRuleQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"tagRuleId"`
		UpdateTagRuleParams
	}{ID: tagRuleID, UpdateTagRuleParams: params})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct{} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}

// DeleteTagRule deletes the tag rule with the specified ID.
func DeleteTagRule(ctx context.Context, gql *graphql.Client, tagRuleID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deleteCloudNativeTagRuleQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"ruleId"`
	}{ID: tagRuleID})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct{} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}
