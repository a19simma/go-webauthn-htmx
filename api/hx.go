package api

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/a19simma/go-webauthn-htmx/pkg/db"
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
		url := fmt.Sprintf("%s/api/users/%s/block", c.BaseURL(), c.Params("username"))
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
		url := fmt.Sprintf("%s/api/users/%s/unblock", c.BaseURL(), c.Params("username"))
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
	hx.Post("/users", func(c *fiber.Ctx) error {
		url := fmt.Sprintf("%s/api/users", c.BaseURL())
		agent := fiber.Post(url)
		agent.Cookie("session_id", c.Cookies("session_id"))
		args := fiber.AcquireArgs()
		args.Set("username", c.FormValue("username"))
		agent.Form(args)
		status, body, errs := agent.Bytes()
		if len(errs) > 0 {
			for _, v := range errs {
				log.Err(v)
			}
		}
		statusText := string(body)
		log.Print(status)
		var statusState string
		switch {
		case status < 299:
			statusState = "succes"
		default:
			statusState = "error"
		}

		return c.Render("components/addUserForm",
			struct {
				Status     string
				StatusText string
			}{
				Status:     statusState,
				StatusText: statusText,
			})
	})
}
