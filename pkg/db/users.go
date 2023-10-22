package db

import (
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/rs/zerolog/log"
)

var Users *gorm.DB

func InitUsers() {
	users, err := gorm.Open(sqlite.Open("users.db"), &gorm.Config{})
	if err != nil {
		log.Fatal().Msg("Failed to open connection to db")
	}
	Users = users

	err = Users.AutoMigrate(&Sessions{}, &Credentials{}, &User{}, &Registration{})
	if err != nil {
		log.Fatal().Msgf("Failed to migrate schema: %v", err.Error())
	}
}

func (user User) WebAuthnID() []byte {
	return user.ID
}

func (user User) WebAuthnName() string {
	return "newUser"
}

func (user User) WebAuthnDisplayName() string {
	return user.Username
}

func (user User) WebAuthnIcon() string {
	return ""
}

type Registration struct {
	ID       []byte `json:"id"`
	Username string
}

func (user User) WebAuthnCredentials() []webauthn.Credential {
	c := []webauthn.Credential{}
	for i := range user.Credentials {
		cred := user.Credentials[i]
		transportStrings := strings.Split(cred.Transport, ",")
		var transport []protocol.AuthenticatorTransport
		for i := range transportStrings {
			transport = append(transport, protocol.AuthenticatorTransport(transportStrings[i]))
		}
		c = append(c, webauthn.Credential{
			ID:              cred.ID,
			PublicKey:       cred.PublicKey,
			AttestationType: cred.AttestationType,
			Transport:       transport,
			Flags:           cred.Flags,
			Authenticator:   cred.Authentication,
		})
	}
	return c
}

type RegistrationStatus int

const (
	Registered RegistrationStatus = iota
	Open
	Blocked
)

func (s RegistrationStatus) String() string {
	return []string{"Registered", "Open", "Blocked"}[s]
}

type Role int

const (
	Admin Role = iota
	Member
)

func (r Role) String() string {
	return []string{"Admin", "Member"}[r]
}

type User struct {
	ID          []byte `json:"id" gorm:"primarykey"`
	Username    string `json:"username"`
	Sessions    []Sessions
	Credentials []Credentials
	Role        Role               `gorm:"type:integer"`
	Status      RegistrationStatus `gorm:"type:integer"`
}

type Credentials struct {
	ID              []byte
	PublicKey       []byte
	AttestationType string
	Transport       string
	Flags           webauthn.CredentialFlags `gorm:"embedded"`
	Authentication  webauthn.Authenticator   `gorm:"embedded"`
	UserID          []byte
	UserUsername    string
}

type Sessions struct {
	Challenge                   string `gorm:"primarykey"`
	UserDisplayName             string
	Expires                     time.Time
	UserVerificationRequirement protocol.UserVerificationRequirement `gorm:"embedded"`
	UserUsername                string
	UserID                      []byte `gorm:"primarykey"`
}
