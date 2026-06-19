// Package service implements GophKeeper business logic.
package service

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Irongoshan-ux/gophkeeper/internal/auth"
	"github.com/Irongoshan-ux/gophkeeper/internal/model"
	"github.com/Irongoshan-ux/gophkeeper/internal/repository"
	"github.com/Irongoshan-ux/gophkeeper/internal/validation"
)

// AuthService handles user registration and login.
type AuthService struct {
	users  repository.UserRepository
	secret string
}

// NewAuthService creates an AuthService.
func NewAuthService(users repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{users: users, secret: jwtSecret}
}

// Register creates a user and returns a JWT token.
func (s *AuthService) Register(ctx context.Context, login, password string) (token, userID string, err error) {
	if err := validation.Credentials(login, password); err != nil {
		return "", "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("hash password: %w", err)
	}
	u, err := s.users.Create(ctx, login, string(hash))
	if err != nil {
		return "", "", err
	}
	token, err = auth.NewToken(s.secret, u.ID)
	if err != nil {
		return "", "", err
	}
	return token, u.ID, nil
}

// Login authenticates a user and returns a JWT token.
func (s *AuthService) Login(ctx context.Context, login, password string) (token, userID string, err error) {
	if err := validation.Credentials(login, password); err != nil {
		return "", "", err
	}
	u, err := s.users.GetByLogin(ctx, login)
	if err != nil {
		if err == repository.ErrNotFound {
			return "", "", repository.ErrNotFound
		}
		return "", "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", "", repository.ErrNotFound
	}
	token, err = auth.NewToken(s.secret, u.ID)
	if err != nil {
		return "", "", err
	}
	return token, u.ID, nil
}

// SecretService manages encrypted secrets for authenticated users.
type SecretService struct {
	secrets repository.SecretRepository
}

// NewSecretService creates a SecretService.
func NewSecretService(secrets repository.SecretRepository) *SecretService {
	return &SecretService{secrets: secrets}
}

// Create stores a new encrypted secret for the user in context.
func (s *SecretService) Create(ctx context.Context, secretType model.SecretType, name string, encrypted []byte) (*model.Secret, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := validation.SecretName(name); err != nil {
		return nil, err
	}
	if len(encrypted) == 0 {
		return nil, fmt.Errorf("encrypted data is required")
	}
	return s.secrets.Create(ctx, &model.Secret{
		UserID:        userID,
		Type:          secretType,
		Name:          name,
		EncryptedData: encrypted,
		Version:       1,
	})
}

// Get returns a secret by id for the authenticated user.
func (s *SecretService) Get(ctx context.Context, id string) (*model.Secret, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.secrets.Get(ctx, userID, id)
}

// List returns all active secrets for the authenticated user.
func (s *SecretService) List(ctx context.Context) ([]*model.Secret, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.secrets.List(ctx, userID)
}

// Update replaces secret fields with optimistic locking via version.
func (s *SecretService) Update(ctx context.Context, id string, secretType model.SecretType, name string, encrypted []byte, version int64) (*model.Secret, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := validation.SecretName(name); err != nil {
		return nil, err
	}
	existing, err := s.secrets.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if version > 0 && existing.Version != version {
		return nil, repository.ErrConflict
	}
	return s.secrets.Update(ctx, &model.Secret{
		ID:            id,
		UserID:        userID,
		Type:          secretType,
		Name:          name,
		EncryptedData: encrypted,
		Version:       existing.Version + 1,
	})
}

// Delete soft-deletes a secret.
func (s *SecretService) Delete(ctx context.Context, id string, version int64) (*model.Secret, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	existing, err := s.secrets.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if version > 0 && existing.Version != version {
		return nil, repository.ErrConflict
	}
	now := time.Now()
	return s.secrets.SoftDelete(ctx, userID, id, now, existing.Version+1)
}

// Sync returns secrets changed since the given time.
func (s *SecretService) Sync(ctx context.Context, since time.Time) ([]*model.Secret, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.secrets.ListSince(ctx, userID, since)
}
