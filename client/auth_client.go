package client

import (
	"context"
	"gitlab.com/iruldev/grpc-class/proto"
	"google.golang.org/grpc"
	"time"
)

type AuthClient struct {
	service  proto.AuthServiceClient
	username string
	password string
}

func NewAuthClient(cc *grpc.ClientConn, username, password string) *AuthClient {
	service := proto.NewAuthServiceClient(cc)
	return &AuthClient{
		service:  service,
		username: username,
		password: password,
	}
}

func (c *AuthClient) Login() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &proto.LoginRequest{
		Username: c.username,
		Password: c.password,
	}

	res, err := c.service.Login(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetAccessToken(), nil
}
