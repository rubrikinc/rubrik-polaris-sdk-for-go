// Code generated by queries_gen.go DO NOT EDIT.

// MIT License
//
// Copyright (c) 2021 Rubrik
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package pcr

// deregisterPrivateContainerRegistry GraphQL query
var deregisterPrivateContainerRegistryQuery = `mutation SdkGolangDeregisterPrivateContainerRegistry($exocomputeAccountId: UUID!) {
    result: deregisterPrivateContainerRegistry(input: {
        exocomputeAccountId: $exocomputeAccountId
    })
}`

// privateContainerRegistry GraphQL query
var privateContainerRegistryQuery = `query SdkGolangPrivateContainerRegistry($exocomputeAccountId: UUID!) {
    result: privateContainerRegistry(input: {
        exocomputeAccountId: $exocomputeAccountId
    }) {
        pcrLatestApprovedBundleVersion
        pcrDetails {
            registryUrl
            imagePullDetails {
                ... on PcrAzureImagePullDetails {
                    customerAppId
                }
                ... on PcrAwsImagePullDetails {
                    awsNativeId
                }
            }
        }
    }
}`

// setPrivateContainerRegistryDetails GraphQL query
var setPrivateContainerRegistryDetailsQuery = `mutation SdkGolangSetPrivateContainerRegistryDetails(
    $cloudType:                CloudType,
    $exocomputeAccountId:      UUID!,
    $registryUrl:              String!,
    $pcrAwsImagePullDetails:   PcrAwsImagePullDetailsInput,
    $pcrAzureImagePullDetails: PcrAzureImagePullDetailsInput
) {
    result: setPrivateContainerRegistry(input: {
        cloudType:                $cloudType,
        exocomputeAccountId:      $exocomputeAccountId,
        registryUrl:              $registryUrl,
        pcrAwsImagePullDetails:   $pcrAwsImagePullDetails,
        pcrAzureImagePullDetails: $pcrAzureImagePullDetails
    })
}`