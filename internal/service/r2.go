package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Service struct {
	client     *s3.Client
	bucketName string
}

func NewR2Service(accountID, accessKeyID, accessKeySecret string, bucketName string) *R2Service {
	cfg := aws.Config{
		Credentials:  credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, ""),
		Region:       "auto",
		BaseEndpoint: aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)),
	}

	client := s3.NewFromConfig(cfg)
	return &R2Service{
		client:     client,
		bucketName: bucketName,
	}
}

func (s *R2Service) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Benzersiz dosya adı oluştur
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	key := filepath.Join(folder, filename)

	// Upload işlemi
	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   src,
		ACL:    "public-read",
	})

	if err != nil {
		return "", err
	}

	// Public URL döndür
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", s.bucketName, key), nil
}

func (s *R2Service) DeleteFile(key string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	return err
}

func (s *R2Service) GetPresignedURL(key string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))

	if err != nil {
		return "", err
	}

	return request.URL, nil
}
