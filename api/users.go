package api

import (
	"net/url"

	"github.com/a19simma/vanilla-js/pkg/db"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func RegisterUserRoutes(router fiber.Router) {
	router.Delete("/:id", func(c *fiber.Ctx) error {
		id, err := url.QueryUnescape(c.Params("id"))
		if err != nil {
			log.Err(err)
			return err
		}
		userId := []byte(id)
		log.Debug().Msgf("Deleting user with id: %s", id)
		result := db.Users.Where("user_id = ?", userId).Delete(db.Credentials{})
		if result.Error != nil {
			log.Err(result.Error)
		}
		if result.RowsAffected == 0 {
			return c.SendStatus(404)
		}

		sess, err := db.LoginSession.Get(c)
		if err != nil {
			log.Err(err)
		}
		err = sess.Destroy()
		if err != nil {
			log.Err(err)
		}

		result = db.Users.Where("id = ?", userId).Delete(db.User{})
		if result.Error != nil {
			log.Err(result.Error)
		}
		if result.RowsAffected == 0 {
			return c.SendStatus(404)
		}
		return nil
	})
}
