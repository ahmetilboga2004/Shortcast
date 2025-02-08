package repository

import (
	"context"
	"shortcast/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewAuthRepository(db *gorm.DB, redis *redis.Client) *AuthRepository {
	return &AuthRepository{
		db:    db,
		redis: redis,
	}
}

func (r *AuthRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *AuthRepository) BlacklistToken(token string, expiry time.Time) error {
	ctx := context.Background()
	duration := time.Until(expiry)

	// Token'ı blacklist'e ekle
	return r.redis.Set(ctx, "blacklist:"+token, true, duration).Err()
}

func (r *AuthRepository) IsTokenBlacklisted(token string) (bool, error) {
	ctx := context.Background()
	exists, err := r.redis.Exists(ctx, "blacklist:"+token).Result()
	return exists == 1, err
}
