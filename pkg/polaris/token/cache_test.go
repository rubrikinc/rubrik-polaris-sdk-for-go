// Copyright 2023 Rubrik, Inc.
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

package token

import (
	"context"
	"crypto/aes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

var (
	// Expires 2030.
	dummyToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNT" +
		"E2MjM5MDIyLCJleHAiOjE5MTM4ODA5OTl9.og3Lk43zo-gCS4pns3KqMO01Cgh2FH7F-u81T6FaxTk"

	// Expired 2021.
	expiredDummyToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF" +
		"0IjoxNTE2MjM5MDIyLCJleHAiOjE2MjI0OTExNTR9.y3TkH5_8Pv7Vde1I-ll2BJ29dX4tYKGIhrAA314VGa0"
)

func TestCacheTokenSource(t *testing.T) {
	t.Setenv("RUBRIK_POLARIS_TOKEN_CACHE_DIR", t.TempDir())

	cache, err := NewCache(&mockSource{}, "key", "suffix", true)
	if err != nil {
		t.Fatal(err)
	}

	// The wrapped token source should return the expired token.
	testToken, err := cache.token(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok := testToken.jwtToken.Raw; tok != expiredDummyToken {
		t.Fatalf("wrong token: %s", tok)
	}

	// The wrapped token source returns the unexpired token.
	testToken, err = cache.token(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok := testToken.jwtToken.Raw; tok != dummyToken {
		t.Fatalf("wrong token: %s", tok)
	}
}

func TestReadWriteCache(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "test-file")
	assertFileNotExist(t, testFile)

	block, err := aes.NewCipher(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}

	// Write token
	testToken, err := fromJWT(dummyToken)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeCache(testFile, testToken, block); err != nil {
		t.Fatal(err)
	}
	assertFileExist(t, testFile)

	// Overwrite token
	if err := writeCache(testFile, testToken, block); err != nil {
		t.Fatal(err)
	}
	assertFileExist(t, testFile)

	// Read token
	testToken, err = readCache(testFile, block)
	if err != nil {
		t.Fatal(err)
	}
	if testToken.jwtToken.Raw != dummyToken {
		t.Fatal("wrong token")
	}
}

func TestReadInvalidTokenFromCache(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "test-file")
	assertFileNotExist(t, testFile)

	buf, err := json.Marshal(cacheEntry{
		Token: []byte("an invalid jwt token"),
		IV:    make([]byte, aes.BlockSize),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, buf, 0666); err != nil {
		t.Fatal(err)
	}
	assertFileExist(t, testFile)

	block, err := aes.NewCipher(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = readCache(testFile, block); !errors.Is(err, errInvalidToken) {
		t.Fatalf("invalid error: %s", err)
	}
}

func TestLockFile(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "test-file")
	assertFileNotExist(t, testFile+"-lock")

	// Lock file, no contention
	unlock, err := lockFile(context.Background(), testFile)
	if err != nil {
		t.Fatal(err)
	}
	assertFileExist(t, testFile+"-lock")
	unlock()
	assertFileNotExist(t, testFile+"-lock")
}

func assertFileExist(t *testing.T, path string) {
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file %q should exist", path)
	}
}

func assertFileNotExist(t *testing.T, path string) {
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("file %q should not exist", path)
	}
}

type mockSource struct {
	call int
}

func (m *mockSource) token(ctx context.Context) (token, error) {
	m.call++
	if m.call > 1 {
		return fromJWT(dummyToken)
	}
	return fromJWT(expiredDummyToken)
}
