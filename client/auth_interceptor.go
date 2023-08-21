package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"time"
)

type AuthInterceptor struct {
	authClient  *AuthClient
	authMethod  map[string]bool
	accessToken string
}

func NewAuthInterceptor(authClient *AuthClient, authMethod map[string]bool, refreshDuration time.Duration) (*AuthInterceptor, error) {
	interceptor := &AuthInterceptor{
		authClient: authClient,
		authMethod: authMethod,
	}

	err := interceptor.scheduleRefreshToken(refreshDuration)
	if err != nil {
		return nil, err
	}

	return interceptor, nil
}

func (c *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		log.Printf("--> unary interceptor: %s", method)
		if c.authMethod[method] {
			return invoker(c.attachToken(ctx), method, req, reply, cc, opts...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (c *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		log.Printf("--> stream interceptor: %s", method)
		if c.authMethod[method] {
			return streamer(c.attachToken(ctx), desc, cc, method, opts...)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func (c *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", c.accessToken)
}

func (c *AuthInterceptor) scheduleRefreshToken(refreshDuration time.Duration) error {
	err := c.refreshToken()
	if err != nil {
		return err
	}

	go func() {
		wait := refreshDuration
		for {
			time.Sleep(wait)
			err := c.refreshToken()
			if err != nil {
				wait = time.Second
			} else {
				wait = refreshDuration
			}
		}
	}()

	return nil
}

func (c *AuthInterceptor) refreshToken() error {
	accessToken, err := c.authClient.Login()
	if err != nil {
		return err
	}
	c.accessToken = accessToken
	log.Printf("token refreshed: %v", accessToken)

	return nil
}
