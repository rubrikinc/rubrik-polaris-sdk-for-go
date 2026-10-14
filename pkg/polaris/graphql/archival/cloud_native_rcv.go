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

package archival

import "slices"

// RCVTier represents the RSC GraphQL RcsTierEnumType enum type.
type RCVTier string

const (
	RCVTierBackup   RCVTier = "BACKUP"
	RCVTierArchive  RCVTier = "ARCHIVE"
	RCVTierRecovery RCVTier = "RECOVERY"
)

// RCVTierValues returns the tiers RCV archival locations support. RECOVERY is
// part of the RSC enum but the create API documents Backup and Archive only.
func RCVTierValues() []RCVTier {
	return slices.Clone([]RCVTier{RCVTierBackup, RCVTierArchive})
}

// RCVRedundancy represents the RSC GraphQL RcvRedundancy enum type.
type RCVRedundancy string

const (
	RCVRedundancyUnknown     RCVRedundancy = "REDUNDANCY_UNKNOWN"
	RCVRedundancySingleZone  RCVRedundancy = "SINGLE_ZONE"
	RCVRedundancyMultiZone   RCVRedundancy = "MULTI_ZONE"
	RCVRedundancyMultiRegion RCVRedundancy = "MULTI_REGION"
)

// RCVRedundancyValues returns the redundancy levels which can be requested when
// creating an archival location. REDUNDANCY_UNKNOWN is only ever read back.
func RCVRedundancyValues() []RCVRedundancy {
	return slices.Clone([]RCVRedundancy{
		RCVRedundancySingleZone,
		RCVRedundancyMultiZone,
		RCVRedundancyMultiRegion,
	})
}

// SourceWorkloadCloud represents the RSC GraphQL SourceWorkloadCloud enum type.
// It is the cloud the archived workloads originate from, which is not
// necessarily the cloud hosting the vault.
type SourceWorkloadCloud string

const (
	SourceWorkloadCloudAWS   SourceWorkloadCloud = "SOURCE_AWS"
	SourceWorkloadCloudAzure SourceWorkloadCloud = "SOURCE_AZURE"
	SourceWorkloadCloudGCP   SourceWorkloadCloud = "SOURCE_GCP"
)

// SourceWorkloadCloudValues returns all source workload clouds.
func SourceWorkloadCloudValues() []SourceWorkloadCloud {
	return slices.Clone([]SourceWorkloadCloud{
		SourceWorkloadCloudAWS,
		SourceWorkloadCloudAzure,
		SourceWorkloadCloudGCP,
	})
}

// LocTemplateType represents the RSC GraphQL CloudNativeLocTemplateType enum
// type. It decides whether an archival location is pinned to a specific region
// or follows the region of the workload being archived.
type LocTemplateType string

const (
	LocTemplateInvalid        LocTemplateType = "INVALID"
	LocTemplateSourceRegion   LocTemplateType = "SOURCE_REGION"
	LocTemplateSpecificRegion LocTemplateType = "SPECIFIC_REGION"
)

// LocTemplateTypeValues returns the location templates an archival location can
// be created with. INVALID is only ever read back.
func LocTemplateTypeValues() []LocTemplateType {
	return slices.Clone([]LocTemplateType{LocTemplateSourceRegion, LocTemplateSpecificRegion})
}

// TargetEncryptionType represents the RSC GraphQL TargetEncryptionTypeEnum enum
// type. It is read back from an archival location and reports which encryption
// mode RSC applied. It covers every archival target type, so most of the values
// never appear on an RCV archival location.
type TargetEncryptionType string

const (
	TargetEncryptionUnknown           TargetEncryptionType = "UNKNOWN_ENCRYPTION_TYPE"
	TargetEncryptionRSAKeyBased       TargetEncryptionType = "RSA_KEY_BASED"
	TargetEncryptionKMSMasterKeyBased TargetEncryptionType = "KMS_MASTER_KEY_BASED"
	TargetEncryptionPasswordBased     TargetEncryptionType = "ENCRYPTION_PASSWORD_BASED"

	// TargetEncryptionUEKMBased has been deprecated by RSC in favour of the
	// two UEKM values below, but it's still read back from archival locations
	// created before the split.
	TargetEncryptionUEKMBased TargetEncryptionType = "UNIFIED_ENCRYPTION_KEY_MGMT_BASED"

	TargetEncryptionUEKMRSABased       TargetEncryptionType = "UEKM_RSA_BASED"
	TargetEncryptionUEKMAWSKMSBased    TargetEncryptionType = "UEKM_AWS_KMS_BASED"
	TargetEncryptionUEKMAKVBased       TargetEncryptionType = "UEKM_AKV_BASED"
	TargetEncryptionUEKMGCPKMSBased    TargetEncryptionType = "UEKM_GCP_KMS_BASED"
	TargetEncryptionSSEDefaultPMK      TargetEncryptionType = "SSE_DEFAULT_PMK"
	TargetEncryptionSSECustomerManaged TargetEncryptionType = "SSE_CMK"
	TargetEncryptionSSECustomerKey     TargetEncryptionType = "SSE_CPK"
)

// RCVTargetEncryptionTypeValues returns the encryption types an RCV archival
// location reports: the platform managed default, the client-side customer
// managed key modes, and server-side encryption with a customer provided key.
func RCVTargetEncryptionTypeValues() []TargetEncryptionType {
	return slices.Clone([]TargetEncryptionType{
		TargetEncryptionSSEDefaultPMK,
		TargetEncryptionUEKMAWSKMSBased,
		TargetEncryptionUEKMAKVBased,
		TargetEncryptionUEKMGCPKMSBased,
		TargetEncryptionUEKMRSABased,
		TargetEncryptionSSECustomerKey,
	})
}
