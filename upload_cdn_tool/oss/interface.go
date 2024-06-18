package oss

import (
	"context"
	"io"
)

type OSS interface {
	UploadObject(ctx context.Context, name string, content io.Reader, contentLength int64) (downloadURL string, err error)
	OpenObject(ctx context.Context, name string) (io.ReadCloser, error)
}
