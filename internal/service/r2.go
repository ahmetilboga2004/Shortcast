package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Service struct {
	client     *s3.Client
	bucketName string
	accountID  string
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
		accountID:  accountID,
	}
}

// UploadFile dosyayı R2'ye yükler ve dosya anahtarını (key) döndürür
func (s *R2Service) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Benzersiz dosya adı oluştur
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	key := strings.Join([]string{folder, filename}, "/")

	fmt.Printf("R2 - Dosya yükleme işlemi başlatıldı. Key: %s\n", key)

	// Upload işlemi
	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   src,
		ACL:    "public-read",
	})

	if err != nil {
		fmt.Printf("R2 - HATA: Dosya yüklenirken hata oluştu. Key: %s, Hata: %v\n", key, err)
		return "", err
	}

	fmt.Printf("R2 - Dosya başarıyla yüklendi. Key: %s\n", key)
	return key, nil
}

func (s *R2Service) DeleteFile(key string) error {
	fmt.Printf("R2 - Dosya silme işlemi başlatıldı. Key: %s\n", key)

	if key == "" {
		fmt.Println("R2 - HATA: Boş key ile silme işlemi yapılamaz")
		return fmt.Errorf("dosya key'i boş olamaz")
	}

	// URL'den key'i çıkar
	if strings.Contains(key, "r2.cloudflarestorage.com") {
		parts := strings.Split(key, "r2.cloudflarestorage.com/")
		if len(parts) > 1 {
			key = parts[1]
		}
	}

	fmt.Printf("R2 - İşlenecek key: %s\n", key)

	// Önce dosyanın var olup olmadığını kontrol et
	_, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		fmt.Printf("R2 - HATA: Dosya bulunamadı veya erişilemedi. Key: %s, Hata: %v\n", key, err)
		return fmt.Errorf("dosya bulunamadı veya erişilemedi: %v", err)
	}

	fmt.Printf("R2 - DeleteObject çağrısı yapılıyor. Bucket: %s, Key: %s\n", s.bucketName, key)
	_, err = s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		fmt.Printf("R2 - HATA: Dosya silinirken hata oluştu. Key: %s, Hata: %v\n", key, err)
		return fmt.Errorf("dosya silinirken hata oluştu: %v", err)
	}

	fmt.Printf("R2 - Dosya başarıyla silindi. Key: %s\n", key)
	return nil
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
