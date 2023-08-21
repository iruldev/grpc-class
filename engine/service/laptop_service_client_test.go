package service

import (
	"bufio"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"gitlab.com/iruldev/grpc-class/engine/repository"
	"gitlab.com/iruldev/grpc-class/proto"
	"gitlab.com/iruldev/grpc-class/sample"
	"gitlab.com/iruldev/grpc-class/serializer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopRepo := repository.NewLaptopRepository()
	serverAddress := startTestLaptopService(t, laptopRepo, nil, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	laptop := sample.NewLaptop()
	expectedID := laptop.Id
	req := &proto.CreateLaptopRequest{Laptop: laptop}

	res, err := laptopClient.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, expectedID, res.Id)

	// check that the laptop is saved to the store
	other, err := laptopRepo.Find(expectedID)
	require.NoError(t, err)
	require.NotNil(t, other)

	// check that the saved laptop is the same as the one we send
	requireSameLaptop(t, laptop, other)
}

func TestClientSearchLaptop(t *testing.T) {
	t.Parallel()

	filter := &proto.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz:   2.2,
		MinRam: &proto.Memory{
			Value: 8,
			Unit:  proto.Memory_GIGABYTE,
		},
	}

	laptopRepo := repository.NewLaptopRepository()
	expectedIDs := make(map[string]bool)

	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()

		switch i {
		case 0:
			laptop.PriceUsd = 2500
		case 1:
			laptop.Cpu.NumberCores = 2
		case 2:
			laptop.Cpu.MinGhz = 2.0
		case 3:
			laptop.Ram = &proto.Memory{
				Value: 4096,
				Unit:  proto.Memory_MEGABYTE,
			}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberCores = 4
			laptop.Cpu.MinGhz = 2.5
			laptop.Cpu.MaxGhz = 4.5
			laptop.Ram = &proto.Memory{
				Value: 16,
				Unit:  proto.Memory_GIGABYTE,
			}
			expectedIDs[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 2000
			laptop.Cpu.NumberCores = 6
			laptop.Cpu.MinGhz = 2.8
			laptop.Cpu.MaxGhz = 5.0
			laptop.Ram = &proto.Memory{
				Value: 64,
				Unit:  proto.Memory_GIGABYTE,
			}
			expectedIDs[laptop.Id] = true
		}

		err := laptopRepo.Save(laptop)
		require.NoError(t, err)
	}

	serverAddress := startTestLaptopService(t, laptopRepo, nil, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	req := &proto.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	found := 0
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Contains(t, expectedIDs, res.GetLaptop().GetId())

		found += 1
	}

	require.Equal(t, len(expectedIDs), found)
}

func TestClientUploadImage(t *testing.T) {
	t.Parallel()

	testImageFolder := "../../tmp"
	laptopRepo := repository.NewLaptopRepository()
	imageRepo := repository.NewImageRepository(testImageFolder)

	laptop := sample.NewLaptop()
	err := laptopRepo.Save(laptop)
	require.NoError(t, err)

	serverAddress := startTestLaptopService(t, laptopRepo, imageRepo, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	imagePath := fmt.Sprintf("%s/laptop.jpg", testImageFolder)
	file, err := os.Open(imagePath)
	require.NoError(t, err)
	defer file.Close()

	stream, err := laptopClient.UploadImage(context.Background())
	require.NoError(t, err)

	imageType := filepath.Ext(imagePath)
	req := &proto.UploadImageRequest{
		Data: &proto.UploadImageRequest_Info{
			Info: &proto.ImageInfo{
				LaptopId:  laptop.GetId(),
				ImageType: imageType,
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	size := 0

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		size += n

		req := &proto.UploadImageRequest{
			Data: &proto.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		require.NoError(t, err)
	}

	res, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.NotZero(t, res.GetId())
	require.EqualValues(t, size, res.GetSize())

	savedImagePath := fmt.Sprintf("%s/%s%s", testImageFolder, res.GetId(), imageType)
	require.FileExists(t, savedImagePath)
	require.NoError(t, os.Remove(savedImagePath))
}

func TestClientRateLaptop(t *testing.T) {
	t.Parallel()

	laptopRepo := repository.NewLaptopRepository()
	ratingRepo := repository.NewRatingRepository()

	laptop := sample.NewLaptop()
	err := laptopRepo.Save(laptop)
	require.NoError(t, err)

	serverAddress := startTestLaptopService(t, laptopRepo, nil, ratingRepo)
	laptopClient := newTestLaptopClient(t, serverAddress)

	stream, err := laptopClient.RateLaptop(context.Background())
	require.NoError(t, err)

	scores := []float64{8, 7.5, 10}
	averages := []float64{8, 7.75, 8.5}

	n := len(scores)
	for i := 0; i < n; i++ {
		req := &proto.RateLaptopRequest{
			LaptopId: laptop.GetId(),
			Score:    scores[i],
		}

		err := stream.Send(req)
		require.NoError(t, err)
	}

	err = stream.CloseSend()
	require.NoError(t, err)

	for idx := 0; ; idx++ {
		res, err := stream.Recv()
		if err == io.EOF {
			require.Equal(t, n, idx)
			return
		}

		require.NoError(t, err)
		require.Equal(t, laptop.GetId(), res.GetLaptopId())
		require.Equal(t, uint32(idx+1), res.GetRatedCount())
		require.Equal(t, averages[idx], res.GetAverageScore())
	}
}

func startTestLaptopService(t *testing.T, laptopRepo repository.LaptopRepository, imageRepo repository.ImageRepository, ratingRepo repository.RatingRepository) string {
	laptopServer := NewLaptopService(laptopRepo, imageRepo, ratingRepo)

	grpcServer := grpc.NewServer()
	proto.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0") // random available port
	require.NoError(t, err)

	go grpcServer.Serve(listener) // block call

	return listener.Addr().String()
}

func newTestLaptopClient(t *testing.T, serverAddress string) proto.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	return proto.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, laptop1 *proto.Laptop, laptop2 *proto.Laptop) {
	json1, err := serializer.ProtobufToJSON(laptop1)
	require.NoError(t, err)

	json2, err := serializer.ProtobufToJSON(laptop1)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}
