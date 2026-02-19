package cluster

import "testing"

func TestTimezoneToFriendlyName(t *testing.T) {
	tests := []struct {
		name     string
		timezone Timezone
		want     string
	}{
		{
			name:     "unspecified timezone returns empty string",
			timezone: TIMEZONE_UNSPECIFIED,
			want:     "",
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
			name:     "timezone with two underscores in city name",
			timezone: TIMEZONE_AMERICA_NEW_YORK,
			want:     "AMERICA/NEW_YORK",
		},
		{
			name:     "Asia timezone with underscore in city name",
			timezone: TIMEZONE_ASIA_HONG_KONG,
			want:     "ASIA/HONG_KONG",
		},
		{
			name:     "Argentina subregion has two slashes",
			timezone: TIMEZONE_AMERICA_ARGENTINA_BUENOS_AIRES,
			want:     "AMERICA/ARGENTINA/BUENOS_AIRES",
		},
		{
			name:     "Indiana subregion has two slashes",
			timezone: TIMEZONE_AMERICA_INDIANA_INDIANAPOLIS,
			want:     "AMERICA/INDIANA/INDIANAPOLIS",
		},
		{
			name:     "Kentucky subregion has two slashes",
			timezone: TIMEZONE_AMERICA_KENTUCKY_LOUISVILLE,
			want:     "AMERICA/KENTUCKY/LOUISVILLE",
		},
		{
			name:     "North Dakota subregion has two slashes",
			timezone: TIMEZONE_AMERICA_NORTH_DAKOTA_NEW_SALEM,
			want:     "AMERICA/NORTH_DAKOTA/NEW_SALEM",
		},
		{
			name:     "North Dakota Center has two slashes",
			timezone: TIMEZONE_AMERICA_NORTH_DAKOTA_CENTER,
			want:     "AMERICA/NORTH_DAKOTA/CENTER",
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
