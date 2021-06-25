//go:generate go run ../queries_gen.go core

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

package core

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CloudAccountAction represents a Polaris cloud account action.
type CloudAccountAction string

const (
	Create              CloudAccountAction = "CREATE"
	Delete              CloudAccountAction = "DELETE"
	UpdateChildAccounts CloudAccountAction = "UPDATE_CHILD_ACCOUNTS"
	UpdatePermissions   CloudAccountAction = "UPDATE_PERMISSIONS"
	UpdateRegions       CloudAccountAction = "UPDATE_REGIONS"
)

// CloudAccountFeature represents a Polaris cloud account feature.
type CloudAccountFeature string

const (
	InvalidFeature        CloudAccountFeature = ""
	AllFeatures           CloudAccountFeature = "ALL"
	AppFlows              CloudAccountFeature = "APP_FLOWS"
	Archival              CloudAccountFeature = "ARCHIVAL"
	CloudAccounts         CloudAccountFeature = "CLOUDACCOUNTS"
	CloudNativeArchival   CloudAccountFeature = "CLOUD_NATIVE_ARCHIVAL"
	CloudNativeProtection CloudAccountFeature = "CLOUD_NATIVE_PROTECTION"
	Exocompute            CloudAccountFeature = "EXOCOMPUTE"
	GCPSharedVPCHost      CloudAccountFeature = "GCP_SHARED_VPC_HOST"
	RDSProtection         CloudAccountFeature = "RDS_PROTECTION"
)

var validFeatures = map[CloudAccountFeature]struct{}{
	InvalidFeature:        {},
	AllFeatures:           {},
	AppFlows:              {},
	Archival:              {},
	CloudAccounts:         {},
	CloudNativeArchival:   {},
	CloudNativeProtection: {},
	Exocompute:            {},
	GCPSharedVPCHost:      {},
	RDSProtection:         {},
}

// ParseFeature returns the CloudAccountFeature matching the given feature
// name. Case insensitive.
func ParseFeature(feature string) (CloudAccountFeature, error) {
	f := CloudAccountFeature(strings.ToUpper(feature))
	if _, ok := validFeatures[f]; ok {
		return f, nil
	}

	return InvalidFeature, errors.New("polaris: invalid feature")
}

// CloudAccountStatus represents a Polaris cloud account status.
type CloudAccountStatus string

const (
	Connected          CloudAccountStatus = "CONNECTED"
	Connecting         CloudAccountStatus = "CONNECTING"
	Disabled           CloudAccountStatus = "DISABLED"
	Disconnected       CloudAccountStatus = "DISCONNECTED"
	MissingPermissions CloudAccountStatus = "MISSING_PERMISSIONS"
)

// TaskChainState represents the state of a Polaris task chain.
type TaskChainState string

const (
	TaskChainInvalid   TaskChainState = ""
	TaskChainCanceled  TaskChainState = "CANCELED"
	TaskChainCanceling TaskChainState = "CANCELING"
	TaskChainFailed    TaskChainState = "FAILED"
	TaskChainReady     TaskChainState = "READY"
	TaskChainRunning   TaskChainState = "RUNNING"
	TaskChainSucceeded TaskChainState = "SUCCEEDED"
	TaskChainUndoing   TaskChainState = "UNDOING"
)

// SLAAssignment represents the type of a SLA assignment in Polaris.
type SLAAssignment string

const (
	Derived    SLAAssignment = "Derived"
	Direct     SLAAssignment = "Direct"
	Unassigned SLAAssignment = "Unassigned"
)

// SLADomain represents a Polaris SLA domain.
type SLADomain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// API wraps around GraphQL clients to give them the Polaris Core API.
type API struct {
	GQL *graphql.Client
}

// Wrap the GraphQL client in the Core API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql}
}

// TaskChain is a collection of sequential tasks that all must complete for the
// task chain to be considered complete.
type TaskChain struct {
	ID          int64          `json:"id"`
	TaskChainID uuid.UUID      `json:"taskchainUuid"`
	State       TaskChainState `json:"state"`
}

// KorgTaskChainStatus returns the task chain for the specified task chain id.
// If the task chain id refers to a task chain that was just created it's state
// might not have reached ready yet. This can be detected by state being
// TaskChainInvalid and error is nil.
func (a API) KorgTaskChainStatus(ctx context.Context, id uuid.UUID) (TaskChain, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/core.KorgTaskChainStatus")

	buf, err := a.GQL.Request(ctx, getKorgTaskchainStatusQuery, struct {
		TaskChainID uuid.UUID `json:"taskchainId,omitempty"`
	}{TaskChainID: id})
	if err != nil {
		return TaskChain{}, err
	}

	a.GQL.Log().Printf(log.Debug, "getKorgTaskchainStatus(%q): %s", id, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				TaskChain TaskChain `json:"taskchain"`
			} `json:"getKorgTaskchainStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return TaskChain{}, err
	}

	return payload.Data.Query.TaskChain, nil
}

// WaitForTaskChain blocks until the Polaris task chain with the specified task
// chain id has completed. When the task chain completes the final state of the
// task chain is returned. The wait parameter specifies the amount of time to
// wait before requesting another task status update.
func (a API) WaitForTaskChain(ctx context.Context, id uuid.UUID, wait time.Duration) (TaskChainState, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/core.WaitForTaskChain")

	for {
		taskChain, err := a.KorgTaskChainStatus(ctx, id)
		if err != nil {
			return TaskChainInvalid, err
		}

		if taskChain.State == TaskChainSucceeded || taskChain.State == TaskChainCanceled || taskChain.State == TaskChainFailed {
			return taskChain.State, nil
		}

		a.GQL.Log().Printf(log.Debug, "Waiting for Polaris task chain: %v", id)

		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return TaskChainInvalid, ctx.Err()
		}
	}
}
