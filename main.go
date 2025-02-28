package main

import (
	_ "shortcast/docs"
	"shortcast/internal/container"
	"shortcast/internal/router"

	"github.com/gofiber/fiber/v2"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @Security					BearerAuth
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	cont := container.NewContainer()

	app := fiber.New(
		fiber.Config{
			BodyLimit: 100 * 1024 * 1024,
		},
	)

	router.SetupAPIRoutes(app, cont)
	router.SetupDocsRoutes(app)

	app.Listen(":8080")

}
