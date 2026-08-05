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

package token

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	internalerrors "github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/errors"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

const (
	// Number of attempts before failing timed-out token requests.
	requestAttempts = 3

	// Per request timeout.
	requestTimeout = 15 * time.Second
)

// timeoutKey is the context key used to pass in a custom request timeout to
// the requestToken function. Intended to be used by unit tests.
var timeoutKey = struct{}{}

// requestToken tries to acquire a token using provided parameters. It returns
// a context.Canceled if the request was canceled, a context.DeadlineExceeded
// if the request exceeds the requestTimeout or a JSONError if the response was
// a JSON error.
func requestToken(ctx context.Context, client *http.Client, tokenURL string, requestBody []byte) ([]byte, error) {
	// Allow a context value with the timeoutKey key to override the default
	// request timeout.
	timeout := requestTimeout
	if timeoutValue := ctx.Value(timeoutKey); timeoutValue != nil {
		timeout = timeoutValue.(time.Duration)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Request an access token from the remote token endpoint.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		// Wrap context related errors.
		return nil, fmt.Errorf("failed to request token: %w", err)
	}
	defer res.Body.Close()

	// Remote responded without a body. For status code 200, this means we are
	// missing the token. For an error, we have no additional details.
	if res.ContentLength == 0 {
		return nil, fmt.Errorf("token response has no body (status code %d)", res.StatusCode)
	}

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		// Wrap context related errors.
		return nil, fmt.Errorf("failed to read token response body (status code %d): %w", res.StatusCode, err)
	}

	// Verify that the content type of the body is JSON. For status code 200,
	// this means we received something that isn't JSON. For an error, we have no
	// additional JSON details.
	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		snippet := string(respBody)
		if len(snippet) > 512 {
			snippet = snippet[:512]
		}
		return nil, fmt.Errorf("token response has Content-Type %s (status code %d): %q",
			contentType, res.StatusCode, strings.TrimSpace(snippet))
	}

	// Remote responded with a JSON document. Try to parse it as an error
	// message.
	var jsonErr internalerrors.JSONError
	if err := json.Unmarshal(respBody, &jsonErr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token response body as an error (status code %d): %s",
			res.StatusCode, err)
	}
	if jsonErr.IsError() {
		// Wrap JSON related errors.
		return nil, fmt.Errorf("token response body is an error (status code %d): %w", res.StatusCode, jsonErr)
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("token response has status code: %s", res.Status)
	}

	return respBody, nil
}

// Deprecated: use RequestWithContext
func Request(client *http.Client, tokenURL string, requestBody []byte, logger log.Logger) ([]byte, error) {
	return RequestWithContext(context.Background(), client, tokenURL, requestBody, logger)
}

// RequestWithContext tries to acquire a token using the provided parameters.
func RequestWithContext(ctx context.Context, client *http.Client, tokenURL string, requestBody []byte, logger log.Logger) ([]byte, error) {
	var err error
	for attempt := 0; attempt < requestAttempts; attempt++ {
		logger.Printf(log.Debug, "Acquire access token (attempt: %d)", attempt+1)

		// Request an access token. Both the request's parent context and the
		// function's child context can result in context errors, don't treat
		// those as request errors.
		var resp []byte
		resp, err = requestToken(ctx, client, tokenURL, requestBody)
		if err == nil {
			return resp, nil
		}
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			var jsonErr internalerrors.JSONError
			if errors.As(err, &jsonErr) {
				// Keep the wrapped JSON related errors.
				return nil, fmt.Errorf("failed to acquire access token: %w", err)
			}

			// Erase all other error types.
			return nil, fmt.Errorf("failed to acquire access token: %s", err)
		}

		// Check if the parent request context has been canceled or timed out,
		// if so, wrap the related error.
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("failed to acquire access token: %w", ctx.Err())
		default:
		}
	}

	return nil, fmt.Errorf("failed to acquire access token after %d attempts", requestAttempts)
}
