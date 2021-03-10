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

package polaris

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

// taskChainID represents the identity of a Polaris task chain. A Polaris task
// chain is a collection of sequential tasks that all must complete for the task
// chain to be considered complete.
type taskChainID string

// taskChainState represents the state of a Polaris task chain.
type taskChainState string

const (
	taskChainInvalid   taskChainState = ""
	taskChainCanceled  taskChainState = "CANCELED"
	taskChainCanceling taskChainState = "CANCELING"
	taskChainFailed    taskChainState = "FAILED"
	taskChainReady     taskChainState = "READY"
	taskChainRunning   taskChainState = "RUNNING"
	taskChainSucceeded taskChainState = "SUCCEEDED"
	taskChainUndoing   taskChainState = "UNDOING"
)

// toTaskChainState returns the task chain state that match the given string,
// if no match is found taskChainInvalid is returned.
func toTaskChainState(status string) taskChainState {
	switch status {
	case "CANCELED":
		return taskChainCanceled
	case "CANCELING":
		return taskChainCanceling
	case "FAILED":
		return taskChainFailed
	case "READY":
		return taskChainReady
	case "RUNNING":
		return taskChainRunning
	case "SUCCEEDED":
		return taskChainSucceeded
	case "UNDOING":
		return taskChainUndoing
	default:
		return taskChainInvalid
	}
}

const (
	// DefaultConfigFile path to where SDK configuration files should be stored
	// by default.
	DefaultConfigFile = "~/.rubrik/polaris-accounts.json"
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
// variables but be all upper case and prepended with RUBRIK_POLARIS, e.g.
// RUBRIK_POLARIS_USERNAME.
func ConfigFromEnv() (Config, error) {
	account := os.Getenv("RUBRIK_POLARIS_ACCOUNT")
	if account != "" {
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

	return Config{URL: url, Username: username, Password: password, LogLevel: logLevel}, nil
}

// ConfigFromFile returns a new Client configuration read from the specified
// path. Configuration files must be in JSON format and the attributes must
// have the same name as the Config variables but be all lower case.
//
// Example configuration file:
//   {
//     "default": {
//       "username": "gopher",
//       "password": "mcgopherface",
//     },
//     "test": {
//       "username": "testie",
//       "password": "mctestieface",
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
		return Config{}, errors.New("polaris: account not found in configuration")
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

// Client is used to make calls to the Polaris platform.
type Client struct {
	url string
	gql *graphql.Client
	log Logger
}

// NewClient returns a new Client with the specified configuration.
func NewClient(config Config, log Logger) (*Client, error) {
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
		level, err := parseLogLevel(logLevel)
		if err != nil {
			return nil, err
		}
		log.SetLogLevel(level)
	} else {
		log = &DiscardLogger{}
	}

	client := &Client{
		url: apiURL,
		gql: graphql.NewClient(apiURL, config.Username, config.Password),
		log: log,
	}

	return client, nil
}

// taskChainState returns the state of the Polaris task chain with the specified
// task chain identity.
func (c *Client) taskChainState(ctx context.Context, id taskChainID) (taskChainState, error) {
	c.log.Print(Trace, "Client.taskChainState")

	buf, err := c.gql.Request(ctx, graphql.CoreTaskchainStatusQuery, struct {
		Filter string `json:"filter,omitempty"`
	}{Filter: string(id)})
	if err != nil {
		return taskChainInvalid, err
	}

	var payload struct {
		Data struct {
			Query struct {
				TaskChain struct {
					ID            int64     `json:"id"`
					TaskchainUUID string    `json:"taskchainUuid"`
					State         string    `json:"state"`
					ProgressedAt  time.Time `json:"progressedAt"`
				} `json:"taskchain"`
			} `json:"getKorgTaskchainStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return taskChainInvalid, err
	}

	return toTaskChainState(payload.Data.Query.TaskChain.State), nil
}

// taskChainWaitFor blocks until the Polaris task chain with the specified task
// chain identity has completed. When the task chain completes the final state
// of the task chain is returned.
func (c *Client) taskChainWaitFor(ctx context.Context, id taskChainID) (taskChainState, error) {
	c.log.Print(Trace, "Client.taskChainWaitFor")

	for {
		status, err := c.taskChainState(ctx, id)
		if err != nil {
			return taskChainInvalid, err
		}

		if status == taskChainCanceled || status == taskChainFailed || status == taskChainSucceeded {
			return status, nil
		}

		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			return taskChainInvalid, ctx.Err()
		}
	}
}
