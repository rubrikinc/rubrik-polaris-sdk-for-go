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
		return err
	}

	if !result.NotDone() {
		return errors.New("polaris: failed to find service principal using ad graph")
	}
	if result.Value().AppDisplayName == nil {
		return errors.New("polaris: failed to lookup the AppDisplayName using ad graph")
	}
	if result.Value().ObjectID == nil {
		return errors.New("polaris: failed to lookup the ObjectID using ad graph")
	}

	objID, err := uuid.Parse(*result.Value().ObjectID)
	if err != nil {
		return err
	}

	principal.appName = *result.Value().AppDisplayName
	principal.objectID = objID

	return nil
}

// azureServicePrincipalFromEnv creates a service principal from the environment
// and information read from the Azure cloud.
func azurePrincipalFromEnv(ctx context.Context, tenantDomain string) (servicePrincipal, error) {
	graphAuthorizer, err := auth.NewAuthorizerFromEnvironmentWithResource(azure.PublicCloud.GraphEndpoint)
	if err != nil {
		return servicePrincipal{}, err
	}

	rmAuthorizer, err := auth.NewAuthorizerFromEnvironmentWithResource(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return servicePrincipal{}, err
	}

	settings, err := auth.GetSettingsFromEnvironment()
	if err != nil {
		return servicePrincipal{}, err
	}

	appID, err := uuid.Parse(settings.Values[auth.ClientID])
	if err != nil {
		return servicePrincipal{}, err
	}

	tenantID, err := uuid.Parse(settings.Values[auth.TenantID])
	if err != nil {
		return servicePrincipal{}, err
	}

	principal := servicePrincipal{
		appID:        appID,
		appSecret:    settings.Values[auth.ClientSecret],
		tenantID:     tenantID,
		tenantDomain: tenantDomain,
		rmAuthorizer: rmAuthorizer,
	}

	if err := azureGraph(ctx, graphAuthorizer, &principal); err != nil {
		return servicePrincipal{}, err
	}

	return principal, nil
}

// azurePrincipalFromFile creates a service principal from the SDK auth file
// and information read from the Azure cloud.
func azurePrincipalFromFile(ctx context.Context, tenantDomain string) (servicePrincipal, error) {
	graphAuthorizer, err := auth.NewAuthorizerFromFile(azure.PublicCloud.GraphEndpoint)
	if err != nil {
		return servicePrincipal{}, err
	}

	rmAuthorizer, err := auth.NewAuthorizerFromFile(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return servicePrincipal{}, err
	}

	settings, err := auth.GetSettingsFromFile()
	if err != nil {
		return servicePrincipal{}, err
	}

	appID, err := uuid.Parse(settings.Values[auth.ClientID])
	if err != nil {
		return servicePrincipal{}, err
	}

	tenantID, err := uuid.Parse(settings.Values[auth.TenantID])
	if err != nil {
		return servicePrincipal{}, err
	}

	principal := servicePrincipal{
		appID:        appID,
		appSecret:    settings.Values[auth.ClientSecret],
		tenantID:     tenantID,
		tenantDomain: tenantDomain,
		rmAuthorizer: rmAuthorizer,
	}

	if err := azureGraph(ctx, graphAuthorizer, &principal); err != nil {
		return servicePrincipal{}, err
	}

	return principal, nil
}

// The Azure SDK requires all parameters be given as environment variables.
// Lock mutex before accessing the environment to prevent race conditions.
var authLocationMutex sync.Mutex

// Default returns a ServicePrincipalFunc that initializes the service
// principal with values from the default SDK auth file or environment
// variables as described by the SDK documentation.
func Default(tenantDomain string) ServicePrincipalFunc {
	return func(ctx context.Context) (servicePrincipal, error) {
		authLocationMutex.Lock()
		defer authLocationMutex.Unlock()

		principal := servicePrincipal{}
		if authLocation := os.Getenv("AZURE_AUTH_LOCATION"); authLocation != "" {
			var err error
			principal, err = azurePrincipalFromFile(ctx, tenantDomain)
			if err != nil {
				return servicePrincipal{}, err
			}
		} else {
			var err error
			principal, err = azurePrincipalFromEnv(ctx, tenantDomain)
			if err != nil {
				return servicePrincipal{}, err
			}
		}

		return principal, nil
	}
}

// SDKAuthFile returns a ServicePrincipalFunc that initializes the service
// principal with values from the SDK auth file and the Azure cloud.
func SDKAuthFile(authFile, tenantDomain string) ServicePrincipalFunc {
	return func(ctx context.Context) (servicePrincipal, error) {
		if strings.HasPrefix(authFile, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return servicePrincipal{}, err
			}

			authFile = filepath.Join(home, strings.TrimPrefix(authFile, "~/"))
		}

		authLocationMutex.Lock()
		defer authLocationMutex.Unlock()

		authLocation := os.Getenv("AZURE_AUTH_LOCATION")
		defer os.Setenv("AZURE_AUTH_LOCATION", authLocation)

		if err := os.Setenv("AZURE_AUTH_LOCATION", authFile); err != nil {
			return servicePrincipal{}, err
		}

		principal, err := azurePrincipalFromFile(ctx, tenantDomain)
		if err != nil {
			return servicePrincipal{}, err
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
			return servicePrincipal{}, err
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
			return servicePrincipal{}, err
		}

		if err := azureGraph(ctx, graphAuthorizer, &principal); err != nil {
			return servicePrincipal{}, err
		}

		return principal, nil
	}
}
