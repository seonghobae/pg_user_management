package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		wantConfig  *Config
	}{
		{
			name: "Valid config with all env vars",
			envVars: map[string]string{
				"PGHOST":     "testhost",
				"PGPORT":     "5433",
				"PGUSER":     "testuser",
				"PGPASSWORD": "testpass",
				"PGDATABASE": "testdb",
				"PGSSLMODE":  "require",
			},
			expectError: false,
			wantConfig: &Config{
				Host:     "testhost",
				Port:     "5433",
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  "require",
			},
		},
		{
			name: "Config with default values",
			envVars: map[string]string{
				"PGPASSWORD": "testpass",
			},
			expectError: false,
			wantConfig: &Config{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "testpass",
				Database: "postgres",
				SSLMode:  "disable",
			},
		},
		{
			name:        "Missing password",
			envVars:     map[string]string{},
			expectError: true,
			wantConfig:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			got, err := NewConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("NewConfig() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("NewConfig() unexpected error: %v", err)
				}
				if got == nil {
					t.Fatal("NewConfig() returned nil config")
				}
				if got.Host != tt.wantConfig.Host {
					t.Errorf("Host = %v, want %v", got.Host, tt.wantConfig.Host)
				}
				if got.Port != tt.wantConfig.Port {
					t.Errorf("Port = %v, want %v", got.Port, tt.wantConfig.Port)
				}
				if got.User != tt.wantConfig.User {
					t.Errorf("User = %v, want %v", got.User, tt.wantConfig.User)
				}
				if got.Password != tt.wantConfig.Password {
					t.Errorf("Password = %v, want %v", got.Password, tt.wantConfig.Password)
				}
				if got.Database != tt.wantConfig.Database {
					t.Errorf("Database = %v, want %v", got.Database, tt.wantConfig.Database)
				}
				if got.SSLMode != tt.wantConfig.SSLMode {
					t.Errorf("SSLMode = %v, want %v", got.SSLMode, tt.wantConfig.SSLMode)
				}
			}
		})
	}
}

func TestConfig_ConnectionString(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   string
	}{
		{
			name: "Standard connection string",
			config: &Config{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "secret",
				Database: "testdb",
				SSLMode:  "disable",
			},
			want: "host=localhost port=5432 user=postgres password=secret dbname=testdb sslmode=disable",
		},
		{
			name: "Connection string with SSL",
			config: &Config{
				Host:     "prod.example.com",
				Port:     "5432",
				User:     "admin",
				Password: "pass123",
				Database: "production",
				SSLMode:  "require",
			},
			want: "host=prod.example.com port=5432 user=admin password=pass123 dbname=production sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ConnectionString()
			if got != tt.want {
				t.Errorf("ConnectionString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		want         string
	}{
		{
			name:         "Environment variable exists",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			want:         "custom",
		},
		{
			name:         "Environment variable does not exist",
			key:          "TEST_VAR_MISSING",
			defaultValue: "default",
			envValue:     "",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			got := getEnv(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
