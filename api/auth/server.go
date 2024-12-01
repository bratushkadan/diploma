package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	auth_pb "github.com/bratushkadan/floral/pb/floral/auth/v1"
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
	auth_pb.UnimplementedUserServiceServer
}

func (s *serverImpl) GetUser(ctx context.Context, req *auth_pb.GetUserRequest) (*auth_pb.GetUserResponse, error) {
	if rand.IntN(10) == 0 {
		return nil, errors.New("service unavailable")
	}
	return &auth_pb.GetUserResponse{
		Id:   req.GetId(),
		Name: "bratushkadan",
	}, nil
}

func RunServer(conf *AuthServerConfig) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", conf.Port))
	if err != nil {
		return fmt.Errorf("failed to serve tcp: %v", err)
	}

	s := grpc.NewServer()

	reflection.Register(s)

	auth_pb.RegisterUserServiceServer(s, &serverImpl{})
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %v", err)
	}

	return nil
}
