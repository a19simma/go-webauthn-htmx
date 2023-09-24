package api

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
)

func RegisterHXRoutes(hx fiber.Router) {
	hx.Post("/test", func(c *fiber.Ctx) error {
		q, err := url.ParseQuery(string(c.Body()))
		if err != nil {
			return err
		}
		_, err = c.WriteString(q.Get("text"))
		return err
	})
	hx.Post("/toast", func(c *fiber.Ctx) error {
		q, err := url.ParseQuery(string(c.Body()))
		if err != nil {
			return err
		}
		return c.Render("components/toast", q.Get("text"))

	})
}
