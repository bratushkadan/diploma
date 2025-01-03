package auth

import (
	"context"
	"fmt"
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
	return &auth_pb.GetUserResponse{
		Id:   req.GetId(),
		Name: "bratushkadan",
	}, nil
}

func (s *serverImpl) GetUsers(ctx context.Context, req *auth_pb.GetUsersRequest) (*auth_pb.GetUsersResponse, error) {
	return &auth_pb.GetUsersResponse{
		Users: []*auth_pb.GetUsersResponse_User{
			{
				Id:   "bratushkadan",
				Name: "Danila Bratushka",
			},
			{
				Id:   "andronidze",
				Name: "Andrey Skochok",
			},
		},
	}, nil
}

func RunServer(conf *AuthServerConfig) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
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
