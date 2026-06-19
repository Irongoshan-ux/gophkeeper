package grpcserver

import (
	"context"

	"github.com/Irongoshan-ux/gophkeeper/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var publicMethods = map[string]bool{
	"/gophkeeper.v1.AuthService/Register": true,
	"/gophkeeper.v1.AuthService/Login":    true,
}

// AuthUnaryInterceptor validates JWT for protected RPCs.
func AuthUnaryInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}
		userID, err := auth.UserIDFromMetadata(ctx, secret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}
		ctx = auth.WithUserID(ctx, userID)
		return handler(ctx, req)
	}
}
