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

package exocompute

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// HealthCheckStatusValue represents the status of an exocompute health check.
type HealthCheckStatusValue string

const (
	HealthCheckStatusCancelled                           HealthCheckStatusValue = "CANCELLED"
	HealthCheckStatusCustomerManagedClusterNotRegistered HealthCheckStatusValue = "CUSTOMER_MANAGED_CLUSTER_NOT_REGISTERED"
	HealthCheckStatusHealthy                             HealthCheckStatusValue = "HEALTHY"
	HealthCheckStatusJobUnsuccessful                     HealthCheckStatusValue = "JOB_UNSUCCESSFUL"
	HealthCheckStatusUnhealthy                           HealthCheckStatusValue = "UNHEALTHY"
	HealthCheckStatusUnknown                             HealthCheckStatusValue = "UNKNOWN"
	HealthCheckStatusVerifying                           HealthCheckStatusValue = "VERIFYING"
)

// HealthCheckStatus holds the status of the latest health check of the
// exocompute cluster serving an exocompute configuration. TaskchainID is the
// ID of the health check job which produced the status.
type HealthCheckStatus struct {
	Status        HealthCheckStatusValue `json:"status"`
	FailureReason string                 `json:"failureReason"`
	LastUpdatedAt string                 `json:"lastUpdatedAt"`
	TaskchainID   string                 `json:"taskchainId"`
}

// TriggerHealthCheck starts an on-demand health check of the exocompute
// cluster serving the exocompute configuration with the specified ID. Returns
// the ID of the health check job. The health check runs asynchronously, its
// result is reported in the HealthCheckStatus of the configuration once the
// TaskchainID matches the returned job ID.
func TriggerHealthCheck(ctx context.Context, gql *graphql.Client, cloudVendor core.CloudVendor, configID uuid.UUID) (string, error) {
	gql.Log().Print(log.Trace)

	query := triggerExocomputeHealthCheckQuery
	buf, err := gql.Request(ctx, query, struct {
		CloudVendor core.CloudVendor `json:"cloudVendor"`
		ConfigID    uuid.UUID        `json:"exocomputeConfigId"`
	}{CloudVendor: cloudVendor, ConfigID: configID})
	if err != nil {
		return "", graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				JobID string `json:"healthCheckJobId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", graphql.UnmarshalError(query, err)
	}
	if payload.Data.Result.JobID == "" {
		return "", graphql.ResponseError(query, errors.New("no health check job id returned"))
	}

	return payload.Data.Result.JobID, nil
}
