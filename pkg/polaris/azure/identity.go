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

package azure

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type identity struct {
	id       string
	internal bool
}

// IdentityFunc returns a subscription identity initialized from the values
// passed to the function creating the IdentityFunc.
type IdentityFunc func(ctx context.Context) (identity, error)

// CloudAccountID returns an IdentityFunc that initializes the identity with
// the specified Polaris cloud account id.
func CloudAccountID(id uuid.UUID) IdentityFunc {
	return func(ctx context.Context) (identity, error) {
		return identity{id.String(), true}, nil
	}
}

// ID returns an IdentityFunc that initializes the identity with the id of the
// specified subscription.
func ID(subscription SubscriptionFunc) IdentityFunc {
	return func(ctx context.Context) (identity, error) {
		config, err := subscription(ctx)
		if err != nil {
			return identity{}, fmt.Errorf("failed to lookup subscription: %v", err)
		}

		return identity{id: config.id.String(), internal: false}, nil
	}
}

// SubscriptionID returns an IdentityFunc that initializes the identity with
// the specified subscription id.
func SubscriptionID(id uuid.UUID) IdentityFunc {
	return func(ctx context.Context) (identity, error) {
		return identity{id.String(), false}, nil
	}
}
