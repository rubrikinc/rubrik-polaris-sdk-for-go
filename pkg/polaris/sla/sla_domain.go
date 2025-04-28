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

package sla

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// DomainByID returns the global SLA domain with the specified ID.
func (a API) DomainByID(ctx context.Context, domainID uuid.UUID) (sla.Domain, error) {
	a.log.Print(log.Trace)

	domains, err := a.Domains(ctx, "")
	if err != nil {
		return sla.Domain{}, err
	}

	for _, domain := range domains {
		if domain.ID == domainID {
			return domain, nil
		}
	}

	return sla.Domain{}, fmt.Errorf("global SLA domain %q %w", domainID, graphql.ErrNotFound)
}

// DomainByName returns the global SLA domain with the specified name.
func (a API) DomainByName(ctx context.Context, name string) (sla.Domain, error) {
	a.log.Print(log.Trace)

	domains, err := a.Domains(ctx, name)
	if err != nil {
		return sla.Domain{}, err
	}

	name = strings.ToLower(name)
	for _, domain := range domains {
		if strings.ToLower(domain.Name) == name {
			return domain, nil
		}
	}

	return sla.Domain{}, fmt.Errorf("global SLA domain %q %w", name, graphql.ErrNotFound)
}

// Domains returns all global SLA domains matching the specified name filter.
func (a API) Domains(ctx context.Context, nameFilter string) ([]sla.Domain, error) {
	a.log.Print(log.Trace)

	var filters []sla.DomainFilter
	if nameFilter != "" {
		filters = append(filters, sla.DomainFilter{
			Field: "NAME",
			Value: nameFilter,
		})
	}
	domains, err := sla.ListDomains(ctx, a.client, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list global SLA domains: %s", err)
	}

	return domains, nil
}

// DomainObjects returns all objects protected by the global SLA domain matching
// the specified name filter. Note that doNotProtect can be used as the domain
// ID to list all objects assigned as Do Not Protect.
func (a API) DomainObjects(ctx context.Context, domainID uuid.UUID, nameFilter string) ([]sla.Object, error) {
	a.log.Print(log.Trace)

	objects, err := sla.ListDomainObjects(ctx, a.client, domainID, sla.ObjectFilter{
		ObjectName:                  nameFilter,
		ProtectionStatus:            sla.StatusUnspecified,
		OnlyDirectlyAssignedObjects: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects for SLA domain %q: %s", domainID, err)
	}

	return objects, nil
}

// CreateDomain creates a new global SLA domain with specified parameters.
func (a API) CreateDomain(ctx context.Context, createParams sla.CreateDomainParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	id, err := sla.CreateDomain(ctx, a.client, createParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create global SLA domain: %s", err)
	}

	return id, nil
}

// DeleteDomain deletes the global SLA domain with the specified ID.
func (a API) DeleteDomain(ctx context.Context, slaID uuid.UUID) error {
	a.log.Print(log.Trace)

	if err := sla.DeleteDomain(ctx, a.client, slaID); err != nil {
		return fmt.Errorf("failed delete global SLA domain: %s", err)
	}

	return nil
}

// AssignDomain assigns the specified global SLA domain according to the
// assignment parameters.
func (a API) AssignDomain(ctx context.Context, assignParams sla.AssignDomainParams) error {
	a.log.Print(log.Trace)

	if err := sla.AssignDomain(ctx, a.client, assignParams); err != nil {
		return fmt.Errorf("failed to assign global SLA domain: %s", err)
	}

	return nil
}
