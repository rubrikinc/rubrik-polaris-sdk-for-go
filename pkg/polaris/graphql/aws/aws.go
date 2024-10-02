//go:generate go run ../queries_gen.go aws

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

// Package aws provides a low-level interface to the AWS GraphQL queries
// provided by the Polaris platform.
package aws

import (
	"fmt"
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Cloud represents the AWS cloud type.
type Cloud string

const (
	ChinaCloud    Cloud = "CHINA"    // Deprecated: use CloudChina.
	GovCloud      Cloud = "GOV"      // Deprecated: use CloudGov.
	StandardCloud Cloud = "STANDARD" // Deprecated: use CloudStandard.
)

const (
	CloudC2S      Cloud = "C2S"
	CloudChina    Cloud = "CHINA"
	CloudGov      Cloud = "GOV"
	CloudSC2S     Cloud = "SC2S"
	CloudStandard Cloud = "STANDARD"
)

// ParseCloud returns the Cloud matching the given cloud string.
func ParseCloud(cloud string) (Cloud, error) {
	c := Cloud(strings.ToUpper(cloud))
	switch c {
	case CloudC2S, CloudChina, CloudGov, CloudSC2S, CloudStandard:
		return c, nil
	default:
		return CloudStandard, fmt.Errorf("invalid cloud: %s", cloud)
	}
}

// ProtectionFeature represents the protection features of an AWS cloud
// account.
type ProtectionFeature string

const (
	EC2 ProtectionFeature = "EC2"
	RDS ProtectionFeature = "RDS"
	S3  ProtectionFeature = "S3"
)

// API wraps around GraphQL client to give it the RSC AWS API.
type API struct {
	Version string // Deprecated: use GQL.DeploymentVersion
	GQL     *graphql.Client
	log     log.Logger
}

// Wrap the GraphQL client in the AWS API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

// Tag represents an AWS tag.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
