// Package grpcserver implements gRPC service handlers.
package grpcserver

import (
	"context"
	"errors"

	gophkeeperv1 "github.com/Irongoshan-ux/gophkeeper/api/gophkeeper/v1"
	"github.com/Irongoshan-ux/gophkeeper/internal/converter"
	"github.com/Irongoshan-ux/gophkeeper/internal/model"
	"github.com/Irongoshan-ux/gophkeeper/internal/repository"
	"github.com/Irongoshan-ux/gophkeeper/internal/service"
	"github.com/Irongoshan-ux/gophkeeper/internal/validation"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements AuthService and SecretService gRPC APIs.
type Server struct {
	gophkeeperv1.UnimplementedAuthServiceServer
	gophkeeperv1.UnimplementedSecretServiceServer
	auth    *service.AuthService
	secrets *service.SecretService
	log     zerolog.Logger
}

// NewServer creates a gRPC server handler.
func NewServer(authSvc *service.AuthService, secretSvc *service.SecretService, log zerolog.Logger) *Server {
	return &Server{auth: authSvc, secrets: secretSvc, log: log}
}

// Register creates a new user account.
func (s *Server) Register(ctx context.Context, req *gophkeeperv1.RegisterRequest) (*gophkeeperv1.AuthResponse, error) {
	token, userID, err := s.auth.Register(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, mapAuthError(err)
	}
	return &gophkeeperv1.AuthResponse{Token: token, UserId: userID}, nil
}

// Login authenticates an existing user.
func (s *Server) Login(ctx context.Context, req *gophkeeperv1.LoginRequest) (*gophkeeperv1.AuthResponse, error) {
	token, userID, err := s.auth.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, mapAuthError(err)
	}
	return &gophkeeperv1.AuthResponse{Token: token, UserId: userID}, nil
}

// CreateSecret stores a new encrypted secret.
func (s *Server) CreateSecret(ctx context.Context, req *gophkeeperv1.CreateSecretRequest) (*gophkeeperv1.Secret, error) {
	secret, err := s.secrets.Create(ctx, converter.ProtoToSecretType(req.GetType()), req.GetName(), req.GetEncryptedData())
	if err != nil {
		return nil, mapSecretError(err)
	}
	return secretToProto(secret), nil
}

// GetSecret returns a secret by id.
func (s *Server) GetSecret(ctx context.Context, req *gophkeeperv1.GetSecretRequest) (*gophkeeperv1.Secret, error) {
	secret, err := s.secrets.Get(ctx, req.GetId())
	if err != nil {
		return nil, mapSecretError(err)
	}
	return secretToProto(secret), nil
}

// ListSecrets returns all active secrets for the user.
func (s *Server) ListSecrets(ctx context.Context, _ *gophkeeperv1.ListSecretsRequest) (*gophkeeperv1.ListSecretsResponse, error) {
	secrets, err := s.secrets.List(ctx)
	if err != nil {
		return nil, mapSecretError(err)
	}
	out := make([]*gophkeeperv1.Secret, 0, len(secrets))
	for _, sec := range secrets {
		out = append(out, secretToProto(sec))
	}
	return &gophkeeperv1.ListSecretsResponse{Secrets: out}, nil
}

// UpdateSecret replaces secret data.
func (s *Server) UpdateSecret(ctx context.Context, req *gophkeeperv1.UpdateSecretRequest) (*gophkeeperv1.Secret, error) {
	secret, err := s.secrets.Update(ctx, req.GetId(), converter.ProtoToSecretType(req.GetType()), req.GetName(), req.GetEncryptedData(), req.GetVersion())
	if err != nil {
		return nil, mapSecretError(err)
	}
	return secretToProto(secret), nil
}

// DeleteSecret soft-deletes a secret.
func (s *Server) DeleteSecret(ctx context.Context, req *gophkeeperv1.DeleteSecretRequest) (*gophkeeperv1.DeleteSecretResponse, error) {
	secret, err := s.secrets.Delete(ctx, req.GetId(), req.GetVersion())
	if err != nil {
		return nil, mapSecretError(err)
	}
	return &gophkeeperv1.DeleteSecretResponse{Id: secret.ID, Version: secret.Version}, nil
}

// Sync returns secrets changed since the client watermark.
func (s *Server) Sync(ctx context.Context, req *gophkeeperv1.SyncRequest) (*gophkeeperv1.SyncResponse, error) {
	since := req.GetSince().AsTime()
	secrets, err := s.secrets.Sync(ctx, since)
	if err != nil {
		return nil, mapSecretError(err)
	}
	out := make([]*gophkeeperv1.Secret, 0, len(secrets))
	for _, sec := range secrets {
		out = append(out, secretToProto(sec))
	}
	return &gophkeeperv1.SyncResponse{Secrets: out}, nil
}

func secretToProto(s *model.Secret) *gophkeeperv1.Secret {
	return &gophkeeperv1.Secret{
		Id:            s.ID,
		Type:          converter.SecretTypeToProto(s.Type),
		Name:          s.Name,
		EncryptedData: s.EncryptedData,
		Version:       s.Version,
		UpdatedAt:     timestamppb.New(s.UpdatedAt),
		Deleted:       s.IsDeleted(),
	}
}

func mapAuthError(err error) error {
	switch {
	case errors.Is(err, repository.ErrConflict):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, repository.ErrNotFound):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, validation.ErrEmptyLogin), errors.Is(err, validation.ErrEmptyPassword):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func mapSecretError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return status.Error(codes.NotFound, "secret not found")
	case errors.Is(err, repository.ErrConflict):
		return status.Error(codes.FailedPrecondition, "version conflict")
	case errors.Is(err, validation.ErrEmptyName):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
