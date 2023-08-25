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
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const (
	lockTimeout = requestAttempts*requestTimeout + 3*time.Second
)

type cache struct {
	block  cipher.Block
	file   string
	source Source
}

// NewCache returns a new cache wrapping the specified token source.
//
// The cache will store authentication tokens in the OS default directory for
// temporary files, this behavior can be overriden by setting the environment
// variable RUBRIK_POLARIS_TOKEN_CACHE_DIR to false.
//
// The cache will also generate a key to encrypt the content of the token cache
// from the RSC account used, this behavior can be overriden by setting the
// environment variable RUBRIK_POLARIS_TOKEN_CACHE_KEY to the encryption key to
// use. The key must be exactly 32 bytes in size and be base64 encoded. Multiple
// processes using the same service account will use the same cached token.
func NewCache(source Source, keyMaterial, suffixMaterial string, allowEnvOverride bool) (*cache, error) {
	key := sha256.Sum256([]byte(keyMaterial))
	suffix := fmt.Sprintf("%x", sha256.Sum256([]byte(suffixMaterial)))
	if allowEnvOverride {
		if tcKey := os.Getenv("RUBRIK_POLARIS_TOKEN_CACHE_KEY"); tcKey != "" {
			tcKeyBuf, err := base64.StdEncoding.DecodeString(tcKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decode provided cache key: %s", err)
			}
			if n := len(tcKeyBuf); n != len(key) {
				return nil, fmt.Errorf("invalid provided cache key size: %d", n)
			}
			copy(key[:], tcKeyBuf)
			suffix += "-env"
		}
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	path := os.TempDir()
	if allowEnvOverride {
		if tcDir := os.Getenv("RUBRIK_POLARIS_TOKEN_CACHE_DIR"); tcDir != "" {
			path = tcDir
		}
	}
	path = filepath.Join(path, fmt.Sprintf("token-%s", suffix))

	return &cache{source: source, block: block, file: path}, nil
}

// token returns the cached token. If the cache is empty or the cached token has
// expired, a new token is fetched from the underlying token source.
func (c *cache) token(ctx context.Context) (token, error) {
	lockCtx, cancel := context.WithTimeout(ctx, lockTimeout)
	defer cancel()
	unlock, err := lockFile(lockCtx, c.file)
	if err != nil {
		return token{}, fmt.Errorf("failed to lock cache: %s", err)
	}
	defer unlock()

	cachedToken, err := readCache(c.file, c.block)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return token{}, fmt.Errorf("failed to read token from cache: %s", err)
	}
	if err == nil && !cachedToken.expired() {
		return cachedToken, nil
	}

	cachedToken, err = c.source.token(ctx)
	if err != nil {
		return token{}, fmt.Errorf("failed to fetch new token: %s", err)
	}

	if err := writeCache(c.file, cachedToken, c.block); err != nil {
		return token{}, fmt.Errorf("failed to write token to cache: %s", err)
	}

	return cachedToken, nil
}

type cacheEntry struct {
	Token string `json:"token"`
	IV    string `json:"iv"`
}

// readCache reads a token from the cache.
func readCache(file string, block cipher.Block) (token, error) {
	buf, err := os.ReadFile(file)
	if err != nil {
		return token{}, fmt.Errorf("failed to read cache entry from cache: %w", err)
	}

	var entry cacheEntry
	if err := json.Unmarshal(buf, &entry); err != nil {
		return token{}, fmt.Errorf("failed to unmarshal cache entry: %s", err)
	}

	iv, err := base64.StdEncoding.DecodeString(entry.IV)
	if err != nil {
		return token{}, fmt.Errorf("failed to decode cache entry salt: %s", err)
	}
	if n := len(iv); n != aes.BlockSize {
		return token{}, fmt.Errorf("invalid iv size: %d", n)
	}
	tokenText, err := base64.StdEncoding.DecodeString(entry.Token)
	if err != nil {
		return token{}, fmt.Errorf("failed to decode cache entry token: %s", err)
	}
	cipher.NewCFBDecrypter(block, iv).XORKeyStream(tokenText, tokenText)

	return fromJWT(string(tokenText))
}

// writeCache writes the specified token to the cache.
func writeCache(file string, token token, block cipher.Block) error {
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return fmt.Errorf("failed to generate random data for cache entry: %s", err)
	}

	buf := make([]byte, len(token.jwtToken.Raw))
	cipher.NewCFBEncrypter(block, iv).XORKeyStream(buf, []byte(token.jwtToken.Raw))

	entry, err := json.Marshal(cacheEntry{
		Token: base64.StdEncoding.EncodeToString(buf),
		IV:    base64.StdEncoding.EncodeToString(iv),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %s", err)
	}

	if err := os.WriteFile(file, entry, 0666); err != nil {
		return fmt.Errorf("failed to write cache entry to cache: %s", err)
	}

	return nil
}

type unlockFunc func()

// lockFile locks the file at the specified path. To unlock the file, call the
// returned unlock function. The call will block until the lock can be obtained.
func lockFile(ctx context.Context, file string) (unlockFunc, error) {
	lockFile := fmt.Sprintf("%s-lock", file)
	for {
		f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL, 0666)
		if err == nil {
			return func() {
				_ = f.Close()
				_ = os.Remove(lockFile)
			}, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("failed to lock file: %q", file)
		}

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
