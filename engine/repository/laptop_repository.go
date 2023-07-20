package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gitlab.com/iruldev/grpc-class/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"sync"
)

var ErrAlreadyExists = errors.New("record already exists")

type LaptopRepository interface {
	Save(laptop *proto.Laptop) error
	Find(id string) (*proto.Laptop, error)
	Search(ctx context.Context, filter *proto.Filter, found func(laptop *proto.Laptop) error) error
}

type LaptopRepositoryImpl struct {
	mutex sync.RWMutex
	data  map[string]*proto.Laptop
}

func NewLaptopRepository() LaptopRepository {
	return &LaptopRepositoryImpl{
		data: make(map[string]*proto.Laptop),
	}
}

func (r *LaptopRepositoryImpl) Save(laptop *proto.Laptop) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	other, err := deepCopy(laptop)
	if err != nil {
		return err
	}

	r.data[other.Id] = other
	return nil
}

func (r *LaptopRepositoryImpl) Find(id string) (*proto.Laptop, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	laptop := r.data[id]
	if laptop == nil {
		return nil, errors.New("record not exists")
	}

	return deepCopy(laptop)
}

func (r *LaptopRepositoryImpl) Search(ctx context.Context, filter *proto.Filter, found func(laptop *proto.Laptop) error) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, laptop := range r.data {
		if ctx.Err() == context.Canceled {
			log.Print("request is canceled")
			return status.Error(codes.Canceled, "request is canceled")
		}

		if ctx.Err() == context.DeadlineExceeded {
			log.Print("deadline is exceeded")
			return status.Error(codes.DeadlineExceeded, "deadline is exceeded")
		}

		if isQualified(filter, laptop) {
			other, err := deepCopy(laptop)
			if err != nil {
				return err
			}

			err = found(other)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isQualified(filter *proto.Filter, laptop *proto.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}

	if laptop.GetCpu().GetNumberCores() < filter.GetMinCpuCores() {
		return false
	}

	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}

	if toBit(laptop.GetRam()) < toBit(filter.GetMinRam()) {
		return false
	}

	return true
}

func toBit(memory *proto.Memory) uint64 {
	value := memory.GetValue()

	switch memory.GetUnit() {
	case proto.Memory_BIT:
		return value
	case proto.Memory_BYTE:
		return value << 3 // 8 = 2^3
	case proto.Memory_KILOBYTE:
		return value << 13 // 1024 * 8 = 2^10 * 2^3 = 2^13
	case proto.Memory_MEGABYTE:
		return value << 23
	case proto.Memory_GIGABYTE:
		return value << 33
	case proto.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}

func deepCopy(laptop *proto.Laptop) (*proto.Laptop, error) {
	// deep copy
	other := &proto.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}
	return other, nil
}
