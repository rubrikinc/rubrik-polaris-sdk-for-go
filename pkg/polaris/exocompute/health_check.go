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
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/exocompute"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// healthCheckPollInterval is the sleep between WaitForHealthCheck polls.
var healthCheckPollInterval = 30 * time.Second

// TriggerHealthCheck starts an on-demand health check of the exocompute
// cluster serving the AWS or Azure exocompute configuration with the specified
// ID. Returns the ID of the health check job, pass it to WaitForHealthCheck to
// wait for the result.
//
// For an RSC managed configuration, RSC rejects the request while a health
// check of the cluster is already running. For a customer managed
// configuration, RSC runs the health check as part of a reconciliation of the
// customer managed cluster.
func (a API) TriggerHealthCheck(ctx context.Context, cloudVendor core.CloudVendor, configID uuid.UUID) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if cloudVendor != core.CloudVendorAWS && cloudVendor != core.CloudVendorAzure {
		return uuid.Nil, fmt.Errorf("unsupported cloud vendor for exocompute health check: %s", cloudVendor)
	}

	jobID, err := exocompute.TriggerHealthCheck(ctx, a.client, cloudVendor, configID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to trigger health check for exocompute configuration %s: %s", configID, err)
	}

	return jobID, nil
}

// WaitForHealthCheck blocks until the health check job with the specified ID
// has finished for the AWS or Azure exocompute configuration with the
// specified ID, polling every 30 seconds. Returns the final health check
// status, which can be any status other than VERIFYING. A status other than
// HEALTHY is not reported as an error.
//
// RSC reports the health check status of the latest exocompute cluster in the
// configuration's region, so the status is only considered once its
// TaskchainID matches the job ID. The state of the job's task chain is checked
// too, so the wait ends with an error if the task chain ends without reporting
// a final status for the job. If the task chain state isn't available, the
// wait relies on the status alone and blocks until the context is done if RSC
// never reports a status for the job, so the context should have a deadline.
//
// Returns an error, together with the last health check status seen for the
// job, if the context is done or the task chain ends before the health check
// has finished, or if the configuration can't be found.
func (a API) WaitForHealthCheck(ctx context.Context, cloudVendor core.CloudVendor, configID, jobID uuid.UUID) (exocompute.HealthCheckStatus, error) {
	a.log.Print(log.Trace)

	if cloudVendor != core.CloudVendorAWS && cloudVendor != core.CloudVendorAzure {
		return exocompute.HealthCheckStatus{}, fmt.Errorf("unsupported cloud vendor for exocompute health check: %s", cloudVendor)
	}

	var last exocompute.HealthCheckStatus
	err := core.Wrap(a.client).WaitForTaskChainWithStatus(ctx, jobID, healthCheckPollInterval, func(ctx context.Context) (bool, error) {
		status, err := a.healthCheckStatus(ctx, cloudVendor, configID)
		if err != nil {
			if ctx.Err() != nil {
				return false, ctx.Err()
			}
			return false, fmt.Errorf("failed to get health check status of exocompute configuration %s: %w", configID, err)
		}

		// A taskchain ID which isn't a UUID, e.g. an empty one, can't belong
		// to the job.
		if id, err := uuid.Parse(status.TaskchainID); err != nil || id != jobID {
			a.log.Printf(log.Debug, "Health check job %s of exocompute configuration %s not reported yet (status: %q, job: %q)",
				jobID, configID, status.Status, status.TaskchainID)
			return false, nil
		}
		last = status
		return status.Status != exocompute.HealthCheckStatusVerifying, nil
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return last, fmt.Errorf("%w while waiting for health check job %s of exocompute configuration %s",
				err, jobID, configID)
		}
		return last, fmt.Errorf("failed to wait for health check job %s of exocompute configuration %s: %w",
			jobID, configID, err)
	}

	return last, nil
}

// healthCheckStatus returns the health check status of the AWS or Azure
// exocompute configuration with the specified ID.
func (a API) healthCheckStatus(ctx context.Context, cloudVendor core.CloudVendor, configID uuid.UUID) (exocompute.HealthCheckStatus, error) {
	switch cloudVendor {
	case core.CloudVendorAWS:
		config, err := a.AWSConfigurationByID(ctx, configID)
		return config.HealthCheckStatus, err
	case core.CloudVendorAzure:
		config, err := a.AzureConfigurationByID(ctx, configID)
		return config.HealthCheckStatus, err
	default:
		return exocompute.HealthCheckStatus{}, fmt.Errorf("unsupported cloud vendor for exocompute health check: %s", cloudVendor)
	}
}
