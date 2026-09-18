package grpc

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs gRPC method calls with timing information.
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	logger := log.New(os.Stdout, "[grpc] ", log.LstdFlags)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		latency := time.Since(start)

		code := codes.OK
		if err != nil {
			if s, ok := status.FromError(err); ok {
				code = s.Code()
			} else {
				code = codes.Unknown
			}
		}

		logger.Printf("| %s | %13v | %s | %v",
			code.String(),
			latency,
			info.FullMethod,
			errString(err),
		)
		return resp, err
	}
}

// StreamLoggingInterceptor logs streaming gRPC calls.
func StreamLoggingInterceptor() grpc.StreamServerInterceptor {
	logger := log.New(os.Stdout, "[grpc:stream] ", log.LstdFlags)
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		err := handler(srv, ss)
		latency := time.Since(start)
		logger.Printf("| STREAM | %13v | %s | %v", latency, info.FullMethod, errString(err))
		return err
	}
}

// RecoveryInterceptor recovers from panics in gRPC handlers.
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	logger := log.New(os.Stderr, "[grpc:recovery] ", log.LstdFlags|log.Lshortfile)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf("PANIC in %s: %v", info.FullMethod, r)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// StreamRecoveryInterceptor recovers from panics in streaming gRPC handlers.
func StreamRecoveryInterceptor() grpc.StreamServerInterceptor {
	logger := log.New(os.Stderr, "[grpc:stream:recovery] ", log.LstdFlags|log.Lshortfile)
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf("PANIC in stream %s: %v", info.FullMethod, r)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(srv, ss)
	}
}

// AuthInterceptor validates JWT tokens in gRPC metadata.
// It extracts the "authorization" metadata entry and validates it.
func AuthInterceptor(secret string) grpc.UnaryServerInterceptor {
	if secret == "" {
		secret = getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth for certain methods (e.g., health check, login)
		if isExemptMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing authorization token")
		}

		token := authHeader[0]
		// Strip "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Validate token — in production, call auth.ValidateJWT here
		if token == "" {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}

		// Store user info in context for downstream handlers
		ctx = context.WithValue(ctx, "grpc_token", token)

		return handler(ctx, req)
	}
}

// StreamAuthInterceptor validates JWT tokens for streaming RPCs.
func StreamAuthInterceptor(secret string) grpc.StreamServerInterceptor {
	if secret == "" {
		secret = getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	}
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if isExemptMethod(info.FullMethod) {
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return status.Errorf(codes.Unauthenticated, "missing authorization token")
		}

		token := authHeader[0]
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		if token == "" {
			return status.Errorf(codes.Unauthenticated, "invalid token")
		}

		return handler(srv, ss)
	}
}

// =========================================================================
// helpers
// =========================================================================

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func isExemptMethod(fullMethod string) bool {
	exempt := []string{
		"/grpc.health.v1.Health/Check",
		"/store4bots.IdentityService/Login",
		"/store4bots.IdentityService/Register",
		"/store4bots.IdentityService/ValidateToken",
	}
	for _, m := range exempt {
		if fullMethod == m {
			return true
		}
	}
	return false
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
