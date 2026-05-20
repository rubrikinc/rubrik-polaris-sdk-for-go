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

import "testing"

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
