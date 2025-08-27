package logsql

import (
	"testing"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
)

// TestSetup tests the plugin setup configuration with various input scenarios.
func TestSetup(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid sqlite3 config",
			input:       `logsql sqlite3 ":memory:"`,
			expectError: false,
		},
		{
			name:        "missing dialect",
			input:       `logsql`,
			expectError: true,
		},
		{
			name:        "missing DSN",
			input:       `logsql sqlite3`,
			expectError: true,
		},
		{
			name:        "extra arguments",
			input:       `logsql sqlite3 ":memory:" extra`,
			expectError: true,
		},
		{
			name:        "unsupported dialect",
			input:       `logsql mysql "user:pass@/dbname"`,
			expectError: true,
		},
		{
			name:        "invalid sqlite3 DSN",
			input:       `logsql sqlite3 "/invalid/path/to/db.sqlite"`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := caddy.NewTestController("dns", tt.input)
			err := setup(c)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				// Check that plugin was added to config
				cfg := dnsserver.GetConfig(c)
				if len(cfg.Plugin) == 0 {
					t.Error("Expected plugin to be added to config")
				}
			}
		})
	}
}

