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
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// TestCreateGcpClusterInputMarshal verifies that CreateGcpClusterInput marshals
// to the exact wire format expected by the RSC createGcpCluster /
// validateCreateGcpClusterInput APIs. This guards against struct tag drift,
// which silently produces inputs the API rejects.
func TestCreateGcpClusterInputMarshal(t *testing.T) {
	input := CreateGcpClusterInput{
		CloudAccountID:       uuid.MustParse("6a6a9c1a-2861-412b-ac00-b48fb74222e2"),
		IsEsType:             true,
		KeepClusterOnFailure: false,
		Region:               "us-east1",
		Zone:                 "us-east1-b",
		Validations:          []ClusterCreateValidations{AllChecks},
		ClusterConfig: GcpClusterConfig{
			ClusterName:    "test-cces",
			UserEmail:      "user@example.com",
			AdminPassword:  "placeholder-secret",
			DNSNameServers: []string{"8.8.8.8"},
			NTPServers:     []string{"pool.ntp.org"},
			NumNodes:       1,
			GcpEsConfig: GcpEsConfigInput{
				BucketName:         "test-bucket",
				Region:             "us-east1",
				ShouldCreateBucket: false,
			},
		},
		VMConfig: GcpVmConfig{
			CDMProduct:       "rubrik-9-4",
			CDMVersion:       "9.4.2",
			DeleteProtection: true,
			InstanceType:     GcpInstanceTypeN2Standard8,
			NetworkConfig: []GcpSubnetInput{{
				HostProject: "host-proj",
				Name:        "default",
				Network:     "default",
				Region:      "us-east1",
			}},
			ServiceAccounts: []GcpServiceAccountInput{{
				Email:  "sa@example.com",
				Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			}},
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal CreateGcpClusterInput: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Required top-level keys must be present with the schema's exact names.
	for _, key := range []string{
		"cloudAccountId", "clusterConfig", "isEsType", "keepClusterOnFailure",
		"region", "validations", "vmConfig", "zone",
	} {
		if _, ok := got[key]; !ok {
			t.Errorf("missing top-level key %q", key)
		}
	}

	// isAzResilient is an optional pointer; it must be omitted when unset.
	if _, ok := got["isAzResilient"]; ok {
		t.Error("isAzResilient should be omitted when nil")
	}

	// clusterConfig: admin password must marshal to its plain value (not
	// redacted) so it reaches the API, and the GCP ES config must be nested
	// under gcpEsConfig with the schema field names.
	clusterConfig, ok := got["clusterConfig"].(map[string]any)
	if !ok {
		t.Fatal("clusterConfig is not an object")
	}
	if clusterConfig["adminPassword"] != "placeholder-secret" {
		t.Errorf("adminPassword = %v, want plain value", clusterConfig["adminPassword"])
	}
	esConfig, ok := clusterConfig["gcpEsConfig"].(map[string]any)
	if !ok {
		t.Fatal("clusterConfig.gcpEsConfig is not an object")
	}
	if esConfig["bucketName"] != "test-bucket" {
		t.Errorf("gcpEsConfig.bucketName = %v, want test-bucket", esConfig["bucketName"])
	}

	// vmConfig: nested network and service account field names must match.
	vmConfig, ok := got["vmConfig"].(map[string]any)
	if !ok {
		t.Fatal("vmConfig is not an object")
	}
	network := vmConfig["networkConfig"].([]any)[0].(map[string]any)
	if network["hostProject"] != "host-proj" {
		t.Errorf("networkConfig[0].hostProject = %v, want host-proj", network["hostProject"])
	}
	serviceAccount := vmConfig["serviceAccounts"].([]any)[0].(map[string]any)
	if serviceAccount["email"] != "sa@example.com" {
		t.Errorf("serviceAccounts[0].email = %v, want sa@example.com", serviceAccount["email"])
	}

	// Optional vmConfig fields must be omitted when empty so the API applies
	// its own defaults.
	for _, key := range []string{"imageId", "labels", "nodeSizeGb", "testImage", "subnetAzConfigs", "vmType"} {
		if _, ok := vmConfig[key]; ok {
			t.Errorf("optional vmConfig key %q should be omitted when empty", key)
		}
	}
}

// TestGcpVmConfigOptionalFieldsMarshal verifies the optional vmConfig fields
// serialize with the correct schema names when populated.
func TestGcpVmConfigOptionalFieldsMarshal(t *testing.T) {
	vmConfig := GcpVmConfig{
		ImageID:    "projects/p/global/images/img",
		Labels:     "key=value",
		NodeSizeGB: 1024,
		VMType:     CCVmConfigDense,
		TestImage: &GcpTestImage{
			ImageName: "img",
			Project:   "p",
		},
	}

	data, err := json.Marshal(vmConfig)
	if err != nil {
		t.Fatalf("failed to marshal GcpVmConfig: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if got["imageId"] != "projects/p/global/images/img" {
		t.Errorf("imageId = %v", got["imageId"])
	}
	if got["labels"] != "key=value" {
		t.Errorf("labels = %v", got["labels"])
	}
	if got["nodeSizeGb"] != float64(1024) {
		t.Errorf("nodeSizeGb = %v, want 1024", got["nodeSizeGb"])
	}
	testImage, ok := got["testImage"].(map[string]any)
	if !ok {
		t.Fatal("testImage is not an object")
	}
	if testImage["imageName"] != "img" || testImage["project"] != "p" {
		t.Errorf("testImage = %v", testImage)
	}
	if got["vmType"] != "DENSE" {
		t.Errorf("vmType = %v, want DENSE", got["vmType"])
	}
}
