// Package crypto provides client-side encryption for secret payloads.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"

	"github.com/Irongoshan-ux/gophkeeper/internal/model"
)

const (
	scryptN = 32768
	scryptR = 8
	scryptP = 1
	keyLen  = 32
)

var (
	// ErrInvalidCiphertext is returned when decryption fails.
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	// ErrEmptyMasterPassword is returned when master password is blank.
	ErrEmptyMasterPassword = errors.New("master password is required")
)

// DeriveKey derives a 32-byte AES key from master password and salt.
func DeriveKey(masterPassword string, salt []byte) ([]byte, error) {
	if masterPassword == "" {
		return nil, ErrEmptyMasterPassword
	}
	return scrypt.Key([]byte(masterPassword), salt, scryptN, scryptR, scryptP, keyLen)
}

// NewSalt generates a random salt for key derivation.
func NewSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// Encrypt serializes payload and encrypts it with AES-GCM.
func Encrypt(key []byte, payload *model.Payload) ([]byte, error) {
	plain, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plain, nil), nil
}

// Decrypt decrypts ciphertext and unmarshals the payload.
func Decrypt(key, ciphertext []byte) (*model.Payload, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, ErrInvalidCiphertext
	}
	nonce, data := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}
	var payload model.Payload
	if err := json.Unmarshal(plain, &payload); err != nil {
		return nil, ErrInvalidCiphertext
	}
	return &payload, nil
}
