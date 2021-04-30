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
)

// Account holds the Polaris account configuration.
type Account struct {
	// Polaris account name.
	Name string

	// Polaris account username.
	Username string

	// Polaris account password.
	Password string

	// Optional Polaris API endpoint. Useful for running the SDK against a test
	// service. Defaults to https://{Account}.my.rubrik.com/api.
	URL string
}

// accountFromEnv returns an Account from the current environment.
func accountFromEnv() Account {
	return Account{
		Name:     os.Getenv("RUBRIK_POLARIS_ACCOUNT_NAME"),
		Username: os.Getenv("RUBRIK_POLARIS_ACCOUNT_USERNAME"),
		Password: os.Getenv("RUBRIK_POLARIS_ACCOUNT_PASSWORD"),
		URL:      os.Getenv("RUBRIK_POLARIS_ACCOUNT_URL"),
	}
}

// AccountFromEnv returns a new Accoount from the user's environment variables.
// Environment variables must have the same name as the Account fields but be
// all upper case and prepended with RUBRIK_POLARIS_ACCOUNT, e.g.
// RUBRIK_POLARIS_ACCOUNT_USERNAME.
func AccountFromEnv() (Account, error) {
	account := accountFromEnv()

	// Validate.
	if account.Name == "" {
		return Account{}, errors.New("polaris: invalid environment variable: RUBRIK_POLARIS_ACCOUNT_NAME")
	}
	if account.Username == "" {
		return Account{}, errors.New("polaris: invalid environment variable: RUBRIK_POLARIS_ACCOUNT_USERNAME")
	}
	if account.Password == "" {
		return Account{}, errors.New("polaris: invalid environment variable: RUBRIK_POLARIS_ACCOUNT_PASSWORD")
	}

	return account, nil
}

// accountFromFile returns an Account from the specified file with the given
// name.
func accountFromFile(file, name string) (Account, error) {
	if strings.HasPrefix(file, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return Account{}, err
		}
		file = filepath.Join(home, strings.TrimPrefix(file, "~/"))
	}

	buf, err := os.ReadFile(file)
	if err != nil {
		return Account{}, err
	}

	var accounts map[string]Account
	if err := json.Unmarshal(buf, &accounts); err != nil {
		return Account{}, err
	}

	account, ok := accounts[name]
	if !ok {
		return Account{}, fmt.Errorf("polaris: account %q not found in %q", name, file)
	}
	account.Name = name

	return account, nil
}

// AccountFromFile returns a new Account read from the specified file. Files
// must be in JSON format and the attributes must have the same name as the
// Account fields but be all lower case. Note that the Name field is used
// as a key for the JSON object. E.g:
//
//   {
//     "account-name-1": {
//       "username": "username-1",
//       "password": "password-1"
//     },
//     "account-name-2": {
//       "username": "username-2",
//       "password": "password-2",
//       "url": "https://polaris-url/api"
//     }
//   }
func AccountFromFile(file, name string) (Account, error) {
	account, err := accountFromFile(file, name)
	if err != nil {
		return Account{}, err
	}

	// Validate.
	if account.Name == "" {
		return Account{}, errors.New("polaris: invalid JSON attribute: name")
	}
	if account.Username == "" {
		return Account{}, errors.New("polaris: invalid JSON attribute: username")
	}
	if account.Password == "" {
		return Account{}, errors.New("polaris: invalid JSON attribute: password")
	}

	return account, nil
}

// DefaultAccount returns a new Account read from the default account file.
// Environment variables can be used to override user information in the file.
// See AccountFromEnv for details. In addition the environment variable
// RUBRIK_POLARIS_ACCOUNT_FILE can be used to override the file that the user
// information is read from.
func DefaultAccount(name string) (Account, error) {
	envAccount := accountFromEnv()

	// Override the given account name.
	if envAccount.Name != "" {
		name = envAccount.Name
	}

	file := DefaultLocalUserFile
	if envFile := os.Getenv("RUBRIK_POLARIS_ACCOUNT_FILE"); envFile != "" {
		file = envFile
	}

	// Ignore errors for now since they might be corrected by what's in the
	// environment.
	fileAccount, _ := accountFromFile(file, name)

	// Merge.
	if envAccount.Name != "" {
		fileAccount.Name = envAccount.Name
	}
	if envAccount.Username != "" {
		fileAccount.Username = envAccount.Username
	}
	if envAccount.Password != "" {
		fileAccount.Password = envAccount.Password
	}
	if envAccount.URL != "" {
		fileAccount.URL = envAccount.URL
	}

	// Validate.
	if fileAccount.Name == "" {
		return Account{}, errors.New("polaris: missing required field: Name")
	}
	if fileAccount.Username == "" {
		return Account{}, errors.New("polaris: missing required field: Username")
	}
	if fileAccount.Password == "" {
		return Account{}, errors.New("polaris: missing required field: Password")
	}

	return fileAccount, nil
}

// ServiceAccount holds the Polaris ServiceAccount configuration.
type ServiceAccount struct {
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	Name           string `json:"name"`
	AccessTokenURI string `json:"access_token_uri"`
}

// serviceAccountFromEnv returns a ServiceAccount from the current environment.
func serviceAccountFromEnv() ServiceAccount {
	return ServiceAccount{
		Name:           os.Getenv("RUBRIK_POLARIS_SERVICEACCOUNT_NAME"),
		ClientID:       os.Getenv("RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTID"),
		ClientSecret:   os.Getenv("RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTSECRET"),
		AccessTokenURI: os.Getenv("RUBRIK_POLARIS_SERVICEACCOUNT_ACCESSTOKENURI"),
	}
}

// ServiceAccountFromEnv returns a new ServiceAccount from the user's
// environment variables. Environment variables must have the same name as the
// ServiceAccount fields but be all upper case and prepended with
// RUBRIK_POLARIS_SERVICEACCOUNT, e.g. RUBRIK_POLARIS_SERVICEACCOUNT_NAME.
func ServiceAccountFromEnv() (ServiceAccount, error) {
	account := serviceAccountFromEnv()

	// Validate.
	if account.Name == "" {
		return ServiceAccount{}, errors.New("polaris: missing environment variable: RUBRIK_POLARIS_SERVICEACCOUNT_NAME")
	}
	if account.ClientID == "" {
		return ServiceAccount{}, errors.New("polaris: missing environment variable: RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTID")
	}
	if account.ClientSecret == "" {
		return ServiceAccount{}, errors.New("polaris: missing environment variable: RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTSECRET")
	}
	if account.AccessTokenURI == "" {
		return ServiceAccount{}, errors.New("polaris: missing environment variable: RUBRIK_POLARIS_SERVICEACCOUNT_ACCESSTOKENURI")
	}

	return account, nil
}

// serviceAccountFromFile returns a ServiceAccount from the specified file.
func serviceAccountFromFile(file string) (ServiceAccount, error) {
	if strings.HasPrefix(file, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return ServiceAccount{}, err
		}
		file = filepath.Join(home, strings.TrimPrefix(file, "~/"))
	}

	buf, err := os.ReadFile(file)
	if err != nil {
		return ServiceAccount{}, err
	}

	var account ServiceAccount
	if err := json.Unmarshal(buf, &account); err != nil {
		return ServiceAccount{}, err
	}

	return account, nil
}

// ServiceAccountFromFile returns a new ServiceAccount read from the specified
// file. Files must be in JSON format and the attributes must have the same
// name as the ServiceAccount fields but be all lower case and have words
// separated by underscores.
func ServiceAccountFromFile(file string) (ServiceAccount, error) {
	account, err := serviceAccountFromFile(file)
	if err != nil {
		return ServiceAccount{}, err
	}

	// Validate.
	if account.Name == "" {
		return ServiceAccount{}, errors.New("polaris: missing JSON attribute: name")
	}
	if account.ClientID == "" {
		return ServiceAccount{}, errors.New("polaris: missing JSON attribute: client_id")
	}
	if account.ClientSecret == "" {
		return ServiceAccount{}, errors.New("polaris: missing JSON attribute: client_secret")
	}
	if account.AccessTokenURI == "" {
		return ServiceAccount{}, errors.New("polaris: missing JSON attribute: access_token_uri")
	}

	return account, nil
}

// DefaultServiceAccount returns a new ServiceAccount read from the default
// service account file. Environment variables can be used to override account
// information in the file. See ServiceAccountFromEnv for details. In addition,
// the environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE can be used to
// override the file that the service account is read from.
func DefaultServiceAccount() (ServiceAccount, error) {
	envAccount := serviceAccountFromEnv()

	file := DefaultServiceAccountFile
	if envFile := os.Getenv("RUBRIK_POLARIS_SERVICEACCOUNT_FILE"); envFile != "" {
		file = envFile
	}

	// Ignore errors for now since they might be corrected by what's in the
	// environment.
	fileAccount, _ := serviceAccountFromFile(file)

	// Merge.
	if envAccount.Name != "" {
		fileAccount.Name = envAccount.Name
	}
	if envAccount.ClientID != "" {
		fileAccount.ClientID = envAccount.ClientID
	}
	if envAccount.ClientSecret != "" {
		fileAccount.ClientSecret = envAccount.ClientSecret
	}
	if envAccount.AccessTokenURI != "" {
		fileAccount.AccessTokenURI = envAccount.AccessTokenURI
	}

	// Validate.
	if fileAccount.Name == "" {
		return ServiceAccount{}, errors.New("polaris: missing required field: Name")
	}
	if fileAccount.ClientID == "" {
		return ServiceAccount{}, errors.New("polaris: missing required field: ClientID")
	}
	if fileAccount.ClientSecret == "" {
		return ServiceAccount{}, errors.New("polaris: missing required field: ClientSecret")
	}
	if fileAccount.AccessTokenURI == "" {
		return ServiceAccount{}, errors.New("polaris: missing required field: AccessTokenURI")
	}

	return fileAccount, nil
}
