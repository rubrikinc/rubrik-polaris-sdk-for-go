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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

// UserAccount holds an RSC local user account configuration. Note that RSC
// user accounts with MFA enabled cannot be used.
type UserAccount struct {
	// Polaris account name.
	Name string

	// Polaris account username.
	Username string

	// Polaris account password.
	Password string

	// Optional Polaris API endpoint. Useful for running the SDK against a test
	// service. Defaults to https://{Name}.my.rubrik.com/api.
	URL string

	envOverride bool
}

func (a *UserAccount) allowEnvOverride() bool {
	return a.envOverride
}

// lookupUserAccount returns a UserAccount from the map of available user
// accounts. If the map contains multiple user accounts and name doesn't match
// any of them, an empty UserAccount with the specified name is returned.
func lookupUserAccount(name string, accounts map[string]UserAccount) UserAccount {
	if len(accounts) == 1 {
		for name, account := range accounts {
			account.Name = name
			return account
		}
	}
	if account, ok := accounts[name]; ok {
		account.Name = name
		return account
	}

	return UserAccount{Name: name}
}

// userAccountFromEnv returns a UserAccount from the current environment.
func userAccountFromEnv(name string) (UserAccount, error) {
	var envKeyFound bool

	var accounts map[string]UserAccount
	if creds, ok := os.LookupEnv("RUBRIK_POLARIS_ACCOUNT_CREDENTIALS"); ok {
		if err := json.Unmarshal([]byte(creds), &accounts); err != nil {
			return UserAccount{}, fmt.Errorf("failed to unmarshal RUBRIK_POLARIS_ACCOUNT_CREDENTIALS: %s", err)
		}
		envKeyFound = true
	}

	if envName, ok := os.LookupEnv("RUBRIK_POLARIS_ACCOUNT_NAME"); ok {
		name = envName
		envKeyFound = true
	}
	account := lookupUserAccount(name, accounts)
	if v, ok := os.LookupEnv("RUBRIK_POLARIS_ACCOUNT_USERNAME"); ok {
		account.Username = v
		envKeyFound = true
	}
	if v, ok := os.LookupEnv("RUBRIK_POLARIS_ACCOUNT_PASSWORD"); ok {
		account.Password = v
		envKeyFound = true
	}
	if v, ok := os.LookupEnv("RUBRIK_POLARIS_ACCOUNT_URL"); ok {
		account.URL = v
		envKeyFound = true
	}

	if !envKeyFound {
		return UserAccount{}, fmt.Errorf("failed to read user account from env: %w", graphql.ErrNotFound)
	}

	return account, nil
}

// UserAccountFromEnv returns a new UserAccount from the current environment.
// The account can be stored as a single JSON encoded environment variable
// (RUBRIK_POLARIS_ACCOUNT_CREDENTIALS) or as multiple plain text environment
// variables (e.g. name, username, etc). When using a single environment
// variable, the JSON content should have the following structure:
//
//	{
//		"<account-name>": {
//			"username": "<username>",
//			"password": "<password>"
//		}
//	}
//
// Or:
//
//	{
//		"<name-1>": {
//			"username": "<username-1>",
//			"password": "<password-1>",
//			"url": "https://<account-name>.my.rubrik.com/api"
//		},
//		"<name-2>": {
//			"username": "<username-2>",
//			"password": "<password-2>",
//			"url": "https://<account-name>.my.rubrik.com/api"
//		}
//	}
//
// The later format is used to hold multiple accounts. The environment variable
// RUBRIK_POLARIS_ACCOUNT_NAME specifies which account to use.
//
// When using multiple environment variables, they must have the same name as
// the public UserAccount fields but be all upper case and prepended with
// RUBRIK_POLARIS_ACCOUNT, e.g. RUBRIK_POLARIS_ACCOUNT_NAME.
//
// Note that RSC user accounts with MFA enabled cannot be used.
func UserAccountFromEnv() (*UserAccount, error) {
	account, err := userAccountFromEnv("")
	if err != nil {
		return nil, err
	}
	account.envOverride = true

	// Validate.
	if account.Name == "" {
		return nil, errors.New("invalid user account name")
	}
	if account.Username == "" {
		return nil, errors.New("invalid user account username")
	}
	if account.Password == "" {
		return nil, errors.New("invalid user account password")
	}

	return &account, nil
}

// userAccountFromFile returns a UserAccount from the specified file with the
// given name.
func userAccountFromFile(file, name string) (UserAccount, error) {
	expFile, err := expandPath(file)
	if err != nil {
		return UserAccount{}, fmt.Errorf("failed to expand file path: %s", err)
	}

	buf, err := os.ReadFile(expFile)
	if err != nil {
		return UserAccount{}, fmt.Errorf("failed to read user account file: %s", err)
	}

	var accounts map[string]UserAccount
	if err := json.Unmarshal(buf, &accounts); err != nil {
		return UserAccount{}, fmt.Errorf("failed to unmarshal user account file: %s", err)
	}

	account, ok := accounts[name]
	if !ok {
		return UserAccount{}, fmt.Errorf("failed to lookup user account %q: %w", name, graphql.ErrNotFound)
	}
	account.Name = name

	return account, nil
}

// UserAccountFromFile returns a new UserAccount read from the specified file.
// The file must be in the JSON format and the attributes must have the same
// name as the public UserAccount fields but be all lower case. Note that the
// name field is used as a key for the JSON object. E.g:
//
//	{
//		"<account-name>": {
//			"username": "<username>",
//			"password": "<password>"
//		}
//	}
//
// Or:
//
//	{
//		"<name-1>": {
//			"username": "<username-1>",
//			"password": "<password-1>",
//			"url": "https://<account-name>.my.rubrik.com/api"
//		},
//		"<name-2>": {
//			"username": "<username-2>",
//			"password": "<password-2>",
//			"url": "https://<account-name>.my.rubrik.com/api"
//		}
//	}
//
// The later format is used to hold multiple accounts. The URL field is
// optional, if it is skipped the URL is constructed from the account name.
//
// If allowEnvOverride is true, environment variables can be used to override
// user information in the file. See AccountFromEnv for details. In addition,
// the environment variable RUBRIK_POLARIS_ACCOUNT_FILE can be used to override
// the file that the user information is read from.
//
// Note that RSC user accounts with MFA enabled cannot be used.
func UserAccountFromFile(file, name string, allowEnvOverride bool) (*UserAccount, error) {
	var envAccount UserAccount
	if allowEnvOverride {
		var err error
		envAccount, err = userAccountFromEnv(name)
		if err != nil && !errors.Is(err, graphql.ErrNotFound) {
			return nil, err
		}

		if envAccount.Name != "" {
			name = envAccount.Name
		}
		if envFile, ok := os.LookupEnv("RUBRIK_POLARIS_ACCOUNT_FILE"); ok {
			file = envFile
		}
	}

	// Ignore errors for now since they might be corrected by what's in the
	// current environment.
	account, fileErr := userAccountFromFile(file, name)
	account.envOverride = allowEnvOverride

	// Merge with current environment.
	if allowEnvOverride {
		if envAccount.Name != "" {
			account.Name = envAccount.Name
		}
		if envAccount.Username != "" {
			account.Username = envAccount.Username
		}
		if envAccount.Password != "" {
			account.Password = envAccount.Password
		}
		if envAccount.URL != "" {
			account.URL = envAccount.URL
		}
	}

	// Validate.
	var msg string
	switch {
	case account.Name == "":
		msg = "invalid user account name"
	case account.Username == "":
		msg = "invalid user account username"
	case account.Password == "":
		msg = "invalid user account password"
	}
	if msg != "" {
		if fileErr != nil {
			msg = fmt.Sprintf("%s (user account file error: %w)", msg, fileErr)
		}
		return nil, errors.New(msg)
	}

	return &account, nil
}

// DefaultUserAccount returns a new UserAccount read from the default account
// file.
//
// If allowEnvOverride is true environment variables can be used to override
// user information in the file. See AccountFromEnv for details. In addition,
// the environment variable RUBRIK_POLARIS_ACCOUNT_FILE can be used to override
// the file that the user information is read from.
//
// Note that RSC user accounts with MFA enabled cannot be used.
func DefaultUserAccount(name string, allowEnvOverride bool) (*UserAccount, error) {
	return UserAccountFromFile(DefaultLocalUserFile, name, allowEnvOverride)
}

// ServiceAccount holds a Polaris ServiceAccount configuration.
type ServiceAccount struct {
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	Name           string `json:"name"`
	AccessTokenURI string `json:"access_token_uri"`

	envOverride bool
}

func (a *ServiceAccount) allowEnvOverride() bool {
	return a.envOverride
}

// serviceAccountFromEnv returns a ServiceAccount from the current environment.
func serviceAccountFromEnv() (ServiceAccount, error) {
	var envKeyFound bool

	var account ServiceAccount
	if creds, ok := os.LookupEnv("RUBRIK_POLARIS_SERVICEACCOUNT_CREDENTIALS"); ok {
		if err := json.Unmarshal([]byte(creds), &account); err != nil {
			return ServiceAccount{}, fmt.Errorf("failed to unmarshal RUBRIK_POLARIS_SERVICEACCOUNT_CREDENTIALS: %s", err)
		}
		envKeyFound = true
	}

	if v, ok := os.LookupEnv("RUBRIK_POLARIS_SERVICEACCOUNT_NAME"); ok {
		account.Name = v
		envKeyFound = true
	}
	if v, ok := os.LookupEnv("RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTID"); ok {
		account.ClientID = v
		envKeyFound = true
	}
	if v, ok := os.LookupEnv("RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTSECRET"); ok {
		account.ClientSecret = v
		envKeyFound = true
	}
	if v, ok := os.LookupEnv("RUBRIK_POLARIS_SERVICEACCOUNT_ACCESSTOKENURI"); ok {
		account.AccessTokenURI = v
		envKeyFound = true
	}

	if !envKeyFound {
		return ServiceAccount{}, fmt.Errorf("failed to read service account from env: %w", graphql.ErrNotFound)
	}

	return account, nil
}

// ServiceAccountFromEnv returns a new ServiceAccount from the current
// environment. The account can be stored as a single environment variable
// (RUBRIK_POLARIS_SERVICEACCOUNT_CREDENTIALS) or as multiple environment
// variables. When using a single environment variable, the content should be
// the RSC service account file downloaded from RSC when creating the service
// account. When using multiple environment variables, they must have the same
// name as the public ServiceAccount fields but be all upper case and prepended
// with RUBRIK_POLARIS_SERVICEACCOUNT, e.g. RUBRIK_POLARIS_SERVICEACCOUNT_NAME.
func ServiceAccountFromEnv() (*ServiceAccount, error) {
	account, err := serviceAccountFromEnv()
	if err != nil {
		return nil, err
	}

	// Validate.
	if account.Name == "" {
		return nil, errors.New("invalid service account name")
	}
	if account.ClientID == "" {
		return nil, errors.New("invalid service account client id")
	}
	if account.ClientSecret == "" {
		return nil, errors.New("invalid service account client secret")
	}
	if account.AccessTokenURI == "" {
		return nil, errors.New("invalid service account access token uri")
	}

	return &account, nil
}

// serviceAccountFromFile returns a ServiceAccount from the specified RSC
// service account file.
func serviceAccountFromFile(file string) (ServiceAccount, error) {
	expFile, err := expandPath(file)
	if err != nil {
		return ServiceAccount{}, fmt.Errorf("failed to expand file path: %s", err)
	}

	buf, err := os.ReadFile(expFile)
	if err != nil {
		return ServiceAccount{}, fmt.Errorf("failed to read service account file: %s", err)
	}

	var account ServiceAccount
	if err := json.Unmarshal(buf, &account); err != nil {
		return ServiceAccount{}, fmt.Errorf("failed to unmarshal service account: %s", err)
	}

	return account, nil
}

// ServiceAccountFromFile returns a new ServiceAccount read from the specified
// RSC service account file.
//
// If allowEnvOverride is true environment variables can be used to override
// account information in the file. See ServiceAccountFromEnv for details. In
// addition, the environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE can be
// used to override the file that the service account is read from.
func ServiceAccountFromFile(file string, allowEnvOverride bool) (*ServiceAccount, error) {
	var envAccount ServiceAccount
	if allowEnvOverride {
		var err error
		envAccount, err = serviceAccountFromEnv()
		if err != nil && !errors.Is(err, graphql.ErrNotFound) {
			return nil, err
		}

		if envFile, ok := os.LookupEnv("RUBRIK_POLARIS_SERVICEACCOUNT_FILE"); ok {
			file = envFile
		}
	}

	// Ignore errors for now since they might be corrected by what's in the
	// current environment.
	account, fileErr := serviceAccountFromFile(file)
	account.envOverride = allowEnvOverride

	// Merge with shell environment.
	if allowEnvOverride {
		if envAccount.Name != "" {
			account.Name = envAccount.Name
		}
		if envAccount.ClientID != "" {
			account.ClientID = envAccount.ClientID
		}
		if envAccount.ClientSecret != "" {
			account.ClientSecret = envAccount.ClientSecret
		}
		if envAccount.AccessTokenURI != "" {
			account.AccessTokenURI = envAccount.AccessTokenURI
		}
	}

	// Validate.
	var msg string
	switch {
	case account.Name == "":
		msg = "invalid service account name"
	case account.ClientID == "":
		msg = "invalid service account client id"
	case account.ClientSecret == "":
		msg = "invalid service account client secret"
	case account.AccessTokenURI == "":
		msg = "invalid service account access token uri"
	}
	if msg != "" {
		if fileErr != nil {
			msg = fmt.Sprintf("%s (service account file error: %s)", msg, fileErr)
		}
		return nil, errors.New(msg)
	}

	return &account, nil
}

// DefaultServiceAccount returns a new ServiceAccount read from the RSC service
// account file at the default service account location.
//
// If allowEnvOverride is true environment variables can be used to override
// account information in the file. See ServiceAccountFromEnv for details. In
// addition, the environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE can be
// used to override the file that the service account is read from.
func DefaultServiceAccount(allowEnvOverride bool) (*ServiceAccount, error) {
	return ServiceAccountFromFile(DefaultServiceAccountFile, allowEnvOverride)
}

func expandPath(file string) (string, error) {
	// Expand the ~ token to the user's home directory.
	if homeToken := fmt.Sprintf("~%c", filepath.Separator); strings.HasPrefix(file, homeToken) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		file = filepath.Join(home, strings.TrimPrefix(file, homeToken))
	}

	// Expand environment variables and make sure that the path is absolute.
	var err error
	file, err = filepath.Abs(os.ExpandEnv(file))
	if err != nil {
		return "", err
	}

	return file, nil
}
