package db

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
	"github.com/rs/zerolog/log"
)

var LoginSession *session.Store

func Init() {
	sessionStore := sqlite3.New()

	LoginSession = session.New(session.Config{
		CookieSameSite: "strict",
		Expiration:     time.Hour * 6,
		Storage:        sessionStore,
	})
}

func DeleteLoginSession(c *fiber.Ctx) error {
	sess, err := LoginSession.Get(c)
	if err != nil {
		log.Err(err)
		return err
	}
	err = sess.Destroy()
	if err != nil {
		log.Err(err)
		return err
	}
	return nil
}

func GetLoginSession(c *fiber.Ctx, username string) (*session.Session, error) {
	sess, err := LoginSession.Get(c)
	if err != nil {
		return nil, err
	}
	sess.Set("username", username)
	err = sess.Save()
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// Validates the Session attached to input context
// returns the username or error
func ValidateLoginSession(c *fiber.Ctx) (username string, err error) {
	s := c.Request().Header.Cookie("session_id")
	sess, err := LoginSession.Get(c)
	if err != nil {
		return "", err
	}

	if sess.ID() != string(s) {
		return "", errors.New("sessions did not match")
	}
	username = sess.Get("username").(string)
	user := &User{}
	db.Where("Username = ?", username).First(&user)
	if user.ID == nil {
		log.Info().Str("Username", username).Str("Session", string(s)).Send()
		return username, errors.New("Username does not exist")
	}

	return username, nil
}
