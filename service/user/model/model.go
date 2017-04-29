package usermodel

import (
	"database/sql"
	"fmt"
	"github.com/hackform/governor/util/hash"
	"github.com/hackform/governor/util/rank"
	"github.com/hackform/governor/util/uid"
	"time"
)

const (
	uidTimeSize = 8
	uidRandSize = 8
	tableName   = "users"
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
		Tags string `json:"auth_tags"`
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
		CreationTime int64  `json:"creation_time"`
	}
)

// New creates a new User Model
func New(username, password, email, firstname, lastname string, r rank.Rank) (*Model, error) {
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
			Tags: r.Stringify(),
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
			CreationTime: time.Now().Unix(),
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
	_, err := db.Exec(fmt.Sprintf("INSERT INTO %s (userid, username, auth_tags, pass_hash, pass_salt, pass_version, email, first_name, last_name, creation_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);", tableName), m.Userid, m.Username, m.Auth.Tags, m.Passhash.Hash, m.Passhash.Salt, m.Passhash.Version, m.Email, m.FirstName, m.LastName, m.CreationTime)
	return err
}

// Update updates the model in the db
func (m *Model) Update(db *sql.DB) error {
	_, err := db.Exec(fmt.Sprintf("UPDATE %s SET (userid, username, auth_tags, pass_hash, pass_salt, pass_version, email, first_name, last_name, creation_time) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) WHERE userid = $1;", tableName), m.Userid, m.Username, m.Auth.Tags, m.Passhash.Hash, m.Passhash.Salt, m.Passhash.Version, m.Email, m.FirstName, m.LastName, m.CreationTime)
	return err
}

// Setup creates a new User table
func Setup(db *sql.DB) error {
	_, err := db.Exec(fmt.Sprintf("CREATE TABLE %s (userid BYTEA PRIMARY KEY, username VARCHAR(255) NOT_NULL, auth_tags TEXT NOT_NULL, pass_hash BYTEA NOT_NULL, pass_salt BYTEA NOT_NULL, pass_version INT NOT_NULL, email VARCHAR(255) NOT_NULL, first_name VARCHAR(255) NOT_NULL, last_name VARCHAR(255) NOT_NULL, creation_time BIGINT NOT_NULL);", tableName))
	return err
}
