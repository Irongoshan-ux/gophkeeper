package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/Irongoshan-ux/gophkeeper/internal/model"
)

// MemoryUserRepository is an in-memory UserRepository for tests.
type MemoryUserRepository struct {
	mu      sync.RWMutex
	users   map[string]*model.User
	byLogin map[string]string
}

// MemorySecretRepository is an in-memory SecretRepository for tests.
type MemorySecretRepository struct {
	mu      sync.RWMutex
	secrets map[string]*model.Secret
}

// NewMemoryRepositories creates empty in-memory stores.
func NewMemoryRepositories() (*MemoryUserRepository, *MemorySecretRepository) {
	return &MemoryUserRepository{
			users:   make(map[string]*model.User),
			byLogin: make(map[string]string),
		}, &MemorySecretRepository{
			secrets: make(map[string]*model.Secret),
		}
}

// Create stores a new user.
func (r *MemoryUserRepository) Create(_ context.Context, login, passwordHash string) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byLogin[login]; ok {
		return nil, ErrConflict
	}
	u := &model.User{
		ID:           uuid.NewString(),
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	r.users[u.ID] = u
	r.byLogin[login] = u.ID
	return u, nil
}

// GetByLogin returns a user by login.
func (r *MemoryUserRepository) GetByLogin(_ context.Context, login string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byLogin[login]
	if !ok {
		return nil, ErrNotFound
	}
	return r.users[id], nil
}

// GetByID returns a user by id.
func (r *MemoryUserRepository) GetByID(_ context.Context, id string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (r *MemorySecretRepository) Create(_ context.Context, secret *model.Secret) (*model.Secret, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if secret.ID == "" {
		secret.ID = uuid.NewString()
	}
	now := time.Now()
	secret.CreatedAt = now
	secret.UpdatedAt = now
	cp := *secret
	r.secrets[secret.ID] = &cp
	return &cp, nil
}

// Get returns a secret by id.
func (r *MemorySecretRepository) Get(_ context.Context, userID, id string) (*model.Secret, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.secrets[id]
	if !ok || s.UserID != userID || s.IsDeleted() {
		return nil, ErrNotFound
	}
	cp := *s
	return &cp, nil
}

// List returns active secrets for a user.
func (r *MemorySecretRepository) List(_ context.Context, userID string) ([]*model.Secret, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*model.Secret
	for _, s := range r.secrets {
		if s.UserID == userID && !s.IsDeleted() {
			cp := *s
			out = append(out, &cp)
		}
	}
	return out, nil
}

// Update replaces a secret.
func (r *MemorySecretRepository) Update(_ context.Context, secret *model.Secret) (*model.Secret, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.secrets[secret.ID]
	if !ok || s.UserID != secret.UserID || s.IsDeleted() {
		return nil, ErrNotFound
	}
	if secret.Version > 0 && s.Version != secret.Version {
		return nil, ErrConflict
	}
	updated := *secret
	updated.Version = s.Version + 1
	updated.UpdatedAt = time.Now()
	r.secrets[secret.ID] = &updated
	cp := updated
	return &cp, nil
}

// SoftDelete marks a secret deleted.
func (r *MemorySecretRepository) SoftDelete(_ context.Context, userID, id string, deletedAt time.Time, expectedVersion int64) (*model.Secret, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.secrets[id]
	if !ok || s.UserID != userID || s.IsDeleted() {
		return nil, ErrNotFound
	}
	if expectedVersion > 0 && s.Version != expectedVersion {
		return nil, ErrConflict
	}
	s.DeletedAt = &deletedAt
	s.UpdatedAt = deletedAt
	s.Version++
	cp := *s
	return &cp, nil
}

// ListSince returns secrets updated after since.
func (r *MemorySecretRepository) ListSince(_ context.Context, userID string, since time.Time) ([]*model.Secret, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*model.Secret
	for _, s := range r.secrets {
		if s.UserID == userID && s.UpdatedAt.After(since) {
			cp := *s
			out = append(out, &cp)
		}
	}
	return out, nil
}
