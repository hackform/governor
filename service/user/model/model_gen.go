// Code generated by go generate. DO NOT EDIT.
package usermodel

import (
	"database/sql"
	"github.com/lib/pq"
)

const (
	modelTableName = "users"
)

const (
	modelSQLSetup = "CREATE TABLE users (userid BYTEA PRIMARY KEY, username VARCHAR(255) NOT NULL UNIQUE, pass_hash BYTEA NOT NULL, email VARCHAR(4096) NOT NULL UNIQUE, first_name VARCHAR(255) NOT NULL, last_name VARCHAR(255) NOT NULL, creation_time BIGINT NOT NULL);"
)

func modelSetup(db *sql.DB) error {
	_, err := db.Exec(modelSQLSetup)
	return err
}

const (
	modelSQLInsert = "INSERT INTO users (userid, username, pass_hash, email, first_name, last_name, creation_time) VALUES ($1, $2, $3, $4, $5, $6, $7);"
)

func modelInsert(db *sql.DB, m *Model) (int, error) {
	_, err := db.Exec(modelSQLInsert, m.Userid, m.Username, m.PassHash, m.Email, m.FirstName, m.LastName, m.CreationTime)
	if err != nil {
		if postgresErr, ok := err.(*pq.Error); ok {
			switch postgresErr.Code {
			case "23505": // unique_violation
				return 3, err
			default:
				return 0, err
			}
		}
	}
	return 0, nil
}

const (
	modelSQLUpdate = "UPDATE users SET (userid, username, pass_hash, email, first_name, last_name, creation_time) = ($1, $2, $3, $4, $5, $6, $7) WHERE userid = $1;"
)

func modelUpdate(db *sql.DB, m *Model) error {
	_, err := db.Exec(modelSQLUpdate, m.Userid, m.Username, m.PassHash, m.Email, m.FirstName, m.LastName, m.CreationTime)
	return err
}

const (
	modelSQLDelete = "DELETE FROM users WHERE userid = $1;"
)

func modelDelete(db *sql.DB, m *Model) error {
	_, err := db.Exec(modelSQLDelete, m.Userid)
	return err
}
