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
}

func NewContainer() *Container {
	cfg := config.LoadConfig()
	db := config.ConnectDB(cfg)
	redis := config.ConnectRedis(cfg)

	// Migrasyon işlemlerini burada yapıyoruz
	err := db.AutoMigrate(
		&model.User{},
		&model.Podcast{},
		&model.Like{},     // Like modelini ekledik
		&model.Comment{},  // Comment modelini ekledik
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
	podcastService := service.NewPodcastService(podcastRepo, userRepo)
	podcastHandler := handler.NewPodcastHandler(podcastService)

	authMiddleware := middleware.NewAuthMiddleware(cfg, authRepo)

	return &Container{
		AuthHandler:    authHandler,
		UserHandler:    userHandler,
		PodcastHandler: podcastHandler,
		AuthMiddleware: authMiddleware,
	}
}
