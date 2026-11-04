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

package cloudapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AWSServices holds the AWS services whose resources are protected as part of
// an application. VPCs, IAM roles and Route 53 hosted zones are always
// included and have no flag. The RSC GraphQL API also has the ebsEnabled flag,
// which the RSC UI never sets, so it's left out.
type AWSServices struct {
	EC2         bool `json:"ec2Enabled"`
	ECR         bool `json:"ecrEnabled"`
	ECS         bool `json:"isEcsEnabled"`
	EKS         bool `json:"eksEnabled"`
	ElastiCache bool `json:"elasticacheEnabled"`
	ELB         bool `json:"elbEnabled"`
	Lambda      bool `json:"lambdaEnabled"`
	RDS         bool `json:"rdsEnabled"`
	S3          bool `json:"s3Enabled"`
	Secrets     bool `json:"secretsEnabled"`
}

// ServiceConfig represents the RSC GraphQL ApplicationServiceConfiguration
// type. RSC replaces the whole service configuration on update, so every flag
// is always sent.
type ServiceConfig struct {
	CloudConfig struct {
		AWS AWSServices `json:"awsConfig"`
	} `json:"applicationCloudConfig"`
}

// NewServiceConfig returns the service configuration for the specified AWS
// services.
func NewServiceConfig(services AWSServices) ServiceConfig {
	var config ServiceConfig
	config.CloudConfig.AWS = services
	return config
}

// MatchingTags represents the RSC GraphQL MatchingTags type. A resource
// carrying any one of the tags is included in the application.
type MatchingTags struct {
	Tags []core.Tag `json:"matchingTagsList"`
}

// CloudRegion represents the RSC GraphQL CloudRegion type for an AWS region.
type CloudRegion struct {
	Region struct {
		AWSRegion aws.RegionEnum `json:"awsRegion"`
	} `json:"region"`
}

// Application represents the RSC GraphQL CloudNativeApplication type.
type Application struct {
	ID                   uuid.UUID        `json:"id"`
	Name                 string           `json:"name"`
	CloudAccountID       uuid.UUID        `json:"cloudNativeAccountId"`
	TargetCloudAccountID uuid.UUID        `json:"targetCloudNativeAccountId"`
	Region               CloudRegion      `json:"region"`
	MatchingTags         MatchingTags     `json:"matchingTags"`
	ResourceARNs         []string         `json:"resourceInclusionList"`
	ExcludedResourceARNs []string         `json:"resourceExclusionList"`
	ServiceConfig        ServiceConfig    `json:"serviceConfigured"`
	RefreshStatus        RefreshStatus    `json:"refreshStatus"`
	LastRefreshAt        *time.Time       `json:"lastRefreshAt"`
	ProtectionStatus     ProtectionStatus `json:"protectionStatus"`
}

// ApplicationByID returns the application with the specified ID. If the
// application doesn't exist, graphql.ErrNotFound is returned.
//
// RSC deletes an application by archiving it, after which the application can
// still be read by ID. An archived application is reported as a relic by the
// hierarchy, which is how it's detected. An ID that never belonged to an
// application is rejected by RSC with a 403 error and no data.
func ApplicationByID(ctx context.Context, gql *graphql.Client, id uuid.UUID) (Application, error) {
	gql.Log().Print(log.Trace)

	query := cloudNativeApplicationQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"id"`
	}{ID: id})
	if err != nil {
		var gqlErr graphql.GQLError
		if errors.As(err, &gqlErr) && gqlErr.Code() == 403 {
			return Application{}, fmt.Errorf("cloud native application %q %w", id, graphql.ErrNotFound)
		}
		return Application{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result    *Application `json:"result"`
			Hierarchy *struct {
				IsRelic bool `json:"isRelic"`
			} `json:"hierarchy"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return Application{}, graphql.UnmarshalError(query, err)
	}
	if payload.Data.Result == nil || (payload.Data.Hierarchy != nil && payload.Data.Hierarchy.IsRelic) {
		return Application{}, fmt.Errorf("cloud native application %q %w", id, graphql.ErrNotFound)
	}

	return *payload.Data.Result, nil
}

// ApplicationRef identifies an application by ID and name.
type ApplicationRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ListAWSApplications returns the IDs and names of the AWS cloud native
// applications whose name contains nameFilter. An empty nameFilter returns all
// applications. Deleted applications are not returned. Use ApplicationByID to
// read an application in full, the list doesn't always hold the resource
// lists while an application refreshes.
func ListAWSApplications(ctx context.Context, gql *graphql.Client, nameFilter string) ([]ApplicationRef, error) {
	gql.Log().Print(log.Trace)

	query := cloudNativeApplicationsQuery
	var refs []ApplicationRef
	var cursor string
	for {
		buf, err := gql.Request(ctx, query, struct {
			After           string `json:"after,omitempty"`
			ApplicationName string `json:"applicationName,omitempty"`
		}{After: cursor, ApplicationName: nameFilter})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []ApplicationRef `json:"nodes"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, graphql.UnmarshalError(query, err)
		}
		refs = append(refs, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return refs, nil
}

// CreateAWSApplicationParams holds the parameters for creating an AWS cloud
// native application. At least one of MatchingTags and ResourceARNs must be
// non-empty.
type CreateAWSApplicationParams struct {
	// CloudAccountID is the RSC cloud account ID of the AWS account the
	// application lives in. The account must be onboarded with the
	// CLOUD_NATIVE_CONFIG_PROTECTION feature.
	CloudAccountID uuid.UUID `json:"cloudNativeAccountId"`

	// TargetCloudAccountID is the RSC cloud account ID of the AWS account the
	// application is recovered to. RSC doesn't validate it.
	TargetCloudAccountID uuid.UUID `json:"targetCloudNativeAccountId"`

	// Name is the name of the application. RSC doesn't enforce the format or
	// uniqueness of the name on create, see ValidateApplicationName.
	Name string `json:"name"`

	// Region is the AWS region of the application.
	Region aws.RegionEnum `json:"awsRegion"`

	// MatchingTags are the tags matching resources to the application.
	MatchingTags []core.Tag `json:"matchingTags"`

	// ResourceARNs are the ARNs of resources explicitly included in the
	// application. Tagged resources, and resources related to the listed
	// ones, are included by RSC and shouldn't be listed. Each ARN must have
	// been discovered by RSC in the account and region.
	ResourceARNs []string `json:"resourceInclusionList"`

	// ExcludedResourceARNs are the ARNs of resources excluded from the
	// application, even when tagged or related to an included resource.
	ExcludedResourceARNs []string `json:"resourceExclusionList"`

	// ServiceConfig holds the AWS services whose resources are protected,
	// see NewServiceConfig.
	ServiceConfig ServiceConfig `json:"serviceConfigured"`
}

// CreateAWSApplication creates an AWS cloud native application and returns
// its ID.
func CreateAWSApplication(ctx context.Context, gql *graphql.Client, params CreateAWSApplicationParams) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	// RSC rejects null for the non-null list inputs.
	if params.MatchingTags == nil {
		params.MatchingTags = []core.Tag{}
	}
	if params.ResourceARNs == nil {
		params.ResourceARNs = []string{}
	}
	if params.ExcludedResourceARNs == nil {
		params.ExcludedResourceARNs = []string{}
	}

	query := createCloudNativeApplicationQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Application struct {
					ID uuid.UUID `json:"id"`
				} `json:"application"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	if payload.Data.Result.Application.ID == uuid.Nil {
		return uuid.Nil, graphql.ResponseError(query, errors.New("no application ID returned"))
	}

	return payload.Data.Result.Application.ID, nil
}

// UpdateAWSApplicationParams holds the parameters for updating an AWS cloud
// native application. Every field is sent, so the parameters must hold the
// complete desired state. An empty ResourceARNs or ExcludedResourceARNs
// clears the list. The tags can't be cleared, RSC ignores an empty list. The
// cloud account and the region can't be changed.
type UpdateAWSApplicationParams struct {
	Name                 string        `json:"name"`
	TargetCloudAccountID uuid.UUID     `json:"targetCloudNativeAccountId"`
	MatchingTags         []core.Tag    `json:"matchingTags"`
	ResourceARNs         []string      `json:"resourceInclusionList"`
	ExcludedResourceARNs []string      `json:"resourceExclusionList"`
	ServiceConfig        ServiceConfig `json:"serviceConfigured"`
}

// UpdateAWSApplication updates the AWS cloud native application with the
// specified ID.
func UpdateAWSApplication(ctx context.Context, gql *graphql.Client, id uuid.UUID, params UpdateAWSApplicationParams) error {
	gql.Log().Print(log.Trace)

	// RSC rejects null for the non-null list inputs, and an empty selection
	// list is how a list is cleared.
	if params.MatchingTags == nil {
		params.MatchingTags = []core.Tag{}
	}
	if params.ResourceARNs == nil {
		params.ResourceARNs = []string{}
	}
	if params.ExcludedResourceARNs == nil {
		params.ExcludedResourceARNs = []string{}
	}

	query := updateCloudNativeApplicationQuery
	if _, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"id"`
		UpdateAWSApplicationParams
	}{ID: id, UpdateAWSApplicationParams: params}); err != nil {
		return graphql.RequestError(query, err)
	}

	return nil
}

// DeleteApplication deletes the cloud native application with the specified
// ID. RSC archives the application, keeping its SLA Domain assignment. Deleting
// an already deleted application succeeds. RSC refuses to delete an
// application while it has config snapshots.
func DeleteApplication(ctx context.Context, gql *graphql.Client, id uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deleteCloudNativeApplicationQuery
	if _, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"id"`
	}{ID: id}); err != nil {
		return graphql.RequestError(query, err)
	}

	return nil
}

// ValidateApplicationName returns the validation status of the specified
// application name. RSC checks the format and that no other application has
// the name.
func ValidateApplicationName(ctx context.Context, gql *graphql.Client, name string) (NameValidationStatus, error) {
	gql.Log().Print(log.Trace)

	query := isValidCloudNativeApplicationNameQuery
	buf, err := gql.Request(ctx, query, struct {
		Name string `json:"name"`
	}{Name: name})
	if err != nil {
		return "", graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Status NameValidationStatus `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Status, nil
}

// HasApplicationSnapshot returns true if the application with the specified
// ID has at least one unexpired config snapshot taken before now.
func HasApplicationSnapshot(ctx context.Context, gql *graphql.Client, id uuid.UUID) (bool, error) {
	gql.Log().Print(log.Trace)

	query := cloudNativeApplicationSnapshotsQuery
	buf, err := gql.Request(ctx, query, struct {
		ApplicationID uuid.UUID `json:"applicationId"`
		BeforeTime    time.Time `json:"beforeTime"`
	}{ApplicationID: id, BeforeTime: time.Now().UTC()})
	if err != nil {
		// RSC reports an application without config snapshots as an error.
		var gqlErr graphql.GQLError
		if errors.As(err, &gqlErr) && len(gqlErr.Errors) > 0 &&
			strings.Contains(gqlErr.Errors[0].Message, "no config snapshot found") {
			return false, nil
		}
		return false, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result *struct {
				ConfigSnapshot *struct {
					SnapshotFID uuid.UUID `json:"snapshotFid"`
				} `json:"configSnapshot"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return false, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result != nil && payload.Data.Result.ConfigSnapshot != nil &&
		payload.Data.Result.ConfigSnapshot.SnapshotFID != uuid.Nil, nil
}

// DeleteSnapshotsOfUnmanagedObjects expires the snapshots of the objects with
// the specified IDs which aren't retained by an SLA Domain. Expiry is
// asynchronous.
func DeleteSnapshotsOfUnmanagedObjects(ctx context.Context, gql *graphql.Client, objectIDs []uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deleteSnapshotsOfUnmanagedObjectsQuery
	buf, err := gql.Request(ctx, query, struct {
		ObjectIDs []uuid.UUID `json:"objectIds"`
	}{ObjectIDs: objectIDs})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"success"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result.Success {
		return graphql.ResponseError(query, errors.New("snapshot deletion not successful"))
	}

	return nil
}
