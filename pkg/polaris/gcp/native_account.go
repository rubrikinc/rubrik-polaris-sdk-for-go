// Copyright 2025 Rubrik, Inc.
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
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gcp "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NativeProjectByID returns the native project with the specified RSC native
// project ID.
func (a API) NativeProjectByID(ctx context.Context, nativeCloudAccountID uuid.UUID) (gcp.NativeProject, error) {
	a.log.Print(log.Trace)

	natives, err := a.NativeProjects(ctx, "")
	if err != nil {
		return gcp.NativeProject{}, fmt.Errorf("failed to list native projects: %s", err)
	}
	for _, native := range natives {
		if native.ID == nativeCloudAccountID {
			return native, nil
		}
	}

	return gcp.NativeProject{}, fmt.Errorf("native project %q %w", nativeCloudAccountID, graphql.ErrNotFound)
}

// NativeProjectByCloudAccountID returns the native project with the specified
// RSC cloud account ID.
func (a API) NativeProjectByCloudAccountID(ctx context.Context, cloudAccountID uuid.UUID) (gcp.NativeProject, error) {
	a.log.Print(log.Trace)

	natives, err := a.NativeProjects(ctx, "")
	if err != nil {
		return gcp.NativeProject{}, fmt.Errorf("failed to list native projects: %s", err)
	}
	for _, native := range natives {
		if native.CloudAccountID == cloudAccountID {
			return native, nil
		}
	}

	return gcp.NativeProject{}, fmt.Errorf("native project for project %q %w", cloudAccountID, graphql.ErrNotFound)
}

// NativeProjects returns all native projects matching the specified filter.
// The filter can be used to search for a substring in project name or number.
func (a API) NativeProjects(ctx context.Context, filter string) ([]gcp.NativeProject, error) {
	return gcp.Wrap(a.client).NativeProjects(ctx, filter)
}
