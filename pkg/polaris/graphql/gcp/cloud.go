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

package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CloudAccount represents a Polaris Cloud Account for GCP. If UseGlobalConfig
// is true the cloud account depends on the default service account.
type CloudAccount struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	ProjectID        string    `json:"projectID"`
	ProjectNumber    int64     `json:"projectNumber"`
	RoleID           string    `json:"roleId"`
	UsesGlobalConfig bool      `json:"usesGlobalConfig"`
}

// Feature represents a Polaris Cloud Account feature for GCP, e.g. Cloud Native
// Protection.
type Feature struct {
	Name   core.Feature `json:"feature"`
	Status core.Status  `json:"status"`
}

// CloudAccountWithFeature hold details about a cloud account and the features
// associated with that account.
type CloudAccountWithFeature struct {
	Account CloudAccount `json:"project"`
	Feature Feature      `json:"featureDetail"`
}

// CloudAccountProjectsByFeature returns the cloud accounts matching the
// specified filter. The filter can be used to search for project id, project
// name and project number.
func (a API) CloudAccountProjectsByFeature(ctx context.Context, feature core.Feature, filter string) ([]CloudAccountWithFeature, error) {
	a.GQL.Log().Print(log.Trace)

	query := allGcpCloudAccountProjectsByFeatureQuery
	if graphql.VersionOlderThan(a.Version, "master-46133", "v20220315") {
		query = gcpCloudAccountListProjectsV0Query
	} else if graphql.VersionOlderThan(a.Version, "master-46713", "v20220412") {
		query = gcpCloudAccountListProjectsQuery
	}
	buf, err := a.GQL.Request(ctx, query, struct {
		Feature core.Feature `json:"feature"`
		Filter  string       `json:"projectSearchText"`
	}{Feature: feature, Filter: filter})
	if err != nil {
		return nil, fmt.Errorf("failed to request CloudAccountListProjects: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "%s(%q, %q): %s", graphql.QueryName(query), feature, filter, string(buf))

	var payload struct {
		Data struct {
			Accounts []CloudAccountWithFeature `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CloudAccountListProjects: %v", err)
	}

	return payload.Data.Accounts, nil
}

// CloudAccountAddManualAuthProject adds the GCP project to Polaris.
func (a API) CloudAccountAddManualAuthProject(ctx context.Context, projectID, projectName string, projectNumber int64, orgName, jwtConfig string, feature core.Feature) error {
	a.GQL.Log().Print(log.Trace)

	query := gcpCloudAccountAddManualAuthProjectQuery
	if graphql.VersionOlderThan(a.Version, "master-46133", "v20220315") {
		query = gcpCloudAccountAddManualAuthProjectV0Query
	}
	_, err := a.GQL.Request(ctx, query, struct {
		Feature   core.Feature `json:"feature"`
		ID        string       `json:"gcpNativeProjectId"`
		Name      string       `json:"gcpProjectName"`
		Number    int64        `json:"gcpProjectNumber"`
		OrgName   string       `json:"organizationName,omitempty"`
		JwtConfig string       `json:"serviceAccountJwtConfigOptional,omitempty"`
	}{Feature: feature, ID: projectID, Name: projectName, Number: projectNumber, OrgName: orgName, JwtConfig: jwtConfig})
	if err != nil {
		return fmt.Errorf("failed to request CloudAccountAddManualAuthProject: %v", err)
	}
	return nil
}

// CloudAccountDeleteProject delete cloud account for the given Polaris cloud
// account id.
func (a API) CloudAccountDeleteProject(ctx context.Context, id uuid.UUID) error {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, gcpCloudAccountDeleteProjectsQuery, struct {
		IDs []uuid.UUID `json:"nativeProtectionProjectUuids"`
	}{IDs: []uuid.UUID{id}})
	if err != nil {
		return fmt.Errorf("failed to request CloudAccountDeleteProject: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "gcpCloudAccountDeleteProjects(%q): %s", []uuid.UUID{id}, string(buf))

	var payload struct {
		Data struct {
			Projects []struct {
				ProjectUUID string `json:"projectUuid"`
				Success     bool   `json:"success"`
				Error       string `json:"error"`
			} `json:"gcpCloudAccountDeleteProjects"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal CloudAccountDeleteProject: %v", err)
	}
	if len(payload.Data.Projects) != 1 {
		return errors.New("expected a single result")
	}

	if !payload.Data.Projects[0].Success {
		return errors.New(payload.Data.Projects[0].Error)
	}

	return nil
}

// FeaturePermissionsForCloudAccount list the permissions needed to enable the
// given Polaris cloud account feature.
func (a API) FeaturePermissionsForCloudAccount(ctx context.Context, feature core.Feature) (permissions []string, err error) {
	a.GQL.Log().Print(log.Trace)

	query := allFeaturePermissionsForGcpCloudAccountQuery
	if graphql.VersionOlderThan(a.Version, "master-46133", "v20220315") {
		query = gcpCloudAccountListPermissionsV0Query
	} else if graphql.VersionOlderThan(a.Version, "master-46713", "v20220412") {
		query = gcpCloudAccountListPermissionsQuery
	}
	buf, err := a.GQL.Request(ctx, query, struct {
		Feature core.Feature `json:"feature"`
	}{Feature: feature})
	if err != nil {
		return nil, fmt.Errorf("failed to request CloudAccountListPermissions: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "%s(%q): %s", graphql.QueryName(query), feature, string(buf))

	var payload struct {
		Data struct {
			Permissions []struct {
				Permission string `json:"permission"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CloudAccountListPermissions: %v", err)
	}

	permissions = make([]string, 0, len(payload.Data.Permissions))
	for _, p := range payload.Data.Permissions {
		permissions = append(permissions, p.Permission)
	}

	return permissions, nil
}

// UpgradeCloudAccountPermissionsWithoutOAuth notifies Polaris that the
// permissions for the GCP service account has been updated for the
// specified Polaris cloud account id and feature.
func (a API) UpgradeCloudAccountPermissionsWithoutOAuth(ctx context.Context, id uuid.UUID, feature core.Feature) error {
	a.GQL.Log().Print(log.Trace)

	query := upgradeGcpCloudAccountPermissionsWithoutOauthQuery
	if graphql.VersionOlderThan(a.Version, "master-46133", "v20220315") {
		query = upgradeGcpCloudAccountPermissionsWithoutOauthV0Query
	}
	buf, err := a.GQL.Request(ctx, query, struct {
		ID      uuid.UUID    `json:"cloudAccountId"`
		Feature core.Feature `json:"feature"`
	}{ID: id, Feature: feature})
	if err != nil {
		return fmt.Errorf("failed to request UpgradeCloudAccountPermissionsWithoutOAuth: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "%s(%q, %q): %s", graphql.QueryName(query), id, feature, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Status struct {
					ProjectUUID uuid.UUID `json:"projectUuuid"`
					Success     bool      `json:"success"`
					Error       string    `json:"error"`
				} `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal UpgradeCloudAccountPermissionsWithoutOAuth: %v", err)
	}
	if !payload.Data.Result.Status.Success {
		return errors.New(payload.Data.Result.Status.Error)
	}

	return nil
}
