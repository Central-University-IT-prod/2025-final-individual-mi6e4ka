package services

import (
	"context"
	"fmt"
	"io"
	"log"

	"git.mi6e4ka.dev/prod-2025/internal/config"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	ctx    context.Context
	client *minio.Client
}

func NewS3Client(config *config.Config) (*S3Client, error) {
	ctx := context.Background()

	endpoint := config.S3.Endpoint
	accessKeyID := config.S3.User
	secretAccessKey := config.S3.Password

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
	exists, err := minioClient.BucketExists(ctx, "prod")
	if err != nil {
		log.Fatalf("failed connect to minio: %v", err)
	}
	if !exists {
		minioClient.MakeBucket(ctx, "prod", minio.MakeBucketOptions{Region: "prod"})
	}
	log.Println("connected to minio")
	return &S3Client{ctx: ctx, client: minioClient}, nil
}

func (s *S3Client) UploadImage(file io.Reader, fileSize int64, fileType string) (string, error) {
	s3Filename := uuid.NewString()
	info, err := s.client.PutObject(context.Background(), "prod", s3Filename, file, fileSize, minio.PutObjectOptions{ContentType: fileType})
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	return info.Key, err
}

func (s *S3Client) GetImage(key string) (io.Reader, int64, string, error) {
	obj, err := s.client.GetObject(context.Background(), "prod", key, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, "", err
	}
	if obj == nil {
		return nil, 0, "", fmt.Errorf("object not found")
	}
	stat, err := obj.Stat()
	if err != nil {
		return nil, 0, "", err
	}
	return obj, stat.Size, stat.ContentType, err
}

func (s *S3Client) DeleteImage(key string) error {
	err := s.client.RemoveObject(context.Background(), "prod", key, minio.RemoveObjectOptions{})
	return err
}
