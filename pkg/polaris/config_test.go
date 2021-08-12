package polaris

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserAccountFromFile(t *testing.T) {
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
