package api

import (
	"errors"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	auth "github.com/a19simma/vanilla-js/pkg"
	"github.com/a19simma/vanilla-js/pkg/db"
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
		_, err = db.GetLoginSession(c)
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

		_, err = db.GetLoginSession(c)
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

		dbsess, err := db.GetLoginSession(c)
		if err != nil {
			return c.SendStatus(204)
		}

		if dbsess.ID() != string(s) {
			return c.SendStatus(412)
		}

		username := dbsess.Get("username")

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

// func FinishLogin(c *fiber.Ctx) error {
// 	user := db.User{Username: c.Params("username")}
// 	db.Users.Where(user).First(&user)
// 	log.Printf(user.Username)
// 	session := db.Sessions{UserID: user.ID}
// 	db.Users.Where("user_id = ?", user.ID).First(&session)
// 	log.Printf(session.UserUsername)
// 	credentials := []db.Credentials{}
// 	db.Users.Where(&db.Credentials{UserID: user.ID}).Find(&credentials)
// 	user.Credentials = credentials
// 	wSession := webauthn.SessionData{
// 		Challenge:            session.Challenge,
// 		UserID:               session.UserID,
// 		AllowedCredentialIDs: [][]byte{},
// 		Expires:              session.Expires,
// 		UserVerification:     session.UserVerificationRequirement,
// 		Extensions:           map[string]interface{}{},
// 	}
//
// 	if err != nil {
// 		log.Error().Err(err).Msg("")
// 		return err
// 	}
// 	_, err = web.ValidateLogin(user, wSession, response)
// 	if err != nil {
// 		log.Error().Msgf("Failed to validate login: %s ", err)
// 		return c.Status(401).SendString("Failed to login, incorrect credentials")
// 	}
// 	sess, err := db.GetLoginSession(c)
// 	if err != nil {
// 		return err
// 	}
// 	sess.Set("username", c.Params("username"))
// 	err = sess.Save()
// 	if err != nil {
// 		return err
// 	}
// 	log.Debug().Msgf("Session was created: %s", sess.ID())
//
// 	err = c.SendString("Login was Successful")
// 	if err != nil {
// 		return err
// 	}
//
// 	log.Debug().Msgf("Login Was Successful for User: %s", user.Username)
// 	return nil
//
// }

// func BeginRegistration(c *fiber.Ctx) error {
// 	user := db.User{Username: c.Params("username")}
// 	existingUser := db.Users.Where(user).First(&user)
// 	if existingUser.RowsAffected > 0 {
// 		return c.Status(409).SendString("Username already exists")
// 	}
// 	id := make([]byte, 32)
// 	_, err := rand.Read(id)
// 	if err != nil {
// 		return err
// 	}
// 	if user.Status != db.Open && user.Username != viper.Get("AUTH_ADMIN_EMAIl") {
// 		return c.Status(401).SendString("Registration not allowed with this email")
// 	}
// 	user.ID = id // Find or create the new user
// 	db.Users.Create(&user)
// 	options, sessionData, err := web.BeginRegistration(user)
// 	if err != nil {
// 		log.Fatal().Err(err)
// 	}
// 	// handle errors if present
// 	// store the sessionData value
//
// 	s := db.Sessions{
// 		Challenge:                   sessionData.Challenge,
// 		Expires:                     sessionData.Expires,
// 		UserVerificationRequirement: sessionData.UserVerification,
// 		UserUsername:                user.Username,
// 		UserID:                      user.ID,
// 	}
// 	db.Users.Delete(&db.Sessions{UserID: user.ID})
// 	db.Users.Create(&s)
// 	log.Printf("user before save: %v session: %v", user, sessionData)
// 	db.Users.Save(&user)
// 	var newUser db.User
// 	db.Users.Find(&newUser)
// 	log.Printf("database user %v", newUser)
// 	return c.JSON(options)
// }

// func FinishRegistration(c *fiber.Ctx, username string) {
// 	body := new(protocol.CredentialCreationResponse)
// 	if err := c.BodyParser(body); err != nil {
// 		return
// 	}
// 	response, err := body.Parse()
//
// 	if err != nil {
// 		// Handle Error and return.
// 		log.Error().Msgf("Failed to register: %v", err)
// 		c.Status(500).SendString("Failed to Register")
// 		return
// 	}
//
// 	user := db.User{Username: c.Params("username")} // Get the user
// 	db.Users.First(&user)
// 	log.Print(user)
// 	// Get the session data stored from the function above
// 	session := db.Sessions{UserID: user.ID}
// 	db.Users.First(&session)
//
// 	wSession := webauthn.SessionData{
// 		Challenge:            session.Challenge,
// 		UserID:               session.UserID,
// 		AllowedCredentialIDs: [][]byte{},
// 		Expires:              session.Expires,
// 		UserVerification:     session.UserVerificationRequirement,
// 		Extensions:           map[string]interface{}{},
// 	}
//
// 	credential, err := web.CreateCredential(user, wSession, response)
// 	if err != nil {
// 		log.Print("credential error: " + err.Error())
//
// 		return
// 	}
//
// 	// If creation was successful, store the credential object
// 	var transport string
// 	for t := range credential.Transport {
// 		transport = transport + "," + string(rune(t))
// 	}
//
// 	err = c.JSON("Registration Success")
// 	if err != nil {
// 		log.Err(err)
// 	}
// 	dCredential := db.Credentials{
// 		ID:              credential.ID,
// 		PublicKey:       credential.PublicKey,
// 		AttestationType: credential.AttestationType,
// 		Transport:       transport,
// 		Flags:           credential.Flags,
// 		Authentication:  credential.Authenticator,
// 		UserUsername:    user.Username,
// 		UserID:          user.ID,
// 	}
//
// 	db.Users.Save(&dCredential)
//
// 	log.Print(user)
// 	if user.Username == viper.Get("AUTH_ADMIN_EMAIL") {
// 		user.Role = db.Admin
// 	} else {
// 		user.Role = db.Member
// 	}
// 	user.Status = db.Registered
// 	db.Users.Save(&user)
// }

func CheckLoginStatus(c *fiber.Ctx) (string, error) {
	s := c.Request().Header.Cookie("session_id")
	if len(s) == 0 {
		return "", errors.New("Session Cookie was not present")
	}

	dbsess, err := db.GetLoginSession(c)
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
		err := db.ValidateLoginSession(c)
		if err != nil {
			log.Printf("failed to validate login: %v", err)
			return c.Redirect("/login", 302)
		}
		return c.Next()
	}
}
