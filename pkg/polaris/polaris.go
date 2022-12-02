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

// Package polaris contains code to interact with the Polaris platform on a
// high level. Relies on the graphql package for low level queries.
package polaris

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

const (
	// DefaultLocalUserFile path to the default local users file.
	DefaultLocalUserFile = "~/.rubrik/polaris-accounts.json"

	// DefaultServiceAccountFile path to the default service account file.
	DefaultServiceAccountFile = "~/.rubrik/polaris-service-account.json"
)

// Client is used to make calls to the Polaris platform.
type Client struct {
	gql *graphql.Client
	log log.Logger
}

// AWS returns the AWS part of the API.
func (c *Client) AWS() aws.API {
	return aws.NewAPI(c.gql)
}

// Azure returns the Azure part of the API.
func (c *Client) Azure() azure.API {
	return azure.NewAPI(c.gql)
}

// GCP returns the GCP part of the API.
func (c *Client) GCP() gcp.API {
	return gcp.NewAPI(c.gql)
}

// Account represents a Polaris account. Implemented by UserAccount and
// ServiceAccount.
type Account interface {
	isAccount()
}

// NewClient returns a new Client from the specified Account. The log level of
// the given logger can be changed at runtime using the environment variable
// RUBRIK_POLARIS_LOGLEVEL.
func NewClient(ctx context.Context, account Account, logger log.Logger) (*Client, error) {
	if serviceAccount, ok := account.(*ServiceAccount); ok {
		return newClientFromServiceAccount(ctx, serviceAccount, http.DefaultTransport, logger)
	}

	if userAccount, ok := account.(*UserAccount); ok {
		return newClientFromUserAccount(ctx, userAccount, logger)
	}

	return nil, errors.New("invalid account type")
}

func NewClientFromServiceAccountWithTransport(ctx context.Context, account Account, transport http.RoundTripper, logger log.Logger) (*Client, error) {
	if serviceAccount, ok := account.(*ServiceAccount); ok {
		return newClientFromServiceAccount(ctx, serviceAccount, transport, logger)
	}
	return nil, errors.New("invalid account type")
}

// newClientFromUserAccount returns a new Client from the specified
// UserAccount. The log level of the given logger can be changed at runtime
// using the environment variable RUBRIK_POLARIS_LOGLEVEL.
func newClientFromUserAccount(ctx context.Context, account *UserAccount, logger log.Logger) (*Client, error) {
	apiURL := account.URL
	if apiURL == "" {
		apiURL = fmt.Sprintf("https://%s.my.rubrik.com/api", account.Name)
	}

	if _, err := url.ParseRequestURI(apiURL); err != nil {
		return nil, fmt.Errorf("invalid url: %v", err)
	}
	if account.Username == "" {
		return nil, errors.New("invalid username")
	}
	if account.Password == "" {
		return nil, errors.New("invalid password")
	}

	if level := os.Getenv("RUBRIK_POLARIS_LOGLEVEL"); level != "" {
		if strings.ToLower(level) != "off" {
			l, err := log.ParseLogLevel(level)
			if err != nil {
				return nil, fmt.Errorf("failed to parse log level: %v", err)
			}
			logger.SetLogLevel(l)
		} else {
			logger = &log.DiscardLogger{}
		}
	}

	logger.Printf(log.Info, "Polaris API URL: %s", apiURL)

	// The gql client is initialized without a version. Query cluster to find
	// out the current version.
	gqlClient := graphql.NewClientFromLocalUser("custom", apiURL, account.Username, account.Password, logger)
	version, err := core.Wrap(gqlClient).DeploymentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment version: %v", err)
	}
	gqlClient.Version = version

	logger.Printf(log.Info, "Polaris version: %s", version)

	client := &Client{
		gql: gqlClient,
		log: logger,
	}

	return client, nil
}

// newClientFromServiceAccount returns a new Client from the specified
// ServiceAccount. The log level of the given logger can be changed at runtime
// using the environment variable RUBRIK_POLARIS_LOGLEVEL.
func newClientFromServiceAccount(ctx context.Context, account *ServiceAccount, transport http.RoundTripper, logger log.Logger) (*Client, error) {
	if account.Name == "" {
		return nil, errors.New("invalid name")
	}
	if account.ClientID == "" {
		return nil, errors.New("invalid client id")
	}
	if account.ClientSecret == "" {
		return nil, errors.New("invalid client secret")
	}
	if _, err := url.ParseRequestURI(account.AccessTokenURI); err != nil {
		return nil, fmt.Errorf("invalid access token uri: %v", err)
	}

	if level := os.Getenv("RUBRIK_POLARIS_LOGLEVEL"); level != "" {
		if strings.ToLower(level) != "off" {
			l, err := log.ParseLogLevel(level)
			if err != nil {
				return nil, fmt.Errorf("failed to parse log level: %v", err)
			}
			logger.SetLogLevel(l)
		} else {
			logger = &log.DiscardLogger{}
		}
	}

	// Extract the API URL from the token access URI.
	i := strings.LastIndex(account.AccessTokenURI, "/")
	if i < 0 {
		return nil, errors.New("invalid access token uri")
	}
	apiURL := account.AccessTokenURI[:i]

	logger.Printf(log.Info, "Polaris API URL: %s", apiURL)

	// The gql client is initialized without a version. Query cluster to find
	// out the current version.
	gqlClient := graphql.NewClientFromServiceAccountWithTransport(
		"custom",
		apiURL,
		account.AccessTokenURI,
		account.ClientID,
		account.ClientSecret,
		transport,
		logger,
	)

	version, err := core.Wrap(gqlClient).DeploymentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment version: %v", err)
	}
	gqlClient.Version = version

	logger.Printf(log.Info, "Polaris version: %s", version)

	client := &Client{
		gql: gqlClient,
		log: logger,
	}

	return client, nil
}

// GQLClient returns the underlaying GraphQL client. Can be used to execute low
// level and raw GraphQL queries against the Polaris platform.
func (c *Client) GQLClient() *graphql.Client {
	return c.gql
}
