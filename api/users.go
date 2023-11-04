package api

import (
	"crypto/rand"
	"errors"
	"net/url"

	"github.com/a19simma/go-webauthn-htmx/pkg/db"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func RegisterUserRoutes(router fiber.Router, userDb db.UserDb) {
	router.Delete("/:username", func(c *fiber.Ctx) error {
		username, err := url.QueryUnescape(c.Params("username"))
		log.Print(username)
		if err != nil {
			log.Err(err)
			return err
		}
		err = userDb.DeleteUser(username)
		if err != nil {
			log.Err(err)
			switch {
			case errors.Is(err, db.ErrNoResults):
				return c.SendStatus(404)
			default:
				return c.SendStatus(500)
			}
		}
		return nil
	})
	router.Post("/:username/block", func(c *fiber.Ctx) error {
		username, err := url.QueryUnescape(c.Params("username"))
		if err != nil {
			log.Err(err)
			return err
		}
		user, err := userDb.GetUser(username)
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
		user.Credentials = nil
		log.Printf("blocked user: %s", user.Username)
		return c.JSON(user)
	})
	router.Post("/:username/unblock", func(c *fiber.Ctx) error {
		username, err := url.QueryUnescape(c.Params("username"))
		if err != nil {
			log.Err(err)
			return err
		}
		user, err := userDb.GetUser(username)
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
		user.Credentials = nil
		log.Printf("unblocked user: %s", user.Username)
		return c.JSON(user)
	})

	router.Post("/", func(c *fiber.Ctx) error {
		username := c.FormValue("username")
		log.Printf("username: %s", username)

		if len(username) == 0 {
			return c.Status(400).SendString("no username")
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
