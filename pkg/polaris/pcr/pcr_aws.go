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
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/pcr"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AWSRegistry returns the AWS private container registry details.
func (a API) AWSRegistry(ctx context.Context, cloudAccountID uuid.UUID) (pcr.AWSRegistry, error) {
	a.log.Print(log.Trace)

	params := pcr.GetAWSRegistryParams{CloudAccountID: cloudAccountID}
	pcrInfo, err := pcr.GetRegistry(ctx, a.client, params)
	if err != nil {
		return pcr.AWSRegistry{}, fmt.Errorf("failed to get private container registry: %s", err)
	}

	return pcrInfo, nil
}

// SetAWSRegistry sets the AWS private container registry.
func (a API) SetAWSRegistry(ctx context.Context, cloudAccountID uuid.UUID, nativeID, registryURL string) error {
	a.log.Print(log.Trace)

	params := pcr.SetAWSRegistryParams{
		CloudAccountID: cloudAccountID,
		NativeID:       nativeID,
		RegistryURL:    registryURL,
	}
	if err := pcr.SetRegistry(ctx, a.client, params); err != nil {
		return fmt.Errorf("failed to set private container registry: %s", err)
	}

	return nil
}
