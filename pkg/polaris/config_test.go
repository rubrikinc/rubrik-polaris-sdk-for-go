package polaris

import (
	"os"
	"path/filepath"
	"testing"
)

func skipOnEnvs(t *testing.T, keys ...string) {
	for _, key := range keys {
		if _, ok := os.LookupEnv(key); ok {
			t.Skipf("Environment variable %q defined", key)
		}
	}
}

func TestUserAccountFromFile(t *testing.T) {
	path, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	file := filepath.Join(path, "config.json")
	data := []byte(`{
		"account": {
			"username": "username",
			"password": "password",
			"url": "https://my-account.dev.my.rubrik-lab.com/api"
		}
	}`)
	if err := os.WriteFile(file, data, 0666); err != nil {
		t.Fatal(err)
	}

	// Test with non-existing file.
	if _, err := UserAccountFromFile("some-non-existing-file", "my-account", false); err == nil {
		t.Fatal("UserAccountFromFile should fail with non-existing file")
	}

	// Test with existing file and existing account.
	account, err := UserAccountFromFile(file, "account", false)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.Username != "username" {
		t.Errorf("invalid username: %v", account.Username)
	}
	if account.Password != "password" {
		t.Errorf("invalid password: %v", account.Password)
	}
	if account.URL != "https://my-account.dev.my.rubrik-lab.com/api" {
		t.Errorf("invalid url: %v", account.URL)
	}

	// Test with existing file and non-existing account.
	if _, err := UserAccountFromFile(file, "non-existing-account", false); err == nil {
		t.Fatal("UserAccountFromFile should fail with non-existing account")
	}
}

func TestSingleUserAccountFromEnvCredentials(t *testing.T) {
	skipOnEnvs(t, "RUBRIK_POLARIS_ACCOUNT_CREDENTIALS", "RUBRIK_POLARIS_ACCOUNT_NAME")

	// Without name.
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_CREDENTIALS", `{
		"account": {
			"username": "username",
			"password": "password"
		}
	}`)
	account, err := UserAccountFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.Username != "username" {
		t.Errorf("invalid username: %v", account.Username)
	}
	if account.Password != "password" {
		t.Errorf("invalid password: %v", account.Password)
	}
	if account.URL != "" {
		t.Errorf("invalid url: %v", account.URL)
	}

	// With name from env (which should be ignored).
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_NAME", "some-account")
	account, err = UserAccountFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.Username != "username" {
		t.Errorf("invalid username: %v", account.Username)
	}
	if account.Password != "password" {
		t.Errorf("invalid password: %v", account.Password)
	}
	if account.URL != "" {
		t.Errorf("invalid url: %v", account.URL)
	}

	// With name from env (which should be ignored) and URL.
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_CREDENTIALS", `{
		"account": {
			"username": "username",
			"password": "password",
			"url": "https://my-account.my.rubrik.com/api"
		}
	}`)
	account, err = UserAccountFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.Username != "username" {
		t.Errorf("invalid username: %v", account.Username)
	}
	if account.Password != "password" {
		t.Errorf("invalid password: %v", account.Password)
	}
	if account.URL != "https://my-account.my.rubrik.com/api" {
		t.Errorf("invalid url: %v", account.URL)
	}
}

func TestMultipleUserAccountsFromEnvCredentials(t *testing.T) {
	skipOnEnvs(t, "RUBRIK_POLARIS_ACCOUNT_CREDENTIALS", "RUBRIK_POLARIS_ACCOUNT_NAME")

	// Without name.
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_CREDENTIALS", `{
		"account-1": {
			"username": "username-1",
			"password": "password-1"
		},
		"account-2": {
			"username": "username-2",
			"password": "password-2"
		}
	}`)
	if _, err := UserAccountFromEnv(); err == nil {
		t.Fatal("name should be required")
	}

	// With correct name from env.
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_NAME", "account-2")
	account, err := UserAccountFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account-2" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.Username != "username-2" {
		t.Errorf("invalid username: %v", account.Username)
	}
	if account.Password != "password-2" {
		t.Errorf("invalid password: %v", account.Password)
	}
	if account.URL != "" {
		t.Errorf("invalid url: %v", account.URL)
	}

	// With correct name from env and URL.
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_CREDENTIALS", `{
		"account-1": {
			"username": "username-1",
			"password": "password-1",
			"url": "https://my-account-1.my.rubrik.com/api"
		},
		"account-2": {
			"username": "username-2",
			"password": "password-2",
			"url": "https://my-account-2.my.rubrik.com/api"
		}
	}`)
	account, err = UserAccountFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account-2" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.Username != "username-2" {
		t.Errorf("invalid username: %v", account.Username)
	}
	if account.Password != "password-2" {
		t.Errorf("invalid password: %v", account.Password)
	}
	if account.URL != "https://my-account-2.my.rubrik.com/api" {
		t.Errorf("invalid url: %v", account.URL)
	}

	// With wrong name from env.
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_NAME", "some-account")
	if _, err = UserAccountFromEnv(); err == nil {
		t.Fatal("matching account name should be required")
	}
}

func TestDefaultUserAccountFromEnv(t *testing.T) {
	skipOnEnvs(t, "RUBRIK_POLARIS_ACCOUNT_NAME", "RUBRIK_POLARIS_ACCOUNT_USERNAME",
		"RUBRIK_POLARIS_ACCOUNT_PASSWORD", "RUBRIK_POLARIS_ACCOUNT_URL")

	// If a user account exists in the default location we skip the test.
	if _, err := DefaultUserAccount("account", false); err == nil {
		t.Skip("Default user account exists")
	}

	t.Setenv("RUBRIK_POLARIS_ACCOUNT_NAME", "account")
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_USERNAME", "username")
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_PASSWORD", "password")
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_URL", "url")

	account, err := DefaultUserAccount("some-account", true)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.Username != "username" {
		t.Errorf("invalid username: %v", account.Username)
	}
	if account.Password != "password" {
		t.Errorf("invalid password: %v", account.Password)
	}
	if account.URL != "url" {
		t.Errorf("invalid url: %v", account.URL)
	}
}

func TestDefaultUserAccountFromEnvCredentials(t *testing.T) {
	skipOnEnvs(t, "RUBRIK_POLARIS_ACCOUNT_CREDENTIALS", "RUBRIK_POLARIS_ACCOUNT_NAME")

	// If a user account exists in the default location we skip the test.
	if _, err := DefaultUserAccount("account", false); err == nil {
		t.Skip("Default user account exists")
	}

	// Single account with name from function (which should be ignored).
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_CREDENTIALS", `{
		"account": {
			"username": "username",
			"password": "password"
		}
	}`)
	account, err := DefaultUserAccount("some-account", true)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.Username != "username" {
		t.Errorf("invalid username: %v", account.Username)
	}
	if account.Password != "password" {
		t.Errorf("invalid password: %v", account.Password)
	}
	if account.URL != "" {
		t.Errorf("invalid url: %v", account.URL)
	}

	// Single account without override.
	if _, err := DefaultUserAccount("account", false); err == nil {
		t.Fatal("no override should require a user account file")
	}

	// Multiple accounts with correct name from a function.
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_CREDENTIALS", `{
		"account-1": {
			"username": "username-1",
			"password": "password-1",
			"url": "https://my-account-1.my.rubrik.com/api"
		},
		"account-2": {
			"username": "username-2",
			"password": "password-2",
			"url": "https://my-account-2.my.rubrik.com/api"
		}
	}`)
	account, err = DefaultUserAccount("account-1", true)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account-1" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.Username != "username-1" {
		t.Errorf("invalid username: %v", account.Username)
	}
	if account.Password != "password-1" {
		t.Errorf("invalid password: %v", account.Password)
	}
	if account.URL != "https://my-account-1.my.rubrik.com/api" {
		t.Errorf("invalid url: %v", account.URL)
	}

	// Multiple accounts with incorrect name from a function.
	if _, err = DefaultUserAccount("some-account", true); err == nil {
		t.Fatal("matching account name should be required")
	}

	// Multiple accounts with incorrect name from a function and correct name
	// from env.
	t.Setenv("RUBRIK_POLARIS_ACCOUNT_NAME", "account-2")
	account, err = DefaultUserAccount("some-account", true)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account-2" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.Username != "username-2" {
		t.Errorf("invalid username: %v", account.Username)
	}
	if account.Password != "password-2" {
		t.Errorf("invalid password: %v", account.Password)
	}
	if account.URL != "https://my-account-2.my.rubrik.com/api" {
		t.Errorf("invalid url: %v", account.URL)
	}

	// Multiple accounts without override.
	if _, err := DefaultUserAccount("account", false); err == nil {
		t.Fatal("no override requires a valid user account file")
	}
}

func TestServiceAccountFromEnv(t *testing.T) {
	skipOnEnvs(t, "RUBRIK_POLARIS_SERVICEACCOUNT_CREDENTIALS")

	t.Setenv("RUBRIK_POLARIS_SERVICEACCOUNT_CREDENTIALS", `{
		"client_id": "client|id",
		"client_secret": "secret",
		"name": "account",
		"access_token_uri": "https://account.my.rubrik.com/api/client_token"
	}`)
	account, err := ServiceAccountFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.ClientID != "client|id" {
		t.Errorf("invalid client id: %v", account.ClientID)
	}
	if account.ClientSecret != "secret" {
		t.Errorf("invalid client secret: %v", account.ClientSecret)
	}
	if account.AccessTokenURI != "https://account.my.rubrik.com/api/client_token" {
		t.Errorf("invalid access token uri: %v", account.AccessTokenURI)
	}
}

func TestDefaultServiceAccountFromEnv(t *testing.T) {
	skipOnEnvs(t, "RUBRIK_POLARIS_SERVICEACCOUNT_NAME", "RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTID",
		"RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTSECRET", "RUBRIK_POLARIS_SERVICEACCOUNT_ACCESSTOKENURI")

	// If a service account exists in the default location we skip the test.
	if _, err := DefaultServiceAccount(false); err == nil {
		t.Skip("Default service account exists")
	}

	t.Setenv("RUBRIK_POLARIS_SERVICEACCOUNT_NAME", "account")
	t.Setenv("RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTID", "client|id")
	t.Setenv("RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTSECRET", "secret")
	t.Setenv("RUBRIK_POLARIS_SERVICEACCOUNT_ACCESSTOKENURI", "accesstokenuri")

	account, err := DefaultServiceAccount(true)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.ClientID != "client|id" {
		t.Errorf("invalid client id: %v", account.ClientID)
	}
	if account.ClientSecret != "secret" {
		t.Errorf("invalid client secret: %v", account.ClientSecret)
	}
	if account.AccessTokenURI != "accesstokenuri" {
		t.Errorf("invalid access token uri: %v", account.AccessTokenURI)
	}
}

func TestDefaultServiceAccountFromEnvCrendentials(t *testing.T) {
	skipOnEnvs(t, "RUBRIK_POLARIS_SERVICEACCOUNT_CREDENTIALS")

	// If a service account exists in the default location we skip the test.
	if _, err := DefaultServiceAccount(false); err == nil {
		t.Skip("Default service account exists")
	}

	t.Setenv("RUBRIK_POLARIS_SERVICEACCOUNT_CREDENTIALS", `{
		"client_id": "client|id",
		"client_secret": "secret",
		"name": "account",
		"access_token_uri": "https://account.my.rubrik.com/api/client_token"
	}`)

	// Overrides non-existent service account file with env.
	account, err := DefaultServiceAccount(true)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "account" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.ClientID != "client|id" {
		t.Errorf("invalid client id: %v", account.ClientID)
	}
	if account.ClientSecret != "secret" {
		t.Errorf("invalid client secret: %v", account.ClientSecret)
	}
	if account.AccessTokenURI != "https://account.my.rubrik.com/api/client_token" {
		t.Errorf("invalid access token uri: %v", account.AccessTokenURI)
	}

	// No override requires a service account file.
	if _, err = DefaultServiceAccount(false); err == nil {
		t.Fatal("no override requires a valid service account file")
	}
}
