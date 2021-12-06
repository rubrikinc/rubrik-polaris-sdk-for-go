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
	"testing"
	"text/template"
	"time"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testnet"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func TestParseFeature(t *testing.T) {
	feature, err := ParseFeature("CLOUD_NATIVE_PROTECTION")
	if err != nil {
		t.Error(err)
	}
	if feature != FeatureCloudNativeProtection {
		t.Errorf("invalid feature: %s", feature)
	}

	feature, err = ParseFeature("cloud_native_protection")
	if err != nil {
		t.Error(err)
	}
	if feature != FeatureCloudNativeProtection {
		t.Errorf("invalid feature: %s", feature)
	}

	feature, err = ParseFeature("cloud-native-protection")
	if err != nil {
		t.Error(err)
	}
	if feature != FeatureCloudNativeProtection {
		t.Errorf("invalid feature: %s", feature)
	}

	feature, err = ParseFeature("invalid-feature")
	if err == nil {
		t.Error("expected test to fail")
	}
	if feature != FeatureInvalid {
		t.Errorf("invalid feature: %s", feature)
	}
}

func TestKorgTaskChainStatus(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/korgtaskchainstatus.json")
	if err != nil {
		t.Fatal(err)
	}

	client, lis := graphql.NewTestClient("john", "doe", log.DiscardLogger{})
	coreAPI := Wrap(client)

	// Respond with status code 200 and a valid body.
	srv := testnet.TestServeJSONWithToken(lis, func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var payload struct {
			Query     string `json:"query"`
			Variables struct {
				TaskChainID string `json:"taskchainId,omitempty"`
			} `json:"variables,omitempty"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			t.Fatal(err)
		}

		err = tmpl.Execute(w, struct {
			ChainState string
			ChainUUID  string
		}{ChainState: "RUNNING", ChainUUID: payload.Variables.TaskChainID})
		if err != nil {
			panic(err)
		}
	})
	defer srv.Shutdown(context.Background())

	id := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	taskChain, err := coreAPI.KorgTaskChainStatus(context.Background(), id)
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
	tmpl, err := template.ParseFiles("testdata/korgtaskchainstatus.json")
	if err != nil {
		t.Fatal(err)
	}

	client, lis := graphql.NewTestClient("john", "doe", log.DiscardLogger{})
	coreAPI := Wrap(client)

	// Respond with status code 200 and a valid body. First 2 reponses have
	// state RUNNING. Third response is SUCCEEDED.
	reqCount := 3
	srv := testnet.TestServeJSONWithToken(lis, func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var payload struct {
			Query     string `json:"query"`
			Variables struct {
				TaskChainID string `json:"taskchainId,omitempty"`
			} `json:"variables,omitempty"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			t.Fatal(err)
		}

		reqCount--
		chainState := "RUNNING"
		if reqCount == 0 {
			chainState = "SUCCEEDED"
		}
		tmpl.Execute(w, struct {
			ChainState string
			ChainUUID  string
		}{ChainState: chainState, ChainUUID: payload.Variables.TaskChainID})
		if err != nil {
			panic(err)
		}
	})
	defer srv.Shutdown(context.Background())

	id := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	state, err := coreAPI.WaitForTaskChain(context.Background(), id, 5*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if state != "SUCCEEDED" {
		t.Errorf("invalid task chain state: %v", state)
	}
}
