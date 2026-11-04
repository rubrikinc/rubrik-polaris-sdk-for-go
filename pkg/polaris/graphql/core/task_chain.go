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

package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

const (
	// The number of attempts before failing to wait for the job when the error
	// returned is a 403, objects not authorized.
	waitAttempts = 50
)

// TaskChainState represents the state of an RSC task chain.
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

// TaskChain is a collection of sequential tasks that all must complete for the
// task chain to be considered complete.
type TaskChain struct {
	ID          int64          `json:"id"`
	TaskChainID uuid.UUID      `json:"taskchainUuid"`
	State       TaskChainState `json:"state"`
}

// KorgTaskChainStatus returns the task chain for the specified task chain id.
// If the task chain id refers to a task chain that was just created, its state
// might not have reached ready yet. This can be detected by state being
// TaskChainInvalid and error is nil.
func (a API) KorgTaskChainStatus(ctx context.Context, taskChainID uuid.UUID) (TaskChain, error) {
	a.log.Print(log.Trace)

	query := getKorgTaskchainStatusQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		TaskChainID uuid.UUID `json:"taskchainId,omitempty"`
	}{TaskChainID: taskChainID})
	if err != nil {
		return TaskChain{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Query struct {
				TaskChain TaskChain `json:"taskchain"`
			} `json:"getKorgTaskchainStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return TaskChain{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Query.TaskChain, nil
}

// WaitForTaskChain blocks until the RSC task chain with the specified task
// chain id has completed. When the task chain completes, the final state of the
// task chain is returned. The wait parameter specifies the amount of time to
// wait before requesting another task status update.
func (a API) WaitForTaskChain(ctx context.Context, taskChainID uuid.UUID, wait time.Duration) (TaskChainState, error) {
	a.log.Print(log.Trace)

	attempt := 0
	for {
		taskChain, err := a.KorgTaskChainStatus(ctx, taskChainID)
		if err != nil {
			var gqlErr graphql.GQLError
			if !errors.As(err, &gqlErr) || len(gqlErr.Errors) < 1 {
				return TaskChainInvalid, fmt.Errorf("failed to get tashchain status for %s: %s", taskChainID, err)
			}
			for _, e := range gqlErr.Errors {
				if e.Extensions.Code == 403 || e.Extensions.Code == 500 {
					continue // Could be a RBAC error that eventually goes away, keep retrying.
				}
				return TaskChainInvalid, fmt.Errorf("unexpected error code when getting tashchain status for %s: %v", taskChainID, err)
			}
			if attempt++; attempt > waitAttempts {
				return TaskChainInvalid, fmt.Errorf("failed to get tashchain status for %s after %d attempts: %s", taskChainID, attempt, err)
			}
			a.log.Printf(log.Debug, "RBAC not ready (attempt: %d)", attempt)
		}

		if taskChain.State == TaskChainSucceeded || taskChain.State == TaskChainCanceled || taskChain.State == TaskChainFailed {
			return taskChain.State, nil
		}

		a.log.Printf(log.Debug, "Waiting for Polaris task chain: %s", taskChainID)

		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return TaskChainInvalid, ctx.Err()
		}
	}
}

// WaitForFeatureDisableTaskChain waits for the feature disable task chain to
// finish. If an error occurs while waiting for the task chain or the task chain
// ends in a failed state, an error is returned.
func (a API) WaitForFeatureDisableTaskChain(ctx context.Context, taskChainID uuid.UUID, featureStatus func(ctx context.Context) (bool, error)) error {
	a.log.Print(log.Trace)

	for {
		// Check the status of the task chain.
		taskChain, err := a.KorgTaskChainStatus(ctx, taskChainID)
		if err != nil {
			var gqlErr graphql.GQLError
			if !errors.As(err, &gqlErr) || len(gqlErr.Errors) < 1 {
				return fmt.Errorf("failed to retrieve taskchain status: %s", err)
			}
			for _, e := range gqlErr.Errors {
				if e.Extensions.Code == 403 || e.Extensions.Code == 500 {
					continue // Could be a RBAC error that eventually goes away, keep retrying.
				}
				return fmt.Errorf("unexpected error code when getting tashchain status for %s: %v", taskChainID, err)
			}

			// If the task chain RBAC is not yet ready, we fall back to checking
			// the status of the account feature.
			if disabled, err := featureStatus(ctx); disabled || err != nil {
				return err
			}

			a.log.Printf(log.Debug, "Task chain RBAC not ready")
		} else {
			if taskChain.State == TaskChainSucceeded {
				return nil
			}
			if taskChain.State == TaskChainCanceled || taskChain.State == TaskChainFailed {
				return fmt.Errorf("taskchain failed: task chain state is %s", taskChain.State)
			}
		}

		a.log.Printf(log.Debug, "Waiting for task chain: %s", taskChainID)
		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// WaitForTaskChainWithStatus blocks until the status function reports that
// the thing produced by the RSC task chain with the specified task chain ID is
// done. The status function is called, followed by a check of the task chain
// state, every wait interval.
//
// When the task chain has ended, i.e. it has succeeded, failed or been
// canceled, the status function is called one final time, since the task
// chain can report its result just before it ends. If the status function
// still isn't done, an error is returned. If the task chain state can't be
// read because RSC rejects the request with a 403 or a 500 error, for example
// because the RBAC of the task chain isn't ready yet, the wait continues using
// only the status function.
//
// Returns the error from the status function if it fails, and ctx.Err() if the
// context is done before the status function reports done.
func (a API) WaitForTaskChainWithStatus(ctx context.Context, taskChainID uuid.UUID, wait time.Duration, status func(ctx context.Context) (bool, error)) error {
	a.log.Print(log.Trace)

	for {
		if done, err := status(ctx); done || err != nil {
			return err
		}

		taskChain, err := a.KorgTaskChainStatus(ctx, taskChainID)
		switch {
		case err != nil && ctx.Err() != nil:
			return ctx.Err()
		case err != nil && isTaskChainStatusUnavailable(err):
			a.log.Printf(log.Debug, "Task chain %s status not available: %s", taskChainID, err)
		case err != nil:
			return fmt.Errorf("failed to get task chain status for %s: %w", taskChainID, err)
		case taskChain.State == TaskChainSucceeded || taskChain.State == TaskChainFailed || taskChain.State == TaskChainCanceled:
			if done, err := status(ctx); done || err != nil {
				return err
			}
			return fmt.Errorf("task chain %s ended with state %s before the status was done", taskChainID, taskChain.State)
		}

		a.log.Printf(log.Debug, "Waiting for task chain %s (state: %q)", taskChainID, taskChain.State)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}

// isTaskChainStatusUnavailable returns true if the error is a GraphQL error
// where every error has the code 403 or 500. RSC returns these when the RBAC of
// a task chain isn't ready yet, or when the task chain's event series can't be
// found.
func isTaskChainStatusUnavailable(err error) bool {
	var gqlErr graphql.GQLError
	if !errors.As(err, &gqlErr) || len(gqlErr.Errors) == 0 {
		return false
	}
	for _, e := range gqlErr.Errors {
		if e.Extensions.Code != 403 && e.Extensions.Code != 500 {
			return false
		}
	}

	return true
}
