package usermodel

import (
	"database/sql"
	"github.com/hackform/governor/util/hash"
	"github.com/hackform/governor/util/rank"
	"github.com/hackform/governor/util/uid"
	"time"
)

const (
	uidTimeSize = 8
	uidRandSize = 8
)

type (
	// Model is the db User model
	Model struct {
		ID
		Auth
		Passhash
		Props
	}

	// ID is user identification
	ID struct {
		Userid   []byte `json:"userid"`
		Username string `json:"username"`
	}

	// Auth manages user permissions
	Auth struct {
		Level uint32   `json:"auth_level"`
		Tags  []string `json:"auth_tags"`
	}

	// Passhash controls the user password
	Passhash struct {
		Hash    []byte `json:"pass_hash"`
		Salt    []byte `json:"pass_salt"`
		Version int    `json:"pass_version"`
	}

	// Props stores user info
	Props struct {
		Email        string `json:"email"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		CreationDate int64  `json:"creation_date"`
	}
)

// New creates a new User Model
func New(username, password, email, firstname, lastname string, level uint32) (*Model, error) {
	mUID, err := uid.NewU(uidTimeSize, uidRandSize)
	if err != nil {
		return nil, err
	}

	mHash, mSalt, mVersion, err := hash.Hash(password, hash.Latest)
	if err != nil {
		return nil, err
	}

	return &Model{
		ID: ID{
			Userid:   mUID.Bytes(),
			Username: username,
		},
		Auth: Auth{
			Level: level,
			Tags:  []string{},
		},
		Passhash: Passhash{
			Hash:    mHash,
			Salt:    mSalt,
			Version: mVersion,
		},
		Props: Props{
			Email:        email,
			FirstName:    firstname,
			LastName:     lastname,
			CreationDate: time.Now().Unix(),
		},
	}, nil
}

// NewBaseUser creates a new Base User Model
func NewBaseUser(username, password, email, firstname, lastname string) (*Model, error) {
	return New(username, password, email, firstname, lastname, rank.BaseUser())
}

// NewAdmin creates a new Admin User Model
func NewAdmin(username, password, email, firstname, lastname string) (*Model, error) {
	return New(username, password, email, firstname, lastname, rank.Admin())
}

// IDBase64 returns the userid as a base64 encoded string
func (m *Model) IDBase64() (string, error) {
	u, err := uid.FromBytes(uidTimeSize, 0, uidRandSize, m.Userid)
	if err != nil {
		return "", err
	}
	return u.Base64(), nil
}

// Insert inserts the model into the db
func (m *Model) Insert(db *sql.DB) error {
	return nil
}

// Update updates the model in the db
func (m *Model) Update(db *sql.DB) error {
	return nil
}

// Setup creates a new User table
func Setup(db *sql.DB) error {
	return nil
}
