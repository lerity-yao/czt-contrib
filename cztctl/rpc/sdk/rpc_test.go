package sdk

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCheckGoctl(t *testing.T) {
	_, lookErr := exec.LookPath("goctl")
	err := CheckGoctl()

	if lookErr != nil {
		// goctl not installed
		if err == nil {
			t.Fatal("expected error when goctl is not installed, got nil")
		}
		if !strings.Contains(err.Error(), "not installed") {
			t.Fatalf("expected error containing 'not installed', got: %v", err)
		}
	} else {
		// goctl installed — should return nil (warning is printed but not an error)
		if err != nil {
			t.Fatalf("expected nil error when goctl is installed, got: %v", err)
		}
	}
}

func TestParseGoctlVersion(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantVer   string
		wantError bool
	}{
		{
			name:      "standard linux output",
			input:     "goctl version 1.7.3 linux/amd64",
			wantVer:   "1.7.3",
			wantError: false,
		},
		{
			name:      "standard darwin output",
			input:     "goctl version 1.10.5 darwin/arm64",
			wantVer:   "1.10.5",
			wantError: false,
		},
		{
			name:      "version only",
			input:     "1.7.3",
			wantVer:   "1.7.3",
			wantError: false,
		},
		{
			name:      "trailing newline",
			input:     "goctl version 1.7.3 linux/amd64\n",
			wantVer:   "1.7.3",
			wantError: false,
		},
		{
			name:      "no valid version",
			input:     "goctl unknown",
			wantVer:   "",
			wantError: true,
		},
		{
			name:      "empty string",
			input:     "",
			wantVer:   "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGoctlVersion(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("parseGoctlVersion(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
				return
			}
			if got != tt.wantVer {
				t.Errorf("parseGoctlVersion(%q) = %q, want %q", tt.input, got, tt.wantVer)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{name: "equal", v1: "1.7.3", v2: "1.7.3", want: 0},
		{name: "major less", v1: "0.9.0", v2: "1.0.0", want: -1},
		{name: "minor less", v1: "1.6.0", v2: "1.7.0", want: -1},
		{name: "patch less", v1: "1.7.2", v2: "1.7.3", want: -1},
		{name: "minor greater", v1: "1.8.0", v2: "1.7.0", want: 1},
		{name: "patch greater", v1: "1.7.4", v2: "1.7.3", want: 1},
		{name: "minor double digit greater", v1: "1.10.0", v2: "1.9.0", want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}
