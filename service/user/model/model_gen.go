// Code generated by go generate. DO NOT EDIT.
package usermodel

import (
	"database/sql"
	"github.com/lib/pq"
	"strconv"
	"strings"
)

const (
	userModelTableName = "users"
)

func userModelSetup(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users (userid VARCHAR(31) PRIMARY KEY, username VARCHAR(255) NOT NULL UNIQUE, pass_hash VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL UNIQUE, first_name VARCHAR(255) NOT NULL, last_name VARCHAR(255) NOT NULL, creation_time BIGINT NOT NULL);")
	return err
}

func userModelGet(db *sql.DB, key string) (*Model, int, error) {
	m := &Model{}
	if err := db.QueryRow("SELECT userid, username, pass_hash, email, first_name, last_name, creation_time FROM users WHERE userid = $1;", key).Scan(&m.Userid, &m.Username, &m.PassHash, &m.Email, &m.FirstName, &m.LastName, &m.CreationTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, 2, err
		}
		if postgresErr, ok := err.(*pq.Error); ok {
			switch postgresErr.Code {
			case "42P01": // undefined_table
				return nil, 4, err
			default:
				return nil, 0, err
			}
		}
		return nil, 0, err
	}
	return m, 0, nil
}

func userModelInsert(db *sql.DB, m *Model) (int, error) {
	_, err := db.Exec("INSERT INTO users (userid, username, pass_hash, email, first_name, last_name, creation_time) VALUES ($1, $2, $3, $4, $5, $6, $7);", m.Userid, m.Username, m.PassHash, m.Email, m.FirstName, m.LastName, m.CreationTime)
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

func userModelUpdate(db *sql.DB, m *Model) error {
	_, err := db.Exec("UPDATE users SET (userid, username, pass_hash, email, first_name, last_name, creation_time) = ($1, $2, $3, $4, $5, $6, $7) WHERE userid = $1;", m.Userid, m.Username, m.PassHash, m.Email, m.FirstName, m.LastName, m.CreationTime)
	return err
}

func userModelDelete(db *sql.DB, m *Model) error {
	_, err := db.Exec("DELETE FROM users WHERE userid = $1;", m.Userid)
	return err
}

func userModelGetModelByUsername(db *sql.DB, key string) (*Model, int, error) {
	m := &Model{}
	if err := db.QueryRow("SELECT userid, username, pass_hash, email, first_name, last_name, creation_time FROM users WHERE username = $1;", key).Scan(&m.Userid, &m.Username, &m.PassHash, &m.Email, &m.FirstName, &m.LastName, &m.CreationTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, 2, err
		}
		if postgresErr, ok := err.(*pq.Error); ok {
			switch postgresErr.Code {
			case "42P01": // undefined_table
				return nil, 4, err
			default:
				return nil, 0, err
			}
		}
		return nil, 0, err
	}
	return m, 0, nil
}

func userModelGetModelByEmail(db *sql.DB, key string) (*Model, int, error) {
	m := &Model{}
	if err := db.QueryRow("SELECT userid, username, pass_hash, email, first_name, last_name, creation_time FROM users WHERE email = $1;", key).Scan(&m.Userid, &m.Username, &m.PassHash, &m.Email, &m.FirstName, &m.LastName, &m.CreationTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, 2, err
		}
		if postgresErr, ok := err.(*pq.Error); ok {
			switch postgresErr.Code {
			case "42P01": // undefined_table
				return nil, 4, err
			default:
				return nil, 0, err
			}
		}
		return nil, 0, err
	}
	return m, 0, nil
}

func userModelGetInfoOrdUserid(db *sql.DB, orderasc bool, limit, offset int) ([]Info, error) {
	order := "DESC"
	if orderasc {
		order = "ASC"
	}
	res := make([]Info, 0, limit)
	rows, err := db.Query("SELECT userid, username, email, first_name, last_name FROM users ORDER BY userid "+order+" LIMIT $1 OFFSET $2;", limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := Info{}
		if err := rows.Scan(&m.Userid, &m.Username, &m.Email, &m.FirstName, &m.LastName); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func userModelGetInfoSetUserid(db *sql.DB, keys []string) ([]Info, error) {
	placeholderStart := 1
	placeholders := make([]string, 0, len(keys))
	for i := range keys {
		placeholders = append(placeholders, "($"+strconv.Itoa(i+placeholderStart)+")")
	}

	args := make([]interface{}, 0, len(keys))
	for _, i := range keys {
		args = append(args, i)
	}

	stmt := "SELECT userid, username, email, first_name, last_name FROM users WHERE userid IN (VALUES " + strings.Join(placeholders, ",") + ");"

	res := make([]Info, 0, len(keys))
	rows, err := db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := Info{}
		if err := rows.Scan(&m.Userid, &m.Username, &m.Email, &m.FirstName, &m.LastName); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
