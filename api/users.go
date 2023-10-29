package api

import (
	"errors"
	"net/url"

	"github.com/a19simma/vanilla-js/pkg/db"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func RegisterUserRoutes(router fiber.Router, userDb db.UserDb) {
	router.Delete("/:id", func(c *fiber.Ctx) error {
		id, err := url.QueryUnescape(c.Params("id"))
		if err != nil {
			log.Err(err)
			return err
		}
		err = userDb.DeleteUser(id)
		if err != nil {
			log.Err(err)
			switch {
			case errors.Is(err, db.ErrNoResults):
				return c.SendStatus(404)
			default:
				return c.SendStatus(500)
			}
		}
		db.DeleteLoginSession(c)
		return nil
	})
}
