//go:generate go run ../queries_gen.go hierarchy

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

// Package hierarchy provides a low-level interface to hierarchy GraphQL queries
// provided by the Polaris platform.
package hierarchy

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API wraps around GraphQL clients to give them the Polaris Hierarchy API.
type API struct {
	GQL *graphql.Client
	log log.Logger
}

// Wrap the GraphQL client in the Hierarchy API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

// ObjectType represents the type of a hierarchy object.
type ObjectType string

// Object represents an RSC hierarchy object with SLA information.
type Object struct {
	ID                  uuid.UUID      `json:"id"`
	Name                string         `json:"name"`
	ObjectType          ObjectType     `json:"objectType"`
	SLAAssignment       sla.Assignment `json:"slaAssignment"`
	ConfiguredSLADomain sla.DomainRef  `json:"configuredSlaDomain"`
	EffectiveSLADomain  sla.DomainRef  `json:"effectiveSlaDomain"`
}

// DoNotProtectSLAID is the special SLA domain ID used to indicate that an
// object should not be protected. This is returned in configuredSlaDomain.ID
// when "Do Not Protect" is directly assigned to an object.
const DoNotProtectSLAID = "DO_NOT_PROTECT"

// UnprotectedSLAID is the special SLA domain ID used to indicate that an
// object is unprotected (no SLA assigned). This is returned in
// effectiveSlaDomain.ID when the object inherits no protection.
const UnprotectedSLAID = "UNPROTECTED"

// ObjectByID returns the hierarchy object with the specified ID.
// This can be used to query any hierarchy object (VMs, databases, tag rules,
// etc.) and retrieve its SLA assignment information including the configured
// and effective SLA domains.
func (a API) ObjectByID(ctx context.Context, fid uuid.UUID) (Object, error) {
	a.log.Print(log.Trace)

	query := hierarchyObjectQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		FID uuid.UUID `json:"fid"`
	}{FID: fid})
	if err != nil {
		return Object{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result Object `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return Object{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}
