// Package app wires and runs the gRPC server.
package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	gophkeeperv1 "github.com/Irongoshan-ux/gophkeeper/api/gophkeeper/v1"
	"github.com/Irongoshan-ux/gophkeeper/internal/config"
	"github.com/Irongoshan-ux/gophkeeper/internal/grpcserver"
	"github.com/Irongoshan-ux/gophkeeper/internal/service"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const shutdownTimeout = 10 * time.Second

// NewGRPCServer builds and registers gRPC services.
func NewGRPCServer(cfg *config.ServerConfig, authSvc *service.AuthService, secretSvc *service.SecretService, log zerolog.Logger) (*grpc.Server, error) {
	var opts []grpc.ServerOption
	if cfg.EnableTLS {
		cert, err := generateSelfSignedCert()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(&cert)))
	}
	opts = append(opts, grpc.UnaryInterceptor(grpcserver.AuthUnaryInterceptor(cfg.JWTSecret)))

	srv := grpc.NewServer(opts...)
	handler := grpcserver.NewServer(authSvc, secretSvc, log)
	gophkeeperv1.RegisterAuthServiceServer(srv, handler)
	gophkeeperv1.RegisterSecretServiceServer(srv, handler)
	return srv, nil
}

// RunGRPC starts the server and stops gracefully when ctx is cancelled.
func RunGRPC(ctx context.Context, srv *grpc.Server, addr string) error {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return fmt.Errorf("grpc server address is required")
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return fmt.Errorf("invalid grpc server address %q: %w", addr, err)
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		stopped := make(chan struct{})
		go func() {
			srv.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			return nil
		case <-time.After(shutdownTimeout):
			srv.Stop()
			return nil
		}
	}
}

func generateSelfSignedCert() (tls.Certificate, error) {
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"gophkeeper"},
		},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost"},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	var certPEM bytes.Buffer
	if err := pem.Encode(&certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return tls.Certificate{}, err
	}

	var keyPEM bytes.Buffer
	if err := pem.Encode(&keyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}); err != nil {
		return tls.Certificate{}, err
	}

	return tls.X509KeyPair(certPEM.Bytes(), keyPEM.Bytes())
}
