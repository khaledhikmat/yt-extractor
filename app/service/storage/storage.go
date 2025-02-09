package storage

import (
	"context"
	"fmt"

	"github.com/khaledhikmat/yt-extractor/service/config"
)

var providers map[string]IService

type storageService struct {
	ConfigSvc config.IService
}

func New(cfgsvc config.IService) IService {
	providers = map[string]IService{
		"s3": newS3(cfgsvc),
	}
	return &storageService{
		ConfigSvc: cfgsvc,
	}
}

func (svc *storageService) NewFile(ctx context.Context, folder, filePath, identifier string) (string, error) {
	r, ok := providers[svc.ConfigSvc.GetStorageProvider()]
	if !ok {
		return "", fmt.Errorf("storage provider %s not found", svc.ConfigSvc.GetStorageProvider())
	}

	return r.NewFile(ctx, folder, filePath, identifier)
}
