package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func main() {
	app := fiber.New()

	// Match any route
	app.Use(func(c *fiber.Ctx) error {

		fmt.Println("🥇 First handler")
		return c.Next()
	})

	// Match all routes starting with /api
	app.Use("/api", func(c *fiber.Ctx) error {
		fmt.Println("🥈 Second handler")
		return c.Next()
	})

	// GET /api/register
	app.Get("/api/list", func(c *fiber.Ctx) error {
		fmt.Println("🥉 Last handler")
		return c.SendString("Hello, World 👋!")
	})

	log.Fatal(app.Listen(":3000"))
}
