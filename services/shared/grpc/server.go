// Package grpc provides shared gRPC client/server infrastructure for Pawradise microservices.
package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ServerConfig holds configuration for a gRPC server.
type ServerConfig struct {
	Port    int
	Service string
}

// StartGRPCServer starts a gRPC server with the given config and registration function.
func StartGRPCServer(cfg ServerConfig, register func(*grpc.Server)) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", cfg.Port, err)
	}

	server := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute,
			MaxConnectionAge:      30 * time.Minute,
			MaxConnectionAgeGrace: 5 * time.Minute,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	register(server)

	log.Printf("[gRPC] %s starting on :%d", cfg.Service, cfg.Port)
	if err := server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}
	return nil
}

// ClientConfig holds configuration for a gRPC client connection.
type ClientConfig struct {
	Target string
}

// NewClientConn creates a new gRPC client connection.
func NewClientConn(cfg ClientConfig) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.Target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", cfg.Target, err)
	}

	return conn, nil
}
