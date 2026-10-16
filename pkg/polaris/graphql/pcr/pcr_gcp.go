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

package pcr

import (
	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
)

// GetGCPRegistryParams holds the parameters for a GCP private container
// registry get operation.
type GetGCPRegistryParams struct {
	CloudAccountID uuid.UUID `json:"exocomputeAccountId"`
}

func (p GetGCPRegistryParams) GetQuery() (string, any, GCPRegistry) {
	return privateContainerRegistryQuery, p, GCPRegistry{}
}

// GCPRegistry holds the result of a GCP private container registry get
// operation.
type GCPRegistry struct {
	LatestApprovedBundleVersion string `json:"pcrLatestApprovedBundleVersion"`
	PCRDetails                  struct {
		RegistryURL      string `json:"registryUrl"`
		ImagePullDetails struct {
			ServiceAccountEmail string `json:"serviceAccountEmail"`
		} `json:"imagePullDetails"`
	} `json:"pcrDetails"`
}

// SetGCPRegistryParams holds the parameters for a GCP private container
// registry set operation. The ownership token is a Google OIDC identity token
// proving ownership of the service account. It is verified by RSC and never
// stored.
type SetGCPRegistryParams struct {
	CloudAccountID               uuid.UUID     `json:"exocomputeAccountId"`
	RegistryURL                  string        `json:"registryUrl"`
	ServiceAccountEmail          string        `json:"serviceAccountEmail"`
	ServiceAccountOwnershipToken secret.String `json:"serviceAccountOwnershipToken"`
}

func (p SetGCPRegistryParams) SetQuery() (string, any, SetGCPRegistryResult) {
	params := struct {
		CloudType        string              `json:"cloudType"`
		CloudAccountID   uuid.UUID           `json:"exocomputeAccountId"`
		RegistryURL      string              `json:"registryUrl"`
		ImagePullDetails gcpImagePullDetails `json:"pcrGcpImagePullDetails"`
	}{
		CloudType:      "GCP",
		CloudAccountID: p.CloudAccountID,
		RegistryURL:    p.RegistryURL,
		ImagePullDetails: gcpImagePullDetails{
			ServiceAccountEmail:          p.ServiceAccountEmail,
			ServiceAccountOwnershipToken: p.ServiceAccountOwnershipToken,
		},
	}
	return setPrivateContainerRegistryDetailsQuery, params, SetGCPRegistryResult{}
}

// SetGCPRegistryResult holds the result of a GCP private container registry
// set operation.
type SetGCPRegistryResult struct{}

type gcpImagePullDetails struct {
	ServiceAccountEmail          string        `json:"serviceAccountEmail"`
	ServiceAccountOwnershipToken secret.String `json:"serviceAccountOwnershipToken"`
}
