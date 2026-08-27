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

package core

import (
	"context"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// HierarchyObjectType represents the type of a hierarchy object.
//
// Deprecated: Use hierarchy.ObjectType instead.
type HierarchyObjectType = hierarchy.ObjectType

// HierarchyObject represents an RSC hierarchy object with SLA information.
//
// Deprecated: Use sla.HierarchyObject instead.
type HierarchyObject = sla.HierarchyObject

// DoNotProtectSLAID is the special SLA domain ID used to indicate that an
// object should not be protected. This is returned in configuredSlaDomain.ID
// when "Do Not Protect" is directly assigned to an object.
//
// Deprecated: Use sla.DoNotProtectSLAID instead.
const DoNotProtectSLAID = sla.DoNotProtectSLAID

// UnprotectedSLAID is the special SLA domain ID used to indicate that an
// object is unprotected (no SLA assigned). This is returned in
// effectiveSlaDomain.ID when the object inherits no protection.
//
// Deprecated: Use sla.UnprotectedSLAID instead.
const UnprotectedSLAID = sla.UnprotectedSLAID

// HierarchyObjectByID returns the hierarchy object with the specified ID.
// This can be used to query any hierarchy object (VMs, databases, tag rules,
// etc.) and retrieve its SLA assignment information including the configured
// and effective SLA domains.
//
// Deprecated: Use sla.ObjectByID instead.
func (a API) HierarchyObjectByID(ctx context.Context, objectID uuid.UUID) (HierarchyObject, error) {
	a.log.Print(log.Trace)

	return sla.ObjectByID(ctx, a.GQL, objectID)
}
