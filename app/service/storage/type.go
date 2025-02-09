package storage

import (
	"context"
)

type IService interface {
	NewFile(ctx context.Context, folder, filePath, identifier string) (string, error)
}
