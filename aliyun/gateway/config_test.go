package gateway

import (
	"strings"
	"testing"
)

func TestConfValidate(t *testing.T) {
	tests := []struct {
		name    string
		conf    Conf
		wantErr string
		wantHost string
	}{
		{
			name: "valid",
			conf: Conf{
				Host:      "https://api.example.com/",
				AppKey:    "app-key",
				AppSecret: "app-secret",
			},
			wantErr:  "",
			wantHost: "https://api.example.com",
		},
		{
			name: "valid http",
			conf: Conf{
				Host:      "http://api.example.com",
				AppKey:    "app-key",
				AppSecret: "app-secret",
			},
			wantErr:  "",
			wantHost: "http://api.example.com",
		},
		{
			name: "missing host",
			conf: Conf{
				AppKey:    "app-key",
				AppSecret: "app-secret",
			},
			wantErr: "host is required",
		},
		{
			name: "invalid host scheme",
			conf: Conf{
				Host:      "ftp://api.example.com",
				AppKey:    "app-key",
				AppSecret: "app-secret",
			},
			wantErr: "host must start with http:// or https://",
		},
		{
			name: "missing app key",
			conf: Conf{
				Host:      "https://api.example.com",
				AppSecret: "app-secret",
			},
			wantErr: "app key is required",
		},
		{
			name: "missing app secret",
			conf: Conf{
				Host:   "https://api.example.com",
				AppKey: "app-key",
			},
			wantErr: "app secret is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conf.Validate()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantHost != "" && tt.conf.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", tt.conf.Host, tt.wantHost)
			}
		})
	}
}
