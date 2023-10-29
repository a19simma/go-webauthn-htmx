package api

import (
	"crypto/rand"
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
		err = db.DeleteLoginSession(c)
		if err != nil {
			return err
		}
		return nil
	})
	router.Post("/:id/block", func(c *fiber.Ctx) error {
		id, err := url.QueryUnescape(c.Params("id"))
		if err != nil {
			log.Err(err)
			return err
		}
		user, err := userDb.GetUser(id)
		if err != nil {
			log.Err(err)
			switch {
			case errors.Is(err, db.ErrNoResults):
				return c.SendStatus(404)
			default:
				return c.SendStatus(500)
			}
		}
		user.Status = db.Blocked
		err = userDb.CreateUser(*user)
		if err != nil {
			return err
		}
		return nil
	})
	router.Post("/:id/unblock", func(c *fiber.Ctx) error {
		id, err := url.QueryUnescape(c.Params("id"))
		if err != nil {
			log.Err(err)
			return err
		}
		user, err := userDb.GetUser(id)
		if err != nil {
			log.Err(err)
			switch {
			case errors.Is(err, db.ErrNoResults):
				return c.SendStatus(404)
			default:
				return c.SendStatus(500)
			}
		}
		creds := userDb.GetUserCredentials(*user)
		if len(creds) > 0 {
			user.Status = db.Registered
		} else {
			user.Status = db.Open
		}
		err = userDb.CreateUser(*user)
		if err != nil {
			return err
		}
		return nil
	})

	router.Post("/", func(c *fiber.Ctx) error {
		username := c.FormValue("username")
		log.Printf("username: %s", username)

		if len(username) == 0 {
			c.Status(400)
			return errors.New("no username")
		}
		id := make([]byte, 32)
		_, err := rand.Read(id)
		if err != nil {
			return err
		}

		user := db.User{ID: id, Username: username, Status: db.Open, Role: db.Member}
		err = userDb.CreateUser(user)
		if err != nil {
			return err
		}

		return nil
	})
}
