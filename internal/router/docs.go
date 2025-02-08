package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func SetupDocsRoutes(app *fiber.App) {
	app.Get("/docs/*", swagger.HandlerDefault)
}
