// Package auth provides JWT helpers and request context utilities.
package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

var (
	// ErrInvalidToken means the JWT is missing, expired, or malformed.
	ErrInvalidToken = errors.New("invalid token")
	// ErrNoUserIDInContext means the authenticated user id is absent from context.
	ErrNoUserIDInContext = errors.New("user id not in context")
)

const tokenExpiry = 24 * time.Hour

type contextKey string

const contextKeyUserID contextKey = "user_id"

type claims struct {
	jwt.RegisteredClaims
	UserID string `json:"uid"`
}

// NewToken signs a JWT for the given user id.
func NewToken(secret, userID string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiry)),
		},
		UserID: userID,
	})
	return t.SignedString(secretBytes(secret))
}

// ParseToken validates a JWT and returns the user id.
func ParseToken(secret, tokenString string) (string, error) {
	t, err := jwt.ParseWithClaims(tokenString, &claims{}, func(_ *jwt.Token) (any, error) {
		return secretBytes(secret), nil
	})
	if err != nil {
		return "", ErrInvalidToken
	}
	c, ok := t.Claims.(*claims)
	if !ok || !t.Valid || c.UserID == "" {
		return "", ErrInvalidToken
	}
	return c.UserID, nil
}

// UserIDFromMetadata extracts and validates Bearer token from gRPC metadata.
func UserIDFromMetadata(ctx context.Context, secret string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrInvalidToken
	}
	vals := md.Get("authorization")
	if len(vals) == 0 || strings.TrimSpace(vals[0]) == "" {
		return "", ErrInvalidToken
	}
	raw := strings.TrimSpace(vals[0])
	if len(raw) > 7 && strings.EqualFold(raw[:7], "bearer ") {
		raw = strings.TrimSpace(raw[7:])
	}
	return ParseToken(secret, raw)
}

// UserIDFromContext returns the authenticated user id set by the auth interceptor.
func UserIDFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(contextKeyUserID)
	if v == nil {
		return "", ErrNoUserIDInContext
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", ErrNoUserIDInContext
	}
	return s, nil
}

// WithUserID attaches userID to context for tests and interceptors.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

// OutgoingContext returns a context with authorization metadata for gRPC clients.
func OutgoingContext(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

func secretBytes(s string) []byte {
	if s == "" {
		return []byte("default-secret")
	}
	return []byte(s)
}
