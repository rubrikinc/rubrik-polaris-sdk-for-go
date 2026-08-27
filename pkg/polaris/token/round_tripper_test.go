// Copyright 2026 Rubrik, Inc.
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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	internalerrors "github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/errors"
)

func TestRoundTripper(t *testing.T) {
	apiSrv := httptest.NewServer(checkBearerAndEchoHandler(t))
	defer apiSrv.Close()

	tokSrv := httptest.NewServer(validTokenHandler(t))
	defer tokSrv.Close()

	// Install the token RoundTripper.
	apiSrv.Client().Transport = NewRoundTripper(apiSrv.Client().Transport, serviceAccountTestSource(tokSrv))

	// Do request.
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, apiSrv.URL, bytes.NewBufferString("body"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := apiSrv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	// Check response.
	if res.StatusCode != http.StatusOK {
		t.Fatal("")
	}
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if s := string(buf); s != "body" {
		t.Fatalf("expected 'body', got '%s'", s)
	}
}

func TestRoundTripperWithJSONError(t *testing.T) {
	apiSrv := httptest.NewServer(checkBearerAndEchoHandler(t))
	defer apiSrv.Close()

	// Returns a JSON error.
	tokSrv := httptest.NewServer(serviceAccountBadCredentialsHandler(t))
	defer tokSrv.Close()

	// Install the token RoundTripper.
	apiSrv.Client().Transport = NewRoundTripper(apiSrv.Client().Transport, serviceAccountTestSource(tokSrv))

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, apiSrv.URL, bytes.NewBufferString("body"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = apiSrv.Client().Do(req)
	var jsonErr internalerrors.JSONError
	if !errors.As(err, &jsonErr) {
		t.Fatal(err)
	}
}

func TestRoundTripperWithCanceledContext(t *testing.T) {
	apiSrv := httptest.NewServer(checkBearerAndEchoHandler(t))
	defer apiSrv.Close()

	tokSrv := httptest.NewServer(validTokenHandler(t))
	defer tokSrv.Close()

	// Install the token RoundTripper.
	apiSrv.Client().Transport = NewRoundTripper(apiSrv.Client().Transport, serviceAccountTestSource(tokSrv))

	// Canceled context.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiSrv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = apiSrv.Client().Do(req)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestRoundTripperWithDeadlineExceededContext(t *testing.T) {
	apiSrv := httptest.NewServer(checkBearerAndEchoHandler(t))
	defer apiSrv.Close()

	tokSrv := httptest.NewServer(validTokenHandler(t))
	defer tokSrv.Close()

	// Install the token RoundTripper.
	apiSrv.Client().Transport = NewRoundTripper(apiSrv.Client().Transport, serviceAccountTestSource(tokSrv))

	// Deadline exceeded context
	ctx, cancel := context.WithTimeout(t.Context(), 0*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiSrv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = apiSrv.Client().Do(req)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func checkBearerAndEchoHandler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check authorization header.
		auth, ok := req.Header["Authorization"]
		if !ok || len(auth) != 1 {
			t.Error("authorization header is missing")
			return
		}
		if !strings.HasPrefix(auth[0], "Bearer") {
			t.Errorf("invalid bearer token: %s", auth[0])
			return
		}

		// Echo body back.
		if _, err := io.Copy(w, req.Body); err != nil {
			t.Errorf("failed to copy request body: %s", err)
		}
	})
}
