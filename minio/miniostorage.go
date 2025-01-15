package minio

import (
	"bytes"
	"context"
	"fmt"

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

func (m *Storage) StoreBytes(ctx context.Context, data []byte, mimeType, objectName string) (string, error) {
	log.Info().
		Str("bucket", m.bucket).
		Str("objectName", objectName).
		Msg("Uploading memory data to MinIO")

	reader := bytes.NewReader(data)
	_, err := m.client.PutObject(
		ctx,
		m.bucket,
		objectName,
		reader,
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: mimeType,
		},
	)
	if err != nil {
		return "", fmt.Errorf("minio put object: %w", err)
	}

	return fmt.Sprintf("minio://%s/%s", m.bucket, objectName), nil
}
