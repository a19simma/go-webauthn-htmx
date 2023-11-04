package pkg

import (
	"crypto/rand"
	"errors"
	"time"

	"github.com/a19simma/go-webauthn-htmx/pkg/db"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	webAuthn                  *webauthn.WebAuthn
	userDb                    db.UserDb
	ErrAlreadyExists          = errors.New("already exists")
	ErrRegistrationNotAllowed = errors.New("registration not allowed")
	ErrLoginBlocked           = errors.New("login has been blocked for this user")
)

type Auth interface {
	BeginLogin(string) (*protocol.CredentialAssertion, error)
	FinishLogin(string, protocol.ParsedCredentialAssertionData) error
	BeginRegistration(string) (*protocol.CredentialCreation, error)
	FinishRegistration(protocol.ParsedCredentialCreationData,
		string) error
}

type AuthImpl struct{}

func (authimpl AuthImpl) FinishLogin(username string, data protocol.ParsedCredentialAssertionData) error {
	user, err := userDb.GetUser(username)
	if err != nil {
		return err
	}
	session := userDb.GetUserSession(*user)
	user.Credentials = userDb.GetUserCredentials(*user)
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
	_, err = webAuthn.ValidateLogin(user, wSession, &data)
	if err != nil {
		log.Err(err)
		return err
	}

	return nil
}

func (authimpl AuthImpl) BeginLogin(username string) (*protocol.CredentialAssertion, error) {
	user, err := userDb.GetUser(username)
	if err != nil {
		return nil, err
	}
	if user.Status == db.Blocked {
		return nil, ErrLoginBlocked
	}
	log.Printf("User Logging in: %v", user)

	options, session, err := webAuthn.BeginLogin(user)
	if err != nil {
		log.Err(err)
		return nil, err
	}
	dSession := db.Sessions{
		Challenge:                   session.Challenge,
		Expires:                     time.Now().Add(time.Minute * 5),
		UserVerificationRequirement: session.UserVerification,
		UserID:                      session.UserID,
	}

	err = userDb.DeleteSessions(*user)
	if err != nil {
		return nil, err
	}
	err = userDb.CreateSession(dSession)
	if err != nil {
		return nil, err
	}
	err = userDb.CreateUser(*user)
	if err != nil {
		return nil, err
	}
	return options, nil
}

func InitAuth(uDb db.UserDb) (Auth, error) {
	origins := []string{"http://localhost:4200"}
	o := viper.GetString("AUTH_ORIGIN")
	if len(o) != 0 {
		origins = append(origins, o)
	}
	wconfig := &webauthn.Config{
		RPDisplayName: "Go Webauthn", // Display Name for your site
		RPID:          "localhost",   // Generally the FQDN for your site
		RPOrigins:     origins,       // The origin URLs allowed for WebAuthn requests
	}

	log.Info().Msgf("allowed origins: %v", origins)

	web, err := webauthn.New(wconfig)
	if err != nil {
		log.Err(err)
		return nil, err
	}
	userDb = uDb
	webAuthn = web

	log.Printf("Initialized Webauthn with config: %v", web)

	return AuthImpl{}, nil
}

func (authImpl AuthImpl) BeginRegistration(username string) (*protocol.CredentialCreation, error) {
	log.Printf("uDb: %v", userDb)
	user, err := userDb.GetUser(username)
	if err != nil && !errors.Is(err, db.ErrNoResults) {
		log.Err(err)
		return nil, err
	}
	if user == nil {
		user = &db.User{Username: username}
		log.Print(user)
	}

	if len(user.ID) == 0 {
		id := make([]byte, 32)
		_, err = rand.Read(id)
		if err != nil {
			return nil, err
		}
		user.ID = id
	}

	if user.Status != db.Open && user.Username != viper.Get("AUTH_ADMIN_EMAIl") {
		return nil, ErrRegistrationNotAllowed
	}

	err = userDb.CreateUser(*user)
	if err != nil {
		return nil, err
	}
	options, sessionData, err := webAuthn.BeginRegistration(user)
	if err != nil {
		log.Err(err)
	}
	s := db.Sessions{
		Challenge:                   sessionData.Challenge,
		Expires:                     sessionData.Expires,
		UserVerificationRequirement: sessionData.UserVerification,
		UserUsername:                user.Username,
		UserID:                      user.ID,
	}

	err = userDb.CreateSession(s)
	if err != nil {
		return nil, err
	}

	return options, nil
}

func (authImpl AuthImpl) FinishRegistration(resp protocol.ParsedCredentialCreationData,
	username string) error {
	user, err := userDb.GetUser(username)
	if err != nil {
		return err
	}
	log.Print(user)

	session := userDb.GetUserSession(*user)

	user.Sessions = append(user.Sessions, session)

	wSession := webauthn.SessionData{
		Challenge:            session.Challenge,
		UserID:               session.UserID,
		AllowedCredentialIDs: [][]byte{},
		Expires:              session.Expires,
		UserVerification:     session.UserVerificationRequirement,
		Extensions:           map[string]interface{}{},
	}

	credential, err := webAuthn.CreateCredential(user, wSession, &resp)
	if err != nil {
		log.Print("credential error: " + err.Error())
		return err
	}

	// If creation was successful, store the credential object
	var transport string
	for t := range credential.Transport {
		transport = transport + "," + string(rune(t))
	}

	if err != nil {
		log.Err(err)
	}
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

	err = userDb.CreateCredentials(dCredential)
	if err != nil {
		return err
	}

	log.Print(user)
	if user.Username == viper.Get("AUTH_ADMIN_EMAIL") {
		user.Role = db.Admin
	} else {
		user.Role = db.Member
	}
	user.Status = db.Registered
	err = userDb.CreateUser(*user)
	if err != nil {
		return err
	}
	return nil
}
