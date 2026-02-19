package cluster

import "testing"

func TestTimezoneToFriendlyName(t *testing.T) {
	tests := []struct {
		name     string
		timezone Timezone
		want     string
	}{
		{
			name:     "unspecified timezone returns Unspecified",
			timezone: TIMEZONE_UNSPECIFIED,
			want:     "Unspecified",
		},
		{
			name:     "UTC timezone returns UTC",
			timezone: TIMEZONE_UTC,
			want:     "UTC",
		},
		{
			name:     "single underscore replaced with slash",
			timezone: TIMEZONE_AMERICA_CHICAGO,
			want:     "AMERICA/CHICAGO",
		},
		{
			name:     "multiple underscores replaced with slashes",
			timezone: TIMEZONE_AMERICA_ARGENTINA_BUENOS_AIRES,
			want:     "AMERICA/ARGENTINA_BUENOS_AIRES",
		},
		{
			name:     "timezone with two underscores",
			timezone: TIMEZONE_AMERICA_NEW_YORK,
			want:     "AMERICA/NEW_YORK",
		},
		{
			name:     "Asia timezone",
			timezone: TIMEZONE_ASIA_HONG_KONG,
			want:     "ASIA/HONG_KONG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TimezoneToFriendlyName(tt.timezone)
			if got != tt.want {
				t.Errorf("TimezoneToFriendlyName(%q) = %q, want %q", tt.timezone, got, tt.want)
			}
		})
	}
}

func TestParseTimezone(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantTz Timezone
		wantOk bool
	}{
		{
			name:   "valid UTC timezone",
			input:  "CLUSTER_TIMEZONE_UTC",
			wantTz: TIMEZONE_UTC,
			wantOk: true,
		},
		{
			name:   "valid America/New_York timezone",
			input:  "CLUSTER_TIMEZONE_AMERICA_NEW_YORK",
			wantTz: TIMEZONE_AMERICA_NEW_YORK,
			wantOk: true,
		},
		{
			name:   "valid unspecified timezone",
			input:  "CLUSTER_TIMEZONE_UNSPECIFIED",
			wantTz: TIMEZONE_UNSPECIFIED,
			wantOk: true,
		},
		{
			name:   "invalid timezone",
			input:  "INVALID_TIMEZONE",
			wantTz: Timezone("INVALID_TIMEZONE"),
			wantOk: false,
		},
		{
			name:   "empty string",
			input:  "",
			wantTz: Timezone(""),
			wantOk: false,
		},
		{
			name:   "timezone without prefix",
			input:  "AMERICA_NEW_YORK",
			wantTz: Timezone("AMERICA_NEW_YORK"),
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTz, gotOk := ParseTimezone(tt.input)
			if gotTz != tt.wantTz {
				t.Errorf("ParseTimezone(%q) timezone = %q, want %q", tt.input, gotTz, tt.wantTz)
			}
			if gotOk != tt.wantOk {
				t.Errorf("ParseTimezone(%q) ok = %v, want %v", tt.input, gotOk, tt.wantOk)
			}
		})
	}
}

func TestIsValidTimezone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid UTC timezone",
			input: "CLUSTER_TIMEZONE_UTC",
			want:  true,
		},
		{
			name:  "valid Asia timezone",
			input: "CLUSTER_TIMEZONE_ASIA_TOKYO",
			want:  true,
		},
		{
			name:  "valid Europe timezone",
			input: "CLUSTER_TIMEZONE_EUROPE_LONDON",
			want:  true,
		},
		{
			name:  "invalid timezone",
			input: "NOT_A_TIMEZONE",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "partial match",
			input: "CLUSTER_TIMEZONE_",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTimezone(tt.input)
			if got != tt.want {
				t.Errorf("IsValidTimezone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
