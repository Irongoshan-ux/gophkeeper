// Package repository defines storage interfaces and implementations.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Irongoshan-ux/gophkeeper/internal/model"
)

var (
	// ErrNotFound is returned when a requested entity does not exist.
	ErrNotFound = errors.New("not found")
	// ErrConflict is returned on unique constraint violations.
	ErrConflict = errors.New("conflict")
)

// UserRepository persists user accounts.
type UserRepository interface {
	Create(ctx context.Context, login, passwordHash string) (*model.User, error)
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
}

// SecretRepository persists encrypted secrets.
type SecretRepository interface {
	Create(ctx context.Context, secret *model.Secret) (*model.Secret, error)
	Get(ctx context.Context, userID, id string) (*model.Secret, error)
	List(ctx context.Context, userID string) ([]*model.Secret, error)
	Update(ctx context.Context, secret *model.Secret) (*model.Secret, error)
	SoftDelete(ctx context.Context, userID, id string, deletedAt time.Time, version int64) (*model.Secret, error)
	ListSince(ctx context.Context, userID string, since time.Time) ([]*model.Secret, error)
}
