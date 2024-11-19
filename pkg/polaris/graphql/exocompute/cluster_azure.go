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

import "github.com/google/uuid"

// AzureClusterConnectionParams holds the parameters for an Azure exocompute
// cluster connection info operation.
type AzureClusterConnectionParams struct {
	ClusterName string    `json:"clusterName"`
	CloudType   string    `json:"cloudType"`
	ConfigID    uuid.UUID `json:"exocomputeConfigId"`
}

func (p AzureClusterConnectionParams) InfoQuery() (string, any, AzureClusterConnectionResult) {
	return exocomputeGetClusterConnectionInfoQuery, p, AzureClusterConnectionResult{}
}

// AzureClusterConnectionResult holds the result of an Azure exocompute cluster
// connection info operation.
type AzureClusterConnectionResult struct {
	Manifest string `json:"clusterSetupYaml"`
}

// ConnectAzureClusterParams holds the parameters for an exocompute Azure
// cluster connect operation.
type ConnectAzureClusterParams AzureClusterConnectionParams

func (p ConnectAzureClusterParams) ConnectQuery() (string, any, ConnectAzureClusterResult) {
	return exocomputeClusterConnectQuery, p, ConnectAzureClusterResult{}
}

// ConnectAzureClusterResult holds the result of an Azure exocompute cluster
// connect operation.
type ConnectAzureClusterResult struct {
	ClusterID uuid.UUID `json:"clusterUuid"`
	AzureClusterConnectionResult
}

// DisconnectAzureClusterParams holds the parameters for an Azure exocompute
// cluster disconnect operation.
type DisconnectAzureClusterParams struct {
	ClusterID uuid.UUID `json:"clusterId"`
	CloudType string    `json:"cloudType"`
}

func (p DisconnectAzureClusterParams) DisconnectQuery() (string, any, DisconnectAzureClusterResult) {
	return disconnectExocomputeClusterQuery, p, DisconnectAzureClusterResult{}
}

// DisconnectAzureClusterResult holds the result of an Azure exocompute cluster
// disconnect operation.
type DisconnectAzureClusterResult struct{}
