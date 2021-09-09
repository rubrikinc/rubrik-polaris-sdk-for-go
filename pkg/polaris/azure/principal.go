// Copyright 2021 Rubrik, Inc.
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

package azure

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/google/uuid"
)

// principalv0 format is used by v0.1.x versions of the provider.
type principalV0 struct {
	AppID        string `json:"app_id"`
	AppName      string `json:"app_name"`
	AppSecret    string `json:"app_secret"`
	TenantID     string `json:"tenant_id"`
	TenantDomain string `json:"tenant_domain"`
}

// decodePrincipalV0 decodes the specified string using the v0 format. Returns
// an error if the decoder encounters unknown fields.
func decodePrincipalV0(text string) (servicePrincipal, error) {
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()

	var v0 principalV0
	if err := decoder.Decode(&v0); err != nil {
		return servicePrincipal{}, err
	}

	appID, err := uuid.Parse(v0.AppID)
	if err != nil {
		return servicePrincipal{}, err
	}

	tenantID, err := uuid.Parse(v0.TenantID)
	if err != nil {
		return servicePrincipal{}, err
	}

	principal := servicePrincipal{
		appID:        appID,
		appName:      v0.AppName,
		appSecret:    v0.AppSecret,
		tenantID:     tenantID,
		tenantDomain: v0.TenantDomain,
	}

	return principal, nil
}

// principalv1 format is used by v0.1.x versions of the provider.
type principalV1 struct {
	AppID        string `json:"appId"`
	AppName      string `json:"appName"`
	AppSecret    string `json:"appSecret"`
	TenantID     string `json:"tenantId"`
	TenantDomain string `json:"tenantDomain"`
}

// decodePrincipalV1 decodes the specified string using the v1 format. Returns
// an error if the decoder encounters unknown fields.
func decodePrincipalV1(text string) (servicePrincipal, error) {
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()

	var v1 principalV1
	if err := decoder.Decode(&v1); err != nil {
		return servicePrincipal{}, err
	}

	appID, err := uuid.Parse(v1.AppID)
	if err != nil {
		return servicePrincipal{}, err
	}

	tenantID, err := uuid.Parse(v1.TenantID)
	if err != nil {
		return servicePrincipal{}, err
	}

	principal := servicePrincipal{
		appID:        appID,
		appName:      v1.AppName,
		appSecret:    v1.AppSecret,
		tenantID:     tenantID,
		tenantDomain: v1.TenantDomain,
	}

	return principal, nil
}

// ServicePrincipal Azure service principal used by Polaris to access one or
// more Azure subscriptions.
type servicePrincipal struct {
	appID        uuid.UUID
	appName      string
	appSecret    string
	tenantID     uuid.UUID
	tenantDomain string
}

// ServicePrincipalFunc returns a service principal initialized from the values
// passed to the function creating the ServicePrincipalFunc.
type ServicePrincipalFunc func(ctx context.Context) (servicePrincipal, error)

// Default returns a ServicePrincipalFunc that initializes the service
// principal with values from the default key file.
func Default() ServicePrincipalFunc {
	return KeyFile(os.Getenv("AZURE_SERVICEPRINCIPAL_LOCATION"))
}

// KeyFile returns a ServicePrincipalFunc that initializes the service
// principal with values from the specified key file.
func KeyFile(keyFile string) ServicePrincipalFunc {
	return func(ctx context.Context) (servicePrincipal, error) {
		buf, err := os.ReadFile(keyFile)
		if err != nil {
			return servicePrincipal{}, err
		}

		if principal, err := decodePrincipalV1(string(buf)); err == nil {
			return principal, nil
		}
		if principal, err := decodePrincipalV0(string(buf)); err == nil {
			return principal, nil
		}

		return servicePrincipal{}, err
	}
}

// ServicePrincipal returns a ServicePrincipalFunc that initializes the service
// principal with the specified values.
func ServicePrincipal(appID uuid.UUID, appName, appSecret string, tenantID uuid.UUID, tenantDomain string) ServicePrincipalFunc {
	return func(ctx context.Context) (servicePrincipal, error) {
		return servicePrincipal{
			appID:        appID,
			appName:      appName,
			appSecret:    appSecret,
			tenantID:     tenantID,
			tenantDomain: tenantDomain,
		}, nil
	}
}
