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

package cloudcluster

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

const (
	testProvisionClusterID = "6f1d4c7a-2b8e-4f3a-9c5d-1e2f3a4b5c6d"
	testProvisionAccountID = "0a9b8c7d-6e5f-4a3b-8c2d-1e0f9a8b7c6d"
)

// Shaped like the configuration RSC records for an Azure cluster create job.
// The admin password, cluster token and storage secret are scrubbed by RSC
// before the configuration is recorded. vm_type is EXTRA_DENSE (2) and
// instance_type is STANDARD_D8S_V5 (3).
const testAzureProvisionConfig = `{
  "account_id": "my-account",
  "user_id": "client|00000000-0000-0000-0000-000000000000",
  "cloud_account_id": "` + testProvisionAccountID + `",
  "is_es_type": true,
  "validations": [1],
  "cluster_config": {
    "cluster_uuid": "` + testProvisionClusterID + `",
    "cluster_name": "my-cluster",
    "num_nodes": 3,
    "dns_name_servers": ["10.0.0.2"],
    "dns_search_domains": ["example.com"],
    "ntp_servers": ["pool.ntp.org"],
    "azure_es_config": {
      "storage_account": "mystorage",
      "resource_group": "my-storage-rg",
      "container_name": "my-container",
      "enable_immutability": true,
      "managed_identity": {"name": "my-identity", "client_id": "11111111-1111-1111-1111-111111111111", "resource_group": "my-identity-rg"}
    }
  },
  "vm_config": {
    "cdm_version": "3.7.31669",
    "cdm_product": "rubrik-cdm-94",
    "vm_type": 2,
    "instance_type": 3,
    "location": "centralus",
    "resource_group": "my-rg",
    "network_resource_group": "my-net-rg",
    "vnet_resource_group": "my-vnet-rg",
    "vnet": "my-vnet",
    "network_security_group": "my-nsg",
    "network_security_resource_group": "my-nsg-rg",
    "subnet_az_configs": [
      {"subnet": "subnet-1", "availability_zone": "1"},
      {"subnet": "subnet-2", "availability_zone": "2"}
    ]
  },
  "is_az_resilient": true
}`

// Shaped like the configuration RSC records for an AWS cluster create job.
// vm_type is left out, which happens for the STANDARD VM type, and
// instance_type is M7I_2XLARGE (10), which the SDK has no constant for.
const testAwsProvisionConfig = `{
  "account_id": "my-account",
  "cloud_account_id": "` + testProvisionAccountID + `",
  "region": "us-east-2",
  "is_es_type": true,
  "keep_cluster_on_failure": true,
  "use_placement_groups": true,
  "cluster_config": {
    "cluster_name": "my-cluster",
    "num_nodes": 1,
    "dns_name_servers": ["169.254.169.253"],
    "ntp_servers": ["169.254.169.123"],
    "dynamic_scaling_enabled": true,
    "aws_es_config": {"bucket_name": "my-bucket", "enable_object_lock": true, "enable_immutability": true}
  },
  "vm_config": {
    "cdm_version": "9.4.0-p2-30507",
    "cdm_product": "rubrik-cdm-product-code",
    "instance_type": 10,
    "instance_profile_name": "my-profile",
    "vpc": "vpc-0123456789abcdef0",
    "subnet": "subnet-0123456789abcdef0",
    "security_groups": ["sg-0123456789abcdef0", "sg-0fedcba9876543210"]
  }
}`

// provisionResponse returns a ccProvisionMetadata response holding the
// specified values.
func provisionResponse(t *testing.T, clusterID, jobType, status, vendor, config string) []byte {
	t.Helper()

	buf, err := json.Marshal(map[string]any{
		"data": map[string]any{
			"result": map[string]any{
				"clusterUuid":     clusterID,
				"jobType":         jobType,
				"status":          status,
				"vendor":          vendor,
				"marshaledConfig": config,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return buf
}

// provisionServer returns a test server responding with the specified
// response. The variables of the request are stored in vars.
func provisionServer(cancel context.CancelCauseFunc, response []byte, vars *map[string]string) *httptest.Server {
	return httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}
		var payload struct {
			Variables map[string]string `json:"variables"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			cancel(err)
			return
		}
		*vars = payload.Variables
		if _, err := w.Write(response); err != nil {
			cancel(err)
		}
	}))
}

func TestProvisionConfigAzure(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	var vars map[string]string
	srv := provisionServer(cancel, provisionResponse(t, testProvisionClusterID, "CLUSTER_CREATE", "COMPLETED", "AZURE", testAzureProvisionConfig), &vars)
	defer srv.Close()

	config, err := Wrap(graphql.NewTestClient(srv)).ProvisionConfig(ctx, uuid.MustParse(testProvisionClusterID))
	if err != nil {
		t.Fatal(err)
	}

	if vars["clusterUuid"] != testProvisionClusterID || vars["jobType"] != "CLUSTER_CREATE" {
		t.Errorf("invalid request variables: %v", vars)
	}
	if config.AWS != nil || config.Azure == nil {
		t.Fatalf("expected only the Azure configuration, got AWS: %v, Azure: %v", config.AWS, config.Azure)
	}
	expected := AzureProvisionConfig{
		CloudAccountID:       uuid.MustParse(testProvisionAccountID),
		IsAzResilient:        true,
		KeepClusterOnFailure: false,
		ClusterConfig: ProvisionClusterConfig{
			ClusterName:      "my-cluster",
			NumNodes:         3,
			DNSNameServers:   []string{"10.0.0.2"},
			DNSSearchDomains: []string{"example.com"},
			NTPServers:       []string{"pool.ntp.org"},
		},
		ImageVersion:                 "3.7.31669",
		SKU:                          "rubrik-cdm-94",
		Region:                       azure.RegionCentralUS,
		InstanceType:                 AzureInstanceTypeStandardD8SV5,
		ResourceGroup:                "my-rg",
		NetworkResourceGroup:         "my-net-rg",
		VnetResourceGroup:            "my-vnet-rg",
		Vnet:                         "my-vnet",
		NetworkSecurityGroup:         "my-nsg",
		NetworkSecurityResourceGroup: "my-nsg-rg",
		SubnetAzConfigs: []SubnetAzConfig{
			{Subnet: "subnet-1", AvailabilityZone: "1"},
			{Subnet: "subnet-2", AvailabilityZone: "2"},
		},
		StorageAccount:               "mystorage",
		StorageResourceGroup:         "my-storage-rg",
		ContainerName:                "my-container",
		EnableImmutability:           true,
		ManagedIdentityName:          "my-identity",
		ManagedIdentityResourceGroup: "my-identity-rg",
		VMType:                       CCVmConfigExtraDense,
		VMTypeKnown:                  true,
	}
	if !reflect.DeepEqual(*config.Azure, expected) {
		t.Errorf("invalid Azure configuration:\n got: %+v\nwant: %+v", *config.Azure, expected)
	}
}

func TestProvisionConfigAws(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	var vars map[string]string
	srv := provisionServer(cancel, provisionResponse(t, testProvisionClusterID, "CLUSTER_CREATE", "COMPLETED", "AWS", testAwsProvisionConfig), &vars)
	defer srv.Close()

	config, err := Wrap(graphql.NewTestClient(srv)).ProvisionConfig(ctx, uuid.MustParse(testProvisionClusterID))
	if err != nil {
		t.Fatal(err)
	}

	if config.AWS == nil || config.Azure != nil {
		t.Fatalf("expected only the AWS configuration, got AWS: %v, Azure: %v", config.AWS, config.Azure)
	}
	expected := AwsProvisionConfig{
		CloudAccountID:       uuid.MustParse(testProvisionAccountID),
		Region:               aws.RegionUsEast2,
		KeepClusterOnFailure: true,
		UsePlacementGroups:   true,
		ClusterConfig: ProvisionClusterConfig{
			ClusterName:           "my-cluster",
			NumNodes:              1,
			DNSNameServers:        []string{"169.254.169.253"},
			NTPServers:            []string{"169.254.169.123"},
			DynamicScalingEnabled: true,
		},
		BucketName:          "my-bucket",
		EnableImmutability:  true,
		EnableObjectLock:    true,
		CDMVersion:          "9.4.0-p2-30507",
		CDMProduct:          "rubrik-cdm-product-code",
		InstanceType:        "M7I_2XLARGE",
		InstanceProfileName: "my-profile",
		VPC:                 "vpc-0123456789abcdef0",
		Subnet:              "subnet-0123456789abcdef0",
		SecurityGroups:      []string{"sg-0123456789abcdef0", "sg-0fedcba9876543210"},
		VMType:              CCVmConfigStandard,
		VMTypeKnown:         false,
	}
	if !reflect.DeepEqual(*config.AWS, expected) {
		t.Errorf("invalid AWS configuration:\n got: %+v\nwant: %+v", *config.AWS, expected)
	}
}

// TestProvisionConfigInstanceTypes verifies that every instance type RSC can
// record in the configuration of a cluster create job is decoded to the
// instance type the RSC API uses for it. The numbers are the values of the
// AwsInstanceType and AzureInstanceType enums in the cluster-mgmt protobuf
// definition.
func TestProvisionConfigInstanceTypes(t *testing.T) {
	awsTypes := map[int]AwsCCInstanceType{
		1: "M5_4XLARGE", 2: "M6I_2XLARGE", 3: "M6I_4XLARGE", 4: "M6I_8XLARGE", 5: "R6I_4XLARGE",
		6: "M6A_2XLARGE", 7: "M6A_4XLARGE", 8: "M6A_8XLARGE", 9: "R6A_4XLARGE",
		10: "M7I_2XLARGE", 11: "M7I_4XLARGE", 12: "M7I_8XLARGE",
		13: "M7A_2XLARGE", 14: "M7A_4XLARGE", 15: "M7A_8XLARGE",
		16: "R7I_4XLARGE", 17: "R7A_4XLARGE",
	}
	for number, want := range awsTypes {
		if got := provisionAwsInstanceTypes[number]; got != want {
			t.Errorf("AWS instance type %d: got %q, want %q", number, got, want)
		}
	}
	if len(provisionAwsInstanceTypes) != len(awsTypes)+1 { // + AWS_TYPE_UNSPECIFIED
		t.Errorf("AWS: got %d instance types, want %d", len(provisionAwsInstanceTypes), len(awsTypes)+1)
	}

	azureTypes := map[int]AzureCCESSupportedInstanceType{
		1: "STANDARD_DS5_V2", 2: "STANDARD_D16S_V5", 3: "STANDARD_D8S_V5", 4: "STANDARD_D32S_V5",
		5: "STANDARD_E16S_V5", 6: "STANDARD_D8AS_V5", 7: "STANDARD_D16AS_V5", 8: "STANDARD_D32AS_V5",
		9: "STANDARD_E16AS_V5", 10: "STANDARD_D8S_V6", 11: "STANDARD_D16S_V6", 12: "STANDARD_E16S_V6",
		13: "STANDARD_D8AS_V6", 14: "STANDARD_D16AS_V6", 15: "STANDARD_E16AS_V6",
	}
	for number, want := range azureTypes {
		if got := provisionAzureInstanceTypes[number]; got != want {
			t.Errorf("Azure instance type %d: got %q, want %q", number, got, want)
		}
	}
	if len(provisionAzureInstanceTypes) != len(azureTypes)+1 { // + TYPE_UNSPECIFIED
		t.Errorf("Azure: got %d instance types, want %d", len(provisionAzureInstanceTypes), len(azureTypes)+1)
	}
}

// TestProvisionConfigInstanceTypesAreComplete verifies that every AWS and Azure
// instance type constant has a number in the maps used to decode the recorded
// configuration of a cluster create job. A constant without a number means
// that a cluster created with the instance type can't be decoded. The lists of
// constants must be kept in sync with TestInstanceTypeEnumCanary in the parent
// package, which verifies them against RSC.
func TestProvisionConfigInstanceTypesAreComplete(t *testing.T) {
	awsTypes := []AwsCCInstanceType{
		AwsInstanceTypeUnspecified, AwsInstanceTypeM5_4XLarge, AwsInstanceTypeM6I_2XLarge,
		AwsInstanceTypeM6I_4XLarge, AwsInstanceTypeM6I_8XLarge, AwsInstanceTypeR6I_4XLarge,
		AwsInstanceTypeM6A_2XLarge, AwsInstanceTypeM6A_4XLarge, AwsInstanceTypeM6A_8XLarge,
		AwsInstanceTypeR6A_4XLarge, AwsInstanceTypeM7I_2XLarge, AwsInstanceTypeM7I_4XLarge,
		AwsInstanceTypeM7I_8XLarge, AwsInstanceTypeR7I_4XLarge, AwsInstanceTypeM7A_2XLarge,
		AwsInstanceTypeM7A_4XLarge, AwsInstanceTypeM7A_8XLarge, AwsInstanceTypeR7A_4XLarge,
	}
	decodable := make(map[AwsCCInstanceType]struct{}, len(provisionAwsInstanceTypes))
	for _, instanceType := range provisionAwsInstanceTypes {
		decodable[instanceType] = struct{}{}
	}
	for _, instanceType := range awsTypes {
		if _, ok := decodable[instanceType]; !ok {
			t.Errorf("AWS instance type %q has no number in provisionAwsInstanceTypes", instanceType)
		}
	}

	azureTypes := []AzureCCESSupportedInstanceType{
		AzureInstanceTypeUnspecified, AzureInstanceTypeStandardDS5V2, AzureInstanceTypeStandardD16SV5,
		AzureInstanceTypeStandardD8SV5, AzureInstanceTypeStandardD32SV5, AzureInstanceTypeStandardE16SV5,
		AzureInstanceTypeStandardD8ASV5, AzureInstanceTypeStandardD16ASV5, AzureInstanceTypeStandardD32ASV5,
		AzureInstanceTypeStandardE16ASV5, AzureInstanceTypeStandardD8SV6, AzureInstanceTypeStandardD16SV6,
		AzureInstanceTypeStandardE16SV6, AzureInstanceTypeStandardD8ASV6, AzureInstanceTypeStandardD16ASV6,
		AzureInstanceTypeStandardE16ASV6,
	}
	azureDecodable := make(map[AzureCCESSupportedInstanceType]struct{}, len(provisionAzureInstanceTypes))
	for _, instanceType := range provisionAzureInstanceTypes {
		azureDecodable[instanceType] = struct{}{}
	}
	for _, instanceType := range azureTypes {
		if _, ok := azureDecodable[instanceType]; !ok {
			t.Errorf("Azure instance type %q has no number in provisionAzureInstanceTypes", instanceType)
		}
	}
}

// TestProvisionConfigErrors covers the responses which don't hold the
// configuration of a cluster create job.
func TestProvisionConfigErrors(t *testing.T) {
	otherClusterID := "9e8d7c6b-5a4f-4e3d-8c2b-1a0f9e8d7c6b"

	tests := []struct {
		name     string
		response func(t *testing.T) []byte
		notFound bool
	}{{
		// RSC has no jobs at all for the cluster.
		name: "NoJobs",
		response: func(t *testing.T) []byte {
			return provisionResponse(t, testProvisionClusterID, "CLUSTER_CREATE", "JOB_NOTFOUND", "", "[]")
		},
		notFound: true,
	}, {
		// RSC has no create job, it returns the configuration of the
		// cluster's latest job together with the JOB_NOTFOUND status.
		name: "NoCreateJob",
		response: func(t *testing.T) []byte {
			return provisionResponse(t, testProvisionClusterID, "CLUSTER_CREATE", "JOB_NOTFOUND", "", testAzureProvisionConfig)
		},
		notFound: true,
	}, {
		name: "WrongJobType",
		response: func(t *testing.T) []byte {
			return provisionResponse(t, testProvisionClusterID, "CLUSTER_RECOVER", "COMPLETED", "AZURE", testAzureProvisionConfig)
		},
		notFound: true,
	}, {
		name: "ClusterIDMismatch",
		response: func(t *testing.T) []byte {
			return provisionResponse(t, otherClusterID, "CLUSTER_CREATE", "COMPLETED", "AZURE", testAzureProvisionConfig)
		},
	}, {
		name: "UnsupportedVendor",
		response: func(t *testing.T) []byte {
			return provisionResponse(t, testProvisionClusterID, "CLUSTER_CREATE", "COMPLETED", "GCP", "{}")
		},
	}, {
		name: "UnknownInstanceType",
		response: func(t *testing.T) []byte {
			return provisionResponse(t, testProvisionClusterID, "CLUSTER_CREATE", "COMPLETED", "AZURE",
				`{"cloud_account_id": "`+testProvisionAccountID+`", "vm_config": {"location": "centralus", "instance_type": 99}}`)
		},
	}, {
		name: "UnknownRegion",
		response: func(t *testing.T) []byte {
			return provisionResponse(t, testProvisionClusterID, "CLUSTER_CREATE", "COMPLETED", "AWS",
				`{"cloud_account_id": "`+testProvisionAccountID+`", "region": "mars-north-1"}`)
		},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer assert.Context(t, ctx, cancel)

			var vars map[string]string
			srv := provisionServer(cancel, test.response(t), &vars)
			defer srv.Close()

			_, err := Wrap(graphql.NewTestClient(srv)).ProvisionConfig(ctx, uuid.MustParse(testProvisionClusterID))
			if err == nil {
				t.Fatal("expected an error")
			}
			if notFound := errors.Is(err, graphql.ErrNotFound); notFound != test.notFound {
				t.Errorf("errors.Is(err, graphql.ErrNotFound) = %v, want %v: %v", notFound, test.notFound, err)
			}
		})
	}
}
