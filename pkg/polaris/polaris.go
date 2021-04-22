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
	// ErrNotFound signals that the specified entity could not be found.
	ErrNotFound = errors.New("not found")

	// ErrNotUnique signals that a request did not result in a unique entity.
	ErrNotUnique = errors.New("not unique")
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

// configFromEnv returns a Config from the current environment variables.
func configFromEnv() Config {
	return Config{
		Account:  os.Getenv("RUBRIK_POLARIS_ACCOUNT"),
		Username: os.Getenv("RUBRIK_POLARIS_USERNAME"),
		Password: os.Getenv("RUBRIK_POLARIS_PASSWORD"),
		URL:      os.Getenv("RUBRIK_POLARIS_URL"),
		LogLevel: os.Getenv("RUBRIK_POLARIS_LOGLEVEL"),
	}
}

// configFromFile returns a Config from the specified file and account name.
func configFromFile(path, account string) (Config, error) {
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

	// Override the given account with the corresponding environment variable.
	if envAccount := os.Getenv("RUBRIK_POLARIS_ACCOUNT"); envAccount != "" {
		account = envAccount
	}

	config, ok := configs[account]
	if !ok {
		return Config{}, fmt.Errorf("polaris: account %q not found in configuration", account)
	}
	config.Account = account

	return config, nil
}

// mergeConfig merges the src configuration with the dest configuration and
// returns the resulting configuration.
func mergeConfig(dest, src Config) (Config, error) {
	if src.Account != "" {
		dest.Account = src.Account
	}
	if src.Username != "" {
		dest.Username = src.Username
	}
	if src.Password != "" {
		dest.Password = src.Password
	}
	if src.URL != "" {
		dest.URL = src.URL
	}
	if src.LogLevel != "" {
		dest.LogLevel = src.LogLevel
	}

	if dest.Account == "" {
		return Config{}, errors.New("polaris: missing required field: account")
	}
	if dest.Username == "" {
		return Config{}, errors.New("polaris: missing required field: username")
	}
	if dest.Password == "" {
		return Config{}, errors.New("polaris: missing required field: password")
	}

	return dest, nil
}

// ConfigFromEnv returns a new Client configuration from the user's environment
// variables. Environment variables must have the same name as the Config
// fields but be all upper case and prepended with RUBRIK_POLARIS, e.g.
// RUBRIK_POLARIS_USERNAME.
func ConfigFromEnv() (Config, error) {
	config := configFromEnv()

	if config.Account == "" {
		return Config{}, errors.New("polaris: missing environment variable: RUBRIK_POLARIS_ACCOUNT")
	}
	if config.Username == "" {
		return Config{}, errors.New("polaris: missing environment variable: RUBRIK_POLARIS_USERNAME")
	}
	if config.Password == "" {
		return Config{}, errors.New("polaris: missing environment variable: RUBRIK_POLARIS_PASSWORD")
	}

	return config, nil
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
//
// The account parameter can be overridden using the RUBRIK_POLARIS_ACCOUNT
// environment variable.
func ConfigFromFile(path, account string) (Config, error) {
	config, err := configFromFile(path, account)
	if err != nil {
		return Config{}, err
	}

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
	config, _ := configFromFile(DefaultConfigFile, account)
	return mergeConfig(config, configFromEnv())
}

// Client is used to make calls to the Polaris platform.
type Client struct {
	url string
	gql *graphql.Client
	log log.Logger
}

// NewClient returns a new Client with the specified configuration. Note that
// when Config.Log is set to false the logger given to NewClient is silently
// replaced by a DiscardLogger.
func NewClient(config Config, logger log.Logger) (*Client, error) {
	return NewClientForApp("custom", config, logger)
}

// NewClientForApp returns a new Client with the specified configuration
// intended to be used in the named application. Note that when Config.Log is
// set to false the logger given to NewClient is silently replaced by a
// DiscardLogger.
func NewClientForApp(app string, config Config, logger log.Logger) (*Client, error) {
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
		gql: graphql.NewClient(app, apiURL, config.Username, config.Password, logger),
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
