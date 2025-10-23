package auth

import (
	"fmt"
)

// AuthMethod represents the authentication method
type AuthMethod string

const (
	MD5         AuthMethod = "md5"
	ScramSHA256 AuthMethod = "scram-sha-256"
)

// ValidateAuthMethod checks if the authentication method is valid
func ValidateAuthMethod(method string) (AuthMethod, error) {
	switch AuthMethod(method) {
	case MD5, ScramSHA256:
		return AuthMethod(method), nil
	default:
		return "", fmt.Errorf("invalid authentication method: %s (must be 'md5' or 'scram-sha-256')", method)
	}
}

// GetPasswordEncryption returns the PostgreSQL password_encryption setting for the auth method
func (am AuthMethod) GetPasswordEncryption() string {
	switch am {
	case MD5:
		return "md5"
	case ScramSHA256:
		return "scram-sha-256"
	default:
		return "scram-sha-256" // default to more secure option
	}
}
