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
// level. Relies on the graphql package for low-level queries.
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
	// DefaultLocalUserFile path to the default local users file.
	DefaultLocalUserFile = "~/.rubrik/polaris-accounts.json"

	// DefaultServiceAccountFile path to the default service account file.
	DefaultServiceAccountFile = "~/.rubrik/polaris-service-account.json"
)

// Account represents a Polaris account. Implemented by UserAccount and
// ServiceAccount.
type Account interface {
	// AccountName returns the RSC account name.
	AccountName() string

	// AccountFQDN returns the fully qualified domain name of the RSC account.
	AccountFQDN() string

	// APIURL returns the RSC account API URL.
	APIURL() string

	// TokenURL returns the RSC account token URL.
	TokenURL() string

	allowEnvOverride() bool

	// Cryptographic material for encrypting cached access tokens.
	cacheKeyMaterial() string
	cacheSuffixMaterial() string
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
	return NewClientWithLogger(account, log.DiscardLogger{})
}

// NewClientWithLogger returns a new Client for the specified Account.
//
// The client will cache authentication tokens by default, this behavior can be
// overridden by setting the environment variable RUBRIK_POLARIS_TOKEN_CACHE to
// false, given that the account specified allows environment variable
// overrides.
func NewClientWithLogger(account Account, logger log.Logger) (*Client, error) {
	cacheToken := true
	if account.allowEnvOverride() {
		if tcUse := os.Getenv("RUBRIK_POLARIS_TOKEN_CACHE"); tcUse != "" {
			if b, err := strconv.ParseBool(tcUse); err != nil {
				cacheToken = b
			}
		}
	}

	var tokenSource token.Source
	switch account := account.(type) {
	case *UserAccount:
		tokenSource = token.NewUserSourceWithLogger(
			http.DefaultClient, account.TokenURL(), account.Username, account.Password, logger)
	case *ServiceAccount:
		tokenSource = token.NewServiceAccountSourceWithLogger(
			http.DefaultClient, account.TokenURL(), account.ClientID, account.ClientSecret, logger)
	default:
		return nil, errors.New("failed to create client: invalid account type")
	}

	if cacheToken {
		var err error
		tokenSource, err = token.NewCache(
			tokenSource, account.cacheKeyMaterial(), account.cacheSuffixMaterial(), account.allowEnvOverride())
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
	level := os.Getenv("RUBRIK_POLARIS_LOGLEVEL")
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
