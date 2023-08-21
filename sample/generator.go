package sample

import (
	"gitlab.com/iruldev/grpc-class/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewKeyboard() *proto.Keyboard {
	keyboard := &proto.Keyboard{
		Layout:  randomKeyboardLayout(),
		Backlit: randomBool(),
	}

	return keyboard
}

func NewCPU() *proto.CPU {
	brand := randomCPUBrand()
	name := randomCPUName(brand)
	numberCores := randomInt(2, 8)
	numberThreads := randomInt(numberCores, 12)
	minGhz := randomFloat64(2.0, 3.5)
	maxGhz := randomFloat64(minGhz, 5.0)
	cpu := &proto.CPU{
		Brand:         brand,
		Name:          name,
		NumberCores:   uint32(numberCores),
		NumberThreads: uint32(numberThreads),
		MinGhz:        minGhz,
		MaxGhz:        maxGhz,
	}
	return cpu
}

func NewGPU() *proto.GPU {
	brand := randomGPUBrand()
	name := randomGPUName(brand)
	minGhz := randomFloat64(1.0, 1.5)
	maxGhz := randomFloat64(minGhz, 2.0)
	memory := &proto.Memory{
		Value: uint64(randomInt(2, 6)),
		Unit:  proto.Memory_GIGABYTE,
	}
	gpu := &proto.GPU{
		Brand:  brand,
		Name:   name,
		MinGhz: minGhz,
		MaxGhz: maxGhz,
		Memory: memory,
	}
	return gpu
}

func NewRAM() *proto.Memory {
	ram := &proto.Memory{
		Value: uint64(randomInt(4, 64)),
		Unit:  proto.Memory_GIGABYTE,
	}
	return ram
}

func NewSSD() *proto.Storage {
	ssd := &proto.Storage{
		Driver: proto.Storage_SSD,
		Memory: &proto.Memory{
			Value: uint64(randomInt(128, 1024)),
			Unit:  proto.Memory_GIGABYTE,
		},
	}
	return ssd
}

func NewHDD() *proto.Storage {
	hdd := &proto.Storage{
		Driver: proto.Storage_HDD,
		Memory: &proto.Memory{
			Value: uint64(randomInt(1, 6)),
			Unit:  proto.Memory_TERABYTE,
		},
	}
	return hdd
}

func NewScreen() *proto.Screen {
	screen := &proto.Screen{
		SizeInch:   randomFloat32(13, 17),
		Resolution: randomScreenResolution(),
		Panel:      randomScreenPanel(),
		Multitouch: randomBool(),
	}
	return screen
}

func NewLaptop() *proto.Laptop {
	brand := randomLaptopBrand()
	name := randomLaptopName(brand)
	laptop := &proto.Laptop{
		Id:          randomID(),
		Brand:       brand,
		Name:        name,
		Cpu:         NewCPU(),
		Ram:         NewRAM(),
		Gpus:        []*proto.GPU{NewGPU()},
		Storages:    []*proto.Storage{NewSSD()},
		Screen:      NewScreen(),
		Keyboard:    NewKeyboard(),
		Weight:      &proto.Laptop_WeightKg{WeightKg: randomFloat64(1.0, 3.0)},
		PriceUsd:    randomFloat64(1500, 3000),
		ReleaseYear: uint32(randomInt(2015, 2022)),
		UpdatedAt:   timestamppb.Now(),
	}

	return laptop
}

func RandomLaptopScore() float64 {
	return float64(randomInt(1, 10))
}
