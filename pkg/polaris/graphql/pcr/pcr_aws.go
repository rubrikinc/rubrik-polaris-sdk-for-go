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

package pcr

import (
	"github.com/google/uuid"
)

// GetAWSRegistryParams holds the parameters for an AWS private container
// registry get operation.
type GetAWSRegistryParams struct {
	CloudAccountID uuid.UUID `json:"exocomputeAccountId"`
}

func (p GetAWSRegistryParams) GetQuery() (string, any, AWSRegistry) {
	return privateContainerRegistryQuery, p, AWSRegistry{}
}

// AWSRegistry holds the result of an AWS private container registry get
// operation.
type AWSRegistry struct {
	LatestApprovedBundleVersion string `json:"pcrLatestApprovedBundleVersion"`
	PCRDetails                  struct {
		RegistryURL      string `json:"registryUrl"`
		ImagePullDetails struct {
			NativeID string `json:"awsNativeId"`
		} `json:"imagePullDetails"`
	} `json:"pcrDetails"`
}

// SetAWSRegistryParams holds the parameters for an AWS private container
// registry set operation.
type SetAWSRegistryParams struct {
	CloudAccountID uuid.UUID `json:"exocomputeAccountId"`
	RegistryURL    string    `json:"registryUrl"`
	NativeID       string    `json:"awsNativeId"`
}

func (p SetAWSRegistryParams) SetQuery() (string, any, SetAWSRegistryResult) {
	params := struct {
		CloudType        string              `json:"cloudType"`
		CloudAccountID   uuid.UUID           `json:"exocomputeAccountId"`
		RegistryURL      string              `json:"registryUrl"`
		ImagePullDetails awsImagePullDetails `json:"pcrAwsImagePullDetails"`
	}{
		CloudType:      "AWS",
		CloudAccountID: p.CloudAccountID,
		RegistryURL:    p.RegistryURL,
		ImagePullDetails: awsImagePullDetails{
			NativeID: p.NativeID,
		},
	}
	return setPrivateContainerRegistryDetailsQuery, params, SetAWSRegistryResult{}
}

// SetAWSRegistryResult holds the result of an AWS private container registry
// set operation.
type SetAWSRegistryResult struct{}

type awsImagePullDetails struct {
	NativeID string `json:"awsNativeId"`
}
