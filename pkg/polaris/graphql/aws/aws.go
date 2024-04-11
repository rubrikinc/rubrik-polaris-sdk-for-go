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

// Package aws provides a low level interface to the AWS GraphQL queries
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
)

// Region represents an AWS region in Polaris.
type Region string

const (
	RegionUnknown      Region = "UNKNOWN_AWS_REGION"
	RegionAfSouth1     Region = "AF_SOUTH_1"
	RegionApEast1      Region = "AP_EAST_1"
	RegionApNorthEast1 Region = "AP_NORTHEAST_1"
	RegionApNorthEast2 Region = "AP_NORTHEAST_2"
	RegionApNorthEast3 Region = "AP_NORTHEAST_3"
	RegionApSouthEast1 Region = "AP_SOUTHEAST_1"
	RegionApSouthEast2 Region = "AP_SOUTHEAST_2"
	RegionApSouthEast3 Region = "AP_SOUTHEAST_3"
	RegionApSouth1     Region = "AP_SOUTH_1"
	RegionCaCentral1   Region = "CA_CENTRAL_1"
	RegionCnNorthWest1 Region = "CN_NORTHWEST_1"
	RegionCnNorth1     Region = "CN_NORTH_1"
	RegionEuCentral1   Region = "EU_CENTRAL_1"
	RegionEuCentral2   Region = "EU_CENTRAL_2"
	RegionEuNorth1     Region = "EU_NORTH_1"
	RegionEuSouth1     Region = "EU_SOUTH_1"
	RegionEuWest1      Region = "EU_WEST_1"
	RegionEuWest2      Region = "EU_WEST_2"
	RegionEuWest3      Region = "EU_WEST_3"
	RegionMeSouth1     Region = "ME_SOUTH_1"
	RegionSaEast1      Region = "SA_EAST_1"
	RegionUsEast1      Region = "US_EAST_1"
	RegionUsEast2      Region = "US_EAST_2"
	RegionUsGovEast1   Region = "US_GOV_EAST_1"
	RegionUsGovWest1   Region = "US_GOV_WEST_1"
	RegionUsWest1      Region = "US_WEST_1"
	RegionUsWest2      Region = "US_WEST_2"
)

// FormatRegion returns the Region as a string formatted in AWS's style, i.e.
// lower case and with hyphen as a separator.
func FormatRegion(region Region) string {
	return strings.ReplaceAll(strings.ToLower(string(region)), "_", "-")
}

// FormatRegions returns the Regions as a slice of strings formatted in AWS's
// style, i.e. lower case and with hyphen as a separator.
func FormatRegions(regions []Region) []string {
	regs := make([]string, 0, len(regions))
	for _, region := range regions {
		regs = append(regs, FormatRegion(region))
	}

	return regs
}

var validRegions = map[Region]struct{}{
	RegionAfSouth1:     {},
	RegionApEast1:      {},
	RegionApNorthEast1: {},
	RegionApNorthEast2: {},
	RegionApNorthEast3: {},
	RegionApSouthEast1: {},
	RegionApSouthEast2: {},
	RegionApSouthEast3: {},
	RegionApSouth1:     {},
	RegionCaCentral1:   {},
	RegionCnNorthWest1: {},
	RegionCnNorth1:     {},
	RegionEuCentral1:   {},
	RegionEuCentral2:   {},
	RegionEuNorth1:     {},
	RegionEuSouth1:     {},
	RegionEuWest1:      {},
	RegionEuWest2:      {},
	RegionEuWest3:      {},
	RegionMeSouth1:     {},
	RegionSaEast1:      {},
	RegionUsEast1:      {},
	RegionUsEast2:      {},
	RegionUsGovEast1:   {},
	RegionUsGovWest1:   {},
	RegionUsWest1:      {},
	RegionUsWest2:      {},
}

// ParseRegion returns the Region matching the given region. Accepts both
// Polaris and AWS style region names.
func ParseRegion(region string) (Region, error) {
	// Polaris region name.
	r := Region(region)
	if _, ok := validRegions[r]; ok {
		return r, nil
	}

	// AWS region name.
	r = Region(strings.ReplaceAll(strings.ToUpper(region), "-", "_"))
	if _, ok := validRegions[r]; ok {
		return r, nil
	}

	return RegionUnknown, fmt.Errorf("invalid aws region: %s", region)
}

// ParseRegions returns the Regions matching the given regions. Accepts both
// Polaris and AWS style region names.
func ParseRegions(regions []string) ([]Region, error) {
	regs := make([]Region, 0, len(regions))

	for _, r := range regions {
		region, err := ParseRegion(r)
		if err != nil {
			return nil, err
		}

		regs = append(regs, region)
	}

	return regs, nil
}

// API wraps around GraphQL clients to give them the RCS AWS API.
type API struct {
	Version string // Deprecated: use GQL.DeploymentVersion
	GQL     *graphql.Client
	log     log.Logger
}

// Wrap the GraphQL client in the AWS API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}
