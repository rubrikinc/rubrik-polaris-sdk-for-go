// Copyright 2024 Rubrik, Inc.
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
	"github.com/google/uuid"
)

// AWSClusterConnectionParams holds the parameters for an AWS exocompute cluster
// connection info operation.
type AWSClusterConnectionParams struct {
	ClusterName string    `json:"clusterName"`
	ConfigID    uuid.UUID `json:"exocomputeConfigId"`
}

func (p AWSClusterConnectionParams) InfoQuery() (string, any, AWSClusterConnectionResult) {
	return awsExocomputeGetClusterConnectionInfoQuery, p, AWSClusterConnectionResult{}
}

// AWSClusterConnectionResult holds the result of an AWS exocompute cluster
// connection info operation.
type AWSClusterConnectionResult struct {
	Command  string `json:"connectionCommand"`
	Manifest string `json:"clusterSetupYaml"`
}

// ConnectAWSClusterParams holds the parameters for an exocompute AWS cluster
// connect operation.
type ConnectAWSClusterParams AWSClusterConnectionParams

func (p ConnectAWSClusterParams) ConnectQuery() (string, any, ConnectAWSClusterResult) {
	return awsExocomputeClusterConnectQuery, p, ConnectAWSClusterResult{}
}

// ConnectAWSClusterResult holds the result of an AWS exocompute cluster connect
// operation.
type ConnectAWSClusterResult struct {
	ClusterID uuid.UUID `json:"clusterUuid"`
	AWSClusterConnectionResult
}

// DisconnectAWSClusterParams holds the parameters for an AWS exocompute cluster
// disconnect operation.
type DisconnectAWSClusterParams struct {
	ClusterID uuid.UUID `json:"clusterId"`
}

func (p DisconnectAWSClusterParams) DisconnectQuery() (string, any, DisconnectAWSClusterResult) {
	return disconnectAwsExocomputeClusterQuery, p, DisconnectAWSClusterResult{}
}

// DisconnectAWSClusterResult holds the result of an AWS exocompute cluster
// disconnect operation.
type DisconnectAWSClusterResult struct{}
