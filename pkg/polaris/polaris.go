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

// Package polaris contains code to interact with the RSC platform on a high
// level. Relies on the graphql package for low level queries.
package polaris

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/token"
)

const (
	// Client environmental variables.
	keyLogLevel         = "RUBRIK_POLARIS_LOGLEVEL"
	keyTokenCache       = "RUBRIK_POLARIS_TOKEN_CACHE"
	keyTokenCacheDir    = "RUBRIK_POLARIS_TOKEN_CACHE_DIR"
	keyTokenCacheSecret = "RUBRIK_POLARIS_TOKEN_CACHE_SECRET"
)

// CacheParams is used to configure the token cache.
//
// Note, if the user account or service account has environment variable
// override enabled, the cache param values can be overridden.
type CacheParams struct {
	Enable bool   // Enable the token cache.
	Dir    string // Directory to store the cached tokens in.
	Secret string // Secret to encrypt the cached tokens with.
}

// Client is used to make calls to the RSC platform.
type Client struct {
	Account Account
	GQL     *graphql.Client
}

// NewClient returns a new Client for the specified Account.
//
// The client will cache authentication tokens by default, this behavior can be
// overridden by setting the environment variable RUBRIK_POLARIS_TOKEN_CACHE to
// false, given that the account specified allows environment variable
// overrides.
func NewClient(account Account) (*Client, error) {
	return NewClientWithLoggerAndCacheParams(account, CacheParams{Enable: true}, log.DiscardLogger{})
}

// NewClientWithCacheParams returns a new Client for the specified Account.
//
// Note, the cache parameters specified can be overridden by environmental
// variables if the specified account allows environment variable overrides.
func NewClientWithCacheParams(account Account, cacheParams CacheParams) (*Client, error) {
	return NewClientWithLoggerAndCacheParams(account, cacheParams, log.DiscardLogger{})
}

// NewClientWithLogger returns a new Client for the specified Account.
//
// The client will cache authentication tokens by default, this behavior can be
// overridden by setting the environment variable RUBRIK_POLARIS_TOKEN_CACHE to
// false, given that the account specified allows environment variable
// overrides.
func NewClientWithLogger(account Account, logger log.Logger) (*Client, error) {
	return NewClientWithLoggerAndCacheParams(account, CacheParams{Enable: true}, logger)
}

// NewClientWithLoggerAndCacheParams returns a new Client for the specified
// Account.
//
// Note, the cache parameters specified can be overridden by environmental
// variables if the specified account allows environment variable overrides.
func NewClientWithLoggerAndCacheParams(account Account, cacheParams CacheParams, logger log.Logger) (*Client, error) {
	if account.allowEnvOverride() {
		if val := os.Getenv(keyTokenCache); val != "" {
			if b, err := strconv.ParseBool(val); err != nil {
				cacheParams.Enable = b
			}
		}
		if val := os.Getenv(keyTokenCacheDir); val != "" {
			cacheParams.Dir = val
		}
		if val := os.Getenv(keyTokenCacheSecret); val != "" {
			cacheParams.Secret = val
		}
	}

	var tokenSource token.Source
	switch account := account.(type) {
	case *ServiceAccount:
		tokenSource = token.NewServiceAccountSourceWithLogger(
			http.DefaultClient, account.TokenURL(), account.ClientID, account.ClientSecret, logger)
	case *UserAccount:
		tokenSource = token.NewUserSourceWithLogger(
			http.DefaultClient, account.TokenURL(), account.Username, account.Password, logger)
	default:
		return nil, errors.New("invalid account type")
	}

	if cacheParams.Enable {
		cacheDir := os.TempDir()
		if cacheParams.Dir != "" {
			cacheDir = cacheParams.Dir
		}

		// When a custom secret is used, we append the custom secret to the
		// suffix material to ensure the token is written to separate file,
		// since it/ will be encrypted with the custom secret.
		keyMaterial := account.cacheKeyMaterial()
		suffixMaterial := account.cacheSuffixMaterial()
		if cacheParams.Secret != "" {
			keyMaterial = cacheParams.Secret
			suffixMaterial += cacheParams.Secret
		}

		var err error
		tokenSource, err = token.NewCacheWithDir(tokenSource, cacheDir, keyMaterial, suffixMaterial)
		if err != nil {
			return nil, fmt.Errorf("failed to create token cache: %s", err)
		}
	}

	return &Client{
		Account: account,
		GQL:     graphql.NewClientWithLogger(account.APIURL(), tokenSource, logger),
	}, nil
}

// SetLogger sets the logger to use.
func (c *Client) SetLogger(logger log.Logger) {
	c.GQL.SetLogger(logger)
}

// SetLogLevelFromEnv sets the log level of the logger to the log level
// specified in the RUBRIK_POLARIS_LOGLEVEL environment variable.
func SetLogLevelFromEnv(logger log.Logger) error {
	level := os.Getenv(keyLogLevel)
	if level == "" {
		return nil
	}
	if strings.ToLower(level) == "off" {
		level = "fatal"
	}

	l, err := log.ParseLogLevel(level)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %s", err)
	}
	logger.SetLogLevel(l)

	return nil
}
