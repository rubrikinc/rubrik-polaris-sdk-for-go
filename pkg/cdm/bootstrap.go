// Copyright 2023 Rubrik, Inc.
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

package cdm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultWaitTime = 10 * time.Second
)

// NodeConfig holds node configuration for the cluster.
type NodeConfig struct {
	Name         string
	ManagementIP string
}

// NTPSymmetricKey holds the NTP servers symmetric key.
type NTPSymmetricKey struct {
	KeyID   int    `json:"keyId"`
	Key     string `json:"key"`
	KeyType string `json:"keyType"`
}

// NTPServerConfig holds NTP server configuration for the bootstrapped cluster.
type NTPServerConfig struct {
	Server       string           `json:"server"`
	SymmetricKey *NTPSymmetricKey `json:"symmetricKey,omitempty"`
}

// CloudStorageLocation holds information about the kind of Rubrik cluster to
// bootstrap.
type CloudStorageLocation interface {
	isCloudStorageConfig()
}

// AzureStorageConfig is used to bootstrap a Rubrik Cloud Cluster Elastic
// Storage (CCES) on Azure.
type AzureStorageConfig struct {
	ConnectionString   string `json:"connectionString"`
	ContainerName      string `json:"containerName"`
	EnableImmutability bool   `json:"isVersionLevelImmutabilitySupported"`
}

func (c AzureStorageConfig) isCloudStorageConfig() {}

// AWSStorageConfig is used to bootstrap a Rubrik Cloud Cluster Elastic Storage
// (CCES) on AWS.
type AWSStorageConfig struct {
	BucketName         string `json:"bucketName"`
	EnableImmutability bool   `json:"isObjectLockingEnabled"`
}

func (c AWSStorageConfig) isCloudStorageConfig() {}

// ClusterConfig
// The kind of cluster to bootstrap is determined by the storageConfig
// parameter. Setting storageConfig to nil starts the bootstrap process for a
// Rubrik Cluster, setting it to AzureStorageConfig starts the process for a
// Rubrik Cloud Cluster Elastic Storage (CCES) on Azure and setting it to
// AWSStorageConfig starts the process for a Rubrik Cloud Cluster Elastic
// Storage (CCES) on AWS.
type ClusterConfig struct {
	ClusterName          string
	ClusterNodes         []NodeConfig
	ManagementGateway    string
	ManagementSubnetMask string
	AdminEmail           string
	AdminPassword        string
	EnableEncryption     bool
	DNSServers           []string
	DNSSearchDomains     []string
	NTPServers           []NTPServerConfig
	CloudStorageLocation CloudStorageLocation
}

// BootstrapClient
type BootstrapClient struct {
	*client
}

// NewBootstrapClient
func NewBootstrapClient(allowInsecureTLS bool) *BootstrapClient {
	return &BootstrapClient{
		client: newClient(allowInsecureTLS),
	}
}

// BootstrapCluster starts the bootstrap process for a Rubrik cluster and
// returns the bootstrap request ID. To wait for the bootstrap process to
// finish, pass the bootstrap request ID to the WaitForBootstrap function.
//
// Bootstrapping a Rubrik cluster requires a single node to have its management
// interface configured.
func (c *BootstrapClient) BootstrapCluster(ctx context.Context, nodeIP string, config ClusterConfig) (int, error) {
	if ok, err := c.IsBootstrapped(ctx, nodeIP); ok || err != nil {
		return 0, err
	}

	// Transform cluster configuration.
	bootstrapConfig := struct {
		Name          string                `json:"name"`
		Encryption    bool                  `json:"enableSoftwareEncryptionAtRest"`
		Admin         admin                 `json:"adminUserInfo"`
		NameServers   []string              `json:"dnsNameservers"`
		SearchDomains []string              `json:"dnsSearchDomains"`
		NTPServers    []NTPServerConfig     `json:"ntpServerConfigs"`
		StorageConfig CloudStorageLocation  `json:"cloudStorageLocation,omitempty"`
		Nodes         map[string]nodeConfig `json:"nodeConfigs"`
	}{
		Name:       config.ClusterName,
		Encryption: config.EnableEncryption,
		Admin: admin{
			ID:       "admin",
			Email:    config.AdminEmail,
			Password: config.AdminPassword,
		},
		NameServers:   config.DNSServers,
		SearchDomains: config.DNSSearchDomains,
		NTPServers:    config.NTPServers,
		StorageConfig: config.CloudStorageLocation,
	}
	for _, node := range config.ClusterNodes {
		bootstrapConfig.Nodes[node.Name] = nodeConfig{
			ManagementIPConfig: managementIPConfig{
				Address: node.ManagementIP,
				Netmask: config.ManagementSubnetMask,
				Gateway: config.ManagementGateway,
			},
		}
	}

	// Start cluster bootstrap.
	endpoint := "/cluster/me/bootstrap"
	buf, code, err := c.post(ctx, nodeIP, endpoint, Internal, bootstrapConfig)
	if err != nil {
		return 0, fmt.Errorf("failed POST request %q: %s", endpoint, err)
	}
	if code != 202 {
		return 0, fmt.Errorf("failed POST request %q: %s", endpoint, http.StatusText(code))
	}

	var bootstrap struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(buf, &bootstrap); err != nil {
		return 0, fmt.Errorf("failed to unmarshal bootstrap status: %s", err)
	}

	return bootstrap.ID, nil
}

// IsBootstrapped returns true if the cluster has been bootstrapped, false
// otherwise.
func (c *BootstrapClient) IsBootstrapped(ctx context.Context, nodeIP string) (bool, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 4*time.Minute)
	defer cancel()

	for {
		endpoint := "/node_management/is_bootstrapped"
		buf, code, err := c.get(reqCtx, nodeIP, endpoint, Internal)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return false, ctx.Err()
			}
			time.Sleep(defaultWaitTime)
			continue
		}
		if code != 200 {
			return false, fmt.Errorf("failed GET request %q: %s", endpoint, http.StatusText(code))
		}

		var isBootstrapped struct {
			Value bool `json:"value"`
		}
		if err := json.Unmarshal(buf, &isBootstrapped); err != nil {
			return false, fmt.Errorf("failed to unmarshal bootstrap status: %s", err)
		}
		return isBootstrapped.Value, nil
	}
}

// WaitForBootstrap blocks until the bootstrapping of the cluster succeeds or
// fails.
func (c *BootstrapClient) WaitForBootstrap(ctx context.Context, nodeIP string, requestID int) error {
	for {
		endpoint := "/cluster/me/bootstrap"
		res, code, err := c.get(ctx, nodeIP, fmt.Sprintf("%s?request_id=%d", endpoint, requestID), Internal)
		if err != nil {
			return fmt.Errorf("failed GET request %q: %s", endpoint, err)
		}
		if code != 200 {
			return fmt.Errorf("failed GET request %q: %s", endpoint, http.StatusText(code))
		}

		var bootstrap struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(res, &bootstrap); err != nil {
			return fmt.Errorf("failed to unmarshal bootstrap status: %s", err)
		}
		switch bootstrap.Status {
		case "IN_PROGRESS":
			time.Sleep(30 * time.Second)
		case "FAILURE", "FAILED":
			return fmt.Errorf("bootstrap failed: %s", bootstrap.Message)
		default:
			return nil
		}
	}
}

type admin struct {
	ID       string `json:"id"`
	Email    string `json:"emailAddress"`
	Password string `json:"password"`
}

type managementIPConfig struct {
	Address string `json:"address"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
}

type nodeConfig struct {
	ManagementIPConfig managementIPConfig `json:"managementIpConfig"`
}

func (c *BootstrapClient) get(ctx context.Context, nodeIP, endpoint string, version APIVersion) ([]byte, int, error) {
	req, err := c.request(ctx, http.MethodGet, nodeIP, version, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(req)
}

func (c *BootstrapClient) post(ctx context.Context, nodeIP, endpoint string, version APIVersion, payload any) ([]byte, int, error) {
	req, err := c.request(ctx, http.MethodPost, nodeIP, version, endpoint, payload)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(req)
}
