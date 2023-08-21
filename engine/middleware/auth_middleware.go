package middleware

import (
	"context"
	"gitlab.com/iruldev/grpc-class/engine/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
)

type AuthMiddleware struct {
	tokenMaker      *service.JWT
	accessibleRoles map[string][]string
}

func NewAuthMiddleware(tokenMaker *service.JWT, accessibleRoles map[string][]string) *AuthMiddleware {
	return &AuthMiddleware{
		tokenMaker:      tokenMaker,
		accessibleRoles: accessibleRoles,
	}
}

func (m *AuthMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Println("--> unary interceptor: ", info.FullMethod)

		err := m.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (m *AuthMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		log.Println("--> stream interceptor: ", info.FullMethod)
		err := m.authorize(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func (m *AuthMiddleware) authorize(ctx context.Context, method string) error {
	accessibleRoles, ok := m.accessibleRoles[method]
	if !ok {
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	value := md["authorization"]
	if len(value) == 0 {
		return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	accessToken := value[0]
	claims, err := m.tokenMaker.Verify(accessToken)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	for _, role := range accessibleRoles {
		if role == claims.Role {
			return nil
		}
	}

	return status.Errorf(codes.PermissionDenied, "no permission to access this RPC")
}
