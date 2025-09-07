package storage

import (
	"context"
	"io"
	"path/filepath"

	"image-gallery/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	client     *minio.Client
	bucketName string
}

func NewMinIOClient(cfg config.StorageConfig) (*MinIOClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}

	minioClient := &MinIOClient{
		client:     client,
		bucketName: cfg.BucketName,
	}

	if err := minioClient.ensureBucket(); err != nil {
		return nil, err
	}

	return minioClient, nil
}

func (m *MinIOClient) ensureBucket() error {
	ctx := context.Background()
	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return err
	}

	if !exists {
		return m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{
			Region: "us-east-1",
		})
	}

	return nil
}

func (m *MinIOClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (m *MinIOClient) GetFile(ctx context.Context, objectName string) (*minio.Object, error) {
	return m.client.GetObject(ctx, m.bucketName, objectName, minio.GetObjectOptions{})
}

func (m *MinIOClient) DeleteFile(ctx context.Context, objectName string) error {
	return m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
}

func (m *MinIOClient) GetFileURL(objectName string) string {
	return filepath.Join("/api/images", objectName, "download")
}