package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	SecretKey     string
	JWTExpiration int
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	R2            R2Config
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Çevre değişkeni yüklenirken hata oluştu: %s", err)
	}

	return &Config{
		Port:          getEnv("PORT", "8080"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "ahmet"),
		DBPassword:    getEnv("DB_PASSWORD", "shortcast"),
		DBName:        getEnv("DB_NAME", "shortcast"),
		SecretKey:     getEnv("SECRET_KEY", "supersecretkey"), // JWT secret key
		JWTExpiration: getEnvAsInt("JWT_EXPIRATION", 3600),    // JWT expiration süresi (saniye olarak)
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),
		R2: R2Config{
			AccountID:       os.Getenv("R2_ACCOUNT_ID"),
			AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
			AccessKeySecret: os.Getenv("R2_ACCESS_KEY_SECRET"),
			BucketName:      os.Getenv("R2_BUCKET_NAME"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	value := getEnv(key, fmt.Sprintf("%d", defaultValue))
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
