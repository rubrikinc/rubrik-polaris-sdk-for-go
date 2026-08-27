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

package cloudcluster

import (
	"encoding/json"
	"testing"
)

// TestVMTypeOmittedWhenEmpty verifies that an unset VMType is dropped from the
// request for every provider's VM config. The empty string is not a valid
// VmType enum member and would be rejected by GraphQL enum coercion, so it must
// be omitted to let the backend apply its default.
func TestVMTypeOmittedWhenEmpty(t *testing.T) {
	tests := map[string]any{
		"aws":   AwsVmConfig{},
		"azure": AzureVMConfig{},
	}

	for name, vmConfig := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(vmConfig)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if _, ok := got["vmType"]; ok {
				t.Errorf("vmType should be omitted when empty, got %v", got["vmType"])
			}
		})
	}

	// When set, vmType must serialize.
	data, err := json.Marshal(AwsVmConfig{VMType: CCVmConfigStandard})
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if got["vmType"] != "STANDARD" {
		t.Errorf("vmType = %v, want STANDARD", got["vmType"])
	}
}
