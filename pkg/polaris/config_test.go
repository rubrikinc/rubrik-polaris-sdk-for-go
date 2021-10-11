package polaris

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// requireEnv skips the current test if specified environment variable is not
// defined or false according to the definition given by strconv.ParseBool.
func requireEnv(t *testing.T, env string, delay time.Duration) {
	val := os.Getenv(env)

	b, err := strconv.ParseBool(val)
	if err == nil && b {
		if delay > 0 {
			time.Sleep(delay)
		}

		return
	}

	t.Skipf("skip due to %q", env)
}

func TestLocalUserFromFile(t *testing.T) {
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
	if _, err := AccountFromFile("some-non-existing-file", "my-account", false); err == nil {
		t.Fatal("LocalUserFromFile should fail with non-existing file")
	}

	// Test with existing file and existing account.
	account, err := AccountFromFile(file, "my-account", false)
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
	if _, err := AccountFromFile(file, "non-existing-account", false); err == nil {
		t.Fatal("LocalUserFromFile should fail with non-existing account")
	}
}
