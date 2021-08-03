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
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Azure permissions.
type Permissions struct {
	Actions        []string
	DataActions    []string
	NotActions     []string
	NotDataActions []string
}

// stringsDiff returns the difference between lhs and rhs, i.e. rhs subtracted
// from lhs.
func stringsDiff(lhs, rhs []string) []string {
	set := make(map[string]struct{})
	for _, s := range lhs {
		set[s] = struct{}{}
	}

	for _, s := range rhs {
		delete(set, s)
	}

	diff := make([]string, 0, len(set))
	for s := range set {
		diff = append(diff, s)
	}

	return diff
}

// stringsIn returns true if slice contains str.
func stringsIn(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}

	return false
}

// azurePermissions returns the permissions assigned to the specified service
// principal.
func azurePermissions(ctx context.Context, authorizer autorest.Authorizer, objectID, subscriptionID uuid.UUID) (Permissions, error) {
	roleClient := authorization.NewRoleAssignmentsClient(subscriptionID.String())
	roleClient.Authorizer = authorizer

	roleResult, err := roleClient.ListComplete(ctx, "")
	if err != nil {
		return Permissions{}, err
	}

	var roles []string
	for roleResult.NotDone() {
		role := roleResult.Value()
		if role.PrincipalID == nil {
			return Permissions{}, errors.New("polaris: principal id of assigned role is nil")
		}
		if role.RoleDefinitionID == nil {
			return Permissions{}, errors.New("polaris: role definition id of assigned role is nil")
		}
		if objectID.String() == *role.PrincipalID {
			roles = append(roles, *role.RoleDefinitionID)
		}

		if err := roleResult.NextWithContext(ctx); err != nil {
			return Permissions{}, err
		}
	}

	roleDefClient := authorization.NewRoleDefinitionsClient(subscriptionID.String())
	roleDefClient.Authorizer = authorizer

	scope := fmt.Sprintf("subscriptions/%s", subscriptionID)
	roleDefResult, err := roleDefClient.ListComplete(ctx, scope, "")
	if err != nil {
		return Permissions{}, err
	}

	var perms Permissions
	for roleDefResult.NotDone() {
		roleDef := roleDefResult.Value()
		if roleDef.ID == nil {
			return Permissions{}, errors.New("polaris: id of role definition is nil")
		}
		if roleDef.Permissions == nil {
			return Permissions{}, errors.New("polaris: permissions of role definitions is nil")
		}
		if stringsIn(roles, *roleDef.ID) {
			for _, perm := range *roleDef.Permissions {
				perms.Actions = append(perms.Actions, stringsDiff(*perm.Actions, perms.Actions)...)
				perms.NotActions = append(perms.Actions, stringsDiff(*perm.Actions, perms.Actions)...)
				perms.DataActions = append(perms.Actions, stringsDiff(*perm.Actions, perms.Actions)...)
				perms.NotDataActions = append(perms.Actions, stringsDiff(*perm.Actions, perms.Actions)...)
			}
		}

		if err := roleDefResult.NextWithContext(ctx); err != nil {
			return Permissions{}, err
		}
	}

	return perms, nil
}

// checkPermissions checks that the specified service principal have the
// required Azure permissions to use the subscription with the given Polaris
// features
func (a API) azureCheckPermissions(ctx context.Context, principal servicePrincipal, subscriptionID uuid.UUID, features []core.Feature) error {
	a.gql.Log().Print(log.Trace, "polaris/azure.azureCheckPermissions")

	perms, err := a.Permissions(ctx, features)
	if err != nil {
		return err
	}

	azurePerms, err := azurePermissions(ctx, principal.rmAuthorizer, principal.objectID, subscriptionID)
	if err != nil {
		return err
	}

	if missing := stringsDiff(perms.Actions, azurePerms.Actions); len(missing) > 0 {
		return fmt.Errorf("polaris: missing action permissions: %v", strings.Join(missing, ","))
	}

	if missing := stringsDiff(perms.DataActions, azurePerms.DataActions); len(missing) > 0 {
		return fmt.Errorf("polaris: missing data action permissions: %v", strings.Join(missing, ","))
	}

	if missing := stringsDiff(perms.NotActions, azurePerms.NotActions); len(missing) > 0 {
		return fmt.Errorf("polaris: missing not action permissions: %v", strings.Join(missing, ","))
	}

	if missing := stringsDiff(perms.NotDataActions, azurePerms.NotDataActions); len(missing) > 0 {
		return fmt.Errorf("polaris: missing not data action permissions: %v", strings.Join(missing, ","))
	}

	return nil
}

// Permissions returns all Azure permissions required to use the specified
// Polaris features.
func (a API) Permissions(ctx context.Context, features []core.Feature) (Permissions, error) {
	a.gql.Log().Print(log.Trace, "polaris/azure.Permissions")

	perms := Permissions{}
	for _, feature := range features {
		permConfig, err := azure.Wrap(a.gql).CloudAccountPermissionConfig(ctx, feature)
		if err != nil {
			return Permissions{}, nil
		}

		for _, perm := range permConfig.RolePermissions {
			perms.Actions = append(perms.Actions, stringsDiff(perm.IncludedActions, perms.Actions)...)
			perms.DataActions = append(perms.DataActions, stringsDiff(perm.IncludedDataActions, perms.DataActions)...)
			perms.NotActions = append(perms.NotActions, stringsDiff(perm.ExcludedActions, perms.NotActions)...)
			perms.NotDataActions = append(perms.NotDataActions, stringsDiff(perm.ExcludedDataActions, perms.NotDataActions)...)
		}
	}

	return perms, nil
}

// PermissionsUpdated should be called after the Azure permissions have been
// updated as a response to an account having the status
// StatusMissingPermissions. This will notify Polaris that the permissions have
// been updated.
func (a API) PermissionsUpdated(ctx context.Context) error {
	a.gql.Log().Print(log.Trace, "polaris/azure.PermissionsUpdated")

	// if prinicipal == nil {
	// 	return errors.New("polaris: prinicipal is not allowed to be nil")
	// }
	// config, err := prinicipal(ctx)
	// if err != nil {
	// 	return err
	// }

	// // nativeID, err := a.toNativeID(ctx, id)
	// // if err != nil {
	// // 	return err
	// // }

	// if len(features) > 0 {
	// 	err = a.azureCheckPermissions(ctx, config, config., features)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// TODO: Invoke Polaris endpoint.

	return nil
}
