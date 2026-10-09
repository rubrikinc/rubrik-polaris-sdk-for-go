//go:generate go run ../queries_gen.go kms

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

// Package kms provides a low level interface to the KMS key manager GraphQL
// queries provided by the RSC platform.
//
// A KMS key manager is a stored connection profile in RSC, giving RSC access to
// a customer owned key management service. Registering one returns an RSC
// generated UUID, which together with a key identifier from that service, such
// as an AWS KMS key ID, selects the key used for client side encryption of an
// archival location.
//
// Only the cloud account provisioned managers, and the Azure managed identity
// variant, are exposed here. The static credential managers, which hold AWS
// access keys or Azure client secrets, belong to the data center archival
// workflow and are currently not part of this package.
package kms

import (
	"slices"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// Type represents the RSC GraphQL KmsType enum type.
type Type string

const (
	TypeUnspecified   Type = "KMS_TYPE_UNSPECIFIED"
	TypeAWSKMS        Type = "AWS_KMS"
	TypeAzureKeyVault Type = "AZURE_KEY_VAULT"
	TypeGCPKMS        Type = "GCP_KMS"
	TypeKMIPServer    Type = "KMIP_SERVER"
)

// TypeValues returns all KMS types which can be registered as key managers.
func TypeValues() []Type {
	return slices.Clone([]Type{TypeAWSKMS, TypeAzureKeyVault, TypeGCPKMS, TypeKMIPServer})
}

// ProvisioningMethod represents the RSC GraphQL ProvisioningMethod enum type.
// It describes how RSC obtains the credentials it uses to reach the key
// management service.
type ProvisioningMethod string

const (
	MethodUnspecified ProvisioningMethod = "PROVISION_METHOD_UNSPECIFIED"

	// MethodCloudAccount resolves credentials from an onboarded RSC cloud
	// account. This is the cloud native archival workflow.
	MethodCloudAccount ProvisioningMethod = "CLOUD_ACCOUNT_PROVISION"

	// MethodDirect uses static credentials stored in RSC. This is the data
	// center archival workflow.
	MethodDirect ProvisioningMethod = "DIRECT_PROVISION"

	// MethodSecure uses a managed identity. Azure only.
	MethodSecure ProvisioningMethod = "SECURE_PROVISION"
)

// ProvisioningMethodValues returns the provisioning methods a key manager of
// the specified cloud vendor can have. PROVISION_METHOD_UNSPECIFIED is part of
// the RSC enum but is never a valid input. An empty slice is returned for cloud
// vendors other than AWS, Azure and GCP.
func ProvisioningMethodValues(vendor core.CloudVendor) []ProvisioningMethod {
	switch vendor {
	case core.CloudVendorAWS:
		return []ProvisioningMethod{MethodCloudAccount, MethodDirect}
	case core.CloudVendorAzure:
		return []ProvisioningMethod{MethodCloudAccount, MethodDirect, MethodSecure}
	case core.CloudVendorGCP:
		return []ProvisioningMethod{MethodCloudAccount}
	default:
		return []ProvisioningMethod{}
	}
}
