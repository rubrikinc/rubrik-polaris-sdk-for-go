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

package access

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// IdentityProvider represents an identity provider configured in RSC.
type IdentityProvider struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	EntityID             string    `json:"entityId"`
	SignInURL            string    `json:"signInUrl"`
	SPInitiatedSignInURL string    `json:"spInitiatedSignInUrl"`
	SPInitiatedTestURL   string    `json:"spInitiatedTestUrl"`
	SignOutURL           string    `json:"signOutUrl"`
	Expiration           time.Time `json:"expirationDate"`
	SigningCertificate   string    `json:"signingCertificate"`
	MetadataJSON         string    `json:"metadataJson"`
	Default              bool      `json:"isDefault"`
	AuthorizedGroups     int       `json:"authorizedGroupsCount"`
	ActiveUsers          int       `json:"activeUserCount"`
	ClaimAttributes      []struct {
		Name string `json:"name"`
		Type string `json:"attributeType"`
	} `json:"idpClaimAttributes"`
}

// ListIdentityProviders returns all identity providers for the current
// organization.
func ListIdentityProviders(ctx context.Context, gql *graphql.Client) ([]IdentityProvider, error) {
	gql.Log().Print(log.Trace)

	query := allCurrentOrgIdentityProvidersQuery
	buf, err := gql.Request(ctx, query, struct{}{})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []IdentityProvider `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}
