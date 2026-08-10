package storage

import (
	"context"
	"io"
)

type StorageService interface {
	UploadFile(ctx context.Context, fileReader io.Reader, filename string, contentType string) (string, error)
}
