package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"gitlab.com/iruldev/grpc-class/engine/repository"
	"gitlab.com/iruldev/grpc-class/proto"
	"gitlab.com/iruldev/grpc-class/sample"
	"gitlab.com/iruldev/grpc-class/serializer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net"
	"testing"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopServer, serverAddress := startTestLaptopService(t, repository.NewLaptopRepository())
	laptopClient := newTestLaptopClient(t, serverAddress)

	laptop := sample.NewLaptop()
	expectedID := laptop.Id
	req := &proto.CreateLaptopRequest{Laptop: laptop}

	res, err := laptopClient.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, expectedID, res.Id)

	// check that the laptop is saved to the store
	other, err := laptopServer.LaptopRepository.Find(expectedID)
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

	laptopRepository := repository.NewLaptopRepository()
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

		err := laptopRepository.Save(laptop)
		require.NoError(t, err)
	}

	_, serverAddress := startTestLaptopService(t, laptopRepository)
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

func startTestLaptopService(t *testing.T, laptopRepository repository.LaptopRepository) (*LaptopService, string) {
	laptopServer := NewLaptopService(laptopRepository)

	grpcServer := grpc.NewServer()
	proto.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0") // random available port
	require.NoError(t, err)

	go grpcServer.Serve(listener) // block call

	return laptopServer, listener.Addr().String()
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
