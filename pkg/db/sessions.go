package db

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
)

var LoginSession *session.Store

func Init() {

	sessionStore := sqlite3.New()

	LoginSession = session.New(session.Config{
		CookieSameSite: "strict",
		Expiration:     time.Minute * 5,
		Storage:        sessionStore,
	})
}
