package api

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/maslotwi/graph-auth/docs"
	"github.com/swaggo/swag"
)

// healthCheck handles API health monitoring.
//
// @Summary             Health Check
// @Description         Returns the health status of the API
// @Tags                System
// @Produce             json
// @Success             200 {object} map[string]string
// @Router              /health [get]
func healthCheck(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

// RunAPIServer is the main api entrypoint
//
// @title               Graph Auth
// @version             1.0
// @description         This is the core REST API for Graph Auth.
// @license.name        MIT
// @license.url         https://opensource.org/licenses/MIT
// @BasePath            /
func RunAPIServer(port int) {
	if envHost := os.Getenv("API_HOST"); envHost != "" {
		docs.SwaggerInfo.Host = envHost
	} else {
		docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", port)
	}

	app := fiber.New()

	api := app.Group("/api")
	api.Get("/health", healthCheck)

	app.Get("/swagger.json", func(c fiber.Ctx) error {
		// Unpack the string and the error
		doc, err := swag.ReadDoc()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to load swagger docs")
		}

		c.Type("json")
		return c.SendString(doc)
	})

	app.Get("/docs", func(c fiber.Ctx) error {
		c.Type("html")
		return c.SendString(`
<!doctype html>
<html>
  <head>
    <title>API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>body { margin: 0; }</style>
  </head>
  <body>
    <script id="api-reference" data-url="/swagger.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>
		`)
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
