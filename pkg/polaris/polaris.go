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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

const (
	// DefaultConfigFile path to where SDK configuration files are stored by
	// default.
	DefaultConfigFile = "~/.rubrik/polaris-accounts.json"
)

var (
	// ErrAccountNotFound signals that the specified account could not be
	// found.
	ErrAccountNotFound = errors.New("polaris: account not found")
)

// Config holds the Client configuration.
type Config struct {
	// Polaris account name.
	Account string

	// Polaris account username.
	Username string

	// Polaris account password.
	Password string

	// Optional Polaris API endpoint. Useful for running the SDK against a test
	// service. Defaults to https://{Account}.my.rubrik.com/api.
	URL string

	// The log level to use. Log levels supported are: trace, debug, info,
	// warn, error, fatal and off. If log level isn't specified it defaults
	// to warn.
	LogLevel string
}

// ConfigFromEnv returns a new Client configuration from the user's environment
// variables. Environment variables must have the same name as the Config
// fields but be all upper case and prepended with RUBRIK_POLARIS, e.g.
// RUBRIK_POLARIS_USERNAME.
func ConfigFromEnv() (Config, error) {
	account := os.Getenv("RUBRIK_POLARIS_ACCOUNT")
	if account == "" {
		return Config{}, errors.New("polaris: missing environment variable: RUBRIK_POLARIS_ACCOUNT")
	}

	username := os.Getenv("RUBRIK_POLARIS_USERNAME")
	if username == "" {
		return Config{}, errors.New("polaris: missing environment variable: RUBRIK_POLARIS_USERNAME")
	}

	password := os.Getenv("RUBRIK_POLARIS_PASSWORD")
	if password == "" {
		return Config{}, errors.New("polaris: missing environment variable: RUBRIK_POLARIS_PASSWORD")
	}

	// Optional environment variables.
	url := os.Getenv("RUBRIK_POLARIS_URL")
	logLevel := os.Getenv("RUBRIK_POLARIS_LOGLEVEL")

	return Config{URL: url, Account: account, Username: username, Password: password, LogLevel: logLevel}, nil
}

// ConfigFromFile returns a new Client configuration read from the specified
// path. Configuration files must be in JSON format and the attributes must
// have the same name as the Config fields but be all lower case. Note that the
// Account field is used as a key for the JSON object. E.g:
//
//   {
//     "account-1": {
//       "username": "username-1",
//       "password": "password-1",
//     },
//     "account-2": {
//       "username": "username-2",
//       "password": "password-2",
//       "loglevel": "debug"
//     }
//   }
func ConfigFromFile(path, account string) (Config, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var configs map[string]Config
	if err := json.Unmarshal(buf, &configs); err != nil {
		return Config{}, err
	}

	config, ok := configs[account]
	if !ok {
		return Config{}, fmt.Errorf("polaris: account %q not found in configuration", account)
	}
	config.Account = account

	if config.Account == "" {
		return Config{}, errors.New("polaris: missing JSON attribute: account")
	}
	if config.Username == "" {
		return Config{}, errors.New("polaris: missing JSON attribute: username")
	}
	if config.Password == "" {
		return Config{}, errors.New("polaris: missing JSON attribute: password")
	}

	return config, nil
}

// DefaultConfig returns a new Client configuration read from the default
// config path. Environment variables can be used to override configuration
// information in the configuration file. See ConfigFromEnv and ConfigFromFile
// for details. Note that the environment variable POLARIS_RUBRIK_ACCOUNT will
// override the account parameter passed in.
func DefaultConfig(account string) (Config, error) {
	if envAccount := os.Getenv("RUBRIK_POLARIS_ACCOUNT"); envAccount != "" {
		account = envAccount
	}

	config, err := ConfigFromFile(DefaultConfigFile, account)
	if err != nil {
		return Config{}, err
	}

	username := os.Getenv("RUBRIK_POLARIS_USERNAME")
	if username != "" {
		config.Username = username
	}

	password := os.Getenv("RUBRIK_POLARIS_PASSWORD")
	if password != "" {
		config.Password = password
	}

	url := os.Getenv("RUBRIK_POLARIS_URL")
	if url != "" {
		config.URL = url
	}

	logLevel := os.Getenv("RUBRIK_POLARIS_LOGLEVEL")
	if logLevel != "" {
		config.LogLevel = logLevel
	}

	return config, nil
}

// Client is used to make calls to the Polaris platform.
type Client struct {
	url string
	gql *graphql.Client
	log log.Logger
}

// NewClient returns a new Client with the specified configuration. Note that
// when Config.Log is set to false the logger given to NewClient is silently
//replaced by a DiscardLogger.
func NewClient(config Config, logger log.Logger) (*Client, error) {
	// Apply default values.
	apiURL := config.URL
	if apiURL == "" {
		apiURL = fmt.Sprintf("https://%s.my.rubrik.com/api", config.Account)
	}

	logLevel := config.LogLevel
	if logLevel == "" {
		logLevel = "warn"
	}

	// Validate configuration.
	if _, err := url.ParseRequestURI(apiURL); err != nil {
		return nil, fmt.Errorf("polaris: invalid url: %w", err)
	}

	if config.Username == "" {
		return nil, errors.New("polaris: invalid username")
	}

	if config.Password == "" {
		return nil, errors.New("polaris: invalid password")
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
		url: apiURL,
		gql: graphql.NewClient(apiURL, config.Username, config.Password, logger),
		log: logger,
	}

	return client, nil
}

// GQLClient returns the underlaying GraphQL client. Can be used to execute low
// level and raw GraphQL queries against the Polaris platform.
func (c *Client) GQLClient() *graphql.Client {
	return c.gql
}
