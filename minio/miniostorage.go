package minio

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

type Storage struct {
	client   *minio.Client
	bucket   string
	basePath string
}

func NewMinIOStorage(endpoint, accessKey, secretKey, bucket, basePath string, useSSL bool) (*Storage, error) {
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		return nil, fmt.Errorf("minio new client: %w", err)
	}

	return &Storage{
		client:   cli,
		bucket:   bucket,
		basePath: basePath,
	}, nil
}

func (m *Storage) StoreFile(ctx context.Context, localPath, mimeType string) (string, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open file for minio upload: %w", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	fileName := filepath.Base(localPath)
	objectName := fileName
	if m.basePath != "" {
		objectName = filepath.Join(m.basePath, fileName)
	}

	log.Info().
		Str("bucket", m.bucket).
		Str("objectName", objectName).
		Msg("Uploading file to MinIO")

	contentType := mimeType

	_, err = m.client.PutObject(
		ctx,
		m.bucket,
		objectName,
		f,
		-1,
		minio.PutObjectOptions{ContentType: contentType},
	)

	if err != nil {
		return "", fmt.Errorf("minio put object: %w", err)
	}

	return fmt.Sprintf("minio://%s/%s", m.bucket, objectName), nil
}
