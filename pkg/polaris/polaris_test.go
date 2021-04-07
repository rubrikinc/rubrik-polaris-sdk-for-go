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
	"strconv"
	"testing"
)

// requireEnv skips the current test if specified environment variable is not
// defined or false according to the definition given by strconv.ParseBool.
func requireEnv(t *testing.T, env string) {
	ok, err := strconv.ParseBool(os.Getenv(env))
	if err != nil || !ok {
		t.Skipf("skip due to %q", env)
	}
}

func TestConfigFromEnv(t *testing.T) {
	config, err := ConfigFromEnv()
	if err == nil {
		t.Fatal("expected ConfigFromEnv to fail")
	}

	if err := os.Setenv("RUBRIK_POLARIS_ACCOUNT", "account"); err != nil {
		t.Fatal(err)
	}
	config, err = ConfigFromEnv()
	if err == nil {
		t.Fatal("expected ConfigFromEnv to fail")
	}

	if err := os.Setenv("RUBRIK_POLARIS_USERNAME", "username"); err != nil {
		t.Fatal(err)
	}
	config, err = ConfigFromEnv()
	if err == nil {
		t.Fatal("expected ConfigFromEnv to fail")
	}

	if err := os.Setenv("RUBRIK_POLARIS_PASSWORD", "password"); err != nil {
		t.Fatal(err)
	}
	config, err = ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if config.Account == "" {
		t.Error("invalid account")
	}
	if config.Username == "" {
		t.Error("invalid username")
	}
	if config.Password == "" {
		t.Error("invalid password")
	}

	if err := os.Setenv("RUBRIK_POLARIS_URL", "url"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("RUBRIK_POLARIS_LOGLEVEL", "logLevel"); err != nil {
		t.Fatal(err)
	}

	config, err = ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if config.URL == "" {
		t.Error("invalid url")
	}
	if config.LogLevel == "" {
		t.Error("invalid log level")
	}
}
