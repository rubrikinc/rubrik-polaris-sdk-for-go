// Copyright 2025 Rubrik, Inc.
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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	aws "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NativeAccountByID returns the native account with the specified RSC cloud
// account ID. Note, AWS uses the same ID for both cloud accounts and native
// cloud accounts.
func (a API) NativeAccountByID(ctx context.Context, cloudAccountID uuid.UUID) (aws.NativeAccount, error) {
	a.log.Print(log.Trace)

	natives, err := a.NativeAccounts(ctx, "")
	if err != nil {
		return aws.NativeAccount{}, fmt.Errorf("failed to list native accounts: %s", err)
	}
	for _, native := range natives {
		if native.ID == cloudAccountID {
			return native, nil
		}
	}

	return aws.NativeAccount{}, fmt.Errorf("native account %q %w", cloudAccountID, graphql.ErrNotFound)
}

// NativeAccounts returns all native accounts matching the specified filter.
// The filter can be used to search for a substring in the account name.
func (a API) NativeAccounts(ctx context.Context, filter string) ([]aws.NativeAccount, error) {
	return aws.Wrap(a.client).NativeAccounts(ctx, filter)
}
