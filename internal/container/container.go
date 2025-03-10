package container

import (
	"log"
	"shortcast/internal/config"
	"shortcast/internal/handler"
	"shortcast/internal/middleware"
	"shortcast/internal/model"
	"shortcast/internal/repository"
	"shortcast/internal/service"
)

type Container struct {
	AuthHandler    *handler.AuthHandler
	UserHandler    *handler.UserHandler
	PodcastHandler *handler.PodcastHandler
	AuthMiddleware *middleware.AuthMiddleware
	R2Service      *service.R2Service
}

func NewContainer() *Container {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config yüklenemedi: %v", err)
	}

	db := config.ConnectDB(cfg)
	redis, err := config.ConnectRedis(cfg)
	if err != nil {
		log.Fatalf("Redis bağlantısı kurulamadı: %v", err)
	}

	// Migrasyon işlemlerini burada yapıyoruz
	err = db.AutoMigrate(
		&model.User{},
		&model.Podcast{},
		&model.Like{},    // Like modelini ekledik
		&model.Comment{}, // Comment modelini ekledik
	)
	if err != nil {
		log.Fatalf("Migrasyon hatası: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserService(userService)

	authRepo := repository.NewAuthRepository(db, redis)
	authService := service.NewAuthService(authRepo, userRepo, cfg)
	authHandler := handler.NewAuthHandler(authService)

	podcastRepo := repository.NewPodcastRepository(db)
	r2Service := service.NewR2Service(
		cfg.R2.AccountID,
		cfg.R2.AccessKeyID,
		cfg.R2.AccessKeySecret,
		cfg.R2.BucketName,
	)
	podcastService := service.NewPodcastService(podcastRepo, userRepo, r2Service, cfg)
	podcastHandler := handler.NewPodcastHandler(podcastService)

	authMiddleware := middleware.NewAuthMiddleware(cfg, authRepo)

	return &Container{
		AuthHandler:    authHandler,
		UserHandler:    userHandler,
		PodcastHandler: podcastHandler,
		AuthMiddleware: authMiddleware,
		R2Service:      r2Service,
	}
}
