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
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/pcr"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// GCPRegistry returns the GCP private container registry details.
func (a API) GCPRegistry(ctx context.Context, cloudAccountID uuid.UUID) (pcr.GCPRegistry, error) {
	a.log.Print(log.Trace)

	params := pcr.GetGCPRegistryParams{CloudAccountID: cloudAccountID}
	pcrInfo, err := pcr.GetRegistry(ctx, a.client, params)
	if err != nil {
		return pcr.GCPRegistry{}, fmt.Errorf("failed to get private container registry: %s", err)
	}

	return pcrInfo, nil
}

// SetGCPRegistry sets the GCP private container registry. The ownership token
// is a Google OIDC identity token, issued for the service account and the
// audience https://rubrik.com/exocompute/pcr/gcp, proving ownership of the
// service account. The token must be less than one hour old.
func (a API) SetGCPRegistry(ctx context.Context, cloudAccountID uuid.UUID, serviceAccountEmail string, ownershipToken secret.String, registryURL string) error {
	a.log.Print(log.Trace)

	params := pcr.SetGCPRegistryParams{
		CloudAccountID:               cloudAccountID,
		RegistryURL:                  registryURL,
		ServiceAccountEmail:          serviceAccountEmail,
		ServiceAccountOwnershipToken: ownershipToken,
	}
	if err := pcr.SetRegistry(ctx, a.client, params); err != nil {
		return fmt.Errorf("failed to set private container registry: %s", err)
	}

	return nil
}
