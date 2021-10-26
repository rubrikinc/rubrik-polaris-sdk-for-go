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

// UserAccount holds a Polaris local user account configuration.
type UserAccount struct {
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

func (a *UserAccount) isAccount() {}

// userAccountFromEnv returns a UserAccount from the current environment.
func userAccountFromEnv() UserAccount {
	return UserAccount{
		Name:     strings.TrimSpace(os.Getenv("RUBRIK_POLARIS_ACCOUNT_NAME")),
		Username: strings.TrimSpace(os.Getenv("RUBRIK_POLARIS_ACCOUNT_USERNAME")),
		Password: strings.TrimSpace(os.Getenv("RUBRIK_POLARIS_ACCOUNT_PASSWORD")),
		URL:      strings.TrimSpace(os.Getenv("RUBRIK_POLARIS_ACCOUNT_URL")),
	}
}

// UserAccountFromEnv returns a new UserAccoount from the user's environment
// variables. Environment variables must have the same name as the UserAccount
// fields but be all upper case and prepended with RUBRIK_POLARIS_ACCOUNT, e.g.
// RUBRIK_POLARIS_ACCOUNT_USERNAME.
func UserAccountFromEnv() (*UserAccount, error) {
	account := userAccountFromEnv()

	// Validate.
	if account.Name == "" {
		return nil, errors.New("polaris: environment variable RUBRIK_POLARIS_ACCOUNT_NAME has an invalid value")
	}
	if account.Username == "" {
		return nil, errors.New("polaris: environment variable RUBRIK_POLARIS_ACCOUNT_USERNAME has an invalid value")
	}
	if account.Password == "" {
		return nil, errors.New("polaris: environment variable RUBRIK_POLARIS_ACCOUNT_PASSWORD has an invalid value")
	}

	return &account, nil
}

// userAccountFromFile returns a UserAccount from the specified file with the
// given name.
func userAccountFromFile(file, name string) (UserAccount, error) {
	if strings.HasPrefix(file, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return UserAccount{}, err
		}
		file = filepath.Join(home, strings.TrimPrefix(file, "~/"))
	}

	buf, err := os.ReadFile(file)
	if err != nil {
		return UserAccount{}, err
	}

	var accounts map[string]UserAccount
	if err := json.Unmarshal(buf, &accounts); err != nil {
		return UserAccount{}, err
	}

	account, ok := accounts[name]
	if !ok {
		return UserAccount{}, fmt.Errorf("polaris: local user account %q not found in %q", name, file)
	}
	account.Name = name

	return account, nil
}

// UserAccountFromFile returns a new UserAccount read from the specified file.
// Files must be in JSON format and the attributes must have the same name as
// the Account fields but be all lower case. Note that the Name field is used
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
//
// If allowEnvOverride is true environment variables can be used to override
// user information in the file. See AccountFromEnv for details. In addition
// the environment variable RUBRIK_POLARIS_ACCOUNT_FILE can be used to override
// the file that the user information is read from.
func UserAccountFromFile(file, name string, allowEnvOverride bool) (*UserAccount, error) {
	envAccount := userAccountFromEnv()

	if allowEnvOverride {
		if envAccount.Name != "" {
			name = envAccount.Name
		}

		if envFile := os.Getenv("RUBRIK_POLARIS_ACCOUNT_FILE"); envFile != "" {
			file = envFile
		}
	}

	// Ignore errors for now since they might be corrected by what's in the
	// shell environment.
	account, err := userAccountFromFile(file, name)

	// Merge with shell environment.
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

	// Validate, note that URL is optional.
	if strings.TrimSpace(account.Name) == "" {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("polaris: local user account has an invalid value for field Name")
	}
	if strings.TrimSpace(account.Username) == "" {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("polaris: local user account has an invalid value for field Username")
	}
	if strings.TrimSpace(account.Password) == "" {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("polaris: local user account has an invalid value for field Password")
	}

	return &account, nil
}

// DefaultUserAccount returns a new UserAccount read from the default account
// file. If allowEnvOverride is true environment variables can be used to
// override user information in the file. See AccountFromEnv for details. In
// addition the environment variable RUBRIK_POLARIS_ACCOUNT_FILE can be used to
// override the file that the user information is read from.
func DefaultUserAccount(name string, allowEnvOverride bool) (*UserAccount, error) {
	return UserAccountFromFile(DefaultLocalUserFile, name, allowEnvOverride)
}

// ServiceAccount holds a Polaris ServiceAccount configuration.
type ServiceAccount struct {
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	Name           string `json:"name"`
	AccessTokenURI string `json:"access_token_uri"`
}

func (a *ServiceAccount) isAccount() {}

// serviceAccountFromEnv returns a ServiceAccount from the current environment.
func serviceAccountFromEnv() ServiceAccount {
	return ServiceAccount{
		Name:           strings.TrimSpace(os.Getenv("RUBRIK_POLARIS_SERVICEACCOUNT_NAME")),
		ClientID:       strings.TrimSpace(os.Getenv("RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTID")),
		ClientSecret:   strings.TrimSpace(os.Getenv("RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTSECRET")),
		AccessTokenURI: strings.TrimSpace(os.Getenv("RUBRIK_POLARIS_SERVICEACCOUNT_ACCESSTOKENURI")),
	}
}

// ServiceAccountFromEnv returns a new ServiceAccount from the user's
// environment variables. Environment variables must have the same name as the
// ServiceAccount fields but be all upper case and prepended with
// RUBRIK_POLARIS_SERVICEACCOUNT, e.g. RUBRIK_POLARIS_SERVICEACCOUNT_NAME.
func ServiceAccountFromEnv() (*ServiceAccount, error) {
	account := serviceAccountFromEnv()

	// Validate.
	if account.Name == "" {
		return nil, errors.New("polaris: environment variable RUBRIK_POLARIS_SERVICEACCOUNT_NAME has an invalid value")
	}
	if account.ClientID == "" {
		return nil, errors.New("polaris: environment variable RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTID has an invalid value")
	}
	if account.ClientSecret == "" {
		return nil, errors.New("polaris: environment variable RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTSECRET has an invalid value")
	}
	if account.AccessTokenURI == "" {
		return nil, errors.New("polaris: environment variable RUBRIK_POLARIS_SERVICEACCOUNT_ACCESSTOKENURI has an invalid value")
	}

	return &account, nil
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
// If allowEnvOverride is true environment variables can be used to override
// account information in the file. See ServiceAccountFromEnv for details. In
// addition, the environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE can be
// used to override the file that the service account is read from.
func ServiceAccountFromFile(file string, allowEnvOverride bool) (*ServiceAccount, error) {
	envAccount := serviceAccountFromEnv()

	if allowEnvOverride {
		if envFile := os.Getenv("RUBRIK_POLARIS_SERVICEACCOUNT_FILE"); envFile != "" {
			file = envFile
		}
	}

	// Ignore errors for now since they might be corrected by what's in the
	// shell environment.
	account, err := serviceAccountFromFile(file)

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
	if strings.TrimSpace(account.Name) == "" {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("polaris: service account has an invalid value for field Name")
	}
	if strings.TrimSpace(account.ClientID) == "" {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("polaris: service account has an invalid value for field ClientID")
	}
	if strings.TrimSpace(account.ClientSecret) == "" {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("polaris: service account has an invalid value for field ClientSecret")
	}
	if strings.TrimSpace(account.AccessTokenURI) == "" {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("polaris: service account has an invalid value for field AccessTokenURI")
	}

	return &account, nil
}

// DefaultServiceAccount returns a new ServiceAccount read from the default
// service account file.
// If allowEnvOverride is true environment variables can be used to override
// account information in the file. See ServiceAccountFromEnv for details. In
// addition, the environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE can be
// used to override the file that the service account is read from.
func DefaultServiceAccount(allowEnvOverride bool) (*ServiceAccount, error) {
	return ServiceAccountFromFile(DefaultServiceAccountFile, allowEnvOverride)
}
