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
func (a API) TriggerHealthCheck(ctx context.Context, cloudVendor core.CloudVendor, configID uuid.UUID) (string, error) {
	a.log.Print(log.Trace)

	if cloudVendor != core.CloudVendorAWS && cloudVendor != core.CloudVendorAzure {
		return "", fmt.Errorf("unsupported cloud vendor for exocompute health check: %s", cloudVendor)
	}

	jobID, err := exocompute.TriggerHealthCheck(ctx, a.client, cloudVendor, configID)
	if err != nil {
		return "", fmt.Errorf("failed to trigger health check for exocompute configuration %s: %s", configID, err)
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
// TaskchainID matches the job ID. RSC only reports the health check status
// when the CNP_EXOCOMPUTE_HEALTH_CHECK_UI_ENABLED feature flag is enabled, if
// it isn't, WaitForHealthCheck blocks until the context is done.
//
// Returns an error if the context is done before the health check has
// finished, together with the last health check status seen for the job, or
// if the configuration can't be found.
func (a API) WaitForHealthCheck(ctx context.Context, cloudVendor core.CloudVendor, configID uuid.UUID, jobID string) (exocompute.HealthCheckStatus, error) {
	a.log.Print(log.Trace)

	var last exocompute.HealthCheckStatus
	for {
		var status exocompute.HealthCheckStatus
		var err error
		switch cloudVendor {
		case core.CloudVendorAWS:
			var config AWSConfiguration
			config, err = a.AWSConfigurationByID(ctx, configID)
			status = config.HealthCheckStatus
		case core.CloudVendorAzure:
			var config AzureConfiguration
			config, err = a.AzureConfigurationByID(ctx, configID)
			status = config.HealthCheckStatus
		default:
			return last, fmt.Errorf("unsupported cloud vendor for exocompute health check: %s", cloudVendor)
		}
		if err != nil {
			if ctx.Err() != nil {
				return last, fmt.Errorf("%w while waiting for health check job %q of exocompute configuration %s",
					ctx.Err(), jobID, configID)
			}
			return last, fmt.Errorf("failed to get health check status of exocompute configuration %s: %w", configID, err)
		}

		if status.TaskchainID == jobID {
			last = status
			if status.Status != exocompute.HealthCheckStatusVerifying {
				return status, nil
			}
		}
		a.log.Printf(log.Debug, "Health check job %q of exocompute configuration %s still running (status: %q, job: %q)",
			jobID, configID, status.Status, status.TaskchainID)

		timer := time.NewTimer(healthCheckPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return last, fmt.Errorf("%w while waiting for health check job %q of exocompute configuration %s",
				ctx.Err(), jobID, configID)
		case <-timer.C:
		}
	}
}
