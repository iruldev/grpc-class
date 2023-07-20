package sample

import (
	"github.com/google/uuid"
	"gitlab.com/iruldev/grpc-class/proto"
	"math/rand"
)

func randomKeyboardLayout() proto.Keyboard_Layout {
	switch rand.Intn(3) {
	case 1:
		return proto.Keyboard_QWERTY
	case 2:
		return proto.Keyboard_QWERTZ
	case 3:
		return proto.Keyboard_AZERTY
	default:
		return proto.Keyboard_UNKNOWN
	}
}

func randomStringFromSet(s ...string) string {
	n := len(s)
	if n == 0 {
		return ""
	}
	return s[rand.Intn(n)]
}

func randomCPUBrand() string {
	return randomStringFromSet("Intel", "AMD")
}

func randomGPUBrand() string {
	return randomStringFromSet("NVIDIA", "AMD")
}

func randomLaptopBrand() string {
	return randomStringFromSet("Apple", "Dell", "Lenovo")
}

func randomCPUName(brand string) string {
	if brand == "Intel" {
		return randomStringFromSet(
			"Xeon E-2286M",
			"Core i9-9980HK",
			"Core i7-9750H",
			"Core i5-9400F",
			"Core i3-1005G1",
		)
	}

	return randomStringFromSet(
		"Ryzen 7 PRO 2700U",
		"Ryzen 5 PRO 3500U",
		"Ryzen 3 PRO 3200GE",
	)
}

func randomGPUName(brand string) string {
	if brand == "NVIDIA" {
		return randomStringFromSet(
			"RTX 2060",
			"RTX 2070",
			"RTX 1660-Ti",
			"GTX 1070",
		)
	}

	return randomStringFromSet(
		"RX 590",
		"RX 580",
		"RX 5700-XT",
		"RX Vega-56",
	)
}

func randomLaptopName(brand string) string {
	switch brand {
	case "Apple":
		return randomStringFromSet("Macbook Air", "Macbook Pro")
	case "Dell":
		return randomStringFromSet("Latitude", "Vostro", "XPS", "Alienware")
	default:
		return randomStringFromSet("Thinkpad XL", "Thinkpad P1", "Thinkpad PS3")
	}
}

func randomScreenPanel() proto.Screen_Panel {
	if rand.Intn(2) == 1 {
		return proto.Screen_IPS
	}
	return proto.Screen_OLED
}

func randomScreenResolution() *proto.Screen_Resolution {
	height := randomInt(1000, 4320)
	width := height * 16 / 9

	resolution := &proto.Screen_Resolution{
		Width:  uint32(width),
		Height: uint32(height),
	}

	return resolution
}

func randomID() string {
	return uuid.New().String()
}

func randomBool() bool {
	return rand.Intn(2) == 1
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func randomFloat64(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func randomFloat32(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}
