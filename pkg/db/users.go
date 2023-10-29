package db

import (
	"errors"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/rs/zerolog/log"
)

var (
	ErrNoResults = errors.New("no results found")
	db           *gorm.DB
)

type UserDb interface {
	GetUser(string) (*User, error)
	GetUsers() []User
	DeleteUser(string) error
	CreateUser(User) error
	CreateSession(Sessions) error
	DeleteSessions(User) error
	GetUserSession(User) Sessions
	GetUserCredentials(User) []Credentials
	CreateCredentials(Credentials) error
}

type UserDbImpl struct{}

func (userdbimpl UserDbImpl) CreateCredentials(c Credentials) error {
	db.Save(c)
	return nil
}

func (userdbimpl UserDbImpl) GetUserSession(user User) Sessions {
	session := Sessions{}
	db.Where("user_id = ?", user.ID).First(&session)
	return session
}

func (userdbimpl UserDbImpl) GetUserCredentials(user User) []Credentials {
	credentials := []Credentials{}
	db.Where("user_id = ?", user.ID).Find(&credentials)
	return credentials
}

func (userdbimpl UserDbImpl) GetUsers() []User {
	users := []User{}
	db.Find(&users).Select("ID", "Username", "Role", "Status")
	return users
}

func (userdbimpl UserDbImpl) DeleteSessions(user User) error {
	db.Where("user_id = ?", user.ID).Delete(&Sessions{})
	return nil
}
func (userdbimpl UserDbImpl) GetUserSessions(user User) (Sessions, error) {
	session := &Sessions{}
	db.Where("user_id = ?", user.ID).First(&session)
	return *session, nil
}

func (userdbimpl UserDbImpl) CreateUser(user User) error {
	log.Printf("saving user: %v", user)
	db.Save(user)
	return nil
}
func (userdbimpl UserDbImpl) CreateSession(session Sessions) error {
	db.Save(session)
	return nil
}

func InitUsers() UserDbImpl {
	usersDb, err := gorm.Open(sqlite.Open("users.db"), &gorm.Config{})
	if err != nil {
		log.Fatal().Msg("Failed to open connection to db")
	}
	db = usersDb

	err = db.AutoMigrate(&Sessions{}, &Credentials{}, &User{}, &Registration{})
	if err != nil {
		log.Fatal().Msgf("Failed to migrate schema: %v", err.Error())
	}
	return UserDbImpl{}
}

func (d UserDbImpl) GetUser(username string) (*User, error) {
	log.Print("Hello from getUser")
	user := User{Username: username}
	result := db.Where(user).Preload("Credentials").First(&user)
	if result.RowsAffected == 0 {
		log.Printf("No results found for user %s", username)
		return nil, ErrNoResults
	}

	return &user, nil
}

func (d UserDbImpl) DeleteUser(id string) error {
	userId := []byte(id)
	log.Debug().Msgf("Deleting user with id: %s", id)
	result := db.Where("user_id = ?", userId).Delete(Credentials{})
	if result.Error != nil {
		log.Err(result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		log.Debug().Str("UserId", id).Msg("Found no credentials to delete to delete")
	}

	result = db.Where("id = ?", userId).Delete(User{})
	if result.Error != nil {
		log.Err(result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNoResults
	}

	return nil
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
