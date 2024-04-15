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
	// Per request timeout
	requestTimeout = 15 * time.Second

	// Number of attempts before failing timed-out token requests
	requestAttempts = 3
)

var errRequestTimeout = fmt.Errorf("token request timeout after %v", requestTimeout)

// requestToken tries to acquire a token using provided parameters. It returns
// errTokenRequestTimeout if the request exceeds tokenRequestTimeout.
func requestToken(ctx context.Context, client *http.Client, tokenURL string, requestBody []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	// Request an access token from the remote token endpoint.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, errRequestTimeout
		}
		return nil, fmt.Errorf("failed to request token: %v", err)
	}
	defer res.Body.Close()
	// Remote responded without a body. For status code 200 this means we are
	// missing the token. For an error we have no additional details.
	if res.ContentLength == 0 {
		return nil, fmt.Errorf("token response has no body (status code %d)", res.StatusCode)
	}

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, errRequestTimeout
		}
		return nil, fmt.Errorf("failed to read token response body (status code %d): %v", res.StatusCode, err)
	}

	// Verify that the content type of the body is JSON. For status code 200
	// this mean we received something that isn't JSON. For an error we have no
	// additional JSON details.
	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		snippet := string(respBody)
		if len(snippet) > 512 {
			snippet = snippet[:512]
		}
		return nil, fmt.Errorf("token response has Content-Type %s (status code %d): %q",
			contentType, res.StatusCode, snippet)
	}

	// Remote responded with a JSON document. Try to parse it as an error
	// message.
	var jsonErr internalerrors.JSONError
	if err := json.Unmarshal(respBody, &jsonErr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token response body as an error (status code %d): %v",
			res.StatusCode, err)
	}
	if jsonErr.IsError() {
		return nil, fmt.Errorf("token response body is an error (status code %d): %w", res.StatusCode, jsonErr)
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("token response has status code: %s", res.Status)
	}

	return respBody, nil
}

// Request tries to acquire a token using the provided parameters.
func Request(client *http.Client, tokenURL string, requestBody []byte, logger log.Logger) ([]byte, error) {
	return RequestWithContext(context.Background(), client, tokenURL, requestBody, logger)
}

// RequestWithContext tries to acquire a token using the provided parameters.
func RequestWithContext(ctx context.Context, client *http.Client, tokenURL string, requestBody []byte, logger log.Logger) ([]byte, error) {
	var err error
	for attempt := 0; attempt < requestAttempts; attempt++ {
		logger.Printf(log.Debug, "Acquire access token (attempt: %d)", attempt+1)

		var resp []byte
		resp, err = requestToken(ctx, client, tokenURL, requestBody)
		if err == nil {
			return resp, nil
		}
		if !errors.Is(err, errRequestTimeout) {
			return nil, fmt.Errorf("failed to acquire access token: %v", err)
		}
	}

	return nil, fmt.Errorf("failed to acquire access token after %d attempts", requestAttempts)
}
