package api

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
)

func healthCheck(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

func RunAPIServer(port int) {
	app := fiber.New()

	app.Get("/health", healthCheck)

	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
