//go:generate go run ../queries_gen.go devops

// Copyright 2026 Rubrik, Inc.
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

// Package devops provides a low-level interface to the DevOps (Azure DevOps and
// GitHub) GraphQL queries provided by the Polaris platform. Only the Azure
// DevOps non-OAuth onboarding and read-back operations are implemented.
package devops

// HostType represents the type of exocompute host used to protect a DevOps
// organization. It maps to the schema DevopsHostType enum.
type HostType string

const (
	// HostTypeUnspecified is the zero value and should not be used.
	HostTypeUnspecified HostType = "HOST_TYPE_UNSPECIFIED"
	// HostTypeCustomer denotes customer-hosted exocompute.
	HostTypeCustomer HostType = "CUSTOMER_HOST"
	// HostTypeRubrik denotes Rubrik-hosted exocompute.
	HostTypeRubrik HostType = "RUBRIK_HOST"
)

// StorageType represents the backup storage type used for a DevOps
// organization. It maps to the schema DevOpsStorageType enum.
type StorageType string

const (
	// StorageTypeUnspecified is the zero value and should not be used.
	StorageTypeUnspecified StorageType = "STORAGE_TYPE_UNSPECIFIED"
	// StorageTypeBYOS denotes customer-hosted storage (bring your own storage).
	StorageTypeBYOS StorageType = "BYOS"
	// StorageTypeRCV denotes Rubrik Cloud Vault storage.
	StorageTypeRCV StorageType = "RCV"
)

// OrgType represents the type of a DevOps organization. It maps to the schema
// DevopsOrgType enum.
type OrgType string

const (
	// OrgTypeUnspecified is the zero value and should not be used.
	OrgTypeUnspecified OrgType = "DEVOPS_ORG_TYPE_UNSPECIFIED"
	// OrgTypeAzureDevOps denotes an Azure DevOps organization.
	OrgTypeAzureDevOps OrgType = "AZURE_DEVOPS"
	// OrgTypeGitHub denotes a GitHub organization.
	OrgTypeGitHub OrgType = "GITHUB"
)

// ConnectionStatus represents the connection status of a DevOps organization.
// It maps to the schema DevopsConnectionStatus enum.
type ConnectionStatus string

const (
	// ConnectionStatusUnspecified is the default value and should not be used.
	ConnectionStatusUnspecified ConnectionStatus = "CONNECTION_STATUS_UNSPECIFIED"
	// ConnectionStatusConnected denotes an organization that is connected and
	// accessible.
	ConnectionStatusConnected ConnectionStatus = "CONNECTION_STATUS_CONNECTED"
	// ConnectionStatusDisconnected denotes an organization that is disconnected
	// or not accessible.
	ConnectionStatusDisconnected ConnectionStatus = "CONNECTION_STATUS_DISCONNECTED"
	// ConnectionStatusMissingPermissions denotes an organization with missing
	// permissions.
	ConnectionStatusMissingPermissions ConnectionStatus = "CONNECTION_STATUS_MISSING_PERMISSIONS"
	// ConnectionStatusConnecting denotes an organization that has been added but
	// is still being set up.
	ConnectionStatusConnecting ConnectionStatus = "CONNECTION_STATUS_CONNECTING"
)

// AuthMechanism represents the authentication mechanism a DevOps organization's
// tenant was onboarded with. It maps to the schema DevopsAuthMechanism enum.
type AuthMechanism string

const (
	// AuthMechanismUnspecified denotes that the mechanism could not be
	// determined.
	AuthMechanismUnspecified AuthMechanism = "DEVOPS_AUTH_MECHANISM_UNSPECIFIED"
	// AuthMechanismOAuth denotes onboarding via OAuth using Rubrik's
	// multi-tenant application.
	AuthMechanismOAuth AuthMechanism = "DEVOPS_AUTH_MECHANISM_OAUTH"
	// AuthMechanismNonOAuth denotes onboarding via non-OAuth using a per-tenant
	// customer-supplied application.
	AuthMechanismNonOAuth AuthMechanism = "DEVOPS_AUTH_MECHANISM_NON_OAUTH"
)
