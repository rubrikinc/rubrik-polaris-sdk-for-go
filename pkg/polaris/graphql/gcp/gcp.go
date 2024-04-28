//go:generate go run ../queries_gen.go gcp

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

// Package gcp provides a low level interface to the GCP GraphQL queries
// provided by the RSC platform.
package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API wraps around GraphQL clients to give them the RSC GCP API.
type API struct {
	Version string // Deprecated: use GQL.DeploymentVersion
	GQL     *graphql.Client
	log     log.Logger
}

// Wrap the GraphQL client in the GCP API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

// DefaultCredentialsServiceAccount gets the default GCP service account name.
// If no default GCP service account has been set an empty string is returned.
func (a API) DefaultCredentialsServiceAccount(ctx context.Context) (name string, err error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, gcpGetDefaultCredentialsServiceAccountQuery, nil)
	if err != nil {
		return "", fmt.Errorf("failed to request gcpGetDefaultCredentialsServiceAccount: %w", err)
	}
	a.log.Printf(log.Debug, "gcpGetDefaultCredentialsServiceAccount(): %s", string(buf))

	var payload struct {
		Data struct {
			Name string `json:"gcpGetDefaultCredentialsServiceAccount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", fmt.Errorf("failed to unmarshal gcpGetDefaultCredentialsServiceAccount: %v", err)
	}

	return payload.Data.Name, nil
}

// SetDefaultServiceAccount sets the default GCP service account. The set
// service account will be used for GCP projects added without a service
// account key file.
func (a API) SetDefaultServiceAccount(ctx context.Context, name, jwtConfig string) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.RequestWithoutLogging(ctx, gcpSetDefaultServiceAccountJwtConfigQuery, struct {
		Name      string `json:"serviceAccountName"`
		JwtConfig string `json:"serviceAccountJwtConfig"`
	}{Name: name, JwtConfig: jwtConfig})
	if err != nil {
		return fmt.Errorf("failed to request gcpSetDefaultServiceAccountJwtConfig: %w", err)
	}
	a.log.Printf(log.Debug, "gcpSetDefaultServiceAccountJwtConfig(%q, %q): %s", name, "<REDACTED>",
		string(buf))

	var payload struct {
		Data struct {
			Success bool `json:"gcpSetDefaultServiceAccountJwtConfig"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal gcpSetDefaultServiceAccountJwtConfig: %v", err)
	}
	if !payload.Data.Success {
		return errors.New("set gcp service account failed")
	}

	return nil
}
