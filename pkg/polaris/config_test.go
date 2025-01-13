package polaris

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindAccount(t *testing.T) {
	dropEnvs(t,
		keyServiceAccountAccessTokenURI,
		keyServiceAccountClientID,
		keyServiceAccountClientSecret,
		keyServiceAccountCredentials,
		keyServiceAccountFile,
		keyServiceAccountName)

	testCases := []struct {
		name      string
		file      string
		text      string
		user      string
		envs      map[string]string
		override  bool
		errPrefix string
	}{{
		name: "Env",
		envs: map[string]string{keyServiceAccountCredentials: toText(serviceAccount{})},
	}, {
		name: "EnvOverride",
		envs: map[string]string{
			keyServiceAccountCredentials: toText(serviceAccount{ClientID: "invalid"}),
			keyServiceAccountClientID:    "id",
		},
	}, {
		name: "EnvError",
		envs: map[string]string{
			keyServiceAccountName:         "name",
			keyServiceAccountClientID:     "id",
			keyServiceAccountClientSecret: "secret",
		},
		errPrefix: "failed to load service account from env: invalid service account access token uri",
	}, {
		name: "File",
		file: toText(serviceAccount{}),
	}, {
		name:     "FileOverride",
		file:     toText(serviceAccount{ClientSecret: "invalid"}),
		envs:     map[string]string{keyServiceAccountClientSecret: "secret"},
		override: true,
	}, {
		name:      "FileError",
		file:      toText(serviceAccount{noClientID: true}),
		errPrefix: "failed to load service account from file: invalid service account client id",
	}, {
		name: "Text",
		text: toText(serviceAccount{}),
	}, {
		name:     "TextOverride",
		text:     toText(serviceAccount{Name: "invalid"}),
		envs:     map[string]string{keyServiceAccountName: "name"},
		override: true,
	}, {
		name:      "TextError",
		text:      toText(serviceAccount{noClientSecret: true}),
		errPrefix: "failed to load service account from text: invalid service account client secret",
	}, {
		name:      "NoDefaultAndNoCredentials",
		errPrefix: "account not found, searched: default service account file and env",
	}, {
		name:      "NoDefaultAndInvalidCredentials",
		text:      "Content of a file not containing an RSC service account",
		errPrefix: "account not found, searched: passed in credentials, default service account file and default user account file",
	}}

	tempDir := t.TempDir()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			for key, value := range testCase.envs {
				t.Setenv(key, value)
			}

			// Create service account test file for test. All test files are
			// removed automatically by the testing package after the test.
			credentials := testCase.text
			if testCase.file != "" {
				credentials = filepath.Join(tempDir, strings.ToLower(testCase.name+".json"))
				if err := os.WriteFile(credentials, []byte(testCase.file), 0666); err != nil {
					t.Fatal(err)
				}
			}

			account, err := FindAccount(credentials, testCase.override)
			if assertErrPrefix(t, err, testCase.errPrefix) {
				return
			}
			assertAccount(t, account)
		})
	}
}

func TestFindAccountWithDefaultUserAccount(t *testing.T) {
	tempDir := t.TempDir()

	t.Setenv("HOME", tempDir)
	t.Setenv("USERPROFILE", tempDir)

	// Create a user account in the default user account file.
	path, err := expandPath(DefaultLocalUserFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	userAccount := `{
		"account": {
			"username": "username",
			"password": "password",
			"url": "https://account.my.rubrik.com/api"
		}
	}`
	if err := os.WriteFile(path, []byte(userAccount), 0666); err != nil {
		t.Fatal(err)
	}

	account, err := FindAccount("account", false)
	if err != nil {
		t.Fatal(err)
	}
	assertAccount(t, account)
}

func TestFindAccountWithDefaultServiceAccount(t *testing.T) {
	tempDir := t.TempDir()

	t.Setenv("HOME", tempDir)
	t.Setenv("USERPROFILE", tempDir)

	// Create a service account in the default service account file.
	path, err := expandPath(DefaultServiceAccountFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(toText(serviceAccount{})), 0666); err != nil {
		t.Fatal(err)
	}

	account, err := FindAccount("", false)
	if err != nil {
		t.Fatal(err)
	}
	assertAccount(t, account)
}

func TestServiceAccountFromEnv(t *testing.T) {
	dropEnvs(t,
		keyServiceAccountAccessTokenURI,
		keyServiceAccountClientID,
		keyServiceAccountClientSecret,
		keyServiceAccountCredentials,
		keyServiceAccountFile,
		keyServiceAccountName)

	testCases := []struct {
		name      string
		envs      map[string]string
		errPrefix string
	}{{
		name: "CredentialsVar",
		envs: map[string]string{keyServiceAccountCredentials: toText(serviceAccount{})},
	}, {
		name: "NoSecretInCredentialsVar",
		envs: map[string]string{
			keyServiceAccountCredentials: toText(serviceAccount{noClientSecret: true}),
		},
		errPrefix: "invalid service account client secret",
	}, {
		name: "InvalidURLInCredentialsVar",
		envs: map[string]string{
			keyServiceAccountCredentials: toText(serviceAccount{AccessTokenURI: "invalid"}),
		},
		errPrefix: "invalid service account access token uri",
	}, {
		name: "MultipleVars",
		envs: map[string]string{
			keyServiceAccountClientID:       "id",
			keyServiceAccountClientSecret:   "secret",
			keyServiceAccountName:           "name",
			keyServiceAccountAccessTokenURI: "https://account.my.rubrik.com/api/client_token",
		},
	}, {
		name: "NoSecretInMultipleVars",
		envs: map[string]string{
			keyServiceAccountClientID:       "id",
			keyServiceAccountName:           "name",
			keyServiceAccountAccessTokenURI: "https://account.my.rubrik.com/api/client_token",
		},
		errPrefix: "invalid service account client secret",
	}, {
		name: "InvalidURLInMultipleVars",
		envs: map[string]string{
			keyServiceAccountClientID:       "id",
			keyServiceAccountClientSecret:   "secret",
			keyServiceAccountName:           "name",
			keyServiceAccountAccessTokenURI: "invalid",
		},
		errPrefix: "invalid service account access token uri",
	}, {
		name: "NameAndCredentialsVars",
		envs: map[string]string{
			keyServiceAccountName:        "name",
			keyServiceAccountCredentials: toText(serviceAccount{Name: "invalid"}),
		},
	}, {
		name:      "NoVars",
		errPrefix: "account not found in env",
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			for key, value := range testCase.envs {
				t.Setenv(key, value)
			}

			account, err := ServiceAccountFromEnv()
			if assertErrPrefix(t, err, testCase.errPrefix) {
				return
			}
			assertAccount(t, account)
		})
	}
}

func TestServiceAccountFromFile(t *testing.T) {
	dropEnvs(t,
		keyServiceAccountAccessTokenURI,
		keyServiceAccountClientID,
		keyServiceAccountClientSecret,
		keyServiceAccountCredentials,
		keyServiceAccountFile,
		keyServiceAccountName)

	testCases := []struct {
		name      string
		file      string
		envs      map[string]string
		override  bool
		noFile    bool
		errPrefix string
	}{{
		name: "File",
		file: toText(serviceAccount{}),
	}, {
		name:     "ClientIDInEnv",
		file:     toText(serviceAccount{ClientID: "invalid"}),
		envs:     map[string]string{keyServiceAccountClientID: "id"},
		override: true,
	}, {
		name: "NameAndCredentialsInEnv",
		file: toText(serviceAccount{
			ClientID:       "invalid",
			ClientSecret:   "invalid",
			Name:           "invalid",
			AccessTokenURI: "invalid",
		}),
		envs: map[string]string{
			keyServiceAccountName:        "name",
			keyServiceAccountCredentials: toText(serviceAccount{Name: "invalid"}),
		},
		override: true,
	}, {
		name:      "FileInEnv",
		file:      toText(serviceAccount{}),
		envs:      map[string]string{keyServiceAccountFile: "/path/to/some/file"},
		override:  true,
		errPrefix: "account not found in file",
	}, {
		name:      "ClientSecretInEnvNoOverride",
		file:      toText(serviceAccount{noClientSecret: true}),
		envs:      map[string]string{keyServiceAccountClientSecret: "secret"},
		errPrefix: "invalid service account client secret",
	}, {
		name:      "NoClientID",
		file:      toText(serviceAccount{noClientID: true}),
		errPrefix: "invalid service account client id",
	}, {
		name:      "InvalidURL",
		file:      toText(serviceAccount{AccessTokenURI: "invalid"}),
		errPrefix: "invalid service account access token uri",
	}, {
		name:      "NoFileName",
		noFile:    true,
		errPrefix: "account not found in file",
	}, {
		name:      "NoFileContent",
		errPrefix: "failed to unmarshal service account file",
	}, {
		name:      "InvalidFileContent",
		file:      "Content of a file not containing an RSC service account",
		errPrefix: "failed to unmarshal service account file",
	}}

	tempDir := t.TempDir()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			for key, value := range testCase.envs {
				t.Setenv(key, value)
			}

			// Create service account test file for test. All test files are
			// removed automatically by the testing package after the test.
			var filename string
			if !testCase.noFile {
				filename = filepath.Join(tempDir, strings.ToLower(testCase.name+".json"))
				if err := os.WriteFile(filename, []byte(testCase.file), 0666); err != nil {
					t.Fatal(err)
				}
			}

			account, err := ServiceAccountFromFile(filename, testCase.override)
			if assertErrPrefix(t, err, testCase.errPrefix) {
				return
			}
			assertAccount(t, account)
		})
	}
}

func TestServiceAccountFromText(t *testing.T) {
	dropEnvs(t,
		keyServiceAccountAccessTokenURI,
		keyServiceAccountClientID,
		keyServiceAccountClientSecret,
		keyServiceAccountCredentials,
		keyServiceAccountFile,
		keyServiceAccountName)

	testCases := []struct {
		name      string
		text      string
		envs      map[string]string
		override  bool
		errPrefix string
	}{{
		name: "Text",
		text: toText(serviceAccount{}),
	}, {
		name:     "ClientIDInEnv",
		text:     toText(serviceAccount{ClientID: "invalid"}),
		envs:     map[string]string{keyServiceAccountClientID: "id"},
		override: true,
	}, {
		name: "NameAndCredentialsInEnv",
		text: toText(serviceAccount{
			ClientID:       "invalid",
			ClientSecret:   "invalid",
			Name:           "invalid",
			AccessTokenURI: "invalid",
		}),
		envs: map[string]string{
			keyServiceAccountName:        "name",
			keyServiceAccountCredentials: toText(serviceAccount{Name: "invalid"}),
		},
		override: true,
	}, {
		name:      "ClientSecretInEnvNoOverride",
		text:      toText(serviceAccount{noClientSecret: true}),
		envs:      map[string]string{keyServiceAccountClientSecret: "secret"},
		errPrefix: "invalid service account client secret",
	}, {
		name:      "NoName",
		text:      toText(serviceAccount{noName: true}),
		errPrefix: "invalid service account name",
	}, {
		name:      "InvalidURL",
		text:      toText(serviceAccount{AccessTokenURI: "invalid"}),
		errPrefix: "invalid service account access token uri",
	}, {
		name:      "EmptyText",
		errPrefix: "account not found in text",
	}, {
		name:      "InvalidText",
		text:      "/path/to/some/file",
		errPrefix: "account not found in text",
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			for key, value := range testCase.envs {
				t.Setenv(key, value)
			}

			account, err := ServiceAccountFromText(testCase.text, testCase.override)
			if assertErrPrefix(t, err, testCase.errPrefix) {
				return
			}
			assertAccount(t, account)
		})
	}
}

func TestUserAccountFromEnv(t *testing.T) {
	dropEnvs(t,
		keyUserAccountCredentials,
		keyUserAccountFile,
		keyUserAccountName,
		keyUserAccountPassword,
		keyUserAccountURL,
		keyUserAccountUsername)

	testCases := []struct {
		name      string
		envs      map[string]string
		override  bool
		errPrefix string
	}{{
		name: "CredentialsVar",
		envs: map[string]string{
			keyUserAccountCredentials: `{
				"account": {
					"username": "username",
					"password": "password",
					"url": "https://account.my.rubrik.com/api"
				}
			}`,
		},
	}, {
		name: "NoPasswordInCredentialsVar",
		envs: map[string]string{
			keyUserAccountCredentials: `{
				"account": {
					"username": "username",
					"url": "https://account.my.rubrik.com/api"
				}
			}`,
		},
		errPrefix: "invalid user account password",
	}, {
		name: "InvalidURLInCredentialsVar",
		envs: map[string]string{
			keyUserAccountCredentials: `{
				"account": {
					"username": "username",
					"password": "password",
					"url": "invalid"
				}
			}`,
		},
		errPrefix: "invalid user account url",
	}, {
		name: "MultipleVars",
		envs: map[string]string{
			keyUserAccountName:     "account",
			keyUserAccountUsername: "username",
			keyUserAccountPassword: "password",
			keyUserAccountURL:      "https://account.my.rubrik.com/api",
		},
	}, {
		name: "MultipleAccountsInCredentialsVar",
		envs: map[string]string{
			keyUserAccountName: "account",
			keyUserAccountCredentials: `{
				"account": {
					"username": "username",
					"password": "password",
					"url": "https://account.my.rubrik.com/api"
				},
				"another-account": {
					"username": "another-username",
					"password": "another-password",
					"url": "https://another-account.my.rubrik.com/api"
				}
			}`,
		},
	}, {
		name:      "NoVars",
		errPrefix: "account not found in env",
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			for key, value := range testCase.envs {
				t.Setenv(key, value)
			}

			account, err := UserAccountFromEnv()
			if assertErrPrefix(t, err, testCase.errPrefix) {
				return
			}
			assertAccount(t, account)
		})
	}
}

func TestUserAccountFromFile(t *testing.T) {
	dropEnvs(t,
		keyUserAccountCredentials,
		keyUserAccountFile,
		keyUserAccountName,
		keyUserAccountPassword,
		keyUserAccountURL,
		keyUserAccountUsername)

	testCases := []struct {
		name      string
		file      string
		envs      map[string]string
		override  bool
		noFile    bool
		errPrefix string
	}{{
		name: "Account",
		file: `{
			"account": {
				"username": "username",
				"password": "password",
				"url": "https://account.my.rubrik.com/api"
			}
		}`,
	}, {
		name: "MultipleAccounts",
		file: `{
			"account": {
				"username": "username",
				"password": "password",
				"url": "https://account.my.rubrik.com/api"
			},
			"another-account": {
				"username": "another-username",
				"password": "another-password",
				"url": "https://another-account.my.rubrik.com/api"
			}
		}`,
	}, {
		name: "AccountNoMatch",
		file: `{
			"another-account": {
				"username": "another-username",
				"password": "another-password",
				"url": "https://another-account.my.rubrik.com/api"
			}
		}`,
		errPrefix: "user account \"account\" not found in user account file",
	}, {
		name: "NoPassword",
		file: `{
			"account": {
				"username": "username",
				"password": "",
				"url": "https://account.my.rubrik.com/api"
			}
		}`,
		errPrefix: "invalid user account password",
	}, {
		name: "UsernameInEnv",
		file: `{
			"account": {
				"username": "invalid",
				"password": "password",
				"url": "https://account.my.rubrik.com/api"
			}
		}`,
		envs:     map[string]string{keyUserAccountUsername: "username"},
		override: true,
	}, {
		name: "URLInEnvNoOverride",
		file: `{
			"account": {
				"username": "invalid",
				"password": "password",
				"url": "invalid"
			}
		}`,
		envs:      map[string]string{keyUserAccountURL: "https://account.my.rubrik.com/api"},
		errPrefix: "invalid user account url",
	}, {
		name:      "NoFile",
		noFile:    true,
		errPrefix: "account not found in file",
	}, {
		name:      "InvalidFileContent",
		file:      "Content of a file not containing an RSC user account",
		errPrefix: "failed to unmarshal user account file",
	}}

	tempDir := t.TempDir()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			for key, value := range testCase.envs {
				t.Setenv(key, value)
			}

			// Create service account test file for test. All test files are
			// removed automatically by the testing package after the test.
			var path string
			if !testCase.noFile {
				path = filepath.Join(tempDir, strings.ToLower(testCase.name+".json"))
				if err := os.WriteFile(path, []byte(testCase.file), 0666); err != nil {
					t.Fatal(err)
				}
			}

			account, err := UserAccountFromFile(path, "account", testCase.override)
			if assertErrPrefix(t, err, testCase.errPrefix) {
				return
			}
			assertAccount(t, account)
		})
	}
}

func TestExpandPath(t *testing.T) {
	path, err := expandPath("~/")
	if err != nil {
		t.Fatal(err)
	}
	if path != os.Getenv("HOME") {
		t.Fatalf("invalid path: %s", path)
	}

	path, err = expandPath("/tmp/file1/../file2")
	if err != nil {
		t.Fatal(err)
	}
	if path != "/tmp/file2" {
		t.Fatalf("invalid path: %s", path)
	}

	path, err = expandPath("/tmp/${HOME}/file1")
	if err != nil {
		t.Fatal(err)
	}
	if path != "/tmp"+os.Getenv("HOME")+"/file1" {
		t.Fatalf("invalid path: %s", path)
	}
}

// assertAccount fails the test if the account is not equal to the expected
// values.
func assertAccount(t *testing.T, account Account) {
	switch account := account.(type) {
	case *ServiceAccount:
		if account.ClientID != "id" {
			t.Fatalf("invalid client id: %s", account.ClientID)
		}
		if account.ClientSecret != "secret" {
			t.Fatalf("invalid client secret: %s", account.ClientSecret)
		}
		if account.Name != "name" {
			t.Fatalf("invalid name: %s", account.Name)
		}
		if account.AccessTokenURI != "https://account.my.rubrik.com/api/client_token" {
			t.Fatalf("invalid access token uri: %s", account.AccessTokenURI)
		}
	case *UserAccount:
		if account.Name != "account" {
			t.Fatalf("invalid name: %s", account.Name)
		}
		if account.Username != "username" {
			t.Fatalf("invalid username: %s", account.Username)
		}
		if account.Password != "password" {
			t.Fatalf("invalid password: %s", account.Password)
		}
		if account.URL != "" && account.URL != "https://account.my.rubrik.com/api" {
			t.Fatalf("invalid url: %s", account.URL)
		}
	default:
		t.Fatalf("invalid account type: %T", account)
	}
}

// assertErrPrefix returns false if the error is nil and the error prefix is
// the empty string. Returns true if the error is not nil and the error message
// matches the error prefix. Otherwise, it fails the test.
func assertErrPrefix(t *testing.T, err error, errPrefix string) bool {
	switch {
	case err == nil && errPrefix == "":
		return false
	case err == nil:
		t.Fatalf("expected error with prefix: %q", errPrefix)
	case errPrefix == "":
		t.Fatalf("unexpected error: %s", err)
	case !strings.HasPrefix(err.Error(), errPrefix):
		t.Fatalf("unexpected error: %s", err)
	}

	return true
}

// dropEnvs removes the environment variables with the given keys while the
// test is running.
func dropEnvs(t *testing.T, keys ...string) {
	for _, key := range keys {
		t.Setenv(key, "")
	}
}

type serviceAccount struct {
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	Name           string `json:"name"`
	AccessTokenURI string `json:"access_token_uri"`

	noClientID       bool
	noClientSecret   bool
	noName           bool
	noAccessTokenURI bool
}

func toText(account serviceAccount) string {
	if account.ClientID == "" && !account.noClientID {
		account.ClientID = "id"
	}
	if account.ClientSecret == "" && !account.noClientSecret {
		account.ClientSecret = "secret"
	}
	if account.Name == "" && !account.noName {
		account.Name = "name"
	}
	if account.AccessTokenURI == "" && !account.noAccessTokenURI {
		account.AccessTokenURI = "https://account.my.rubrik.com/api/client_token"
	}

	buf, err := json.Marshal(account)
	if err != nil {
		panic(err)
	}

	return string(buf)
}
