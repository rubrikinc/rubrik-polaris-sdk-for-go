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

package access

import (
	"context"
	"fmt"
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/access"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// IdentityProviders returns all identity providers for the current
// organization.
func (a API) IdentityProviders(ctx context.Context) ([]access.IdentityProvider, error) {
	a.log.Print(log.Trace)

	providers, err := access.ListIdentityProviders(ctx, a.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity providers: %s", err)
	}

	return providers, nil
}

// IdentityProviderByID returns the identity provider with the specified ID.
func (a API) IdentityProviderByID(ctx context.Context, id string) (access.IdentityProvider, error) {
	a.log.Print(log.Trace)

	providers, err := a.IdentityProviders(ctx)
	if err != nil {
		return access.IdentityProvider{}, err
	}

	lowerID := strings.ToLower(id)
	for _, provider := range providers {
		if strings.ToLower(provider.ID) == lowerID {
			return provider, nil
		}
	}

	return access.IdentityProvider{}, fmt.Errorf("identity provider %q %w", id, graphql.ErrNotFound)
}

// IdentityProviderByName returns the identity provider with the specified name.
// Returns an error if multiple identity providers share the same name.
func (a API) IdentityProviderByName(ctx context.Context, name string) (access.IdentityProvider, error) {
	a.log.Print(log.Trace)

	providers, err := a.IdentityProviders(ctx)
	if err != nil {
		return access.IdentityProvider{}, err
	}

	lowerName := strings.ToLower(name)
	var matches []access.IdentityProvider
	for _, provider := range providers {
		if strings.ToLower(provider.Name) == lowerName {
			matches = append(matches, provider)
		}
	}

	switch len(matches) {
	case 0:
		return access.IdentityProvider{}, fmt.Errorf("identity provider %q %w", name, graphql.ErrNotFound)
	case 1:
		return matches[0], nil
	default:
		return access.IdentityProvider{}, fmt.Errorf("multiple identity providers found with name %q, use IdentityProviderByID instead", name)
	}
}
