package utils

import (
	"context"
	"fmt"
	"time"

	"shortcast/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	r2Client *s3.Client
	r2Config aws.Config
)

func InitR2Client(cfg *config.Config) {
	r2Config = aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.R2.AccessKeyID,
			cfg.R2.AccessKeySecret,
			"",
		),
		Region: "auto",
	}

	r2Client = s3.New(s3.Options{
		BaseEndpoint: aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2.AccountID)),
		Region:       "auto",
		Credentials:  r2Config.Credentials,
		UsePathStyle: true,
	})
}

// GenerateSignedURL, dosya anahtarı (key) için imzalı URL oluşturur
// Örnek: audio/12345_music.mp3 -> https://accountid.r2.cloudflarestorage.com/bucket/audio/12345_music.mp3?imza...
func GenerateSignedURL(fileKey string, cfg *config.Config) (string, error) {
	if r2Client == nil {
		InitR2Client(cfg)
	}

	presignClient := s3.NewPresignClient(r2Client)

	request, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(cfg.R2.BucketName),
		Key:    aws.String(fileKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Hour * 24 // URL geçerlilik süresi 24 saat
	})

	if err != nil {
		return "", fmt.Errorf("imzalı URL oluşturulamadı: %w", err)
	}

	return request.URL, nil
}

// ExtractKeyFromURL, tam R2 URL'sinden dosya anahtarını (key) çıkarır
func ExtractKeyFromURL(fullURL string, cfg *config.Config) string {
	baseURL := fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/",
		cfg.R2.AccountID,
		cfg.R2.BucketName)

	if len(fullURL) > len(baseURL) && fullURL[:len(baseURL)] == baseURL {
		return fullURL[len(baseURL):]
	}
	return fullURL
}

// GenerateSignedURLs, birden fazla dosya anahtarı için imzalı URL'ler oluşturur
func GenerateSignedURLs(fileKeys []string, cfg *config.Config) (map[string]string, error) {
	result := make(map[string]string)

	for _, key := range fileKeys {
		if key == "" {
			continue
		}

		signedURL, err := GenerateSignedURL(key, cfg)
		if err != nil {
			return nil, err
		}

		result[key] = signedURL
	}

	return result, nil
}
