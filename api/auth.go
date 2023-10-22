package api

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/a19simma/vanilla-js/pkg/db"
)

var (
	web *webauthn.WebAuthn
	err error
)

func RegisterAuthRoutes(auth fiber.Router) {

	wconfig := &webauthn.Config{
		RPDisplayName: "Go Webauthn",                     // Display Name for your site
		RPID:          "localhost",                       // Generally the FQDN for your site
		RPOrigins:     []string{"http://localhost:4200"}, // The origin URLs allowed for WebAuthn requests
	}

	if web, err = webauthn.New(wconfig); err != nil {
		fmt.Println(err)
	}

	auth.Get("/register/begin/:username", BeginRegistration)
	auth.Post("/verify-registration/:username", func(c *fiber.Ctx) error {
		FinishRegistration(c, c.Params("username"))
		return nil
	})
	auth.Get("/generate-authentication-options/:username", func(c *fiber.Ctx) error {
		return BeginLogin(c)
	})
	auth.Post("/verify-authentication/:username", func(c *fiber.Ctx) error {
		return FinishLogin(c)
	})
	auth.Get("/users/:username", func(c *fiber.Ctx) error {
		user := db.User{Username: c.Params("username")}
		db.Users.First(&db.Credentials{}, &user)
		err := c.JSON(user)
		log.Print(user.Credentials)
		log.Print(user.Sessions)
		return err
	})
	auth.Get("/keys/:username", func(c *fiber.Ctx) error {
		credentials := []db.Credentials{}
		db.Users.Where(&db.Credentials{UserUsername: c.Params("username")}).Find(&credentials)
		return c.JSON(credentials)
	})
	auth.Get("/status", func(c *fiber.Ctx) error {
		s := c.Request().Header.Cookie("session_id")
		if len(s) == 0 {
			return c.SendStatus(204)
		}

		dbsess, err := db.LoginSession.Get(c)
		if err != nil {
			return c.SendStatus(204)
		}

		if dbsess.ID() != string(s) {
			return c.SendStatus(412)
		}

		username := dbsess.Get("username")

		return c.SendString(fmt.Sprint(username))
	})
	auth.Get("/logout", func(c *fiber.Ctx) error {
		sess, err := db.LoginSession.Get(c)
		if err != nil {
			return err
		}
		err = sess.Destroy()
		if err != nil {
			return err
		}
		c.Response().Header.Set("HX-Redirect", "/login")
		return c.Render("components/loginCard", nil)
	})

}

func BeginLogin(c *fiber.Ctx) error {
	user := db.User{Username: c.Params("username")}
	log.Info().Msgf("%v", db.Users)
	db.Users.Preload("Credentials").First(&user)

	log.Debug().Any("User Logging in: ", user).Msg("")

	options, session, err := web.BeginLogin(user)
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	dSession := db.Sessions{
		Challenge:                   session.Challenge,
		Expires:                     session.Expires,
		UserVerificationRequirement: session.UserVerification,
		UserID:                      session.UserID,
	}

	log.Debug().Any("Session to be saved in database", dSession).Msg("")
	db.Users.Where("user_id = ?", user.ID).Delete(&db.Sessions{})
	log.Debug().Interface("Removing Old sessions for UserID: %s", user)

	db.Users.Save(&dSession)
	db.Users.Save(&user)
	return c.JSON(options)
}

func FinishLogin(c *fiber.Ctx) error {
	body := new(protocol.CredentialAssertionResponse)
	if err := c.BodyParser(body); err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	response, err := body.Parse()
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}

	user := db.User{Username: c.Params("username")}
	db.Users.First(&user)
	session := db.Sessions{UserID: user.ID}
	db.Users.First(&session)
	credentials := []db.Credentials{}
	db.Users.Where(&db.Credentials{UserID: user.ID}).Find(&credentials)
	user.Credentials = credentials
	wSession := webauthn.SessionData{
		Challenge:            session.Challenge,
		UserID:               session.UserID,
		AllowedCredentialIDs: [][]byte{},
		Expires:              session.Expires,
		UserVerification:     session.UserVerificationRequirement,
		Extensions:           map[string]interface{}{},
	}

	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	_, err = web.ValidateLogin(user, wSession, response)
	if err != nil {
		log.Error().Msgf("Failed to validate login: %s ", err)
		return c.JSON(struct{ Status string }{Status: "Failed"})
	}
	sess, err := db.LoginSession.Get(c)
	if err != nil {
		return err
	}
	sess.Set("username", c.Params("username"))
	err = sess.Save()
	if err != nil {
		return err
	}
	log.Debug().Msgf("Session was created: %s", sess.ID())

	err = c.SendString("Login was Successful")
	if err != nil {
		return err
	}

	log.Debug().Msgf("Login Was Successful for User: %s", user.Username)
	return nil

}

func BeginRegistration(c *fiber.Ctx) error {
	existingUser := db.Users.Take(&db.User{Username: c.Params("username")})
	if existingUser.RowsAffected > 0 {
		return c.Status(409).SendString("Username already exists")
	}
	id := make([]byte, 32)
	_, err := rand.Read(id)
	if err != nil {
		return err
	}
	user := db.User{Username: c.Params("username"), ID: id} // Find or create the new user
	db.Users.Create(&user)
	options, sessionData, err := web.BeginRegistration(user)
	if err != nil {
		log.Fatal().Err(err)
	}
	// handle errors if present
	// store the sessionData value

	s := db.Sessions{
		Challenge:                   sessionData.Challenge,
		Expires:                     sessionData.Expires,
		UserVerificationRequirement: sessionData.UserVerification,
		UserUsername:                user.Username,
		UserID:                      user.ID,
	}
	db.Users.Delete(&db.Sessions{UserID: user.ID})
	db.Users.Create(&s)
	log.Printf("user before save: %v session: %v", user, sessionData)
	db.Users.Save(&user)
	var newUser db.User
	db.Users.Find(&newUser)
	log.Printf("database user %v", newUser)
	return c.JSON(options)
}

func FinishRegistration(c *fiber.Ctx, username string) {
	body := new(protocol.CredentialCreationResponse)
	if err := c.BodyParser(body); err != nil {
		return
	}
	response, err := body.Parse()

	if err != nil {
		// Handle Error and return.
		log.Error().Msgf("Failed to register: %v", err)
		c.Status(500).SendString("Failed to Register")
		return
	}

	user := db.User{Username: c.Params("username")} // Get the user
	db.Users.First(&user)
	log.Print(user)
	// Get the session data stored from the function above
	session := db.Sessions{UserID: user.ID}
	db.Users.First(&session)

	wSession := webauthn.SessionData{
		Challenge:            session.Challenge,
		UserID:               session.UserID,
		AllowedCredentialIDs: [][]byte{},
		Expires:              session.Expires,
		UserVerification:     session.UserVerificationRequirement,
		Extensions:           map[string]interface{}{},
	}

	credential, err := web.CreateCredential(user, wSession, response)
	if err != nil {
		log.Print("credential error: " + err.Error())

		return
	}

	// If creation was successful, store the credential object
	var transport string
	for t := range credential.Transport {
		transport = transport + "," + string(rune(t))
	}

	err = c.JSON("Registration Success")
	dCredential := db.Credentials{
		ID:              credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Transport:       transport,
		Flags:           credential.Flags,
		Authentication:  credential.Authenticator,
		UserUsername:    user.Username,
		UserID:          user.ID,
	}

	db.Users.Save(&dCredential)

	// Pseudocode to add the user credential.
	log.Print(user)
	db.Users.Save(&user)
}

func JSONResponse(w http.ResponseWriter, d interface{}, c int) {
	dj, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	fmt.Fprintf(w, "%s", dj)
}

func CheckLoginStatus(c *fiber.Ctx) (string, error) {
	s := c.Request().Header.Cookie("session_id")
	if len(s) == 0 {
		return "", errors.New("Session Cookie was not present")
	}

	dbsess, err := db.LoginSession.Get(c)
	if err != nil {
		return "", errors.New("Session Was not found in db")
	}

	if dbsess.ID() != string(s) {
		return "", errors.New("Session did not match the one in db")
	}

	username := dbsess.Get("username").(string)
	return username, nil
}

func NewLoginRedirect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		s := c.Request().Header.Cookie("session_id")
		if len(s) == 0 {
			return c.Redirect("/login", 302)
		}
		dbsess, err := db.LoginSession.Get(c)
		if err != nil {
			return c.Redirect("/login", 302)
		}

		if dbsess.ID() != string(s) {
			log.Debug().Msgf("Session Cookie did not match: 1: %v 2: %v", dbsess.ID(), string(s))
			return c.Redirect("/login", 302)
		}

		log.Debug().Msgf("Session Cookie matched, successfully authenticated: 1: %v 2: %v",
			dbsess.ID(), string(s))

		return c.Next()
	}
}
