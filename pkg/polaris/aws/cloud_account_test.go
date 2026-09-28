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

package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// TestToCloudAccountKeepsConfigProtection guards against the Cloud Native
// Config Protection feature being dropped from SupportedFeatures. The list
// filters every account read, so a missing entry makes an onboarded feature
// invisible to AccountByID and friends rather than failing loudly.
func TestToCloudAccountKeepsConfigProtection(t *testing.T) {
	account := toCloudAccount(aws.CloudAccountWithFeatures{
		Features: []aws.Feature{{
			Feature: core.FeatureCloudNativeConfigProtection.Name,
			Status:  core.StatusConnecting,
		}},
	})
	if _, ok := account.Feature(core.FeatureCloudNativeConfigProtection); !ok {
		t.Error("CLOUD_NATIVE_CONFIG_PROTECTION dropped by toCloudAccount")
	}
}

// TestDisableFeatureDisablesConfigProtection verifies that disabling the Cloud
// Native Config Protection feature starts a CONFIG native account disable job.
// RSC refuses to disable Cloud Discovery while the feature is still enabled, so
// skipping it makes account removal fail.
func TestDisableFeatureDisablesConfigProtection(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	var disabled []string
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}
		var payload struct {
			Query     string `json:"query"`
			Variables struct {
				Feature string `json:"awsNativeProtectionFeature"`
			} `json:"variables"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}
		switch {
		case strings.Contains(payload.Query, "startAwsNativeAccountDisableJob"):
			disabled = append(disabled, payload.Variables.Feature)
			if _, err := fmt.Fprint(w, `{"data":{"startAwsNativeAccountDisableJob":{"error":"","jobId":"01a0e796-2fe3-75a3-9f1a-aa6cadec1888"}}}`); err != nil {
				cancel(err)
			}
		case strings.Contains(payload.Query, "getKorgTaskchainStatus"):
			if _, err := fmt.Fprint(w, `{"data":{"getKorgTaskchainStatus":{"taskchain":{"id":1,"state":"SUCCEEDED","taskchainUuid":"01a0e796-2fe3-75a3-9f1a-aa6cadec1888"}}}}`); err != nil {
				cancel(err)
			}
		default:
			cancel(fmt.Errorf("unexpected query: %s", payload.Query))
		}
	}))
	defer srv.Close()

	account := CloudAccount{
		ID: uuid.MustParse("e381bc19-ea11-493c-9751-2252116aef7d"),
		Features: []Feature{{
			Feature: core.FeatureCloudNativeConfigProtection,
			Status:  core.StatusConnected,
		}},
	}
	if err := Wrap(mockClient(srv)).disableFeature(ctx, account, core.FeatureCloudNativeConfigProtection, false); err != nil {
		t.Fatal(err)
	}
	if len(disabled) != 1 || disabled[0] != string(aws.Config) {
		t.Errorf("expected one %s disable job, got %v", aws.Config, disabled)
	}
}

func TestFeatureOnboardingMode(t *testing.T) {
	tests := []struct {
		name     string
		feature  Feature
		expected OnboardingMode
	}{{
		name:     "Empty stack ARN is IAM",
		feature:  Feature{},
		expected: OnboardingModeIAM,
	}, {
		name:     "Non-empty stack ARN is CFT",
		feature:  Feature{StackArn: "arn:aws:cloudformation:us-east-1:123456789012:stack/rsc"},
		expected: OnboardingModeCFT,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if mode := tc.feature.OnboardingMode(); mode != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, mode)
			}
		})
	}
}

func TestCloudAccountOnboardingMode(t *testing.T) {
	cft := Feature{StackArn: "arn:aws:cloudformation:us-east-1:123456789012:stack/rsc"}
	iam := Feature{}

	tests := []struct {
		name     string
		account  CloudAccount
		expected OnboardingMode
	}{{
		name:     "No features is IAM",
		account:  CloudAccount{},
		expected: OnboardingModeIAM,
	}, {
		name:     "Single IAM feature is IAM",
		account:  CloudAccount{Features: []Feature{iam}},
		expected: OnboardingModeIAM,
	}, {
		name:     "Multiple IAM features is IAM",
		account:  CloudAccount{Features: []Feature{iam, iam, iam}},
		expected: OnboardingModeIAM,
	}, {
		name:     "Single CFT feature is CFT",
		account:  CloudAccount{Features: []Feature{cft}},
		expected: OnboardingModeCFT,
	}, {
		name:     "Multiple CFT features is CFT",
		account:  CloudAccount{Features: []Feature{cft, cft}},
		expected: OnboardingModeCFT,
	}, {
		name:     "Mixed CFT and IAM features is CFT",
		account:  CloudAccount{Features: []Feature{iam, cft, iam}},
		expected: OnboardingModeCFT,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if mode := tc.account.OnboardingMode(); mode != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, mode)
			}
		})
	}
}
