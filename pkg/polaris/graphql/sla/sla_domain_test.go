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

package sla

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestObjectSpecificConfigsOmitZeroRetention verifies that object-specific
// SLA config structs marshal zero-valued RetentionDuration fields as omitted,
// not as {"duration":0,"unit":""}. The latter is rejected by the GraphQL
// backend because "" is not a valid RetentionUnit enum.
func TestObjectSpecificConfigsOmitZeroRetention(t *testing.T) {
	tests := []struct {
		name        string
		value       any
		mustHave    []string
		mustNotHave []string
	}{
		{
			name: "SapHanaConfig with only IncrementalFrequency",
			value: SapHanaConfig{
				IncrementalFrequency: RetentionDuration{Duration: 1, Unit: Days},
			},
			mustHave:    []string{`"incrementalFrequency"`},
			mustNotHave: []string{`"logRetention"`, `"differentialFrequency"`, `"unit":""`},
		},
		{
			name:        "SapHanaConfig empty",
			value:       SapHanaConfig{},
			mustNotHave: []string{`"incrementalFrequency"`, `"logRetention"`, `"differentialFrequency"`, `"unit":""`},
		},
		{
			name: "DB2Config with only LogRetention",
			value: DB2Config{
				LogRetention: RetentionDuration{Duration: 7, Unit: Days},
			},
			mustHave:    []string{`"logRetention"`},
			mustNotHave: []string{`"incrementalFrequency"`, `"differentialFrequency"`, `"unit":""`},
		},
		{
			name: "InformixSlaConfig with only LogFrequency",
			value: InformixSlaConfig{
				LogFrequency: RetentionDuration{Duration: 1, Unit: Hours},
			},
			mustHave: []string{`"logFrequency"`},
			mustNotHave: []string{
				`"incrementalFrequency"`, `"incrementalRetention"`, `"logRetention"`, `"unit":""`,
			},
		},
		{
			name: "OracleConfig with only Frequency and LogRetention",
			value: OracleConfig{
				Frequency:    RetentionDuration{Duration: 1, Unit: Days},
				LogRetention: RetentionDuration{Duration: 7, Unit: Days},
			},
			mustHave:    []string{`"frequency"`, `"logRetention"`},
			mustNotHave: []string{`"hostLogRetention"`, `"unit":""`},
		},
		{
			name:        "SapHanaStorageSnapshotConfig empty",
			value:       SapHanaStorageSnapshotConfig{},
			mustNotHave: []string{`"frequency"`, `"retention"`, `"unit":""`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}
			s := string(data)
			for _, want := range tt.mustHave {
				if !strings.Contains(s, want) {
					t.Errorf("expected JSON to contain %s, got %s", want, s)
				}
			}
			for _, unwant := range tt.mustNotHave {
				if strings.Contains(s, unwant) {
					t.Errorf("expected JSON to not contain %s, got %s", unwant, s)
				}
			}
		})
	}
}
