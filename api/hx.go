package api

import "github.com/gofiber/fiber/v2"

func RegisterHXRoutes(hx fiber.Router) {
	hx.Post("/test", func(c *fiber.Ctx) error {
		_, err := c.WriteString("Hello!")
		return err
	})
}
