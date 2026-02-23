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
	"testing"
)

func TestTagPairUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  TagPair
	}{
		{
			name:  "sentinel only is replaced with empty slice",
			input: `{"key":"env","matchAllTagValues":true,"values":["ALL_TAG_OR_LABEL_VALUES_OPTION"]}`,
			want:  TagPair{Key: "env", MatchAllTagValues: true, Values: []string{}},
		},
		{
			name:  "normal values are preserved unchanged",
			input: `{"key":"env","matchAllTagValues":false,"values":["prod","staging","dev"]}`,
			want:  TagPair{Key: "env", MatchAllTagValues: false, Values: []string{"prod", "staging", "dev"}},
		},
		{
			name:  "empty values slice is preserved",
			input: `{"key":"env","matchAllTagValues":true,"values":[]}`,
			want:  TagPair{Key: "env", MatchAllTagValues: true, Values: []string{}},
		},
		{
			name:  "absent values field results in nil slice",
			input: `{"key":"env","matchAllTagValues":false}`,
			want:  TagPair{Key: "env", MatchAllTagValues: false, Values: nil},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got TagPair
			if err := json.Unmarshal([]byte(tc.input), &got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Key != tc.want.Key {
				t.Errorf("Key: got %q, want %q", got.Key, tc.want.Key)
			}
			if got.MatchAllTagValues != tc.want.MatchAllTagValues {
				t.Errorf("MatchAllTagValues: got %v, want %v", got.MatchAllTagValues, tc.want.MatchAllTagValues)
			}
			if len(got.Values) != len(tc.want.Values) {
				t.Errorf("Values length: got %d (%v), want %d (%v)", len(got.Values), got.Values, len(tc.want.Values), tc.want.Values)
				return
			}
			for i := range tc.want.Values {
				if got.Values[i] != tc.want.Values[i] {
					t.Errorf("Values[%d]: got %q, want %q", i, got.Values[i], tc.want.Values[i])
				}
			}
		})
	}
}
