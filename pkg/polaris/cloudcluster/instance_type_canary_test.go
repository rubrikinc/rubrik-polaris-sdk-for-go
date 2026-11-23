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
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlcloudcluster "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cloudcluster"
)

// enumNames returns the values as a slice of strings. The instance types of the
// cloud vendors are distinct types.
func enumNames[T ~string](values []T) []string {
	names := make([]string, 0, len(values))
	for _, value := range values {
		names = append(names, string(value))
	}
	return names
}

// TestInstanceTypeEnumCanary acts as a canary test to catch changes to the
// instance type enums of the Cloud Cluster in RSC early. A new instance type is
// added to the enum when RSC starts supporting it, which is usually before the
// CDM version it requires is released, so the instance types RSC currently
// offers can be a subset of the enum. Cloud Cluster configurations and the
// recorded configuration of a cluster create job can hold any value of the
// enum, which is why the SDK needs to know all of them.
//
// The test needs no cloud credentials. To run it against an RSC instance, the
// following environment variables need to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
//
// When the test reports a value in RSC which is missing from the SDK:
//  1. Add a constant for it to the instance type constants of the cloud vendor
//     in the graphql/cloudcluster package, and to its AwsInstanceTypes,
//     AzureInstanceTypes or GcpInstanceTypes function.
//  2. For AWS and Azure, add the number of the value in the AwsInstanceType or
//     AzureInstanceType enum of the cluster-mgmt protobuf definition to
//     provisionAwsInstanceTypes or provisionAzureInstanceTypes in the
//     graphql/cloudcluster package, and to the matching test of its numbers.
//     Without it, the recorded configuration of a cluster created with the
//     instance type can't be decoded.
//  3. Add it to the instance types accepted by terraform-provider-rubrik, if
//     RSC offers it for a CDM version.
func TestInstanceTypeEnumCanary(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	tests := []struct {
		name   string
		enum   string
		values []string
	}{
		{name: "AWS", enum: "AwsInstanceType", values: enumNames(gqlcloudcluster.AwsInstanceTypes())},
		{name: "Azure", enum: "AzureInstanceType", values: enumNames(gqlcloudcluster.AzureInstanceTypes())},
		{name: "GCP", enum: "GcpInstanceType", values: enumNames(gqlcloudcluster.GcpInstanceTypes())},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rscValues, err := graphql.EnumValuesAsSet(t.Context(), client.GQL, test.enum)
			if err != nil {
				t.Fatal(err)
			}

			sdkValues := make(map[string]struct{}, len(test.values))
			for _, value := range test.values {
				sdkValues[value] = struct{}{}
			}

			for value := range sdkValues {
				if _, ok := rscValues[value]; !ok {
					t.Errorf("RSC enum %s: %q exist in the SDK but not in RSC", test.enum, value)
				}
			}
			for value := range rscValues {
				if _, ok := sdkValues[value]; !ok {
					t.Errorf("RSC enum %s: %q exist in RSC but not in the SDK, see the documentation of "+
						"TestInstanceTypeEnumCanary for what to update", test.enum, value)
				}
			}
		})
	}
}
