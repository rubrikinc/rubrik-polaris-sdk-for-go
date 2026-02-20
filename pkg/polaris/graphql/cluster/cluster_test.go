package cluster

import (
	"testing"
)

func TestParseTimeZone(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:      "ValidTimezoneAmericaNewYork",
			input:     "America/New_York",
			expected:  "CLUSTER_TIMEZONE_AMERICA_NEW_YORK",
			expectErr: false,
		},
		{
			name:      "ValidTimezoneEuropeLondon",
			input:     "Europe/London",
			expected:  "CLUSTER_TIMEZONE_EUROPE_LONDON",
			expectErr: false,
		},
		{
			name:      "ValidTimezoneUTC",
			input:     "UTC",
			expected:  "CLUSTER_TIMEZONE_UTC",
			expectErr: false,
		},
		{
			name:      "InvalidTimezone",
			input:     "Invalid/Timezone",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "EmptyString",
			input:     "",
			expected:  "CLUSTER_TIMEZONE_UTC",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeZone(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result.String() != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result.String())
				}
			}
		})
	}
}

func TestTimezoneString(t *testing.T) {
	tests := []struct {
		name     string
		timezone Timezone
		expected string
	}{
		{
			name:     "TimezoneWithSlash",
			timezone: Timezone("America/New_York"),
			expected: "CLUSTER_TIMEZONE_AMERICA_NEW_YORK",
		},
		{
			name:     "TimezoneUTC",
			timezone: Timezone("UTC"),
			expected: "CLUSTER_TIMEZONE_UTC",
		},
		{
			name:     "EmptyTimezone",
			timezone: Timezone(""),
			expected: "CLUSTER_TIMEZONE_UNSPECIFIED",
		},
		{
			name:     "TimezoneWithMultipleSlashes",
			timezone: Timezone("America/Indiana/Indianapolis"),
			expected: "CLUSTER_TIMEZONE_AMERICA_INDIANA_INDIANAPOLIS",
		},
		{
			name:     "LowercaseTimezone",
			timezone: Timezone("europe/london"),
			expected: "CLUSTER_TIMEZONE_EUROPE_LONDON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.timezone.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
