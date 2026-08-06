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
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	internalerrors "github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/errors"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func TestRequestWithContext(t *testing.T) {
	srv := httptest.NewServer(validTokenHandler(t))
	defer srv.Close()

	_, err := RequestWithContext(t.Context(), srv.Client(), srv.URL, []byte{}, &log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRequestWithContextNoResponse(t *testing.T) {
	srvCtx, srvCancel := context.WithCancel(t.Context())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-srvCtx.Done()
	}))
	defer srv.Close()
	defer srvCancel()

	ctx := context.WithValue(t.Context(), timeoutKey{}, 10*time.Millisecond)
	_, err := RequestWithContext(ctx, srv.Client(), srv.URL, []byte{}, &log.DiscardLogger{})
	if err == nil || err.Error() != "failed to acquire access token after 3 attempts" {
		t.Fatal(err)
	}
}

func TestRequestWithContextAndJSONError(t *testing.T) {
	srv := httptest.NewServer(serviceAccountBadCredentialsHandler(t))
	defer srv.Close()

	_, err := RequestWithContext(t.Context(), srv.Client(), srv.URL, []byte{}, &log.DiscardLogger{})
	var jsonErr internalerrors.JSONError
	if !errors.As(err, &jsonErr) {
		t.Fatal(err)
	}
}

func TestRequestWithCanceledContext(t *testing.T) {
	srv := httptest.NewServer(validTokenHandler(t))
	defer srv.Close()

	// Canceled context.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := RequestWithContext(ctx, srv.Client(), srv.URL, []byte{}, &log.DiscardLogger{})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestRequestWithDeadlineExceededContext(t *testing.T) {
	srv := httptest.NewServer(validTokenHandler(t))
	defer srv.Close()

	// Deadline exceeded context.
	ctx, cancel := context.WithTimeout(t.Context(), 0*time.Second)
	defer cancel()

	_, err := RequestWithContext(ctx, srv.Client(), srv.URL, []byte{}, &log.DiscardLogger{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}
