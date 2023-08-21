package service

import (
	"context"
	"gitlab.com/iruldev/grpc-class/engine/repository"
	"gitlab.com/iruldev/grpc-class/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	proto.UnimplementedAuthServiceServer
	UserRepository repository.UserRepository
	TokenMaker     *JWT
}

func NewAuthService(userRepository repository.UserRepository, tokenMaker *JWT) *AuthService {
	return &AuthService{
		UserRepository: userRepository,
		TokenMaker:     tokenMaker,
	}
}

func (s *AuthService) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	user, err := s.UserRepository.Find(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot find user: %v", err)
	}

	if user == nil || !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.NotFound, "incorrect username/password")
	}

	token, err := s.TokenMaker.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &proto.LoginResponse{AccessToken: token}
	return res, nil
}
