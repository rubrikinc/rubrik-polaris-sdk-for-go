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

package kms

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CreateGCPKeyManagerParams holds the parameters for creating a cloud account
// provisioned GCP Cloud KMS key manager. RSC resolves the service account
// credentials from the cloud account at call time, so no credentials are
// stored.
type CreateGCPKeyManagerParams struct {
	// CloudAccountID is the RSC cloud account ID of the GCP project holding the
	// key ring. The project must be onboarded with the
	// CLOUD_NATIVE_UEM_KEY_MANAGEMENT feature in a connected state.
	CloudAccountID string `json:"cloudAccountId"`

	// Name is the name of the key manager, unique across all key managers.
	Name string `json:"instanceName"`

	// KeyRing is the short name of the Cloud KMS key ring, not the fully
	// qualified resource path. RSC verifies that the key ring exists.
	KeyRing string `json:"keyRing"`

	// Location is the Cloud KMS location of the key ring.
	Location gcp.KMSLocationEnum `json:"location"`

	Description string `json:"description,omitempty"`
}

// CreateGCPKeyManager creates a cloud account provisioned GCP Cloud KMS key
// manager and returns its ID. The cloud account, the key ring and the location
// cannot be changed after creation.
func CreateGCPKeyManager(ctx context.Context, gql *graphql.Client, params CreateGCPKeyManagerParams) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	query := addGcpKmsWithCloudAccountQuery
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				KeyManagerID uuid.UUID `json:"kmsKeyVaultId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}

	if payload.Data.Result.KeyManagerID == uuid.Nil {
		return uuid.Nil, graphql.ResponseError(query, fmt.Errorf("no key manager ID returned"))
	}

	return payload.Data.Result.KeyManagerID, nil
}

// UpdateGCPKeyManagerParams holds the parameters for updating a cloud account
// provisioned GCP Cloud KMS key manager. An empty field leaves the current
// value unchanged.
type UpdateGCPKeyManagerParams struct {
	Name        string `json:"instanceName,omitempty"`
	Description string `json:"description,omitempty"`
}

// UpdateGCPKeyManager updates the cloud account provisioned GCP Cloud KMS key
// manager with the specified ID.
func UpdateGCPKeyManager(ctx context.Context, gql *graphql.Client, keyManagerID uuid.UUID, params UpdateGCPKeyManagerParams) error {
	gql.Log().Print(log.Trace)

	query := updateGcpKmsWithCloudAccountQuery
	if _, err := gql.Request(ctx, query, struct {
		KeyManagerID uuid.UUID `json:"kmsKeyVaultId"`
		UpdateGCPKeyManagerParams
	}{KeyManagerID: keyManagerID, UpdateGCPKeyManagerParams: params}); err != nil {
		return graphql.RequestError(query, err)
	}

	return nil
}

// GCPCryptoKey represents a crypto key in the key ring of a GCP Cloud KMS key
// manager.
type GCPCryptoKey struct {
	// Name is the short crypto key name, not the fully qualified resource
	// path.
	Name string `json:"cryptoKeyName"`

	// Purpose is the purpose of the crypto key, as reported by GCP Cloud KMS.
	// Only keys with the ENCRYPT_DECRYPT purpose can be used for encryption.
	Purpose string `json:"purpose"`

	// PrimaryVersionID is the version ID of the primary version of the crypto
	// key. Empty for keys whose purpose has no primary version.
	PrimaryVersionID string `json:"primaryVersionId"`
}

// GCPCryptoKeysInKMS returns the crypto keys in the key ring of the key
// manager. Every key in the key ring is returned, including keys whose purpose
// does not allow encryption.
func GCPCryptoKeysInKMS(ctx context.Context, gql *graphql.Client, keyManagerID uuid.UUID) ([]GCPCryptoKey, error) {
	gql.Log().Print(log.Trace)

	query := allCryptoKeyIdsInGcpKmsQuery
	buf, err := gql.Request(ctx, query, struct {
		KeyManagerID uuid.UUID `json:"kmsKeyVaultId"`
	}{KeyManagerID: keyManagerID})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []GCPCryptoKey `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// GCPCryptoKeyVersion represents a single version of a crypto key in GCP Cloud
// KMS.
type GCPCryptoKeyVersion struct {
	// Version is the version ID, not the fully qualified resource path.
	Version string `json:"cryptoKeyVersion"`

	// State is the state of the version. Only ENABLED versions can be used for
	// encryption.
	State     string     `json:"state"`
	CreatedAt *time.Time `json:"createdAt"`
}

// GCPCryptoKeyVersionsInKMS returns the versions of the named crypto key in the
// key ring of the key manager. Like an Azure key, a GCP crypto key is referred
// to by name and version, so the version is part of what identifies the key to
// an archival location.
func GCPCryptoKeyVersionsInKMS(ctx context.Context, gql *graphql.Client, keyManagerID uuid.UUID, cryptoKeyName string) ([]GCPCryptoKeyVersion, error) {
	gql.Log().Print(log.Trace)

	query := allKeyVersionsForCryptoKeyInGcpKmsQuery
	buf, err := gql.Request(ctx, query, struct {
		KeyManagerID  uuid.UUID `json:"kmsKeyVaultId"`
		CryptoKeyName string    `json:"cryptoKeyName"`
	}{KeyManagerID: keyManagerID, CryptoKeyName: cryptoKeyName})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []GCPCryptoKeyVersion `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// GCPKeyRingsInProject returns the short names of the key rings in the Cloud
// KMS location of the GCP project with the specified RSC cloud account ID.
func GCPKeyRingsInProject(ctx context.Context, gql *graphql.Client, cloudAccountID uuid.UUID, location gcp.Region) ([]string, error) {
	gql.Log().Print(log.Trace)

	query := allKeyRingsInGcpProjectQuery
	buf, err := gql.Request(ctx, query, struct {
		CloudAccountID uuid.UUID           `json:"cloudAccountId"`
		Location       gcp.KMSLocationEnum `json:"location"`
	}{CloudAccountID: cloudAccountID, Location: location.ToKMSLocationEnum()})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []struct {
				KeyRing string `json:"keyRing"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	keyRings := make([]string, 0, len(payload.Data.Result))
	for _, keyRing := range payload.Data.Result {
		keyRings = append(keyRings, keyRing.KeyRing)
	}

	return keyRings, nil
}
