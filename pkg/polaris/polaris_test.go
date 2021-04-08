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
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// requireEnv skips the current test if specified environment variable is not
// defined or false according to the definition given by strconv.ParseBool.
func requireEnv(t *testing.T, env string) {
	val := os.Getenv(env)

	n, err := strconv.ParseInt(val, 10, 64)
	if err == nil && n > 0 {
		return
	}

	b, err := strconv.ParseBool(val)
	if err == nil && b {
		return
	}

	t.Skipf("skip due to %q", env)
}

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

func TestConfigFromEnv(t *testing.T) {
	restore := resetEnv()
	defer restore()

	if _, err := ConfigFromEnv(); err == nil {
		t.Fatal("ConfigFromEnv should fail with missing environment variables")
	}

	if err := os.Setenv("RUBRIK_POLARIS_ACCOUNT", "my-account"); err != nil {
		t.Fatal(err)
	}
	if _, err := ConfigFromEnv(); err == nil {
		t.Fatal("ConfigFromEnv should fail with missing environment variables")
	}

	if err := os.Setenv("RUBRIK_POLARIS_USERNAME", "john.doe@rubrik.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := ConfigFromEnv(); err == nil {
		t.Fatal("ConfigFromEnv should fail with missing environment variables")
	}

	if err := os.Setenv("RUBRIK_POLARIS_PASSWORD", "Janedoe!"); err != nil {
		t.Fatal(err)
	}
	config, err := ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if config.Account != "my-account" {
		t.Errorf("invalid account: %v", config.Account)
	}
	if config.Username != "john.doe@rubrik.com" {
		t.Errorf("invalid username: %v", config.Username)
	}
	if config.Password != "Janedoe!" {
		t.Errorf("invalid password: %v", config.Password)
	}

	if err := os.Setenv("RUBRIK_POLARIS_URL", "https://my-account.dev.my.rubrik-lab.com/api"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("RUBRIK_POLARIS_LOGLEVEL", "error"); err != nil {
		t.Fatal(err)
	}
	config, err = ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if config.URL != "https://my-account.dev.my.rubrik-lab.com/api" {
		t.Errorf("invalid username: %v", config.URL)
	}
	if config.LogLevel != "error" {
		t.Errorf("invalid password: %v", config.LogLevel)
	}
}

func TestConfigFromFile(t *testing.T) {
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
	if _, err := ConfigFromFile("some-non-existing-file", "my-account"); err == nil {
		t.Fatal("ConfigFromFile should fail with non-existing file")
	}

	// Test with existing file and existing account.
	config, err := ConfigFromFile(file, "my-account")
	if err != nil {
		t.Fatal(err)
	}
	if config.Account != "my-account" {
		t.Errorf("invalid account: %v", config.Account)
	}
	if config.Username != "john.doe@rubrik.com" {
		t.Errorf("invalid username: %v", config.Username)
	}
	if config.Password != "Janedoe!" {
		t.Errorf("invalid password: %v", config.Password)
	}
	if config.URL != "https://my-account.dev.my.rubrik-lab.com/api" {
		t.Errorf("invalid url: %v", config.URL)
	}
	if config.LogLevel != "error" {
		t.Errorf("invalid log level: %v", config.LogLevel)
	}

	// Test with existing file and non-existing account.
	if _, err := ConfigFromFile(file, "non-existing-account"); err == nil {
		t.Fatal("ConfigFromFile should fail with non-existing account")
	}

	// Test with existing file and non-existing account overridden by
	// environment variable.
	if err := os.Setenv("RUBRIK_POLARIS_ACCOUNT", "my-account"); err != nil {
		t.Fatal(err)
	}
	if _, err := ConfigFromFile(file, "non-existing-account"); err != nil {
		t.Fatal(err)
	}
}

func TestMergeConfig(t *testing.T) {
	config, err := mergeConfig(Config{}, Config{})
	if err == nil {
		t.Fatal("mergeConfig should fail when the resulting configuration is invalid")
	}

	config1 := Config{
		Username: "jane.doe@rubrik.com",
		Password: "Johndoe!",
		URL:      "https://my-account.dev.my.rubrik-lab.com/api",
		LogLevel: "info",
	}
	config2 := Config{
		Account:  "my-account",
		Username: "john.doe@rubrik.com",
		Password: "Janedoe!",
		LogLevel: "error",
	}

	config, err = mergeConfig(config1, config2)
	if err != nil {
		t.Fatal(err)
	}
	if config.Account != "my-account" {
		t.Errorf("invalid account: %v", config.Account)
	}
	if config.Username != "john.doe@rubrik.com" {
		t.Errorf("invalid username: %v", config.Username)
	}
	if config.Password != "Janedoe!" {
		t.Errorf("invalid password: %v", config.Password)
	}
	if config.URL != "https://my-account.dev.my.rubrik-lab.com/api" {
		t.Errorf("invalid url: %v", config.URL)
	}
	if config.LogLevel != "error" {
		t.Errorf("invalid log level: %v", config.LogLevel)
	}
}
