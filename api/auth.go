package api

import (
	"errors"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	auth "github.com/a19simma/go-webauthn-htmx/pkg"
	"github.com/a19simma/go-webauthn-htmx/pkg/db"
)

var (
	authSvc auth.Auth
)

func RegisterAuthRoutes(r fiber.Router, userDb db.UserDb) {
	log.Printf("Router uDb: %v", userDb)
	auth, err := auth.InitAuth(userDb)
	authSvc = auth
	if err != nil {
		log.Fatal().Err(err)
	}

	r.Get("/register/begin/:username", func(c *fiber.Ctx) error {
		options, err := authSvc.BeginRegistration(c.Params("username"))
		if err != nil {
			return err
		}
		return c.JSON(options)
	})
	r.Post("/verify-registration/:username", func(c *fiber.Ctx) error {
		body := new(protocol.CredentialCreationResponse)
		if err := c.BodyParser(body); err != nil {
			return err
		}
		response, err := body.Parse()
		if err != nil {
			return err
		}
		err = authSvc.FinishRegistration(*response, c.Params("username"))
		if err != nil {
			return err
		}
		_, err = db.GetLoginSession(c, c.Params("username"))
		if err != nil {
			log.Err(err)
		}

		return c.JSON("Registration Success")
	})
	r.Get("/generate-authentication-options/:username", func(c *fiber.Ctx) error {
		resp, err := authSvc.BeginLogin(c.Params("username"))
		if err != nil {
			log.Err(err)
			return err
		}
		return c.JSON(resp)
	})
	r.Post("/verify-authentication/:username", func(c *fiber.Ctx) error {
		body := new(protocol.CredentialAssertionResponse)
		if err := c.BodyParser(body); err != nil {
			log.Err(err)
			return err
		}
		response, err := body.Parse()
		if err != nil {
			log.Err(err)
			return err
		}
		err = authSvc.FinishLogin(c.Params("username"), *response)
		if err != nil {
			return err
		}

		_, err = db.GetLoginSession(c, c.Params("username"))
		if err != nil {
			return err
		}

		return nil
	})
	r.Get("/status", func(c *fiber.Ctx) error {
		s := c.Request().Header.Cookie("session_id")
		if len(s) == 0 {
			return c.SendStatus(204)
		}

		username, err := db.ValidateLoginSession(c)
		if err != nil {
			return c.SendStatus(204)
		}

		return c.SendString(fmt.Sprint(username))
	})
	r.Get("/logout", func(c *fiber.Ctx) error {
		err := db.DeleteLoginSession(c)
		if err != nil {
			return err
		}
		c.Response().Header.Set("HX-Redirect", "/login")
		return c.Render("components/loginCard", nil)
	})

}

func CheckLoginStatus(c *fiber.Ctx) (string, error) {
	s := c.Request().Header.Cookie("session_id")
	if len(s) == 0 {
		return "", errors.New("Session Cookie was not present")
	}

	username, err := db.ValidateLoginSession(c)
	if err != nil {
		log.Err(err)
		return "", err
	}
	return username, nil
}

func NewLoginRedirect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		s := c.Request().Header.Cookie("session_id")
		if len(s) == 0 {
			return c.Redirect("/login", 302)
		}
		_, err := db.ValidateLoginSession(c)
		if err != nil {
			log.Printf("failed to validate login: %v", err)
			return c.Redirect("/login", 302)
		}
		return c.Next()
	}
}
