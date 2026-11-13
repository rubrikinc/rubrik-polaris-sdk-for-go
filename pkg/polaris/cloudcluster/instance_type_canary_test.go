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
//     in the graphql/cloudcluster package, and to the list in this test.
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
	}{{
		name: "AWS",
		enum: "AwsInstanceType",
		values: enumNames([]gqlcloudcluster.AwsCCInstanceType{
			gqlcloudcluster.AwsInstanceTypeUnspecified,
			gqlcloudcluster.AwsInstanceTypeM5_4XLarge,
			gqlcloudcluster.AwsInstanceTypeM6I_2XLarge,
			gqlcloudcluster.AwsInstanceTypeM6I_4XLarge,
			gqlcloudcluster.AwsInstanceTypeM6I_8XLarge,
			gqlcloudcluster.AwsInstanceTypeR6I_4XLarge,
			gqlcloudcluster.AwsInstanceTypeM6A_2XLarge,
			gqlcloudcluster.AwsInstanceTypeM6A_4XLarge,
			gqlcloudcluster.AwsInstanceTypeM6A_8XLarge,
			gqlcloudcluster.AwsInstanceTypeR6A_4XLarge,
			gqlcloudcluster.AwsInstanceTypeM7I_2XLarge,
			gqlcloudcluster.AwsInstanceTypeM7I_4XLarge,
			gqlcloudcluster.AwsInstanceTypeM7I_8XLarge,
			gqlcloudcluster.AwsInstanceTypeR7I_4XLarge,
			gqlcloudcluster.AwsInstanceTypeM7A_2XLarge,
			gqlcloudcluster.AwsInstanceTypeM7A_4XLarge,
			gqlcloudcluster.AwsInstanceTypeM7A_8XLarge,
			gqlcloudcluster.AwsInstanceTypeR7A_4XLarge,
		}),
	}, {
		name: "Azure",
		enum: "AzureInstanceType",
		values: enumNames([]gqlcloudcluster.AzureCCESSupportedInstanceType{
			gqlcloudcluster.AzureInstanceTypeUnspecified,
			gqlcloudcluster.AzureInstanceTypeStandardDS5V2,
			gqlcloudcluster.AzureInstanceTypeStandardD16SV5,
			gqlcloudcluster.AzureInstanceTypeStandardD8SV5,
			gqlcloudcluster.AzureInstanceTypeStandardD32SV5,
			gqlcloudcluster.AzureInstanceTypeStandardE16SV5,
			gqlcloudcluster.AzureInstanceTypeStandardD8ASV5,
			gqlcloudcluster.AzureInstanceTypeStandardD16ASV5,
			gqlcloudcluster.AzureInstanceTypeStandardD32ASV5,
			gqlcloudcluster.AzureInstanceTypeStandardE16ASV5,
			gqlcloudcluster.AzureInstanceTypeStandardD8SV6,
			gqlcloudcluster.AzureInstanceTypeStandardD16SV6,
			gqlcloudcluster.AzureInstanceTypeStandardE16SV6,
			gqlcloudcluster.AzureInstanceTypeStandardD8ASV6,
			gqlcloudcluster.AzureInstanceTypeStandardD16ASV6,
			gqlcloudcluster.AzureInstanceTypeStandardE16ASV6,
		}),
	}, {
		name: "GCP",
		enum: "GcpInstanceType",
		values: enumNames([]gqlcloudcluster.GcpCCInstanceType{
			gqlcloudcluster.GcpInstanceTypeUnspecified,
			gqlcloudcluster.GcpInstanceTypeN2Standard8,
			gqlcloudcluster.GcpInstanceTypeN2Standard16,
			gqlcloudcluster.GcpInstanceTypeN2Highmem16,
			gqlcloudcluster.GcpInstanceTypeN2DStandard8,
			gqlcloudcluster.GcpInstanceTypeN2DStandard16,
			gqlcloudcluster.GcpInstanceTypeN2DHighmem16,
			gqlcloudcluster.GcpInstanceTypeN4Standard8,
			gqlcloudcluster.GcpInstanceTypeN4Standard16,
			gqlcloudcluster.GcpInstanceTypeN4Highmem16,
			gqlcloudcluster.GcpInstanceTypeN4DStandard8,
			gqlcloudcluster.GcpInstanceTypeN4DStandard16,
			gqlcloudcluster.GcpInstanceTypeN4DHighmem16,
		}),
	}}
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
