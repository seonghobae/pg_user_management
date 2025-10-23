package auth

import (
	"testing"
)

func TestValidateAuthMethod(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		wantMethod  AuthMethod
		expectError bool
	}{
		{
			name:        "Valid MD5 method",
			method:      "md5",
			wantMethod:  MD5,
			expectError: false,
		},
		{
			name:        "Valid SCRAM-SHA-256 method",
			method:      "scram-sha-256",
			wantMethod:  ScramSHA256,
			expectError: false,
		},
		{
			name:        "Invalid method",
			method:      "invalid",
			wantMethod:  "",
			expectError: true,
		},
		{
			name:        "Empty method",
			method:      "",
			wantMethod:  "",
			expectError: true,
		},
		{
			name:        "Case sensitive - uppercase",
			method:      "MD5",
			wantMethod:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAuthMethod(tt.method)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateAuthMethod() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateAuthMethod() unexpected error: %v", err)
				}
				if got != tt.wantMethod {
					t.Errorf("ValidateAuthMethod() = %v, want %v", got, tt.wantMethod)
				}
			}
		})
	}
}

func TestAuthMethod_GetPasswordEncryption(t *testing.T) {
	tests := []struct {
		name       string
		authMethod AuthMethod
		want       string
	}{
		{
			name:       "MD5 encryption",
			authMethod: MD5,
			want:       "md5",
		},
		{
			name:       "SCRAM-SHA-256 encryption",
			authMethod: ScramSHA256,
			want:       "scram-sha-256",
		},
		{
			name:       "Invalid method defaults to scram-sha-256",
			authMethod: AuthMethod("invalid"),
			want:       "scram-sha-256",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.authMethod.GetPasswordEncryption()
			if got != tt.want {
				t.Errorf("GetPasswordEncryption() = %v, want %v", got, tt.want)
			}
		})
	}
}
