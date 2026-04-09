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

package dspm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/dspm"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// PolicyByID returns the data security policy with the specified ID.
func (a API) PolicyByID(ctx context.Context, id uuid.UUID) (dspm.Policy, error) {
	a.log.Print(log.Trace)

	policy, err := dspm.PolicyByID(ctx, a.client, id)
	if err != nil {
		return dspm.Policy{}, fmt.Errorf("failed to get data security policy %q: %w", id, err)
	}

	return policy, nil
}

// PolicyByName returns the data security policy with the specified name.
func (a API) PolicyByName(ctx context.Context, name string) (dspm.Policy, error) {
	a.log.Print(log.Trace)

	policies, err := dspm.Policies(ctx, a.client)
	if err != nil {
		return dspm.Policy{}, fmt.Errorf("failed to get data security policies: %v", err)
	}

	var matches []dspm.Policy
	lowerName := strings.ToLower(name)
	for _, policy := range policies {
		if strings.ToLower(policy.Name) == lowerName {
			matches = append(matches, policy)
		}
	}

	switch len(matches) {
	case 0:
		return dspm.Policy{}, fmt.Errorf("data security policy %q: %w", name, graphql.ErrNotFound)
	case 1:
		return matches[0], nil
	default:
		return dspm.Policy{}, fmt.Errorf("multiple data security policies named %q found", name)
	}
}

// Policies returns all data security policies.
func (a API) Policies(ctx context.Context) ([]dspm.Policy, error) {
	a.log.Print(log.Trace)

	policies, err := dspm.Policies(ctx, a.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get data security policies: %v", err)
	}

	return policies, nil
}

// CreatePolicy creates a new data security policy. Returns the ID of the
// created policy.
func (a API) CreatePolicy(ctx context.Context, input dspm.CreateInput) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	id, err := dspm.CreatePolicy(ctx, a.client, input)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to create data security policy %q: %v", input.Name, err)
	}

	return id, nil
}

// UpdatePolicy updates a data security policy.
func (a API) UpdatePolicy(ctx context.Context, input dspm.UpdateInput) error {
	a.log.Print(log.Trace)

	if err := dspm.UpdatePolicy(ctx, a.client, input); err != nil {
		return fmt.Errorf("failed to update data security policy %q: %v", input.ID, err)
	}

	return nil
}

// DeletePolicy deletes the data security policy with the specified ID.
func (a API) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	a.log.Print(log.Trace)

	if err := dspm.DeletePolicy(ctx, a.client, id); err != nil {
		return fmt.Errorf("failed to delete data security policy %q: %v", id, err)
	}

	return nil
}

// FilterTypes returns the available filter types for the specified resource
// type.
func (a API) FilterTypes(ctx context.Context, resourceType dspm.ResourceType) ([]dspm.FilterType, error) {
	a.log.Print(log.Trace)

	types, err := dspm.FilterTypes(ctx, a.client, resourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get filter types for %q: %v", resourceType, err)
	}

	return types, nil
}

// FilterValues returns the possible relationships and values for the specified
// filter type.
func (a API) FilterValues(ctx context.Context, filterType dspm.FilterType, searchTerm string) (dspm.FilterMetadata, error) {
	a.log.Print(log.Trace)

	meta, err := dspm.FilterValues(ctx, a.client, filterType, searchTerm)
	if err != nil {
		return dspm.FilterMetadata{}, fmt.Errorf("failed to get filter values for %q: %v", filterType, err)
	}

	return meta, nil
}
