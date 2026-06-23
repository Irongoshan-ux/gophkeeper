// Package model defines domain entities for GophKeeper.
package model

import "time"

// SecretType identifies the kind of stored secret.
type SecretType int16

const (
	SecretTypeUnspecified SecretType = 0
	SecretTypeCredentials SecretType = 1
	SecretTypeText        SecretType = 2
	SecretTypeBinary      SecretType = 3
	SecretTypeCard        SecretType = 4
	SecretTypeOTP         SecretType = 5
)

// User is a registered account.
type User struct {
	ID           string
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}

// Secret is an encrypted entry owned by a user.
type Secret struct {
	ID            string
	UserID        string
	Type          SecretType
	Name          string
	EncryptedData []byte
	Version       int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// IsDeleted reports whether the secret was soft-deleted.
func (s *Secret) IsDeleted() bool {
	return s.DeletedAt != nil
}

// Payload is the plaintext structure encrypted on the client before upload.
type Payload struct {
	Metadata string            `json:"metadata,omitempty"`
	Fields   map[string]string `json:"fields,omitempty"`
	Binary   []byte            `json:"binary,omitempty"`
}
