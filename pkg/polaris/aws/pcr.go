// Copyright 2021 Rubrik, Inc.
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

package aws

import (
	"context"
	"fmt"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// SetPrivateContainerRegistry sets the private container registry details for
// the given AWS account.
func (a API) SetPrivateContainerRegistry(ctx context.Context, id IdentityFunc, url string, nativeID string) error {
	a.log.Print(log.Trace)

	cloudAccountID, err := a.toCloudAccountID(ctx, id)
	if err != nil {
		return err
	}

	if err := aws.Wrap(a.client).SetPrivateContainerRegistryDetails(ctx, cloudAccountID, url, nativeID); err != nil {
		return fmt.Errorf("failed to set private container registrys: %s", err)
	}

	return nil
}

// PrivateContainerRegistry returns the private container registry details for
// the given AWS account.
func (a API) PrivateContainerRegistry(ctx context.Context, id IdentityFunc) (nativeID, url string, err error) {
	a.log.Print(log.Trace)

	cloudAccountID, err := a.toCloudAccountID(ctx, id)
	if err != nil {
		return "", "", err
	}

	nativeID, url, err = aws.Wrap(a.client).PrivateContainerRegistry(ctx, cloudAccountID)
	if err != nil {
		return "", "", fmt.Errorf("failed to read private container registrys: %s", err)
	}

	return nativeID, url, err
}
