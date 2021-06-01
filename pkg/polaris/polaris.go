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
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

const (
	// DefaultLocalUserFile path to the default local users file.
	DefaultLocalUserFile = "~/.rubrik/polaris-accounts.json"

	// DefaultServiceAccountFile path to the default service account file.
	DefaultServiceAccountFile = "~/.rubrik/polaris-service-account.json"
)

var (
	// ErrNotFound signals that the specified entity could not be found.
	ErrNotFound = errors.New("not found")

	// ErrNotUnique signals that a request did not result in a unique entity.
	ErrNotUnique = errors.New("not unique")
)

// Client is used to make calls to the Polaris platform.
type Client struct {
	gql *graphql.Client
	log log.Logger
}

// NewClient returns a new Client from the specified Account. The log level of
// the given logger can be changed at runtime using the environment variable
// RUBRIK_POLARIS_LOGLEVEL.
func NewClient(account Account, logger log.Logger) (*Client, error) {
	apiURL := account.URL
	if apiURL == "" {
		apiURL = fmt.Sprintf("https://%s.my.rubrik.com/api", account.Name)
	}

	if _, err := url.ParseRequestURI(apiURL); err != nil {
		return nil, fmt.Errorf("polaris: invalid url: %w", err)
	}
	if account.Username == "" {
		return nil, errors.New("polaris: invalid username")
	}
	if account.Password == "" {
		return nil, errors.New("polaris: invalid password")
	}

	logLevel := "warn"
	if level := os.Getenv("RUBRIK_POLARIS_LOGLEVEL"); level != "" {
		logLevel = level
	}
	if strings.ToLower(logLevel) != "off" {
		level, err := log.ParseLogLevel(logLevel)
		if err != nil {
			return nil, err
		}
		logger.SetLogLevel(level)
	} else {
		logger = &log.DiscardLogger{}
	}

	client := &Client{
		gql: graphql.NewClientFromLocalUser("custom", apiURL, account.Username, account.Password, logger),
		log: logger,
	}

	return client, nil
}

// NewClientFromServiceAccount returns a new Client from the specified
// ServiceAccount. The log level of the given logger can be changed at runtime
// using the environment variable RUBRIK_POLARIS_LOGLEVEL.
func NewClientFromServiceAccount(account ServiceAccount, logger log.Logger) (*Client, error) {
	if account.Name == "" {
		return nil, errors.New("polaris: invalid name")
	}
	if account.ClientID == "" {
		return nil, errors.New("polaris: invalid client id")
	}
	if account.ClientSecret == "" {
		return nil, errors.New("polaris: invalid client secret")
	}
	if _, err := url.ParseRequestURI(account.AccessTokenURI); err != nil {
		return nil, fmt.Errorf("polaris: invalid access token uri: %w", err)
	}

	logLevel := "warn"
	if level := os.Getenv("RUBRIK_POLARIS_LOGLEVEL"); level != "" {
		logLevel = level
	}
	if strings.ToLower(logLevel) != "off" {
		level, err := log.ParseLogLevel(logLevel)
		if err != nil {
			return nil, err
		}
		logger.SetLogLevel(level)
	} else {
		logger = &log.DiscardLogger{}
	}

	// Extract the API URL from the token access URI.
	i := strings.LastIndex(account.AccessTokenURI, "/")
	if i < 0 {
		return nil, errors.New("polaris: invalid access token uri")
	}
	apiURL := account.AccessTokenURI[:i]

	gql := graphql.NewClientFromServiceAccount("custom", apiURL, account.AccessTokenURI, account.ClientID,
		account.ClientSecret, logger)
	client := &Client{
		gql: gql,
		log: logger,
	}

	return client, nil

}

// GQLClient returns the underlaying GraphQL client. Can be used to execute low
// level and raw GraphQL queries against the Polaris platform.
func (c *Client) GQLClient() *graphql.Client {
	return c.gql
}

func fromPolarisRegionNames(polarisNames []string) []string {
	names := make([]string, 0, len(polarisNames))
	for _, name := range polarisNames {
		names = append(names, strings.ReplaceAll(strings.ToLower(name), "_", "-"))
	}

	return names
}

func toPolarisRegionNames(names ...string) []string {
	polarisNames := make([]string, 0, len(names))
	for _, name := range names {
		polarisNames = append(polarisNames, strings.ReplaceAll(strings.ToUpper(name), "-", "_"))
	}

	return polarisNames
}
