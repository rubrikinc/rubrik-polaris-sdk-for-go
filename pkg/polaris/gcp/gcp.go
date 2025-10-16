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

// Package gcp provides a high level interface to the GCP part of the RSC
// platform.
package gcp

import (
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for GCP project management.
type API struct {
	client *graphql.Client
	log    log.Logger
}

// Deprecated: use Wrap instead.
func NewAPI(gql *graphql.Client) API {
	return API{client: gql, log: gql.Log()}
}

// Wrap the RSC client in the azure API.
func Wrap(client *polaris.Client) API {
	return API{client: client.GQL, log: client.GQL.Log()}
}

// WrapGQL wraps the GQL client in the aws API.
func WrapGQL(gql *graphql.Client) API {
	return API{client: gql, log: gql.Log()}
}
