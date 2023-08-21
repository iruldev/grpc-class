package main

import (
	"flag"
	"fmt"
	"gitlab.com/iruldev/grpc-class/engine/repository"
	"gitlab.com/iruldev/grpc-class/engine/service"
	"gitlab.com/iruldev/grpc-class/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("start server on port %d", *port)

	laptopRepo := repository.NewLaptopRepository()
	imageRepo := repository.NewImageRepository("img")
	ratingRepo := repository.NewRatingRepository()
	laptopServer := service.NewLaptopService(laptopRepo, imageRepo, ratingRepo)
	grpcServer := grpc.NewServer()

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
