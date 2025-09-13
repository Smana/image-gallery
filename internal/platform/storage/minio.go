package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"image-gallery/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	client     *minio.Client
	bucketName string
}

func NewMinIOClient(cfg config.StorageConfig) (*MinIOClient, error) {
	var creds *credentials.Credentials
	
	// Use AWS credentials chain if no static credentials are provided
	// This supports EKS Pod Identity, IAM roles, AWS credentials file, etc.
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		creds = credentials.NewChainCredentials([]credentials.Provider{
			&credentials.EnvAWS{},        // AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
			&credentials.FileAWSCredentials{}, // ~/.aws/credentials
			&credentials.IAM{},           // EC2/ECS/EKS IAM roles
		})
	} else {
		// Fall back to static credentials for local development
		creds = credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, "")
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  creds,
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

// ListObjects lists objects in the bucket
func (m *MinIOClient) ListObjects(ctx context.Context, prefix string, maxKeys int) ([]MinIOObjectInfo, error) {
	if maxKeys <= 0 {
		maxKeys = 1000
	}
	
	fmt.Printf("DEBUG: MinIO ListObjects - bucket=%s, prefix=%s, maxKeys=%d\n", m.bucketName, prefix, maxKeys)
	
	options := minio.ListObjectsOptions{
		Prefix:     prefix,
		MaxKeys:    maxKeys,
		Recursive:  true,
	}
	
	objectCh := m.client.ListObjects(ctx, m.bucketName, options)
	
	var objects []MinIOObjectInfo
	for object := range objectCh {
		if object.Err != nil {
			fmt.Printf("DEBUG: MinIO ListObjects error: %v\n", object.Err)
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		
		fmt.Printf("DEBUG: Found object: %s (size: %d, type: %s)\n", object.Key, object.Size, object.ContentType)
		
		objects = append(objects, MinIOObjectInfo{
			Key:          object.Key,
			Size:         object.Size,
			ContentType:  object.ContentType,
			LastModified: object.LastModified,
			ETag:         object.ETag,
			UserMetadata: object.UserMetadata,
		})
	}
	
	fmt.Printf("DEBUG: Total objects found: %d\n", len(objects))
	return objects, nil
}

// MinIOObjectInfo represents information about a stored object
type MinIOObjectInfo struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	LastModified time.Time         `json:"last_modified"`
	ETag         string            `json:"etag"`
	UserMetadata map[string]string `json:"user_metadata,omitempty"`
}
