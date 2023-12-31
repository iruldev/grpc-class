package repository

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"os"
	"sync"
)

type ImageRepository interface {
	Save(laptopID, imageType string, imageData bytes.Buffer) (string, error)
}

type ImageRepositoryImpl struct {
	mutex       sync.RWMutex
	imageFolder string
	images      map[string]*ImageInfo
}

type ImageInfo struct {
	LaptopID string
	Type     string
	Path     string
}

func NewImageRepository(imageFolder string) ImageRepository {
	return &ImageRepositoryImpl{
		imageFolder: imageFolder,
		images:      make(map[string]*ImageInfo),
	}
}

func (r *ImageRepositoryImpl) Save(laptopID, imageType string, imageData bytes.Buffer) (string, error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate image id: %w", err)
	}

	imagePath := fmt.Sprintf("%s/%s%s", r.imageFolder, imageID, imageType)

	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: %w", err)
	}

	_, err = imageData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write image to file: %w", err)
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.images[imageID.String()] = &ImageInfo{
		LaptopID: laptopID,
		Type:     imageType,
		Path:     imagePath,
	}

	return imageID.String(), nil
}
