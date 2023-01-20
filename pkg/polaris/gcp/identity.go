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
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

type identityKind int

const (
	internalID identityKind = iota
	externalID
	externalNumber
)

type identity struct {
	id   string
	kind identityKind
}

// IdentityFunc returns a project identity initialized from the values passed
// to the function creating the IdentityFunc.
type IdentityFunc func(ctx context.Context) (identity, error)

// CloudAccountID returns an IdentityFunc that initializes the identity with
// the specified Polaris cloud account id.
func CloudAccountID(cloudAccountID uuid.UUID) IdentityFunc {
	return func(ctx context.Context) (identity, error) {
		return identity{id: cloudAccountID.String(), kind: internalID}, nil
	}
}

// ID returns an IdentityFunc that initializes the identity with the id of the
// specified project.
func ID(project ProjectFunc) IdentityFunc {
	return func(ctx context.Context) (identity, error) {
		config, err := project(ctx)
		if err != nil {
			return identity{}, fmt.Errorf("failed to lookup project: %v", err)
		}

		return identity{id: config.id, kind: externalID}, nil
	}
}

// ProjectID returns an IdentityFunc that initializes the identity with the
// specified project id.
func ProjectID(projectID string) IdentityFunc {
	return func(ctx context.Context) (identity, error) {
		if len(projectID) < 6 || len(projectID) > 30 {
			return identity{}, errors.New("invalid GCP project id")
		}

		return identity{id: projectID, kind: externalID}, nil
	}
}

// ProjectNumber returns an IdentityFunc that initializes the identity with the
// specified project number.
func ProjectNumber(number int64) IdentityFunc {
	return func(ctx context.Context) (identity, error) {
		return identity{id: strconv.FormatInt(number, 10), kind: externalNumber}, nil
	}
}
