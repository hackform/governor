// Code generated by go generate. DO NOT EDIT.
package sessionmodel

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"strings"
)

const (
	sessionModelTableName = "usersessions"
)

func sessionModelSetup(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS usersessions (sessionid VARCHAR(31) PRIMARY KEY, userid VARCHAR(31) NOT NULL, keyhash VARCHAR(127) NOT NULL);")
	return err
}

func sessionModelGet(db *sql.DB, key string) (*Model, int, error) {
	m := &Model{}
	if err := db.QueryRow("SELECT sessionid, userid, keyhash FROM usersessions WHERE sessionid = $1;", key).Scan(&m.SessionID, &m.Userid, &m.KeyHash); err != nil {
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

func sessionModelInsert(db *sql.DB, m *Model) (int, error) {
	_, err := db.Exec("INSERT INTO usersessions (sessionid, userid, keyhash) VALUES ($1, $2, $3);", m.SessionID, m.Userid, m.KeyHash)
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

func sessionModelInsertBulk(db *sql.DB, models []*Model, allowConflict bool) (int, error) {
	conflictSQL := ""
	if allowConflict {
		conflictSQL = " ON CONFLICT DO NOTHING"
	}
	placeholders := make([]string, 0, len(models))
	args := make([]interface{}, 0, len(models)*3)
	for c, m := range models {
		n := c * 3
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", n+1, n+2, n+3))
		args = append(args, m.SessionID, m.Userid, m.KeyHash)
	}
	_, err := db.Exec("INSERT INTO usersessions (sessionid, userid, keyhash) VALUES "+strings.Join(placeholders, ", ")+conflictSQL+";", args...)
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

func sessionModelUpdate(db *sql.DB, m *Model) error {
	_, err := db.Exec("UPDATE usersessions SET (sessionid, userid, keyhash) = ($1, $2, $3) WHERE sessionid = $1;", m.SessionID, m.Userid, m.KeyHash)
	return err
}

func sessionModelDelete(db *sql.DB, m *Model) error {
	_, err := db.Exec("DELETE FROM usersessions WHERE sessionid = $1;", m.SessionID)
	return err
}
