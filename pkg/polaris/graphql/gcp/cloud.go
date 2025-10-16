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
	"strings"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CloudAccount represents an RSC Cloud Account for GCP. If UseGlobalConfig
// is true the cloud account depends on the default service account.
type CloudAccount struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	ProjectID        string    `json:"projectID"`
	ProjectNumber    int64     `json:"projectNumber"`
	RoleID           string    `json:"roleId"`
	UsesGlobalConfig bool      `json:"usesGlobalConfig"`
}

// Feature represents an RSC Cloud Account feature for GCP, e.g. Cloud Native
// Protection.
type Feature struct {
	Feature string      `json:"feature"`
	Status  core.Status `json:"status"`
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
	a.log.Print(log.Trace)

	query := allGcpCloudAccountProjectsByFeatureQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Feature string `json:"feature"`
		Filter  string `json:"projectSearchText"`
	}{Feature: feature.Name, Filter: filter})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Accounts []CloudAccountWithFeature `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Accounts, nil
}

// CloudAccountAddManualAuthProject adds the GCP project and features to RSC.
func (a API) CloudAccountAddManualAuthProject(ctx context.Context, projectID, projectName string, projectNumber int64, orgName, jwtConfig string, features []core.Feature) error {
	a.log.Print(log.Trace)

	featuresWithoutPG, featuresWithPG, err := core.FilterFeaturesOnPermissionGroups(features)
	if err != nil {
		return fmt.Errorf("failed to add project: %s", err)
	}

	query := gcpCloudAccountAddManualAuthProjectQuery
	if _, err := a.GQL.Request(ctx, query, struct {
		ID             string         `json:"gcpNativeProjectId"`
		Name           string         `json:"gcpProjectName"`
		Number         int64          `json:"gcpProjectNumber"`
		OrgName        string         `json:"organizationName,omitempty"`
		JwtConfig      secret.String  `json:"serviceAccountJwtConfig,omitempty"`
		JwtConfigOpt   secret.String  `json:"serviceAccountJwtConfigOptional,omitempty"`
		Features       []string       `json:"features,omitempty"`
		FeaturesWithPG []core.Feature `json:"featuresWithPG,omitempty"`
	}{
		ID:             projectID,
		Name:           projectName,
		Number:         projectNumber,
		OrgName:        orgName,
		JwtConfig:      secret.String(jwtConfig),
		JwtConfigOpt:   secret.String(jwtConfig),
		Features:       featuresWithoutPG,
		FeaturesWithPG: featuresWithPG,
	}); err != nil {
		return graphql.RequestError(query, err)
	}

	return nil
}

type DeleteProjectFeatureInput struct {
	CloudAccountIDs []uuid.UUID `json:"cloudAccountIds"`
	DeleteSnapshots bool        `json:"deleteSnapshots"`
	Feature         string      `json:"feature"`
}

type DeleteProjectJob struct {
	ID             uuid.UUID
	CloudAccountID uuid.UUID
	Feature        core.Feature
}

// CloudAccountDeleteProjectV2 removes the feature from the project with the
// specified cloud account ID.
func (a API) CloudAccountDeleteProjectV2(ctx context.Context, cloudAccountID uuid.UUID, features []core.Feature, deleteSnapshots bool) ([]DeleteProjectJob, error) {
	a.log.Print(log.Trace)

	featureInputs := make([]DeleteProjectFeatureInput, 0, len(features))
	for _, feature := range features {
		featureInputs = append(featureInputs, DeleteProjectFeatureInput{
			CloudAccountIDs: []uuid.UUID{cloudAccountID},
			DeleteSnapshots: deleteSnapshots,
			Feature:         feature.Name,
		})
	}

	query := gcpCloudAccountDeleteProjectsV2Query
	buf, err := a.GQL.Request(ctx, query, struct {
		Features []DeleteProjectFeatureInput `json:"features"`
	}{Features: featureInputs})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Errors []struct {
					Error          string `json:"error"`
					RubrikObjectID string `json:"rubrikObjectId"`
				} `json:"errors"`
				JobIDs []struct {
					JobID          uuid.UUID `json:"jobId"`
					RubrikObjectID string    `json:"rubrikObjectId"`
				} `json:"jobIds"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}
	if len(payload.Data.Result.Errors) > 0 {
		return nil, graphql.ResponseError(query, errors.New(payload.Data.Result.Errors[0].Error))
	}

	jobs := make([]DeleteProjectJob, 0, len(payload.Data.Result.JobIDs))
	for _, job := range payload.Data.Result.JobIDs {
		parts := strings.Split(job.RubrikObjectID, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid object id: %s", job.RubrikObjectID)
		}

		cloudAccountID, err := uuid.Parse(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid cloud account id: %s", parts[0])
		}

		jobs = append(jobs, DeleteProjectJob{
			ID:             job.JobID,
			CloudAccountID: cloudAccountID,
			Feature:        core.Feature{Name: parts[1]},
		})
	}

	return jobs, nil
}

// FeaturePermissionsForCloudAccount list the permissions needed to enable the
// given RSC cloud account feature.
func (a API) FeaturePermissionsForCloudAccount(ctx context.Context, feature core.Feature) (permissions []string, err error) {
	a.log.Print(log.Trace)

	query := allFeaturePermissionsForGcpCloudAccountQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Feature string `json:"feature"`
	}{Feature: feature.Name})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Permissions []struct {
				Permission string `json:"permission"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	permissions = make([]string, 0, len(payload.Data.Permissions))
	for _, p := range payload.Data.Permissions {
		permissions = append(permissions, p.Permission)
	}

	return permissions, nil
}

// UpgradeCloudAccountPermissionsWithoutOAuth notifies RSC that the permissions
// for the GCP service account has been updated for the specified RSC cloud
// account ID and feature.
func (a API) UpgradeCloudAccountPermissionsWithoutOAuth(ctx context.Context, cloudAccountID uuid.UUID, feature core.Feature) error {
	a.log.Print(log.Trace)

	query := upgradeGcpCloudAccountPermissionsWithoutOauthQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID      uuid.UUID `json:"cloudAccountId"`
		Feature string    `json:"feature"`
	}{ID: cloudAccountID, Feature: feature.Name})
	if err != nil {
		return graphql.RequestError(query, err)
	}

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
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result.Status.Success {
		return graphql.ResponseError(query, errors.New(payload.Data.Result.Status.Error))
	}

	return nil
}
