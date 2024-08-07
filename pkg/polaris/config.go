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
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultLocalUserFile path to the default local users file.
	DefaultLocalUserFile = "~/.rubrik/polaris-accounts.json"

	// DefaultServiceAccountFile path to the default service account file.
	DefaultServiceAccountFile = "~/.rubrik/polaris-service-account.json"

	// UserAccount environment variables.
	keyUserAccountCredentials = "RUBRIK_POLARIS_ACCOUNT_CREDENTIALS"
	keyUserAccountFile        = "RUBRIK_POLARIS_ACCOUNT_FILE"
	keyUserAccountName        = "RUBRIK_POLARIS_ACCOUNT_NAME"
	keyUserAccountPassword    = "RUBRIK_POLARIS_ACCOUNT_PASSWORD"
	keyUserAccountURL         = "RUBRIK_POLARIS_ACCOUNT_URL"
	keyUserAccountUsername    = "RUBRIK_POLARIS_ACCOUNT_USERNAME"

	// ServiceAccount environment variables.
	keyServiceAccountAccessTokenURI = "RUBRIK_POLARIS_SERVICEACCOUNT_ACCESSTOKENURI"
	keyServiceAccountClientID       = "RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTID"
	keyServiceAccountClientSecret   = "RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTSECRET"
	keyServiceAccountCredentials    = "RUBRIK_POLARIS_SERVICEACCOUNT_CREDENTIALS"
	keyServiceAccountFile           = "RUBRIK_POLARIS_SERVICEACCOUNT_FILE"
	keyServiceAccountName           = "RUBRIK_POLARIS_SERVICEACCOUNT_NAME"
)

var (
	// errAccountNotFound signals that no trace of the account was found.
	errAccountNotFound accountNotFound
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
	cacheKeyMaterial() string
	cacheSuffixMaterial() string
}

// FindAccount looks for a valid Account using the passed in credentials string
// with the following algorithm:
//
// When credentials string contains a value:
//
//  1. Try and read the service account by interpreting the credentials string as
//     a service account.
//
//  2. Try and read the service account by interpreting the credentials string as
//     a path to a file holding the service account.
//
//  3. If the default user account file exist, try and read the user account from
//     the file by interpreting the credentials string as the user account name.
//
// When the credentials string is empty:
//
//  1. If the default service account file exist, try and read the service
//     account from the file.
//
//  2. If env contains a service account, try and read the service account from
//     env.
//
// Note, when reading an account from file, environment variables can be used to
// override account information.
func FindAccount(credentials string, allowEnvOverride bool) (Account, error) {
	if credentials == "" {
		serviceAccount, err := DefaultServiceAccount(allowEnvOverride)
		if err == nil {
			return serviceAccount, nil
		}
		if !errors.Is(err, errAccountNotFound) {
			return nil, fmt.Errorf("failed to load service account from default file: %s", err)
		}

		serviceAccount, err = ServiceAccountFromEnv()
		if err == nil {
			return serviceAccount, nil
		}
		if !errors.Is(err, errAccountNotFound) {
			return nil, fmt.Errorf("failed to load service account from env: %s", err)
		}

		return nil, errors.New("failed to load account, searched: default service account file and env")
	}

	serviceAccount, err := ServiceAccountFromText(credentials, allowEnvOverride)
	if err == nil {
		return serviceAccount, nil
	}
	if !errors.Is(err, errAccountNotFound) {
		return nil, fmt.Errorf("failed to load service account from text: %s", err)
	}

	serviceAccount, err = ServiceAccountFromFile(credentials, allowEnvOverride)
	if err == nil {
		return serviceAccount, nil
	}
	if !errors.Is(err, errAccountNotFound) {
		return nil, fmt.Errorf("failed to load service account from file: %s", err)
	}

	userAccount, err := DefaultUserAccount(credentials, allowEnvOverride)
	if err == nil {
		return userAccount, nil
	}
	if !errors.Is(err, errAccountNotFound) {
		return nil, fmt.Errorf("failed to load user account from default file: %s", err)
	}

	return nil, errors.New("failed to load account, searched: passed in credentials, default service account file and default user account file")
}

// UserAccount holds an RSC local user account configuration. Depending on how
// the local user account is stored, the Name field might hold the RSC account
// name.
//
// Note, RSC user accounts with MFA enabled cannot be used.
type UserAccount struct {
	Name     string // User account name.
	Username string // RSC account username.
	Password string // RSC account password.

	// Optional RSC API endpoint. Useful for running the SDK against a test
	// service. When omitted, it defaults to https://{Name}.my.rubrik.com/api.
	URL string

	accountName string
	accountFQDN string
	apiURL      string
	envOverride bool
	tokenURL    string
}

// AccountName returns the RSC account name. Note, this might not be the same
// as the name of the UserAccount.
func (a *UserAccount) AccountName() string {
	return a.accountName
}

// AccountFQDN returns the fully qualified domain name of the RSC account.
func (a *UserAccount) AccountFQDN() string {
	return a.accountFQDN
}

// APIURL returns the RSC account API URL.
func (a *UserAccount) APIURL() string {
	return a.apiURL
}

// TokenURL returns the RSC account token URL.
func (a *UserAccount) TokenURL() string {
	return a.tokenURL
}

func (a *UserAccount) allowEnvOverride() bool {
	return a.envOverride
}

func (a *UserAccount) cacheKeyMaterial() string {
	return a.Name + a.URL + a.Username + a.Password
}

func (a *UserAccount) cacheSuffixMaterial() string {
	return a.Name + a.Username
}

// DefaultUserAccount returns a new UserAccount read from the default account
// file.
//
// If allowEnvOverride is true environment variables can be used to override
// user information in the file. See UserAccountFromEnv for details.
// In addition, the environment variable RUBRIK_POLARIS_ACCOUNT_FILE can be used
// to override the file that the user information is read from.
//
// Note that RSC user accounts with MFA enabled cannot be used.
func DefaultUserAccount(name string, allowEnvOverride bool) (*UserAccount, error) {
	return UserAccountFromFile(DefaultLocalUserFile, name, allowEnvOverride)
}

// UserAccountFromEnv returns a new UserAccount from the current environment.
// The account can be stored as a single JSON encoded environment variable
// (RUBRIK_POLARIS_ACCOUNT_CREDENTIALS) or as multiple plain text environment
// variables (e.g. name, username, etc.). When using a single environment
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
	account := userAccountFromEnv("")
	account.envOverride = true

	if account.Name == "" && account.Username == "" && account.Password == "" && account.URL == "" {
		return nil, fmt.Errorf("%wuser account not found in env", errAccountNotFound)
	}

	if err := initUserAccount(&account); err != nil {
		return nil, err
	}

	return &account, nil
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
// user information in the file. See UserAccountFromEnv for details.
// In addition, the environment variable RUBRIK_POLARIS_ACCOUNT_FILE can be used
// to override the file that the user information is read from.
//
// Note that RSC user accounts with MFA enabled cannot be used.
func UserAccountFromFile(file, name string, allowEnvOverride bool) (*UserAccount, error) {
	var envAccount UserAccount
	if allowEnvOverride {
		envAccount = userAccountFromEnv(name)
		if envAccount.Name != "" {
			name = envAccount.Name
		}

		if val := os.Getenv(keyUserAccountFile); val != "" {
			file = val
		}
	}

	account, err := userAccountFromFile(file, name)
	if err != nil {
		return nil, err
	}

	account.envOverride = allowEnvOverride
	if allowEnvOverride {
		overrideUserAccount(&account, envAccount)
	}

	if err := initUserAccount(&account); err != nil {
		return nil, err
	}

	return &account, nil
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

// userAccountFromEnv returns the named UserAccount from the current
// environment.
func userAccountFromEnv(name string) UserAccount {
	var accounts map[string]UserAccount
	if val := os.Getenv(keyUserAccountCredentials); val != "" {
		var credAccounts map[string]UserAccount
		if err := json.Unmarshal([]byte(val), &credAccounts); err == nil {
			accounts = credAccounts
		}
	}
	if val := os.Getenv(keyUserAccountName); val != "" {
		name = val
	}

	account := lookupUserAccount(name, accounts)
	if val := os.Getenv(keyUserAccountUsername); val != "" {
		account.Username = val
	}
	if val := os.Getenv(keyUserAccountPassword); val != "" {
		account.Password = val
	}
	if val := os.Getenv(keyUserAccountURL); val != "" {
		account.URL = val
	}

	return account
}

// userAccountFromFile returns the named UserAccount from the specified file.
func userAccountFromFile(file, name string) (UserAccount, error) {
	expFile, err := expandPath(file)
	if err != nil {
		return UserAccount{}, fmt.Errorf("failed to expand user account file path: %s", err)
	}

	// Stat the file to determine if it exists, stat doesn't require access
	// permissions on the file.
	if info, err := os.Stat(expFile); err != nil || info.IsDir() {
		if err == nil {
			err = fmt.Errorf("user account file is a directory")
		}
		return UserAccount{}, fmt.Errorf("%wfailed to access user account file: %s", errAccountNotFound, err)
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
		return UserAccount{}, fmt.Errorf("user account %q not found in user account file: %s", name, expFile)
	}
	account.Name = name

	return account, nil
}

// initUserAccount validates the user account data and initializes the
// additional fields.
func initUserAccount(account *UserAccount) error {
	if account.Name == "" {
		return errors.New("invalid user account name")
	}
	if account.Username == "" {
		return errors.New("invalid user account username")
	}
	if account.Password == "" {
		return errors.New("invalid user account password")
	}
	if account.URL == "" {
		account.URL = fmt.Sprintf("https://%s.my.rubrik.com/api", account.Name)
	}

	// Derive fields.
	u, err := url.ParseRequestURI(account.URL)
	if err != nil {
		return fmt.Errorf("invalid user account url: %s", err)
	}
	fqdn := u.Hostname()
	i := strings.Index(fqdn, ".")
	if i == -1 {
		return errors.New("invalid url: no account name found")
	}
	account.accountName = fqdn[:i]
	account.accountFQDN = fqdn
	account.apiURL = account.URL
	account.tokenURL = account.URL + "/session"

	return nil
}

// overrideUserAccount overrides the fields of the UserAccount with valid
// information from the other UserAccount.
func overrideUserAccount(account *UserAccount, other UserAccount) {
	if other.Name != "" {
		account.Name = other.Name
	}
	if other.Username != "" {
		account.Username = other.Username
	}
	if other.Password != "" {
		account.Password = other.Password
	}
	if other.URL != "" {
		account.URL = other.URL
	}
}

// ServiceAccount holds an RSC ServiceAccount configuration. The Name field
// holds the name of the service account and not the name of the RSC account.
type ServiceAccount struct {
	ClientID       string `json:"client_id"`        // Client ID.
	ClientSecret   string `json:"client_secret"`    // Client secret.
	Name           string `json:"name"`             // Service account name.
	AccessTokenURI string `json:"access_token_uri"` // Access token URI.

	accountName string
	accountFQDN string
	apiURL      string
	envOverride bool
	tokenURL    string
}

// AccountName returns the RSC account name. Note, this might not be the same
// as the name of the ServiceAccount.
func (a *ServiceAccount) AccountName() string {
	return a.accountName
}

// AccountFQDN returns the fully qualified domain name of the RSC account.
func (a *ServiceAccount) AccountFQDN() string {
	return a.accountFQDN
}

// APIURL returns the RSC account API URL.
func (a *ServiceAccount) APIURL() string {
	return a.apiURL
}

// TokenURL returns the RSC account token URL.
func (a *ServiceAccount) TokenURL() string {
	return a.tokenURL
}

func (a *ServiceAccount) allowEnvOverride() bool {
	return a.envOverride
}

func (a *ServiceAccount) cacheKeyMaterial() string {
	return a.Name + a.AccessTokenURI + a.ClientID + a.ClientSecret
}

func (a *ServiceAccount) cacheSuffixMaterial() string {
	return a.Name + a.ClientID
}

// DefaultServiceAccount returns a new ServiceAccount read from the RSC service
// account file at the default service account location.
//
// If allowEnvOverride is true, environment variables can be used to override
// account information in the file. See ServiceAccountFromEnv for details. In
// addition, the environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE can be
// used to override the file that the service account is read from.
func DefaultServiceAccount(allowEnvOverride bool) (*ServiceAccount, error) {
	return ServiceAccountFromFile(DefaultServiceAccountFile, allowEnvOverride)
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
	account := serviceAccountFromEnv()
	account.envOverride = true

	if account.Name == "" && account.ClientID == "" && account.ClientSecret == "" && account.AccessTokenURI == "" {
		return nil, fmt.Errorf("%wservice account not found in env", errAccountNotFound)
	}

	if err := initServiceAccount(&account); err != nil {
		return nil, err
	}

	return &account, nil
}

// ServiceAccountFromFile returns a new ServiceAccount read from the specified
// RSC service account file.
//
// If allowEnvOverride is true environment variables can be used to override
// account information in the file. See ServiceAccountFromEnv for details. In
// addition, the environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE can be
// used to override the file that the service account is read from.
func ServiceAccountFromFile(file string, allowEnvOverride bool) (*ServiceAccount, error) {
	if allowEnvOverride {
		if val := os.Getenv(keyServiceAccountFile); val != "" {
			file = val
		}
	}

	account, err := serviceAccountFromFile(file)
	if err != nil {
		return nil, err
	}

	account.envOverride = allowEnvOverride
	if allowEnvOverride {
		overrideServiceAccount(&account, serviceAccountFromEnv())
	}

	if err := initServiceAccount(&account); err != nil {
		return nil, err
	}

	return &account, nil

}

// ServiceAccountFromText returns a new ServiceAccount read from the specified
// text containing an RSC service account.
//
// If allowEnvOverride is true environment variables can be used to override
// account information in the file. See ServiceAccountFromEnv for details. In
// addition, the environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE can be
// used to override the file that the service account is read from.
func ServiceAccountFromText(text string, allowEnvOverride bool) (*ServiceAccount, error) {
	account, err := serviceAccountFromString(text)
	if err != nil {
		return nil, err
	}

	account.envOverride = allowEnvOverride
	if allowEnvOverride {
		overrideServiceAccount(&account, serviceAccountFromEnv())
	}

	if err := initServiceAccount(&account); err != nil {
		return nil, err
	}

	return &account, nil
}

// serviceAccountFromEnv returns a ServiceAccount from the current environment.
func serviceAccountFromEnv() ServiceAccount {
	var account ServiceAccount
	if val := os.Getenv(keyServiceAccountCredentials); val != "" {
		var credAccount ServiceAccount
		if err := json.Unmarshal([]byte(val), &credAccount); err == nil {
			account = credAccount
		}
	}

	if val := os.Getenv(keyServiceAccountName); val != "" {
		account.Name = val
	}
	if val := os.Getenv(keyServiceAccountClientID); val != "" {
		account.ClientID = val
	}
	if val := os.Getenv(keyServiceAccountClientSecret); val != "" {
		account.ClientSecret = val
	}
	if val := os.Getenv(keyServiceAccountAccessTokenURI); val != "" {
		account.AccessTokenURI = val
	}

	return account
}

// serviceAccountFromFile returns a ServiceAccount from the specified service
// account file.
func serviceAccountFromFile(file string) (ServiceAccount, error) {
	expFile, err := expandPath(file)
	if err != nil {
		return ServiceAccount{}, fmt.Errorf("failed to expand service account file path: %s", err)
	}

	// Stat the file to determine if it exists, stat doesn't require access
	// permissions on the file.
	if info, err := os.Stat(expFile); err != nil || info.IsDir() {
		if err == nil {
			err = fmt.Errorf("service account file is a directory")
		}
		return ServiceAccount{}, fmt.Errorf("%wfailed to access service account file: %s", errAccountNotFound, err)
	}
	buf, err := os.ReadFile(expFile)
	if err != nil {
		return ServiceAccount{}, fmt.Errorf("failed to read service account file: %s", err)
	}

	var account ServiceAccount
	if err := json.Unmarshal(buf, &account); err != nil {
		return ServiceAccount{}, fmt.Errorf("failed to unmarshal service account file: %s", err)
	}

	return account, nil
}

// serviceAccountFromString returns a ServiceAccount from the specified JSON
// encoded string.
func serviceAccountFromString(s string) (ServiceAccount, error) {
	var account ServiceAccount
	if err := json.Unmarshal([]byte(s), &account); err != nil {
		return ServiceAccount{}, fmt.Errorf("%wfailed to unmarshal service account text: %s", errAccountNotFound, err)
	}

	return account, nil
}

// initServiceAccount validates the service account data and initializes the
// additional fields.
func initServiceAccount(account *ServiceAccount) error {
	if account.Name == "" {
		return errors.New("invalid service account name")
	}
	if account.ClientID == "" {
		return errors.New("invalid service account client id")
	}
	if account.ClientSecret == "" {
		return errors.New("invalid service account client secret")
	}

	// Derive account name and FQDN.
	u, err := url.ParseRequestURI(account.AccessTokenURI)
	if err != nil {
		return fmt.Errorf("invalid service account access token uri: %s", err)
	}
	fqdn := u.Hostname()
	i := strings.Index(fqdn, ".")
	if i == -1 {
		return errors.New("invalid service account access token uri: no account name found")
	}
	account.accountName = fqdn[:i]
	account.accountFQDN = fqdn

	// Derive API URL and token URL.
	i = strings.LastIndex(account.AccessTokenURI, "/")
	if i < 0 {
		return errors.New("invalid service account access token uri: malformed path")
	}
	account.apiURL = account.AccessTokenURI[:i]
	account.tokenURL = account.AccessTokenURI

	return nil
}

// overrideServiceAccount overrides the fields of the ServiceAccount with valid
// information from the other ServiceAccount.
func overrideServiceAccount(account *ServiceAccount, other ServiceAccount) {
	if other.Name != "" {
		account.Name = other.Name
	}
	if other.ClientID != "" {
		account.ClientID = other.ClientID
	}
	if other.ClientSecret != "" {
		account.ClientSecret = other.ClientSecret
	}
	if other.AccessTokenURI != "" {
		account.AccessTokenURI = other.AccessTokenURI
	}
}

// accountNotFound is an error type that signals that an account was not found.
type accountNotFound struct{}

func (e accountNotFound) Error() string {
	return ""
}

func expandPath(file string) (string, error) {
	// Expand the ~ token to the user's home directory. This should never fail
	// unless the shell environment is broken.
	if homeToken := fmt.Sprintf("~%c", filepath.Separator); strings.HasPrefix(file, homeToken) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		file = filepath.Join(home, strings.TrimPrefix(file, homeToken))
	}

	// Expand environment variables and make sure that the path is absolute.
	// This should never fail unless the shell environment is broken.
	var err error
	file, err = filepath.Abs(os.ExpandEnv(file))
	if err != nil {
		return "", err
	}

	return file, nil
}
