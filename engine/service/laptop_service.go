package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/iruldev/grpc-class/engine/repository"
	"gitlab.com/iruldev/grpc-class/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type LaptopService struct {
	proto.UnimplementedLaptopServiceServer
	LaptopRepository repository.LaptopRepository
}

func NewLaptopService(laptopRepository repository.LaptopRepository) *LaptopService {
	return &LaptopService{LaptopRepository: laptopRepository}
}

func (s *LaptopService) CreateLaptop(ctx context.Context, req *proto.CreateLaptopRequest) (*proto.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("receive a create-laptop request with id: %s", laptop.Id)

	if len(laptop.Id) > 0 {
		//	Check if it's a valid UUID
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop ID is not a valid UUID: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot generate a new laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	//time.Sleep(6 * time.Second)

	if ctx.Err() == context.Canceled {
		log.Print("request is canceled")
		return nil, status.Error(codes.Canceled, "request is canceled")
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Print("deadline is exceeded")
		return nil, status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	}

	// save the laptop to store
	err := s.LaptopRepository.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, repository.ErrAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "cannot save laptop to the db: %v", err)
	}

	log.Printf("saved laptop with id: %s", laptop.Id)
	res := &proto.CreateLaptopResponse{Id: laptop.Id}
	return res, nil
}

func (s *LaptopService) SearchLaptop(req *proto.SearchLaptopRequest, stream proto.LaptopService_SearchLaptopServer) error {
	filter := req.GetFilter()
	log.Printf("receive a search-laptop request with filter: %v", filter)

	err := s.LaptopRepository.Search(stream.Context(), filter, func(laptop *proto.Laptop) error {
		res := &proto.SearchLaptopResponse{Laptop: laptop}
		err := stream.Send(res)
		if err != nil {
			return err
		}

		log.Printf("sent laptop with id: %s", laptop.GetId())
		return nil
	})

	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}
