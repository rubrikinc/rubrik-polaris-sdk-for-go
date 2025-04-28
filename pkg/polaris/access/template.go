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

package access

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/access"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// RoleTemplateByID returns the role template with the specified role template
// ID.
func (a API) RoleTemplateByID(ctx context.Context, roleTemplateID uuid.UUID) (access.RoleTemplate, error) {
	templates, err := a.RoleTemplates(ctx, "")
	if err != nil {
		return access.RoleTemplate{}, err
	}

	for _, template := range templates {
		if template.ID == roleTemplateID {
			return template, nil
		}
	}

	return access.RoleTemplate{}, fmt.Errorf("role template %q %w", roleTemplateID, graphql.ErrNotFound)
}

// RoleTemplateByName returns the role template with a name exactly matching the
// specified name.
func (a API) RoleTemplateByName(ctx context.Context, roleTemplateName string) (access.RoleTemplate, error) {
	templates, err := a.RoleTemplates(ctx, roleTemplateName)
	if err != nil {
		return access.RoleTemplate{}, err
	}

	for _, template := range templates {
		if template.Name == roleTemplateName {
			return template, nil
		}
	}

	return access.RoleTemplate{}, fmt.Errorf("role template %q %w", roleTemplateName, graphql.ErrNotFound)
}

// RoleTemplates returns the role templates matching the specified role template
// name filter. The name filter matches all role templates that has the
// specified name filter as part of their name.
func (a API) RoleTemplates(ctx context.Context, nameFilter string) ([]access.RoleTemplate, error) {
	a.client.Log().Print(log.Trace)

	templates, err := access.ListRoleTemplates(ctx, a.client, nameFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get role templates with filter %q: %s", nameFilter, err)
	}

	return templates, nil
}
