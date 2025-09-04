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

package core

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

func TestParseFeatureNoValidation(t *testing.T) {
	if feature := ParseFeatureNoValidation("CLOUD_NATIVE_PROTECTION"); !feature.Equal(FeatureCloudNativeProtection) {
		t.Errorf("invalid feature: %s", feature)
	}

	if feature := ParseFeatureNoValidation("cloud_native_protection"); !feature.Equal(FeatureCloudNativeProtection) {
		t.Errorf("invalid feature: %s", feature)
	}

	if feature := ParseFeatureNoValidation("cloud-native-protection"); !feature.Equal(FeatureCloudNativeProtection) {
		t.Errorf("invalid feature: %s", feature)
	}
}

func TestKorgTaskChainStatus(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/korg_taskchain_status_response.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	// Respond with status code 200 and a valid body.
	srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}

		var payload struct {
			Query     string `json:"query"`
			Variables struct {
				TaskChainID string `json:"taskchainId,omitempty"`
			} `json:"variables,omitempty"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}

		if err := tmpl.Execute(w, struct {
			ChainState string
			ChainUUID  string
		}{ChainState: "RUNNING", ChainUUID: payload.Variables.TaskChainID}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	id := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	taskChain, err := Wrap(graphql.NewTestClient(srv)).KorgTaskChainStatus(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if taskChain.ID != 12761540 {
		t.Errorf("invalid task chain id: %v", taskChain.ID)
	}
	if taskChain.TaskChainID != id {
		t.Errorf("invalid task chain uuid: %v", taskChain.TaskChainID)
	}
	if taskChain.State != TaskChainRunning {
		t.Errorf("invalid task chain state: %v", taskChain.State)
	}
}

func TestWaitForTaskChain(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/korg_taskchain_status_response.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	// Respond with status code 200 and a valid body. The First 2 responses have
	// state RUNNING. The Third response is SUCCEEDED.
	reqCount := 3
	srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}

		var payload struct {
			Query     string `json:"query"`
			Variables struct {
				TaskChainID string `json:"taskchainId,omitempty"`
			} `json:"variables,omitempty"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}

		reqCount--
		chainState := "RUNNING"
		if reqCount == 0 {
			chainState = "SUCCEEDED"
		}
		if err := tmpl.Execute(w, struct {
			ChainState string
			ChainUUID  string
		}{ChainState: chainState, ChainUUID: payload.Variables.TaskChainID}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	id := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	state, err := Wrap(graphql.NewTestClient(srv)).WaitForTaskChain(ctx, id, 5*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if state != "SUCCEEDED" {
		t.Errorf("invalid task chain state: %v", state)
	}
}

func TestFormatTimestampToRFC3339(t *testing.T) {
	now := time.Now()
	timestamp := FormatTimestampToRFC3339(now)
	if timestamp != now.UTC().Format("2006-01-02T15:04:05.000Z") {
		t.Errorf("invalid timestamp: %v", timestamp)
	}
}
