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

import "github.com/google/uuid"

// GetAzureRegistryParams holds the parameters for an Azure private container
// registry get operation.
type GetAzureRegistryParams struct {
	CloudAccountID uuid.UUID `json:"exocomputeAccountId"`
}

func (p GetAzureRegistryParams) GetQuery() (string, any, AzureRegistry) {
	return privateContainerRegistryQuery, p, AzureRegistry{}
}

// AzureRegistry holds the result of an Azure private container registry get
// operation.
type AzureRegistry struct {
	LatestApprovedBundleVersion string `json:"pcrLatestApprovedBundleVersion"`
	PCRDetails                  struct {
		RegistryURL      string `json:"registryUrl"`
		ImagePullDetails struct {
			CustomerAppId string `json:"customerAppId"`
		} `json:"imagePullDetails"`
	} `json:"pcrDetails"`
}

// SetAzureRegistryParams holds the parameters for an Azure private container
// registry set operation.
type SetAzureRegistryParams struct {
	CloudAccountID uuid.UUID `json:"exocomputeAccountId"`
	RegistryURL    string    `json:"registryUrl"`
	CustomerAppId  uuid.UUID `json:"customerAppId"`
}

func (p SetAzureRegistryParams) SetQuery() (string, any, SetAzureRegistryResult) {
	params := struct {
		CloudType        string                `json:"cloudType"`
		CloudAccountID   uuid.UUID             `json:"exocomputeAccountId"`
		RegistryURL      string                `json:"registryUrl"`
		ImagePullDetails azureImagePullDetails `json:"pcrAzureImagePullDetails"`
	}{
		CloudType:      "AZURE",
		CloudAccountID: p.CloudAccountID,
		RegistryURL:    p.RegistryURL,
		ImagePullDetails: azureImagePullDetails{
			CustomerAppId: p.CustomerAppId,
		},
	}
	return setPrivateContainerRegistryDetailsQuery, params, SetAzureRegistryResult{}
}

// SetAzureRegistryResult holds the result of an Azure private container
// registry set operation.
type SetAzureRegistryResult struct{}

type azureImagePullDetails struct {
	CustomerAppId uuid.UUID `json:"customerAppId"`
}
