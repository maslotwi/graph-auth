package api

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/maslotwi/graph-auth/docs"
	"github.com/maslotwi/graph-auth/helpers/environment"
	"github.com/swaggo/swag"
)

// healthCheck handles API health monitoring.
//
// @Summary             Health Check
// @Description         Returns the health status of the API
// @Tags                System
// @Produce             json
// @Success             200 {object} map[string]string
// @Router              /api/health [get]
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
func RunAPIServer() {
	docs.SwaggerInfo.Host = environment.BaseUrl

	app := fiber.New()

	// CORS: allow the Vite dev server in development; override via CORS_ORIGINS env var.
	allowedOrigins := os.Getenv("CORS_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: strings.Split(allowedOrigins, ","),
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
	}))
	RegisterOAuthRoutes(app)
	RegisterDelegationRoutes(app)

	// API routes
	api := app.Group("/api")
	api.Get("/health", healthCheck)

	// Swagger docs (kept at root so /docs and /swagger.json are always reachable)
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
		return c.SendString(`<!doctype html>
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
</html>`)
	})

	// Serve built frontend as SPA in production (only when frontend/dist exists).
	distPath := "./frontend/dist"
	if _, err := os.Stat(distPath); err == nil {
		app.Use("/", static.New(distPath, static.Config{
			NotFoundHandler: func(c fiber.Ctx) error {
				return c.SendFile(distPath + "/index.html")
			},
		}))
	}

	log.Fatal(app.Listen(fmt.Sprintf(":" + environment.Port)))
}
