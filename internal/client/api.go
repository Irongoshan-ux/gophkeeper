// Package client provides a gRPC client for the GophKeeper CLI.
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	gophkeeperv1 "github.com/Irongoshan-ux/gophkeeper/api/gophkeeper/v1"
	"github.com/Irongoshan-ux/gophkeeper/internal/auth"
	"github.com/Irongoshan-ux/gophkeeper/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// API wraps gRPC calls to AuthService and SecretService.
type API struct {
	conn   *grpc.ClientConn
	auth   gophkeeperv1.AuthServiceClient
	secret gophkeeperv1.SecretServiceClient
	token  string
}

// Options configures the gRPC client connection.
type Options struct {
	Address  string
	Insecure bool
	Token    string
	Dialer   func(context.Context, string) (net.Conn, error) // optional, for tests
}

// Connect dials the GophKeeper server.
func Connect(opts Options) (*API, error) {
	var creds credentials.TransportCredentials
	if opts.Insecure {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: true}) //nolint:gosec // dev self-signed
	}
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	if opts.Dialer != nil {
		dialOpts = append(dialOpts, grpc.WithContextDialer(opts.Dialer))
	}
	conn, err := grpc.NewClient(opts.Address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial server: %w", err)
	}
	return &API{
		conn:   conn,
		auth:   gophkeeperv1.NewAuthServiceClient(conn),
		secret: gophkeeperv1.NewSecretServiceClient(conn),
		token:  opts.Token,
	}, nil
}

// Close closes the underlying connection.
func (a *API) Close() error {
	return a.conn.Close()
}

// SetToken updates the JWT used for authenticated calls.
func (a *API) SetToken(token string) {
	a.token = token
}

// Register creates a new account and stores the returned token.
func (a *API) Register(ctx context.Context, login, password string) (string, string, error) {
	resp, err := a.auth.Register(ctx, &gophkeeperv1.RegisterRequest{Login: login, Password: password})
	if err != nil {
		return "", "", err
	}
	a.token = resp.GetToken()
	return resp.GetToken(), resp.GetUserId(), nil
}

// Login authenticates and stores the returned token.
func (a *API) Login(ctx context.Context, login, password string) (string, string, error) {
	resp, err := a.auth.Login(ctx, &gophkeeperv1.LoginRequest{Login: login, Password: password})
	if err != nil {
		return "", "", err
	}
	a.token = resp.GetToken()
	return resp.GetToken(), resp.GetUserId(), nil
}

// CreateSecret uploads an encrypted secret.
func (a *API) CreateSecret(ctx context.Context, secretType model.SecretType, name string, encrypted []byte) (*model.Secret, error) {
	resp, err := a.secret.CreateSecret(a.authCtx(ctx), &gophkeeperv1.CreateSecretRequest{
		Type:          modelTypeToProto(secretType),
		Name:          name,
		EncryptedData: encrypted,
	})
	if err != nil {
		return nil, err
	}
	return protoToSecret(resp), nil
}

// GetSecret downloads a secret by id.
func (a *API) GetSecret(ctx context.Context, id string) (*model.Secret, error) {
	resp, err := a.secret.GetSecret(a.authCtx(ctx), &gophkeeperv1.GetSecretRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return protoToSecret(resp), nil
}

// ListSecrets returns all active secrets.
func (a *API) ListSecrets(ctx context.Context) ([]*model.Secret, error) {
	resp, err := a.secret.ListSecrets(a.authCtx(ctx), &gophkeeperv1.ListSecretsRequest{})
	if err != nil {
		return nil, err
	}
	out := make([]*model.Secret, 0, len(resp.GetSecrets()))
	for _, s := range resp.GetSecrets() {
		out = append(out, protoToSecret(s))
	}
	return out, nil
}

// UpdateSecret replaces secret data.
func (a *API) UpdateSecret(ctx context.Context, id string, secretType model.SecretType, name string, encrypted []byte, version int64) (*model.Secret, error) {
	resp, err := a.secret.UpdateSecret(a.authCtx(ctx), &gophkeeperv1.UpdateSecretRequest{
		Id:            id,
		Type:          modelTypeToProto(secretType),
		Name:          name,
		EncryptedData: encrypted,
		Version:       version,
	})
	if err != nil {
		return nil, err
	}
	return protoToSecret(resp), nil
}

// DeleteSecret soft-deletes a secret.
func (a *API) DeleteSecret(ctx context.Context, id string, version int64) error {
	_, err := a.secret.DeleteSecret(a.authCtx(ctx), &gophkeeperv1.DeleteSecretRequest{Id: id, Version: version})
	return err
}

// Sync returns secrets changed since the given time.
func (a *API) Sync(ctx context.Context, since time.Time) ([]*model.Secret, error) {
	resp, err := a.secret.Sync(a.authCtx(ctx), &gophkeeperv1.SyncRequest{
		Since: timestamppb.New(since),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*model.Secret, 0, len(resp.GetSecrets()))
	for _, s := range resp.GetSecrets() {
		out = append(out, protoToSecret(s))
	}
	return out, nil
}

func (a *API) authCtx(ctx context.Context) context.Context {
	if a.token == "" {
		return ctx
	}
	return auth.OutgoingContext(ctx, a.token)
}

func modelTypeToProto(t model.SecretType) gophkeeperv1.SecretType {
	switch t {
	case model.SecretTypeCredentials:
		return gophkeeperv1.SecretType_SECRET_TYPE_CREDENTIALS
	case model.SecretTypeText:
		return gophkeeperv1.SecretType_SECRET_TYPE_TEXT
	case model.SecretTypeBinary:
		return gophkeeperv1.SecretType_SECRET_TYPE_BINARY
	case model.SecretTypeCard:
		return gophkeeperv1.SecretType_SECRET_TYPE_CARD
	case model.SecretTypeOTP:
		return gophkeeperv1.SecretType_SECRET_TYPE_OTP
	default:
		return gophkeeperv1.SecretType_SECRET_TYPE_UNSPECIFIED
	}
}

func protoToSecret(s *gophkeeperv1.Secret) *model.Secret {
	sec := &model.Secret{
		ID:            s.GetId(),
		Type:          protoTypeToModel(s.GetType()),
		Name:          s.GetName(),
		EncryptedData: s.GetEncryptedData(),
		Version:       s.GetVersion(),
	}
	if s.GetUpdatedAt() != nil {
		sec.UpdatedAt = s.GetUpdatedAt().AsTime()
	}
	if s.GetDeleted() {
		t := sec.UpdatedAt
		sec.DeletedAt = &t
	}
	return sec
}

func protoTypeToModel(t gophkeeperv1.SecretType) model.SecretType {
	switch t {
	case gophkeeperv1.SecretType_SECRET_TYPE_CREDENTIALS:
		return model.SecretTypeCredentials
	case gophkeeperv1.SecretType_SECRET_TYPE_TEXT:
		return model.SecretTypeText
	case gophkeeperv1.SecretType_SECRET_TYPE_BINARY:
		return model.SecretTypeBinary
	case gophkeeperv1.SecretType_SECRET_TYPE_CARD:
		return model.SecretTypeCard
	case gophkeeperv1.SecretType_SECRET_TYPE_OTP:
		return model.SecretTypeOTP
	default:
		return model.SecretTypeUnspecified
	}
}
