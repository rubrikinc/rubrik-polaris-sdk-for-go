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
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws/arn"

	"github.com/google/uuid"
)

type identity struct {
	id       string
	internal bool
}

// IdentityFunc returns a project identity initialized from the values passed
// to the function creating the IdentityFunc.
type IdentityFunc func(ctx context.Context) (identity, error)

// AccountID returns an IdentityFunc that initializes the identity with the
// specified AWS account id.
func AccountID(awsAccountID string) IdentityFunc {
	return func(ctx context.Context) (identity, error) {
		if !verifyAccountID(awsAccountID) {
			return identity{}, errors.New("invalid AWS id")
		}

		return identity{id: awsAccountID, internal: false}, nil
	}
}

// CloudAccountID returns an IdentityFunc that initializes the identity with
// the specified RSC cloud account id.
func CloudAccountID(cloudAccountID uuid.UUID) IdentityFunc {
	return func(ctx context.Context) (identity, error) {
		return identity{id: cloudAccountID.String(), internal: true}, nil
	}
}

// ID returns an IdentityFunc that initializes the identity with the id of the
// specified account.
func ID(account AccountFunc) IdentityFunc {
	return func(ctx context.Context) (identity, error) {
		config, err := account(ctx)
		if err != nil {
			return identity{}, fmt.Errorf("failed to lookup account id: %v", err)
		}

		return identity{id: config.id, internal: false}, nil
	}
}

// Role returns an IdentityFunc that initializes the identity with the specified
// AWS account id.
func Role(roleARN string) IdentityFunc {
	return func(ctx context.Context) (identity, error) {
		arn, err := arn.Parse(roleARN)
		if err != nil {
			return identity{}, fmt.Errorf("failed to parse role ARN: %v", err)
		}
		if !verifyAccountID(arn.AccountID) {
			return identity{}, errors.New("invalid AWS id")
		}

		return identity{id: arn.AccountID, internal: false}, nil
	}
}

// verifyAccountID returns true if the AWS account id is valid.
func verifyAccountID(awsAccountID string) bool {
	if len(awsAccountID) != 12 {
		return false
	}
	if _, err := strconv.ParseInt(awsAccountID, 10, 64); err != nil {
		return false
	}

	return true
}
