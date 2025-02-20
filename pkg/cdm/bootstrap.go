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

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

const (
	defaultWait = 30 * time.Second
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

// StorageConfig holds information about the kind of Rubrik cluster to
// bootstrap.
type StorageConfig interface {
	isCloudStorageConfig()
}

// CDMStorageConfig is used to boostrap a physical Rubrik cluster.
type CDMStorageConfig struct {
	EnableEncryption bool
}

func (c CDMStorageConfig) isCloudStorageConfig() {}

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

// ClusterConfig holds the configuration for the cluster.
//
// The kind of cluster to bootstrap is determined by the storageConfig
// parameter. Setting storageConfig to CDMStorageConfig starts the bootstrap
// process for a physical Rubrik Cluster, setting it to AWSStorageConfig starts
// the process for a Rubrik Cloud Cluster Elastic Storage (CCES) on AWS and
// setting it to AzureStorageConfig starts the process for a Rubrik Cloud
// Cluster Elastic Storage (CCES) on Azure.
type ClusterConfig struct {
	ClusterName          string
	ClusterNodes         []NodeConfig
	ManagementGateway    string
	ManagementSubnetMask string
	AdminEmail           string
	AdminPassword        string
	DNSServers           []string
	DNSSearchDomains     []string
	NTPServers           []NTPServerConfig
	StorageConfig        StorageConfig
}

// BootstrapClient is used to make bootstrap API calls to the CDM platform.
type BootstrapClient struct {
	*client
	Log log.Logger
}

// NewBootstrapClient creates a new bootstrap client from the provided Rubrik
// cluster credentials.
func NewBootstrapClient(allowInsecureTLS bool) *BootstrapClient {
	return NewBootstrapClientWithLogger(allowInsecureTLS, log.DiscardLogger{})
}

// NewBootstrapClientWithLogger creates a new bootstrap client from the provided
// Rubrik cluster credentials.
func NewBootstrapClientWithLogger(allowInsecureTLS bool, logger log.Logger) *BootstrapClient {
	return &BootstrapClient{
		client: newClient(allowInsecureTLS),
		Log:    logger,
	}
}

// BootstrapCluster starts the bootstrap process for a Rubrik cluster and
// returns the bootstrap request ID. To wait for the bootstrap process to
// finish, pass the bootstrap request ID to the WaitForBootstrap function.
//
// The cluster can be rebooted at any time when this function runs, the timeout
// parameter controls how long we wait for the cluster to become responsive
// again.
//
// Bootstrapping a Rubrik cluster requires a single node to have its management
// interface configured.
func (c *BootstrapClient) BootstrapCluster(ctx context.Context, nodeIP string, config ClusterConfig, timeout time.Duration) (int, error) {
	c.Log.Print(log.Trace)

	ok, err := c.IsBootstrapped(ctx, nodeIP, timeout)
	if err != nil {
		return 0, fmt.Errorf("failed to check cluster bootstrap status: %s", err)
	}
	if ok {
		return 0, errors.New("cluster is already bootstrapped")
	}

	// Encryption can only be enabled on physical Rubrik clusters.
	var enableEncryption bool
	if cdmConfig, ok := config.StorageConfig.(CDMStorageConfig); ok {
		enableEncryption = cdmConfig.EnableEncryption
	}

	// Transform cluster configuration.
	bootstrapConfig := struct {
		Name          string                `json:"name"`
		Encryption    bool                  `json:"enableSoftwareEncryptionAtRest"`
		Admin         admin                 `json:"adminUserInfo"`
		NameServers   []string              `json:"dnsNameservers"`
		SearchDomains []string              `json:"dnsSearchDomains"`
		NTPServers    []NTPServerConfig     `json:"ntpServerConfigs"`
		StorageConfig StorageConfig         `json:"cloudStorageLocation,omitempty"`
		Nodes         map[string]nodeConfig `json:"nodeConfigs"`
	}{
		Name:       config.ClusterName,
		Encryption: enableEncryption,
		Admin: admin{
			ID:       "admin",
			Email:    config.AdminEmail,
			Password: config.AdminPassword,
		},
		NameServers:   config.DNSServers,
		SearchDomains: config.DNSSearchDomains,
		NTPServers:    config.NTPServers,
		StorageConfig: wrapCloudStorageProvider(config.StorageConfig),
		Nodes:         make(map[string]nodeConfig, len(config.ClusterNodes)),
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

	endpoint := "/cluster/me/bootstrap"
	buf, code, err := c.post(ctx, nodeIP, endpoint, Internal, bootstrapConfig)
	if err != nil {
		return 0, fmt.Errorf("failed POST request %q: %s", endpoint, err)
	}

	// We unmarshal the response from the bootstrap request early to capture
	// potential error messages. Unmarshal errors are ignored at this time as
	// certain responses could contain malformed JSON object.
	var bootstrap struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	}
	jsonErr := json.Unmarshal(buf, &bootstrap)

	if code != 202 {
		msg := fmt.Sprintf("%s (%d)", http.StatusText(code), code)
		if bootstrap.Status != "" {
			msg = fmt.Sprintf("%s: %s", msg, bootstrap.Status)
		}

		return 0, fmt.Errorf("failed POST request %q: %s", endpoint, msg)
	}

	// If the bootstrap request was successful the response should contain a
	// valid JSON object.
	if jsonErr != nil {
		return 0, fmt.Errorf("failed to unmarshal bootstrap status: %s", jsonErr)
	}

	return bootstrap.ID, nil
}

// IsBootstrapped returns true if the cluster has been bootstrapped, false
// otherwise. The cluster can be rebooted at any time when this function runs,
// the timeout parameter controls how long we wait for the cluster to become
// responsive again.
func (c *BootstrapClient) IsBootstrapped(ctx context.Context, nodeIP string, timeout time.Duration) (bool, error) {
	c.Log.Print(log.Trace)

	var failure time.Time
	for {
		buf, code, err := c.get(ctx, nodeIP, "/node_management/is_bootstrapped", Internal)
		if err == nil && code == 200 {
			var isBootstrapped struct {
				Value bool `json:"value"`
			}
			if err := json.Unmarshal(buf, &isBootstrapped); err != nil {
				return false, fmt.Errorf("failed to unmarshal bootstrap status: %s", err)
			}

			return isBootstrapped.Value, nil
		}

		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		if err == nil {
			err = errors.New(http.StatusText(code))
		}
		if failure.IsZero() {
			failure = time.Now()
		}
		if time.Since(failure) > timeout {
			return false, err
		}

		c.Log.Printf(log.Debug, "Request returned: %s, retrying", err)
		time.Sleep(defaultWait)
	}
}

// WaitForBootstrap blocks until the bootstrapping of the cluster succeeds or
// fails. The cluster can be rebooted at any time when this function runs, the
// timeout parameter controls how long we wait for the cluster to become
// responsive again.
func (c *BootstrapClient) WaitForBootstrap(ctx context.Context, nodeIP string, requestID int, timeout time.Duration) error {
	c.Log.Print(log.Trace)

	var failure time.Time
	for {
		res, code, err := c.get(ctx, nodeIP, fmt.Sprintf("/cluster/me/bootstrap?request_id=%d", requestID), Internal)
		if err == nil && code == 200 {
			failure = time.Time{}

			var bootstrap struct {
				Status  string `json:"status"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(res, &bootstrap); err != nil {
				return fmt.Errorf("failed to unmarshal bootstrap status: %s", err)
			}
			switch bootstrap.Status {
			case "IN_PROGRESS":
				c.Log.Print(log.Debug, "Bootstrap in progress")
				time.Sleep(defaultWait)
				continue
			case "FAILURE", "FAILED":
				return fmt.Errorf("bootstrap failed: %s", bootstrap.Message)
			default:
				return nil
			}
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err == nil {
			err = errors.New(http.StatusText(code))
		}
		if failure.IsZero() {
			failure = time.Now()
		}
		if time.Since(failure) > timeout {
			return err
		}

		c.Log.Printf(log.Debug, "Request returned: %s, retrying", err)
		time.Sleep(defaultWait)
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

// wrapCloudStorageProvider wraps the CloudStorageLocation in a cloud storage
// provider, as expected by CDM.
func wrapCloudStorageProvider(config StorageConfig) StorageConfig {
	switch config.(type) {
	case AWSStorageConfig, *AWSStorageConfig:
		return struct {
			StorageConfig `json:"awsStorageConfig"`
		}{StorageConfig: config}
	case AzureStorageConfig, *AzureStorageConfig:
		return struct {
			StorageConfig `json:"azureStorageConfig"`
		}{StorageConfig: config}
	default:
		return nil
	}
}

func (c *BootstrapClient) get(ctx context.Context, nodeIP, endpoint string, version APIVersion) ([]byte, int, error) {
	c.Log.Print(log.Trace)

	req, err := c.request(ctx, http.MethodGet, nodeIP, version, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(req)
}

func (c *BootstrapClient) post(ctx context.Context, nodeIP, endpoint string, version APIVersion, payload any) ([]byte, int, error) {
	c.Log.Print(log.Trace)

	req, err := c.request(ctx, http.MethodPost, nodeIP, version, endpoint, payload)
	if err != nil {
		return nil, 0, err
	}

	return c.doRequest(req)
}
