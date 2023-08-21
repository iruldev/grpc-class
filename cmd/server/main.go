package main

import (
	"flag"
	"fmt"
	"gitlab.com/iruldev/grpc-class/engine/middleware"
	"gitlab.com/iruldev/grpc-class/engine/model/entity"
	"gitlab.com/iruldev/grpc-class/engine/repository"
	"gitlab.com/iruldev/grpc-class/engine/service"
	"gitlab.com/iruldev/grpc-class/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"
)

func seedUsers(userRepo repository.UserRepository) error {
	err := createUser(userRepo, "admin1", "secret", "admin")
	if err != nil {
		return err
	}
	return createUser(userRepo, "user1", "secret", "user")
}

func createUser(userRepo repository.UserRepository, username, password, role string) error {
	user, err := entity.NewUser(username, password, role)
	if err != nil {
		return err
	}
	return userRepo.Save(user)
}

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
)

func accessibleRoles() map[string][]string {
	const laptopServicePath = "/grpc.class.LaptopService/"
	return map[string][]string{
		laptopServicePath + "CreateLaptop": {"admin"},
		laptopServicePath + "UploadImage":  {"admin"},
		laptopServicePath + "RateLaptop":   {"admin", "user"},
	}
}

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("start server on port %d", *port)

	userRepo := repository.NewUserRepository()
	err := seedUsers(userRepo)
	if err != nil {
		log.Fatal("cannot seed users")
	}

	tokenMaker := service.NewJWTService(secretKey, tokenDuration)
	authServer := service.NewAuthService(userRepo, tokenMaker)

	laptopRepo := repository.NewLaptopRepository()
	imageRepo := repository.NewImageRepository("img")
	ratingRepo := repository.NewRatingRepository()
	laptopServer := service.NewLaptopService(laptopRepo, imageRepo, ratingRepo)

	interceptor := middleware.NewAuthMiddleware(tokenMaker, accessibleRoles())
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	)

	proto.RegisterAuthServiceServer(grpcServer, authServer)
	proto.RegisterLaptopServiceServer(grpcServer, laptopServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}

	reflection.Register(grpcServer)

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
