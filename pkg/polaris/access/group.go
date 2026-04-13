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
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/access"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// SSOGroupByID returns the SSO group with the specified SSO group ID.
func (a API) SSOGroupByID(ctx context.Context, id string) (access.SSOGroup, error) {
	a.client.Log().Print(log.Trace)

	groups, err := a.SSOGroups(ctx, "")
	if err != nil {
		return access.SSOGroup{}, err
	}

	lowerID := strings.ToLower(id)
	for _, group := range groups {
		if strings.ToLower(group.ID) == lowerID {
			return group, nil
		}
	}

	return access.SSOGroup{}, fmt.Errorf("SSO group %q %w", id, graphql.ErrNotFound)
}

// SSOGroupByName returns the SSO group with the specified SSO group name.
// Returns an error if multiple groups with the same name exist across different
// auth domains. Use SSOGroupByNameAndAuthDomain to disambiguate.
func (a API) SSOGroupByName(ctx context.Context, name string) (access.SSOGroup, error) {
	a.client.Log().Print(log.Trace)

	groups, err := a.SSOGroups(ctx, name)
	if err != nil {
		return access.SSOGroup{}, err
	}

	lowerName := strings.ToLower(name)
	var matches []access.SSOGroup
	for _, group := range groups {
		if strings.ToLower(group.Name) == lowerName {
			matches = append(matches, group)
		}
	}

	switch len(matches) {
	case 0:
		return access.SSOGroup{}, fmt.Errorf("SSO group %q %w", name, graphql.ErrNotFound)
	case 1:
		return matches[0], nil
	default:
		return access.SSOGroup{}, fmt.Errorf("multiple SSO groups named %q found, use SSOGroupByNameAndAuthDomain to disambiguate", name)
	}
}

// SSOGroupByNameAndAuthDomain returns the SSO group with the specified name in
// the specified auth domain. Returns graphql.ErrNotFound if no matching group
// is found. Returns an error if multiple groups match.
func (a API) SSOGroupByNameAndAuthDomain(ctx context.Context, name, authDomainID string) (access.SSOGroup, error) {
	a.client.Log().Print(log.Trace)

	idp, err := a.IdentityProviderByID(ctx, authDomainID)
	if err != nil {
		return access.SSOGroup{}, err
	}
	groups, err := access.ListSSOGroups(ctx, a.client, access.SSOGroupFilter{
		Name:          name,
		AuthDomainIDs: []string{authDomainID},
	})
	if err != nil {
		return access.SSOGroup{}, fmt.Errorf("failed to get SSO groups with name %q and auth domain %q: %s", name, authDomainID, err)
	}

	lowerName := strings.ToLower(name)
	lowerDomainName := strings.ToLower(idp.Name)
	var matches []access.SSOGroup
	for _, group := range groups {
		if strings.ToLower(group.Name) == lowerName && strings.ToLower(group.DomainName) == lowerDomainName {
			matches = append(matches, group)
		}
	}

	switch len(matches) {
	case 0:
		return access.SSOGroup{}, fmt.Errorf("SSO group %q in auth domain %q %w", name, authDomainID, graphql.ErrNotFound)
	case 1:
		return matches[0], nil
	default:
		return access.SSOGroup{}, fmt.Errorf("multiple SSO groups named %q found in auth domain %q", name, authDomainID)
	}
}

// SSOGroups returns the SSO groups matching the specified SSO group name
// filter.
func (a API) SSOGroups(ctx context.Context, nameFilter string) ([]access.SSOGroup, error) {
	a.client.Log().Print(log.Trace)

	groups, err := access.ListSSOGroups(ctx, a.client, access.SSOGroupFilter{
		Name: nameFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get SSO groups with filter %q: %s", nameFilter, err)
	}

	return groups, nil
}

// AssignSSOGroupRole assigns the role to the SSO group with the specified SSO
// group ID.
func (a API) AssignSSOGroupRole(ctx context.Context, ssoGroupID string, roleID uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	if err := access.AssignRoles(ctx, a.client, access.AssignRoleParams{
		RoleIDs:  []uuid.UUID{roleID},
		GroupIDs: []string{ssoGroupID},
	}); err != nil {
		return fmt.Errorf("failed to assign role %q to SSO group %q: %s", roleID, ssoGroupID, err)
	}

	return nil
}

// AssignSSOGroupRoles assigns the roles to the SSO group with the specified SSO
// group ID.
func (a API) AssignSSOGroupRoles(ctx context.Context, ssoGroupID string, roleIDs []uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	if err := access.AssignRoles(ctx, a.client, access.AssignRoleParams{
		RoleIDs:  roleIDs,
		GroupIDs: []string{ssoGroupID},
	}); err != nil {
		return fmt.Errorf("failed to assign roles %s to SSO group %q: %s", joinUUIDs(roleIDs), ssoGroupID, err)
	}

	return nil
}

// UnassignSSOGroupRole unassigns the role from the SSO group with the specified
// SSO group ID. Returns graphql.ErrNotFound if the user does not exist.
func (a API) UnassignSSOGroupRole(ctx context.Context, ssoGroupID string, roleID uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	group, err := a.SSOGroupByID(ctx, ssoGroupID)
	if err != nil {
		return err
	}

	roleIDs := make([]uuid.UUID, 0, len(group.Roles))
	for _, role := range group.Roles {
		if role.ID != roleID {
			roleIDs = append(roleIDs, role.ID)
		}
	}
	if err := access.ReplaceRoles(ctx, a.client, access.ReplaceRoleParams{
		RoleIDs:  roleIDs,
		GroupIDs: []string{ssoGroupID},
	}); err != nil {
		return fmt.Errorf("failed to unassign role %q from SSO group %q: %s", roleID, ssoGroupID, err)
	}

	return nil
}

// UnassignSSOGroupRoles unassigns the roles from the SSO group with the
// specified SSO group ID. Returns graphql.ErrNotFound if the user does not
// exist.
func (a API) UnassignSSOGroupRoles(ctx context.Context, ssoGroupID string, roleIDs []uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	group, err := a.SSOGroupByID(ctx, ssoGroupID)
	if err != nil {
		return err
	}

	keepRoleIDs := make([]uuid.UUID, 0, len(group.Roles))
	for _, role := range group.Roles {
		if !slices.Contains(roleIDs, role.ID) {
			keepRoleIDs = append(keepRoleIDs, role.ID)
		}
	}
	if err := access.ReplaceRoles(ctx, a.client, access.ReplaceRoleParams{
		RoleIDs:  keepRoleIDs,
		GroupIDs: []string{ssoGroupID},
	}); err != nil {
		return fmt.Errorf("failed to unassign role %s from SSO group %q: %s", joinUUIDs(roleIDs), ssoGroupID, err)
	}

	return nil
}

// CreateSSOGroup creates a new SSO group with the specified name, roles, and
// auth domain.
func (a API) CreateSSOGroup(ctx context.Context, groupName string, roleIDs []uuid.UUID, authDomainID string) error {
	a.client.Log().Print(log.Trace)

	if err := access.InviteSSOGroup(ctx, a.client, access.InviteSSOGroupParams{
		GroupName:    groupName,
		RoleIDs:      roleIDs,
		AuthDomainID: authDomainID,
	}); err != nil {
		return fmt.Errorf("failed to create SSO group %q: %s", groupName, err)
	}

	return nil
}

// DeleteSSOGroup deletes the SSO group with the specified ID.
func (a API) DeleteSSOGroup(ctx context.Context, groupID string) error {
	a.client.Log().Print(log.Trace)

	if err := access.DeleteSSOGroups(ctx, a.client, []string{groupID}); err != nil {
		return fmt.Errorf("failed to delete SSO group %q: %s", groupID, err)
	}

	return nil
}

// ReplaceSSOGroupRoles replaces all the roles for the SSO group with the
// specified SSO group ID.
func (a API) ReplaceSSOGroupRoles(ctx context.Context, ssoGroupID string, newRoleIDs []uuid.UUID) error {
	a.client.Log().Print(log.Trace)

	if err := access.ReplaceRoles(ctx, a.client, access.ReplaceRoleParams{
		RoleIDs:  newRoleIDs,
		GroupIDs: []string{ssoGroupID},
	}); err != nil {
		return fmt.Errorf("failed to replace roles for SSO group %q: %s", ssoGroupID, err)
	}

	return nil
}
