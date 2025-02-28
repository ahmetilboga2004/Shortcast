package router

import (
	"fmt"
	"os"
	"path/filepath"
	"shortcast/internal/container"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupAPIRoutes(app *fiber.App, cont *container.Container) {
	currentDir, _ := os.Getwd()
	uploadPath := filepath.Join(currentDir, "uploads")
	fmt.Printf("Static dosya yolu: %s\n", uploadPath)

	// Static middleware'i sadeleştirelim
	app.Static("/uploads", uploadPath, fiber.Static{
		Browse: true,
	})

	// CORS ayarlarını app seviyesinde yapalım
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000, http://localhost:8080",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
		MaxAge:           3000,
	}))

	// API routes
	api := app.Group("/api")

	api.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	auth := api.Group("/auth")
	auth.Post("/login", cont.AuthMiddleware.GuestMiddleware(), cont.AuthHandler.Login)
	auth.Post("/register", cont.AuthMiddleware.GuestMiddleware(), cont.AuthHandler.Register)
	auth.Post("/logout", cont.AuthMiddleware.JWTMiddleware(), cont.AuthHandler.Logout)

	user := api.Group("/users")
	user.Use(cont.AuthMiddleware.JWTMiddleware())
	user.Get("/:id", cont.UserHandler.GetByID)
	user.Get("/:user_id/podcasts", cont.PodcastHandler.GetUserPodcasts)

	podcast := api.Group("/podcasts")
	podcast.Use(cont.AuthMiddleware.JWTMiddleware())

	// Önce spesifik route'ları tanımla
	podcast.Get("/liked", cont.PodcastHandler.GetLikedPodcasts)
	podcast.Get("/discover", cont.PodcastHandler.DiscoverPodcasts)
	podcast.Get("/category/:category", cont.PodcastHandler.GetPodcastsByCategory)
	podcast.Get("/file/*", cont.PodcastHandler.GetFileContent)

	// Sonra parametreli route'ları tanımla
	podcast.Get("/:id", cont.PodcastHandler.GetPodcastByID)
	podcast.Post("/:id/like", cont.PodcastHandler.LikePodcast)
	podcast.Post("/:id/comments", cont.PodcastHandler.AddComment)
	podcast.Get("/:id/comments", cont.PodcastHandler.GetComments)
	podcast.Put("/:id", cont.PodcastHandler.UpdatePodcast)
	podcast.Delete("/:id", cont.PodcastHandler.DeletePodcast)
	podcast.Put("/:id/cover", cont.PodcastHandler.UpdatePodcastCover)

	// En son genel route'ları tanımla
	podcast.Post("/", cont.PodcastHandler.UploadPodcast)
}
