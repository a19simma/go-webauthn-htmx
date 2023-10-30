package api

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/a19simma/vanilla-js/pkg/db"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
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
	hx.Get("/login", func(c *fiber.Ctx) error {
		username, err := CheckLoginStatus(c)
		if err != nil {
			return c.Render("login", struct{ Status bool }{Status: false})
		}
		c.Response().Header.Set("HX-Redirect", "/")
		return c.Render("components/loginCard",
			struct {
				Status   bool
				Username string
			}{
				Status:   true,
				Username: username,
			})
	})
	hx.Delete("/users/:id", func(c *fiber.Ctx) error {
		url := fmt.Sprintf("/api/users/%s", c.Params("id"))
		agent := fiber.Delete(url)
		status, _, errs := agent.Bytes()
		if len(errs) > 0 {
			log.Err(errs[0])
		}
		if status > 299 {
			return c.SendStatus(status)
		}

		return c.SendStatus(200)
	})
	hx.Post("/users/:username/block", func(c *fiber.Ctx) error {
		url := fmt.Sprintf("http://localhost:4200/api/users/%s/block", c.Params("username"))
		log.Print(url)
		agent := fiber.Post(url)
		agent.Cookie("session_id", c.Cookies("session_id"))
		status, body, errs := agent.Bytes()
		log.Print(status, body, errs)
		if len(errs) > 0 {
			log.Err(errs[0])
		}
		if status > 299 {
			return c.SendStatus(status)
		}

		var user db.User
		err := json.Unmarshal(body, &user)
		if err != nil {
			log.Err(err)
		}
		return c.Render("components/userTableRow",
			struct {
				ID       string
				Status   db.RegistrationStatus
				Username string
				Role     db.Role
			}{
				ID:       string(user.ID),
				Status:   user.Status,
				Username: user.Username,
				Role:     user.Role,
			})
	})
	hx.Post("/users/:username/unblock", func(c *fiber.Ctx) error {
		url := fmt.Sprintf("http://localhost:4200/api/users/%s/unblock", c.Params("username"))
		agent := fiber.Post(url)
		agent.Cookie("session_id", c.Cookies("session_id"))
		status, body, errs := agent.Bytes()
		log.Print(status)
		if len(errs) > 0 {
			log.Err(errs[0])
		}
		if status > 299 {
			return c.SendStatus(status)
		}
		var user db.User
		err := json.Unmarshal(body, &user)
		if err != nil {
			log.Err(err)
		}

		return c.Render("components/userTableRow",
			struct {
				ID       string
				Status   db.RegistrationStatus
				Username string
				Role     db.Role
			}{
				ID:       string(user.ID),
				Status:   user.Status,
				Username: user.Username,
				Role:     user.Role,
			})
	})
}
