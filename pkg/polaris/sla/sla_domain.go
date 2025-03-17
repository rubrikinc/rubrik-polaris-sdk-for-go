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

// GlobalSLADomainByID returns the global SLA domain with the specified ID.
func (a API) GlobalSLADomainByID(ctx context.Context, slaID uuid.UUID) (sla.GlobalSLADomain, error) {
	a.log.Print(log.Trace)

	slaDomains, err := a.GlobalSLADomains(ctx, "")
	if err != nil {
		return sla.GlobalSLADomain{}, err
	}

	for _, slaDomain := range slaDomains {
		if slaDomain.ID == slaID {
			return slaDomain, nil
		}
	}

	return sla.GlobalSLADomain{}, fmt.Errorf("global SLA domain %q %w", slaID, graphql.ErrNotFound)
}

// GlobalSLADomainByName returns the global SLA domain with the specified name.
func (a API) GlobalSLADomainByName(ctx context.Context, name string) (sla.GlobalSLADomain, error) {
	a.log.Print(log.Trace)

	slaDomains, err := a.GlobalSLADomains(ctx, name)
	if err != nil {
		return sla.GlobalSLADomain{}, err
	}

	name = strings.ToLower(name)
	for _, slaDomain := range slaDomains {
		if strings.ToLower(slaDomain.Name) == name {
			return slaDomain, nil
		}
	}

	return sla.GlobalSLADomain{}, fmt.Errorf("global SLA domain %q %w", name, graphql.ErrNotFound)
}

// GlobalSLADomains returns all global SLA domains matching the specified name
// filter.
func (a API) GlobalSLADomains(ctx context.Context, nameFilter string) ([]sla.GlobalSLADomain, error) {
	a.log.Print(log.Trace)

	var filters []sla.SLADomainFilter
	if nameFilter != "" {
		filters = append(filters, sla.SLADomainFilter{
			Field: "NAME",
			Value: nameFilter,
		})
	}
	slaDomains, err := sla.ListSLADomains(ctx, a.client, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list global SLA domains: %s", err)
	}

	return slaDomains, nil
}

// GlobalSLADomainProtectedObjects returns all objects protected by the global
// SLA domain matching the specified name filter.
func (a API) GlobalSLADomainProtectedObjects(ctx context.Context, slaID uuid.UUID, nameFilter string) ([]sla.ProtectedObject, error) {
	a.log.Print(log.Trace)

	objects, err := sla.ListSLADomainProtectedObjects(ctx, a.client, slaID, sla.ProtectedObjectFilter{
		ObjectName:                      nameFilter,
		ProtectionStatus:                sla.ProtectionStatusUnspecified,
		ShowOnlyDirectlyAssignedObjects: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list protected objects for SLA domain %q: %s", slaID, err)
	}

	return objects, nil
}

// CreateGlobalSLADomain creates a new global SLA domain with specified
// parameters.
func (a API) CreateGlobalSLADomain(ctx context.Context, createParams sla.CreateGlobalSLAParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	id, err := sla.CreateGlobalSLADomain(ctx, a.client, createParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create global SLA domain: %s", err)
	}

	return id, nil
}

// DeleteGlobalSLADomain deletes the global SLA domain with the specified ID.
func (a API) DeleteGlobalSLADomain(ctx context.Context, slaID uuid.UUID) error {
	a.log.Print(log.Trace)

	if err := sla.DeleteGlobalSLADomain(ctx, a.client, slaID); err != nil {
		return fmt.Errorf("failed delete global SLA domain: %s", err)
	}

	return nil
}

// AssignSLADomain assigns the specified global SLA domain according to the
// assignment parameters.
func (a API) AssignSLADomain(ctx context.Context, assignParams sla.AssignSLAParams) error {
	a.log.Print(log.Trace)

	if err := sla.AssignSLADomain(ctx, a.client, assignParams); err != nil {
		return fmt.Errorf("failed to assign global SLA domain: %s", err)
	}

	return nil
}
