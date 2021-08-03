package polaris

import (
	"os"
	"path/filepath"
	"testing"
)

// resetEnv resets the environment variables related to the Polaris SDK. The
// returned restore function can be used to restore the environment variables.
func resetEnv() (restore func()) {
	account := os.Getenv("RUBRIK_POLARIS_ACCOUNT")
	os.Unsetenv("RUBRIK_POLARIS_ACCOUNT")

	username := os.Getenv("RUBRIK_POLARIS_USERNAME")
	os.Unsetenv("RUBRIK_POLARIS_USERNAME")

	password := os.Getenv("RUBRIK_POLARIS_PASSWORD")
	os.Unsetenv("RUBRIK_POLARIS_PASSWORD")

	url := os.Getenv("RUBRIK_POLARIS_URL")
	os.Unsetenv("RUBRIK_POLARIS_URL")

	logLevel := os.Getenv("RUBRIK_POLARIS_LOGLEVEL")
	os.Unsetenv("RUBRIK_POLARIS_LOGLEVEL")

	return func() {
		os.Setenv("RUBRIK_POLARIS_ACCOUNT", account)
		os.Setenv("RUBRIK_POLARIS_USERNAME", username)
		os.Setenv("RUBRIK_POLARIS_PASSWORD", password)
		os.Setenv("RUBRIK_POLARIS_URL", url)
		os.Setenv("RUBRIK_POLARIS_LOGLEVEL", logLevel)
	}
}

func TestUserAccountFromEnv(t *testing.T) {
	restore := resetEnv()
	defer restore()

	if _, err := UserAccountFromEnv(); err == nil {
		t.Fatal("UserAccountFromEnv should fail with missing environment variables")
	}

	if err := os.Setenv("RUBRIK_POLARIS_ACCOUNT_NAME", "my-account"); err != nil {
		t.Fatal(err)
	}
	if _, err := UserAccountFromEnv(); err == nil {
		t.Fatal("UserAccountFromEnv should fail with missing environment variables")
	}

	if err := os.Setenv("RUBRIK_POLARIS_ACCOUNT_USERNAME", "john.doe@rubrik.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := UserAccountFromEnv(); err == nil {
		t.Fatal("UserAccountFromEnv should fail with missing environment variables")
	}

	if err := os.Setenv("RUBRIK_POLARIS_ACCOUNT_PASSWORD", "Janedoe!"); err != nil {
		t.Fatal(err)
	}
	account, err := UserAccountFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "my-account" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.Username != "john.doe@rubrik.com" {
		t.Errorf("invalid username: %v", account.Username)
	}
	if account.Password != "Janedoe!" {
		t.Errorf("invalid password: %v", account.Password)
	}

	if err := os.Setenv("RUBRIK_POLARIS_ACCOUNT_URL", "https://my-account.dev.my.rubrik-lab.com/api"); err != nil {
		t.Fatal(err)
	}
	account, err = UserAccountFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if account.URL != "https://my-account.dev.my.rubrik-lab.com/api" {
		t.Errorf("invalid username: %v", account.URL)
	}
}

func TestUserAccountFromFile(t *testing.T) {
	restore := resetEnv()
	defer restore()

	path, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	file := filepath.Join(path, "config.json")
	data := []byte(`{
		"my-account": {
			"username": "john.doe@rubrik.com",
			"password": "Janedoe!",
			"url": "https://my-account.dev.my.rubrik-lab.com/api",
			"loglevel": "error"
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
	account, err := UserAccountFromFile(file, "my-account", false)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "my-account" {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.Username != "john.doe@rubrik.com" {
		t.Errorf("invalid username: %v", account.Username)
	}
	if account.Password != "Janedoe!" {
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
