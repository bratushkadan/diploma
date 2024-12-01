package auth

import (
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/bratushkadan/floral/pb/floral/v1"
)

type AuthServerConfig struct {
	Port int
}

func NewAuthServerConfig(port int) *AuthServerConfig {
	return &AuthServerConfig{
		Port: port,
	}
}

type serverImpl struct {
	floral.UnimplementedUserServiceServer
}

func RunServer(conf *AuthServerConfig) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", conf.Port))
	if err != nil {
		return fmt.Errorf("failed to serve tcp: %v", err)
	}

	s := grpc.NewServer()

	floral.RegisterUserServiceServer(s, &serverImpl{})
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %v", err)
	}

	return nil
}
