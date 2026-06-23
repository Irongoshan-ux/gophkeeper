package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgerrcode"

	"github.com/Irongoshan-ux/gophkeeper/internal/model"
	"github.com/Irongoshan-ux/gophkeeper/internal/repository/sqlc/db"
)

const pgUniqueUsersLogin = "users_login_key"

// PostgresUserRepository implements UserRepository.
type PostgresUserRepository struct {
	q *db.Queries
}

// PostgresSecretRepository implements SecretRepository.
type PostgresSecretRepository struct {
	q *db.Queries
}

// NewPostgresRepositories creates PostgreSQL-backed user and secret repositories.
func NewPostgresRepositories(pool *pgxpool.Pool) (*PostgresUserRepository, *PostgresSecretRepository) {
	q := db.New(pool)
	return &PostgresUserRepository{q: q}, &PostgresSecretRepository{q: q}
}

func uuidFromString(s string) (pgtype.UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	var p pgtype.UUID
	copy(p.Bytes[:], u[:])
	p.Valid = true
	return p, nil
}

func uuidToString(p pgtype.UUID) string {
	if !p.Valid {
		return ""
	}
	return uuid.UUID(p.Bytes).String()
}

func secretFromRow(row db.Secret) *model.Secret {
	s := &model.Secret{
		ID:            uuidToString(row.ID),
		UserID:        uuidToString(row.UserID),
		Type:          model.SecretType(row.Type),
		Name:          row.Name,
		EncryptedData: row.EncryptedData,
		Version:       row.Version,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		s.DeletedAt = &t
	}
	return s
}

// Create stores a new user account.
func (r *PostgresUserRepository) Create(ctx context.Context, login, passwordHash string) (*model.User, error) {
	u, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation && e.ConstraintName == pgUniqueUsersLogin {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &model.User{
		ID:           uuidToString(u.ID),
		Login:        u.Login,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}, nil
}

// GetByLogin returns a user by login.
func (r *PostgresUserRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	u, err := r.q.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by login: %w", err)
	}
	return &model.User{
		ID:           uuidToString(u.ID),
		Login:        u.Login,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}, nil
}

// GetByID returns a user by id.
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	pid, err := uuidFromString(id)
	if err != nil {
		return nil, ErrNotFound
	}
	u, err := r.q.GetUserByID(ctx, pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &model.User{
		ID:           uuidToString(u.ID),
		Login:        u.Login,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}, nil
}

func (r *PostgresSecretRepository) Create(ctx context.Context, secret *model.Secret) (*model.Secret, error) {
	uid, err := uuidFromString(secret.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	now := time.Now()
	row, err := r.q.CreateSecret(ctx, db.CreateSecretParams{
		UserID:        uid,
		Type:          int16(secret.Type),
		Name:          secret.Name,
		EncryptedData: secret.EncryptedData,
		Version:       secret.Version,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		return nil, fmt.Errorf("create secret: %w", err)
	}
	return secretFromRow(row), nil
}

// Get returns a secret by id for the given user.
func (r *PostgresSecretRepository) Get(ctx context.Context, userID, id string) (*model.Secret, error) {
	uid, err := uuidFromString(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	sid, err := uuidFromString(id)
	if err != nil {
		return nil, ErrNotFound
	}
	row, err := r.q.GetSecret(ctx, db.GetSecretParams{ID: sid, UserID: uid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get secret: %w", err)
	}
	return secretFromRow(row), nil
}

// List returns active secrets for a user.
func (r *PostgresSecretRepository) List(ctx context.Context, userID string) ([]*model.Secret, error) {
	uid, err := uuidFromString(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	rows, err := r.q.ListSecrets(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	out := make([]*model.Secret, 0, len(rows))
	for _, row := range rows {
		out = append(out, secretFromRow(row))
	}
	return out, nil
}

// Update replaces secret fields and bumps version.
func (r *PostgresSecretRepository) Update(ctx context.Context, secret *model.Secret) (*model.Secret, error) {
	uid, err := uuidFromString(secret.UserID)
	if err != nil {
		return nil, ErrNotFound
	}
	sid, err := uuidFromString(secret.ID)
	if err != nil {
		return nil, ErrNotFound
	}
	now := time.Now()
	row, err := r.q.UpdateSecret(ctx, db.UpdateSecretParams{
		ID:              sid,
		UserID:          uid,
		Name:            secret.Name,
		EncryptedData:   secret.EncryptedData,
		Type:            int16(secret.Type),
		UpdatedAt:       now,
		ExpectedVersion: secret.Version,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, r.secretUpdateConflictOrNotFound(ctx, secret.UserID, secret.ID)
		}
		return nil, fmt.Errorf("update secret: %w", err)
	}
	return secretFromRow(row), nil
}

// SoftDelete marks a secret as deleted.
func (r *PostgresSecretRepository) SoftDelete(ctx context.Context, userID, id string, deletedAt time.Time, expectedVersion int64) (*model.Secret, error) {
	uid, err := uuidFromString(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	sid, err := uuidFromString(id)
	if err != nil {
		return nil, ErrNotFound
	}
	row, err := r.q.SoftDeleteSecret(ctx, db.SoftDeleteSecretParams{
		ID:              sid,
		UserID:          uid,
		DeletedAt:       pgtype.Timestamptz{Time: deletedAt, Valid: true},
		ExpectedVersion: expectedVersion,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, r.secretUpdateConflictOrNotFound(ctx, userID, id)
		}
		return nil, fmt.Errorf("delete secret: %w", err)
	}
	return secretFromRow(row), nil
}

func (r *PostgresSecretRepository) secretUpdateConflictOrNotFound(ctx context.Context, userID, id string) error {
	if _, err := r.Get(ctx, userID, id); err != nil {
		return ErrNotFound
	}
	return ErrConflict
}

// ListSince returns secrets changed after the given timestamp.
func (r *PostgresSecretRepository) ListSince(ctx context.Context, userID string, since time.Time) ([]*model.Secret, error) {
	uid, err := uuidFromString(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	rows, err := r.q.ListSecretsSince(ctx, db.ListSecretsSinceParams{
		UserID: uid,
		UpdatedAt: since,
	})
	if err != nil {
		return nil, fmt.Errorf("list secrets since: %w", err)
	}
	out := make([]*model.Secret, 0, len(rows))
	for _, row := range rows {
		out = append(out, secretFromRow(row))
	}
	return out, nil
}
