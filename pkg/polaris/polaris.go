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
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
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

// Client is used to make calls to the RSC platform.
type Client struct {
	GQL *graphql.Client
	Log log.Logger // Deprecated: use the logger passed in to NewClientWithLogger.
}

// Account represents a Polaris account. Implemented by UserAccount and
// ServiceAccount.
type Account interface {
	isAccount()
}

// Deprecated: The context and the logger parameters will be dropped in the next
// release, use NewClientWithLogger
func NewClient(ctx context.Context, account Account, logger log.Logger) (*Client, error) {
	return NewClientWithLogger(account, logger)
}

// NewClientWithLogger returns a new Client for the specified Account.
func NewClientWithLogger(account Account, logger log.Logger) (*Client, error) {
	var client *Client

	var err error
	switch account := account.(type) {
	case *UserAccount:
		client, err = newClientFromUserAccount(account, logger)
	case *ServiceAccount:
		client, err = newClientFromServiceAccount(account, logger)
	default:
		err = errors.New("invalid account type")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return client, nil
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
		return fmt.Errorf("failed to parse log level: %v", err)
	}
	logger.SetLogLevel(l)

	return nil
}

// newClientFromUserAccount returns a new Client from the specified UserAccount.
func newClientFromUserAccount(account *UserAccount, logger log.Logger) (*Client, error) {
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

	tokenSource := token.NewUserSourceWithLogger(http.DefaultClient, apiURL, account.Username, account.Password, logger)

	client := &Client{
		GQL: graphql.NewClientWithLogger(apiURL, tokenSource, logger),
		Log: logger,
	}

	return client, nil
}

// newClientFromServiceAccount returns a new Client from the specified
// ServiceAccount.
func newClientFromServiceAccount(account *ServiceAccount, logger log.Logger) (*Client, error) {
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

	// Extract the API URL from the token access URI.
	i := strings.LastIndex(account.AccessTokenURI, "/")
	if i < 0 {
		return nil, errors.New("invalid access token uri")
	}
	apiURL := account.AccessTokenURI[:i]

	tokenSource := token.NewServiceAccountSourceWithLogger(
		http.DefaultClient, account.AccessTokenURI, account.ClientID, account.ClientSecret, logger)

	client := &Client{
		GQL: graphql.NewClientWithLogger(apiURL, tokenSource, logger),
		Log: logger,
	}

	return client, nil
}
