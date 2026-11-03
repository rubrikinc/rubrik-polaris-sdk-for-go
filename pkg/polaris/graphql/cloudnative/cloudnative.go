//go:generate go run ../queries_gen.go cloudnative

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

// Package cloudnative provides a low level interface to the cloud native
// application GraphQL queries provided by the RSC platform.
//
// A cloud native application, shown as a Cloud Application in the RSC UI,
// groups the resources of one AWS account and region, selected by tags and by
// explicitly listed ARNs, so that their configuration can be protected by an
// SLA Domain with the AWS_CONFIG_OBJECT_TYPE object type. The account must be
// onboarded with the CLOUD_NATIVE_CONFIG_PROTECTION feature.
package cloudnative

// RefreshStatus represents the RSC GraphQL CloudNativeApplicationRefreshStatus
// enum type.
type RefreshStatus string

const (
	RefreshStatusUnspecified RefreshStatus = "APP_REFRESH_STATUS_UNSPECIFIED"
	RefreshStatusRefreshing  RefreshStatus = "APP_REFRESHING"
	RefreshStatusFailed      RefreshStatus = "APP_REFRESH_FAILED"
	RefreshStatusSuccess     RefreshStatus = "APP_REFRESH_SUCCESS"
)

// ProtectionStatus represents the RSC GraphQL CloudNativeProtectionStatus enum
// type.
type ProtectionStatus string

const (
	ProtectionStatusUnspecified          ProtectionStatus = "PROTECTION_STATUS_UNSPECIFIED"
	ProtectionStatusUnprotected          ProtectionStatus = "PROTECTION_STATUS_UNPROTECTED"
	ProtectionStatusPartiallyProtected   ProtectionStatus = "PROTECTION_STATUS_PARTIALLY_PROTECTED"
	ProtectionStatusProtected            ProtectionStatus = "PROTECTION_STATUS_PROTECTED"
	ProtectionStatusProtectedWithWarning ProtectionStatus = "PROTECTION_STATUS_PROTECTED_WITH_WARNINGS"
)

// NameValidationStatus represents the RSC GraphQL
// CloudNativeApplicationNameValidationStatus enum type.
type NameValidationStatus string

const (
	NameValidationUnspecified NameValidationStatus = "APP_NAME_VALIDATION_UNSPECIFIED"
	NameValid                 NameValidationStatus = "APP_NAME_VALID"
	NameInvalidFormat         NameValidationStatus = "APP_NAME_INVALID_FORMAT"
	NameNonUnique             NameValidationStatus = "APP_NAME_NON_UNIQUE"
)
