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

// Package cloudapp provides a high level interface for working with cloud
// native applications.
//
// A cloud native application, shown as a Cloud Application in the RSC UI,
// groups the resources of one AWS account and region so that their
// configuration can be protected by an SLA Domain with the
// AWS_CONFIG_OBJECT_TYPE object type. The application ID is also the object ID
// used when assigning the SLA Domain.
package cloudapp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlcloudapp "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cloudapp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	gqlaws "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	gqlsla "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for cloud native application management.
type API struct {
	client *graphql.Client
	log    log.Logger
}

// Wrap the RSC client in the cloud native application API.
func Wrap(client *polaris.Client) API {
	return API{client: client.GQL, log: client.GQL.Log()}
}

// WrapGQL wraps the GQL client in the cloud native application API.
func WrapGQL(client *graphql.Client) API {
	return API{client: client, log: client.Log()}
}

// ApplicationByID returns the application with the specified ID. If the
// application doesn't exist, or has been deleted, graphql.ErrNotFound is
// returned.
func (a API) ApplicationByID(ctx context.Context, id uuid.UUID) (gqlcloudapp.Application, error) {
	a.log.Print(log.Trace)

	app, err := gqlcloudapp.ApplicationByID(ctx, a.client, id)
	if err != nil {
		return gqlcloudapp.Application{}, fmt.Errorf("failed to get cloud native application %q: %w", id, err)
	}

	return app, nil
}

// ListAWSApplications returns the IDs and names of the AWS cloud native
// applications whose name contains nameFilter. An empty nameFilter returns all
// applications.
func (a API) ListAWSApplications(ctx context.Context, nameFilter string) ([]gqlcloudapp.ApplicationRef, error) {
	a.log.Print(log.Trace)

	refs, err := gqlcloudapp.ListAWSApplications(ctx, a.client, nameFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list cloud native applications: %s", err)
	}

	return refs, nil
}

// CreateAWSApplication creates an AWS cloud native application and returns
// its ID.
//
// RSC validates neither the application name nor the target cloud account on
// create, so both are checked here: the name must be valid and unused, as
// ValidateApplicationName reports, and the target cloud account must differ
// from the cloud account, as the RSC UI requires.
func (a API) CreateAWSApplication(ctx context.Context, params gqlcloudapp.CreateAWSApplicationParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if params.Region.Region == gqlaws.RegionUnknown {
		return uuid.Nil, errors.New("invalid AWS region")
	}
	if params.TargetCloudAccountID == params.CloudAccountID {
		return uuid.Nil, errors.New("the target cloud account must differ from the cloud account")
	}
	if len(params.MatchingTags) == 0 && len(params.ResourceARNs) == 0 {
		return uuid.Nil, errors.New("at least one matching tag or resource ARN is required")
	}

	status, err := gqlcloudapp.ValidateApplicationName(ctx, a.client, params.Name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to validate cloud native application name %q: %s", params.Name, err)
	}
	if status != gqlcloudapp.NameValid {
		return uuid.Nil, fmt.Errorf("invalid cloud native application name %q: %s", params.Name, status)
	}

	// RSC persists the application before it checks the cloud account, so an
	// application created in an account without a connected
	// CLOUD_NATIVE_CONFIG_PROTECTION feature is left behind, without an ID,
	// and breaks the application list once the account is removed.
	account, err := aws.WrapGQL(a.client).AccountByID(ctx, params.CloudAccountID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get cloud account %q: %s", params.CloudAccountID, err)
	}
	feature, ok := account.Feature(core.FeatureCloudNativeConfigProtection)
	if !ok || feature.Status != core.StatusConnected {
		return uuid.Nil, fmt.Errorf("the %s feature of cloud account %q is not connected",
			core.FeatureCloudNativeConfigProtection.Name, params.CloudAccountID)
	}
	if !feature.HasRegion(params.Region.Region.Name()) {
		return uuid.Nil, fmt.Errorf("the %s feature of cloud account %q is not enabled in region %s",
			core.FeatureCloudNativeConfigProtection.Name, params.CloudAccountID, params.Region.Region.Name())
	}

	id, err := gqlcloudapp.CreateAWSApplication(ctx, a.client, params)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create cloud native application %q: %s", params.Name, err)
	}

	return id, nil
}

// UpdateAWSApplication updates the AWS cloud native application with the
// specified ID. The parameters must hold the complete desired state of the
// application.
func (a API) UpdateAWSApplication(ctx context.Context, id uuid.UUID, params gqlcloudapp.UpdateAWSApplicationParams) error {
	a.log.Print(log.Trace)

	if params.TargetCloudAccountID == uuid.Nil {
		return errors.New("the target cloud account is required")
	}

	if err := gqlcloudapp.UpdateAWSApplication(ctx, a.client, id, params); err != nil {
		return fmt.Errorf("failed to update cloud native application %q: %s", id, err)
	}

	return nil
}

// ValidateApplicationName returns the validation status of the specified
// application name.
func (a API) ValidateApplicationName(ctx context.Context, name string) (gqlcloudapp.NameValidationStatus, error) {
	a.log.Print(log.Trace)

	status, err := gqlcloudapp.ValidateApplicationName(ctx, a.client, name)
	if err != nil {
		return "", fmt.Errorf("failed to validate cloud native application name %q: %s", name, err)
	}

	return status, nil
}

// RemoveApplication removes the cloud native application with the specified
// ID. RSC refuses to remove an application while it has config snapshots. If
// deleteSnapshots is true, the application is first set to Do Not Protect
// with its existing snapshots expiring immediately, its remaining unmanaged
// snapshots are deleted, and RemoveApplication waits until RSC reports no
// config snapshot before removing the application. The wait ends when ctx is
// done. Snapshots of the workloads in the application belong to the workloads
// and are not affected.
func (a API) RemoveApplication(ctx context.Context, id uuid.UUID, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	if deleteSnapshots {
		if err := a.deleteApplicationSnapshots(ctx, id); err != nil {
			return fmt.Errorf("failed to delete snapshots of cloud native application %q: %s", id, err)
		}
	}

	if err := gqlcloudapp.DeleteApplication(ctx, a.client, id); err != nil {
		return fmt.Errorf("failed to remove cloud native application %q: %s", id, err)
	}

	return nil
}

func (a API) deleteApplicationSnapshots(ctx context.Context, id uuid.UUID) error {
	if err := gqlsla.AssignDomain(ctx, a.client, gqlsla.AssignDomainParams{
		DomainAssignType:          gqlsla.DoNotProtect,
		ObjectIDs:                 []uuid.UUID{id},
		ExistingSnapshotRetention: gqlsla.ExpireImmediately,
	}); err != nil {
		return err
	}
	if err := gqlcloudapp.DeleteSnapshotsOfUnmanagedObjects(ctx, a.client, []uuid.UUID{id}); err != nil {
		return err
	}

	// RSC expires the snapshots asynchronously.
	for {
		hasSnapshot, err := gqlcloudapp.HasApplicationSnapshot(ctx, a.client, id)
		if err != nil {
			return err
		}
		if !hasSnapshot {
			return nil
		}

		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			return fmt.Errorf("config snapshots still present: %w", ctx.Err())
		}
	}
}
