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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/google/uuid"
)

// ServicePrincipal Azure service principal used by Polaris to access one or
// more Azure subscriptions.
type servicePrincipal struct {
	appID        uuid.UUID
	appName      string
	appSecret    string
	tenantID     uuid.UUID
	tenantDomain string
	objectID     uuid.UUID
	rmAuthorizer autorest.Authorizer
}

// ServicePrincipalFunc returns a service principal initialized from the values
// passed to the function creating the ServicePrincipalFunc.
type ServicePrincipalFunc func(ctx context.Context) (servicePrincipal, error)

// azureGraph looks up the app display name and object id for the specified
// service principal in Azure AD Graph.
func azureGraph(ctx context.Context, authorizer autorest.Authorizer, principal *servicePrincipal) error {
	client := graphrbac.NewServicePrincipalsClient(principal.tenantID.String())
	client.Authorizer = authorizer

	// This filter should allow the query to run with very few permissions.
	filter := fmt.Sprintf("servicePrincipalNames/any(c:c eq '%s')", principal.appID)
	result, err := client.ListComplete(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get Azure service principal names using Graph: %v", err)
	}

	if !result.NotDone() {
		return errors.New("failed to find Azure service principal using Graph")
	}
	if result.Value().AppDisplayName == nil {
		return errors.New("failed to lookup Azure service principal app display name using Graph")
	}
	if result.Value().ObjectID == nil {
		return errors.New("failed to lookup Azure service principal object id using Graph")
	}

	objID, err := uuid.Parse(*result.Value().ObjectID)
	if err != nil {
		return fmt.Errorf("failed to parse Azure service principal object id: %v", err)
	}

	principal.appName = *result.Value().AppDisplayName
	principal.objectID = objID

	return nil
}

// azureServicePrincipalFromAzEnv creates a service principal from the
// environment and information read from the Azure cloud.
func azurePrincipalFromAzEnv(ctx context.Context, tenantDomain string) (servicePrincipal, error) {
	graphAuthorizer, err := auth.NewAuthorizerFromEnvironmentWithResource(azure.PublicCloud.GraphEndpoint)
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to create Azure Graph authorizer from env: %v", err)
	}

	rmAuthorizer, err := auth.NewAuthorizerFromEnvironmentWithResource(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return servicePrincipal{},
			fmt.Errorf("failed to create Azure Resource Manager authorizer from env: %v", err)
	}

	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to get Azure auth settings from env: %v", err)
	}

	appID, err := uuid.Parse(settings.Values[auth.ClientID])
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to parse Azure client id: %v", err)
	}

	tenantID, err := uuid.Parse(settings.Values[auth.TenantID])
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to parse Azure tenant id: %v", err)
	}

	principal := servicePrincipal{
		appID:        appID,
		appSecret:    settings.Values[auth.ClientSecret],
		tenantID:     tenantID,
		tenantDomain: tenantDomain,
		rmAuthorizer: rmAuthorizer,
	}

	if err := azureGraph(ctx, graphAuthorizer, &principal); err != nil {
		return servicePrincipal{},
			fmt.Errorf("failed to lookup app display name and object id for service principal: %v", err)
	}

	return principal, nil
}

// azurePrincipalFromAzFile creates a service principal from the SDK auth file
// and information read from the Azure cloud.
func azurePrincipalFromAzFile(ctx context.Context, tenantDomain string) (servicePrincipal, error) {
	graphAuthorizer, err := auth.NewAuthorizerFromFile(azure.PublicCloud.GraphEndpoint)
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to create Azure Graph authorizer from file: %v", err)
	}

	rmAuthorizer, err := auth.NewAuthorizerFromFile(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return servicePrincipal{},
			fmt.Errorf("failed to create Azure Resource Manager authorizer from file: %v", err)
	}

	settings, err := auth.GetSettingsFromFile()
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to get Azure auth settings from file: %v", err)
	}

	appID, err := uuid.Parse(settings.Values[auth.ClientID])
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to parse Azure client id: %v", err)
	}

	tenantID, err := uuid.Parse(settings.Values[auth.TenantID])
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to parse Azure tenant id: %v", err)
	}

	principal := servicePrincipal{
		appID:        appID,
		appSecret:    settings.Values[auth.ClientSecret],
		tenantID:     tenantID,
		tenantDomain: tenantDomain,
		rmAuthorizer: rmAuthorizer,
	}

	if err := azureGraph(ctx, graphAuthorizer, &principal); err != nil {
		return servicePrincipal{},
			fmt.Errorf("failed to lookup app display name and object id for service principal: %v", err)
	}

	return principal, nil
}

// principalAzureError is used to tag errors originating from Azure while
// loading a service principal from file. The main use case is to capture errors
// relating to expired secrets.
type principalAzureError struct {
	err error
}

func (e principalAzureError) Error() string {
	return e.err.Error()
}

func (e principalAzureError) Is(target error) bool {
	if _, ok := target.(principalAzureError); ok {
		return true
	}
	if _, ok := target.(*principalAzureError); ok {
		return true
	}

	return false
}

func (e principalAzureError) Unwrap() error {
	return e.err
}

// principalV0 format is used by v0.1.x versions of the Polaris SDK.
type principalV0 struct {
	AppID        string `json:"app_id"`
	AppName      string `json:"app_name"`
	AppSecret    string `json:"app_secret"`
	TenantID     string `json:"tenant_id"`
	TenantDomain string `json:"tenant_domain"`
}

// decodePrincipalV0 decodes the specified string using the v0 format. Returns
// an error if the decoder encounters unknown fields.
func decodePrincipalV0(ctx context.Context, data, tenantDomain string) (servicePrincipal, error) {
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.DisallowUnknownFields()

	var v0 principalV0
	if err := decoder.Decode(&v0); err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to unmarshal v0 service principal: %v", err)
	}
	appID, err := uuid.Parse(v0.AppID)
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to parse service principal app id: %v", err)
	}
	tenantID, err := uuid.Parse(v0.TenantID)
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to parse service principal tenant id: %v", err)
	}
	if tenantDomain != v0.TenantDomain {
		return servicePrincipal{}, fmt.Errorf("tenant domain mismatch: %s != %s", tenantDomain, v0.TenantDomain)
	}

	rmConfig := auth.NewClientCredentialsConfig(v0.AppID, v0.AppSecret, tenantDomain)
	rmAuthorizer, err := rmConfig.Authorizer()
	if err != nil {
		err = fmt.Errorf("failed to get Azure Resource Manager authorizer: %v", err)
		return servicePrincipal{}, principalAzureError{err: err}
	}

	principal := servicePrincipal{
		appID:        appID,
		appName:      v0.AppName,
		appSecret:    v0.AppSecret,
		tenantID:     tenantID,
		tenantDomain: tenantDomain,
		rmAuthorizer: rmAuthorizer,
	}

	graphConfig := auth.ClientCredentialsConfig{
		ClientID:     v0.AppID,
		ClientSecret: v0.AppSecret,
		TenantID:     v0.TenantID,
		Resource:     azure.PublicCloud.GraphEndpoint,
		AADEndpoint:  azure.PublicCloud.ActiveDirectoryEndpoint,
	}
	graphAuthorizer, err := graphConfig.Authorizer()
	if err != nil {
		err = fmt.Errorf("failed to get Azure Graph authorizer: %v", err)
		return servicePrincipal{}, principalAzureError{err: err}
	}

	if err := azureGraph(ctx, graphAuthorizer, &principal); err != nil {
		err = fmt.Errorf("failed to lookup app display name and object id for service principal: %v", err)
		return servicePrincipal{}, principalAzureError{err: err}
	}

	if v0.AppName != "" {
		principal.appName = v0.AppName
	}

	return principal, nil
}

// principalV1 format is used by v0.2.x versions of the Polaris SDK.
type principalV1 struct {
	AppID        string `json:"appId"`
	AppName      string `json:"appName"`
	AppSecret    string `json:"appSecret"`
	TenantID     string `json:"tenantId"`
	TenantDomain string `json:"tenantDomain"`
}

// decodePrincipalV1 decodes the specified string using the v1 format. Returns
// an error if the decoder encounters unknown fields.
func decodePrincipalV1(ctx context.Context, data, tenantDomain string) (servicePrincipal, error) {
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.DisallowUnknownFields()

	var v1 principalV1
	if err := decoder.Decode(&v1); err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to unmarshal v1 service principal: %v", err)
	}
	appID, err := uuid.Parse(v1.AppID)
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to parse service principal app id: %v", err)
	}
	tenantID, err := uuid.Parse(v1.TenantID)
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to parse service principal tenant id: %v", err)
	}
	if tenantDomain != v1.TenantDomain {
		return servicePrincipal{}, fmt.Errorf("tenant domain mismatch: %s != %s", tenantDomain, v1.TenantDomain)
	}

	rmConfig := auth.NewClientCredentialsConfig(v1.AppID, v1.AppSecret, tenantDomain)
	rmAuthorizer, err := rmConfig.Authorizer()
	if err != nil {
		err = fmt.Errorf("failed to get Azure Resource Manager authorizer: %v", err)
		return servicePrincipal{}, principalAzureError{err: err}
	}

	principal := servicePrincipal{
		appID:        appID,
		appName:      v1.AppName,
		appSecret:    v1.AppSecret,
		tenantID:     tenantID,
		tenantDomain: tenantDomain,
		rmAuthorizer: rmAuthorizer,
	}

	graphConfig := auth.ClientCredentialsConfig{
		ClientID:     v1.AppID,
		ClientSecret: v1.AppSecret,
		TenantID:     v1.TenantID,
		Resource:     azure.PublicCloud.GraphEndpoint,
		AADEndpoint:  azure.PublicCloud.ActiveDirectoryEndpoint,
	}
	graphAuthorizer, err := graphConfig.Authorizer()
	if err != nil {
		err = fmt.Errorf("failed to get Azure Graph authorizer: %v", err)
		return servicePrincipal{}, principalAzureError{err: err}
	}

	if err := azureGraph(ctx, graphAuthorizer, &principal); err != nil {
		err = fmt.Errorf("failed to lookup app display name and object id for service principal: %v", err)
		return servicePrincipal{}, principalAzureError{err: err}

	}

	if v1.AppName != "" {
		principal.appName = v1.AppName
	}

	return principal, nil
}

// principalV2 format is used by v0.3.x versions of the Polaris SDK.
type principalV2 struct {
	AppID     string `json:"appId"`
	AppName   string `json:"appName"`
	AppSecret string `json:"appSecret"`
	TenantID  string `json:"tenantId"`
}

// decodePrincipalV2 decodes the specified string using the v2 format. Returns
// an error if the decoder encounters unknown fields.
func decodePrincipalV2(ctx context.Context, data, tenantDomain string) (servicePrincipal, error) {
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.DisallowUnknownFields()

	var v2 principalV2
	if err := decoder.Decode(&v2); err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to unmarshal v2 service principal: %v", err)
	}
	appID, err := uuid.Parse(v2.AppID)
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to parse service principal app id: %v", err)
	}
	tenantID, err := uuid.Parse(v2.TenantID)
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to parse service principal tenant id: %v", err)
	}

	rmConfig := auth.NewClientCredentialsConfig(v2.AppID, v2.AppSecret, tenantDomain)
	rmAuthorizer, err := rmConfig.Authorizer()
	if err != nil {
		err = fmt.Errorf("failed to get Azure Resource Manager authorizer: %v", err)
		return servicePrincipal{}, principalAzureError{err: err}
	}

	principal := servicePrincipal{
		appID:        appID,
		appName:      v2.AppName,
		appSecret:    v2.AppSecret,
		tenantID:     tenantID,
		tenantDomain: tenantDomain,
		rmAuthorizer: rmAuthorizer,
	}

	graphConfig := auth.ClientCredentialsConfig{
		ClientID:     v2.AppID,
		ClientSecret: v2.AppSecret,
		TenantID:     v2.TenantID,
		Resource:     azure.PublicCloud.GraphEndpoint,
		AADEndpoint:  azure.PublicCloud.ActiveDirectoryEndpoint,
	}
	graphAuthorizer, err := graphConfig.Authorizer()
	if err != nil {
		err = fmt.Errorf("failed to get Azure Graph authorizer: %v", err)
		return servicePrincipal{}, principalAzureError{err: err}
	}

	if err := azureGraph(ctx, graphAuthorizer, &principal); err != nil {
		err = fmt.Errorf("failed to lookup app display name and object id for service principal: %v", err)
		return servicePrincipal{}, principalAzureError{err: err}
	}

	if v2.AppName != "" {
		principal.appName = v2.AppName
	}

	return principal, nil
}

// azurePrincipalFromKeyFile creates a service principal from the specified key
// file and information read from the Azure cloud.
func azurePrincipalFromKeyFile(ctx context.Context, keyFile, tenantDomain string) (servicePrincipal, error) {
	if strings.HasPrefix(keyFile, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return servicePrincipal{}, fmt.Errorf("failed to get home dir: %v", err)
		}

		keyFile = filepath.Join(home, strings.TrimPrefix(keyFile, "~/"))
	}

	buf, err := os.ReadFile(keyFile)
	if err != nil {
		return servicePrincipal{}, fmt.Errorf("failed to read key file: %v", err)
	}

	principal, err := decodePrincipalV2(ctx, string(buf), tenantDomain)
	if err == nil {
		return principal, nil
	}
	if errors.Is(err, principalAzureError{}) {
		return servicePrincipal{}, err
	}

	principal, err = decodePrincipalV1(ctx, string(buf), tenantDomain)
	if err == nil {
		return principal, nil
	}
	if errors.Is(err, principalAzureError{}) {
		return servicePrincipal{}, err
	}

	principal, err = decodePrincipalV0(ctx, string(buf), tenantDomain)
	if err == nil {
		return principal, nil
	}
	if errors.Is(err, principalAzureError{}) {
		return servicePrincipal{}, err
	}

	return servicePrincipal{}, fmt.Errorf("unrecognized file format: %v", keyFile)
}

// The Azure SDK requires all parameters be given as environment variables.
// Lock mutex before accessing the environment to prevent race conditions.
var authLocationMutex sync.Mutex

// Default returns a ServicePrincipalFunc that initializes the service
// principal with values from the default SDK auth file, the default key file
// or environment variables as described by the SDK documentation.
func Default(tenantDomain string) ServicePrincipalFunc {
	return func(ctx context.Context) (servicePrincipal, error) {
		authLocationMutex.Lock()
		defer authLocationMutex.Unlock()

		var principal servicePrincipal
		var err error
		switch {
		case os.Getenv("AZURE_AUTH_LOCATION") != "":
			principal, err = azurePrincipalFromAzFile(ctx, tenantDomain)
			if err != nil {
				return servicePrincipal{}, fmt.Errorf("failed to read service principal from file: %v", err)
			}
		case os.Getenv("AZURE_SERVICEPRINCIPAL_LOCATION") != "":
			keyfile := os.Getenv("AZURE_SERVICEPRINCIPAL_LOCATION")
			principal, err = azurePrincipalFromKeyFile(ctx, keyfile, tenantDomain)
			if err != nil {
				return servicePrincipal{}, fmt.Errorf("failed to read service principal from file: %v", err)
			}
		default:
			principal, err = azurePrincipalFromAzEnv(ctx, tenantDomain)
			if err != nil {
				return servicePrincipal{}, fmt.Errorf("failed to read service principal from env: %v", err)
			}
		}

		return principal, nil
	}
}

// KeyFile returns a ServicePrincipalFunc that initializes the service
// principal with values from the specified key file and the Azure cloud.
func KeyFile(keyFile, tenantDomain string) ServicePrincipalFunc {
	return func(ctx context.Context) (servicePrincipal, error) {
		return azurePrincipalFromKeyFile(ctx, keyFile, tenantDomain)
	}
}

// SDKAuthFile returns a ServicePrincipalFunc that initializes the service
// principal with values from the SDK auth file and the Azure cloud.
func SDKAuthFile(authFile, tenantDomain string) ServicePrincipalFunc {
	return func(ctx context.Context) (servicePrincipal, error) {
		if strings.HasPrefix(authFile, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return servicePrincipal{}, fmt.Errorf("failed to get home dir: %v", err)
			}

			authFile = filepath.Join(home, strings.TrimPrefix(authFile, "~/"))
		}

		authLocationMutex.Lock()
		defer authLocationMutex.Unlock()

		authLocation := os.Getenv("AZURE_AUTH_LOCATION")
		defer os.Setenv("AZURE_AUTH_LOCATION", authLocation)

		if err := os.Setenv("AZURE_AUTH_LOCATION", authFile); err != nil {
			return servicePrincipal{}, fmt.Errorf("failed to set env AZURE_AUTH_LOCATION: %v", err)
		}

		principal, err := azurePrincipalFromAzFile(ctx, tenantDomain)
		if err != nil {
			return servicePrincipal{}, fmt.Errorf("failed to read service principal from file: %v", err)
		}

		return principal, nil
	}
}

// ServicePrincipal returns a ServicePrincipalFunc that initializes the service
// principal with the specified values.
func ServicePrincipal(appID uuid.UUID, appSecret string, tenantID uuid.UUID, tenantDomain string) ServicePrincipalFunc {
	return func(ctx context.Context) (servicePrincipal, error) {
		rmConfig := auth.NewClientCredentialsConfig(appID.String(), appSecret, tenantID.String())
		rmAuthorizer, err := rmConfig.Authorizer()
		if err != nil {
			return servicePrincipal{}, fmt.Errorf("failed to get Azure Resource Manager authorizer: %v", err)
		}

		principal := servicePrincipal{
			appID:        appID,
			appSecret:    appSecret,
			tenantID:     tenantID,
			tenantDomain: tenantDomain,
			rmAuthorizer: rmAuthorizer,
		}

		graphConfig := auth.ClientCredentialsConfig{
			ClientID:     appID.String(),
			ClientSecret: appSecret,
			TenantID:     tenantID.String(),
			Resource:     azure.PublicCloud.GraphEndpoint,
			AADEndpoint:  azure.PublicCloud.ActiveDirectoryEndpoint,
		}
		graphAuthorizer, err := graphConfig.Authorizer()
		if err != nil {
			return servicePrincipal{}, fmt.Errorf("failed to get Azure Graph authorizer: %v", err)
		}

		if err := azureGraph(ctx, graphAuthorizer, &principal); err != nil {
			return servicePrincipal{},
				fmt.Errorf("failed to lookup app display name and object id for service principal: %v", err)
		}

		return principal, nil
	}
}
