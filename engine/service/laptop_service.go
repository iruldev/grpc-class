package service

import (
	"bytes"
	"context"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/iruldev/grpc-class/engine/repository"
	"gitlab.com/iruldev/grpc-class/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
)

const maxImageSize = 1 << 20 // 1 megabyte

type LaptopService struct {
	proto.UnimplementedLaptopServiceServer
	LaptopRepository repository.LaptopRepository
	ImageRepository  repository.ImageRepository
	RatingRepository repository.RatingRepository
}

func NewLaptopService(
	laptopRepository repository.LaptopRepository,
	imageRepository repository.ImageRepository,
	ratingRepository repository.RatingRepository,
) *LaptopService {
	return &LaptopService{
		LaptopRepository: laptopRepository,
		ImageRepository:  imageRepository,
		RatingRepository: ratingRepository,
	}
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

func (s *LaptopService) UploadImage(stream proto.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive image info: %v", err))
	}

	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("receive an upload-image request for laptop %s with image type %s", laptopID, imageType)

	laptop, err := s.LaptopRepository.Find(laptopID)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot find laptop: %v", err))
	}
	if laptop == nil {
		return logError(status.Errorf(codes.InvalidArgument, "laptop %s doesn't exist", laptopID))
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		if err := contextError(stream.Context()); err != nil {
			return err
		}
		log.Print("waiting to receive more data ")

		req, err = stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chuck := req.GetChunkData()
		size := len(chuck)

		log.Printf("received a chunk with size: %d", size)

		imageSize += size

		if imageSize > maxImageSize {
			return logError(status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize))
		}

		_, err = imageData.Write(chuck)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chuck data: %v", err))
		}
	}

	imageID, err := s.ImageRepository.Save(laptopID, imageType, imageData)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot save image to db: %v", err))
	}

	res := &proto.UploadImageRespons{
		Id:   imageID,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}

	log.Printf("send image with id: %s, size: %d", imageID, imageSize)

	return nil
}

func (s *LaptopService) RateLaptop(stream proto.LaptopService_RateLaptopServer) error {
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive stream request: %v", err))
		}
		laptopID := req.GetLaptopId()
		score := req.GetScore()

		log.Printf("received a rate-laptop request: id = %s, score = %.2f", laptopID, score)

		found, err := s.LaptopRepository.Find(laptopID)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot find laptop: %v", err))
		}
		if found == nil {
			return logError(status.Errorf(codes.NotFound, "laptop id %s is not found", laptopID))
		}

		rating, err := s.RatingRepository.Add(laptopID, score)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot add rating to the store: %v", err))
		}

		res := &proto.RateLaptopResponse{
			LaptopId:     laptopID,
			RatedCount:   rating.Count,
			AverageScore: rating.Sum / float64(rating.Count),
		}

		err = stream.Send(res)
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot send stream response: %v", err))
		}
	}
	return nil
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is canceled"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}
