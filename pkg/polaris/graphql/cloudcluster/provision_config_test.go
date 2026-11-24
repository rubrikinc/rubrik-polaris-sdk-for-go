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

// Shaped like the masked GcpClusterCreateReq RSC records for a Multi-AZ cluster.
const testGcpProvisionConfig = `{
  "cloud_account_id": "` + testProvisionAccountID + `",
  "region": "us-west1",
  "zone": "us-west1-a",
  "is_es_type": true,
  "is_az_resilient": true,
  "keep_cluster_on_failure": true,
  "cluster_config": {
    "cluster_name": "my-cluster",
    "num_nodes": 3,
    "dns_name_servers": ["169.254.169.254"],
    "dns_search_domains": ["example.com"],
    "ntp_servers": ["time.google.com"],
    "gcp_es_config": {"bucket_name": "my-bucket", "region": "us-west1"}
  },
  "vm_config": {
    "instance_type": 6,
    "cdm_version": "9.4.0-p2-30507",
    "cdm_product": "rubrik-cdm",
    "delete_protection": true,
    "network_config": [
      {"network": "my-vpc", "host_project": "shared-vpc", "region": "us-west1", "name": "subnet-a"},
      {"network": "my-vpc", "host_project": "shared-vpc", "region": "us-west1", "name": "subnet-b"},
      {"network": "my-vpc", "host_project": "shared-vpc", "region": "us-west1", "name": "subnet-c"}
    ],
    "subnet_az_configs": [
      {"subnet": "subnet-a", "availability_zone": "us-west1-a"},
      {"subnet": "subnet-b", "availability_zone": "us-west1-b"},
      {"subnet": "subnet-c", "availability_zone": "us-west1-c"}
    ],
    "service_accounts": [{"email": "node@example.iam.gserviceaccount.com", "scopes": ["https://www.googleapis.com/auth/cloud-platform"]}]
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

func requestProvisionConfig(t *testing.T, vendor, config string) ProvisionConfig {
	t.Helper()

	ctx, cancel := context.WithCancelCause(t.Context())
	defer assert.Context(t, ctx, cancel)

	var vars map[string]string
	srv := provisionServer(cancel, provisionResponse(t, testProvisionClusterID, "CLUSTER_CREATE", "COMPLETED", vendor, config), &vars)
	defer srv.Close()

	result, err := Wrap(graphql.NewTestClient(srv)).ProvisionConfig(ctx, uuid.MustParse(testProvisionClusterID))
	if err != nil {
		t.Fatal(err)
	}
	if vars["clusterUuid"] != testProvisionClusterID || vars["jobType"] != "CLUSTER_CREATE" {
		t.Errorf("invalid request variables: %v", vars)
	}
	if result.ClusterUUID != testProvisionClusterID || result.JobType != "CLUSTER_CREATE" ||
		result.Status != "COMPLETED" || result.Vendor != vendor || result.MarshaledConfig != config {
		t.Errorf("invalid response fields: %+v", result)
	}
	return result
}

// The API layer returns the configuration as RSC recorded it, with enums as
// numbers and without any conversion.
func TestProvisionConfigAzure(t *testing.T) {
	config := requestProvisionConfig(t, "AZURE", testAzureProvisionConfig)
	if config.AWS != nil || config.Azure == nil || config.GCP != nil {
		t.Fatalf("expected only the Azure configuration, got %+v", config)
	}

	vmType := 2
	expected := AzureProvisionConfig{
		CloudAccountID:       testProvisionAccountID,
		IsAzResilient:        true,
		KeepClusterOnFailure: false,
		ClusterConfig: AzureProvisionClusterConfig{
			ProvisionClusterConfig: ProvisionClusterConfig{
				ClusterName:      "my-cluster",
				NumNodes:         3,
				DNSNameServers:   []string{"10.0.0.2"},
				DNSSearchDomains: []string{"example.com"},
				NTPServers:       []string{"pool.ntp.org"},
			},
			AzureEsConfig: AzureProvisionEsConfig{
				StorageAccount:     "mystorage",
				ResourceGroup:      "my-storage-rg",
				ContainerName:      "my-container",
				EnableImmutability: true,
				ManagedIdentity:    AzureProvisionManagedIdentity{Name: "my-identity", ResourceGroup: "my-identity-rg"},
			},
		},
		VMConfig: AzureProvisionVMConfig{
			CDMVersion:                   "3.7.31669",
			CDMProduct:                   "rubrik-cdm-94",
			VMType:                       &vmType,
			Location:                     "centralus",
			InstanceType:                 3,
			ResourceGroup:                "my-rg",
			NetworkResourceGroup:         "my-net-rg",
			VnetResourceGroup:            "my-vnet-rg",
			Vnet:                         "my-vnet",
			NetworkSecurityGroup:         "my-nsg",
			NetworkSecurityResourceGroup: "my-nsg-rg",
			SubnetAzConfigs: []ProvisionSubnetAzConfig{
				{Subnet: "subnet-1", AvailabilityZone: "1"},
				{Subnet: "subnet-2", AvailabilityZone: "2"},
			},
		},
	}
	if !reflect.DeepEqual(*config.Azure, expected) {
		t.Errorf("invalid Azure configuration:\n got: %+v\nwant: %+v", *config.Azure, expected)
	}
}

func TestProvisionConfigAws(t *testing.T) {
	config := requestProvisionConfig(t, "AWS", testAwsProvisionConfig)
	if config.AWS == nil || config.Azure != nil || config.GCP != nil {
		t.Fatalf("expected only the AWS configuration, got %+v", config)
	}

	expected := AwsProvisionConfig{
		CloudAccountID:       testProvisionAccountID,
		Region:               "us-east-2",
		KeepClusterOnFailure: true,
		UsePlacementGroups:   true,
		ClusterConfig: AwsProvisionClusterConfig{
			ProvisionClusterConfig: ProvisionClusterConfig{
				ClusterName:           "my-cluster",
				NumNodes:              1,
				DNSNameServers:        []string{"169.254.169.253"},
				NTPServers:            []string{"169.254.169.123"},
				DynamicScalingEnabled: true,
			},
			AwsEsConfig: AwsProvisionEsConfig{BucketName: "my-bucket", EnableImmutability: true, EnableObjectLock: true},
		},
		VMConfig: AwsProvisionVMConfig{
			CDMVersion:          "9.4.0-p2-30507",
			CDMProduct:          "rubrik-cdm-product-code",
			InstanceType:        10,
			InstanceProfileName: "my-profile",
			VPC:                 "vpc-0123456789abcdef0",
			Subnet:              "subnet-0123456789abcdef0",
			SecurityGroups:      []string{"sg-0123456789abcdef0", "sg-0fedcba9876543210"},
		},
	}
	if !reflect.DeepEqual(*config.AWS, expected) {
		t.Errorf("invalid AWS configuration:\n got: %+v\nwant: %+v", *config.AWS, expected)
	}
}

func TestProvisionConfigGcp(t *testing.T) {
	config := requestProvisionConfig(t, "GCP", testGcpProvisionConfig)
	if config.AWS != nil || config.Azure != nil || config.GCP == nil {
		t.Fatalf("expected only the GCP configuration, got %+v", config)
	}

	expected := GcpProvisionConfig{
		CloudAccountID:       testProvisionAccountID,
		Region:               "us-west1",
		Zone:                 "us-west1-a",
		IsAzResilient:        true,
		KeepClusterOnFailure: true,
		ClusterConfig: GcpProvisionClusterConfig{
			ProvisionClusterConfig: ProvisionClusterConfig{
				ClusterName:      "my-cluster",
				NumNodes:         3,
				DNSNameServers:   []string{"169.254.169.254"},
				DNSSearchDomains: []string{"example.com"},
				NTPServers:       []string{"time.google.com"},
			},
			GcpEsConfig: GcpProvisionEsConfig{BucketName: "my-bucket"},
		},
		VMConfig: GcpProvisionVMConfig{
			CDMVersion:       "9.4.0-p2-30507",
			CDMProduct:       "rubrik-cdm",
			InstanceType:     6,
			DeleteProtection: true,
			NetworkConfig: []GcpProvisionNetworkConfig{
				{Network: "my-vpc", HostProject: "shared-vpc", Region: "us-west1", Name: "subnet-a"},
				{Network: "my-vpc", HostProject: "shared-vpc", Region: "us-west1", Name: "subnet-b"},
				{Network: "my-vpc", HostProject: "shared-vpc", Region: "us-west1", Name: "subnet-c"},
			},
			SubnetAzConfigs: []ProvisionSubnetAzConfig{
				{Subnet: "subnet-a", AvailabilityZone: "us-west1-a"},
				{Subnet: "subnet-b", AvailabilityZone: "us-west1-b"},
				{Subnet: "subnet-c", AvailabilityZone: "us-west1-c"},
			},
			ServiceAccounts: []GcpProvisionServiceAccount{{Email: "node@example.iam.gserviceaccount.com"}},
		},
	}
	if !reflect.DeepEqual(*config.GCP, expected) {
		t.Errorf("invalid GCP configuration:\n got: %+v\nwant: %+v", *config.GCP, expected)
	}
}

// A cloud vendor unknown to the SDK isn't an error, the marshaled
// configuration is returned as is.
func TestProvisionConfigUnknownVendor(t *testing.T) {
	config := requestProvisionConfig(t, "OCI", "{}")
	if config.AWS != nil || config.Azure != nil || config.GCP != nil {
		t.Errorf("expected no vendor configuration, got %+v", config)
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
		name: "EmptyConfig",
		response: func(t *testing.T) []byte {
			return provisionResponse(t, testProvisionClusterID, "CLUSTER_CREATE", "COMPLETED", "AWS", "[]")
		},
	}, {
		name: "InvalidConfig",
		response: func(t *testing.T) []byte {
			return provisionResponse(t, testProvisionClusterID, "CLUSTER_CREATE", "COMPLETED", "AWS", "{")
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
