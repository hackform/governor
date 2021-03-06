// Code generated by go generate forge model v0.3; DO NOT EDIT.

package model

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

const (
	sessionModelTableName = "usersessions"
)

func sessionModelSetup(db *sql.DB) (int, error) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS usersessions (sessionid VARCHAR(63) PRIMARY KEY, userid VARCHAR(31) NOT NULL, keyhash VARCHAR(127) NOT NULL, time BIGINT NOT NULL, auth_time BIGINT NOT NULL, ipaddr VARCHAR(63), user_agent VARCHAR(1023));")
	if err != nil {
		return 0, err
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS usersessions_userid_index ON usersessions (userid);")
	if err != nil {
		if postgresErr, ok := err.(*pq.Error); ok {
			switch postgresErr.Code {
			case "42501": // insufficient_privilege
				return 5, err
			default:
				return 0, err
			}
		}
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS usersessions_time_index ON usersessions (time);")
	if err != nil {
		if postgresErr, ok := err.(*pq.Error); ok {
			switch postgresErr.Code {
			case "42501": // insufficient_privilege
				return 5, err
			default:
				return 0, err
			}
		}
	}
	return 0, nil
}

func sessionModelInsert(db *sql.DB, m *Model) (int, error) {
	_, err := db.Exec("INSERT INTO usersessions (sessionid, userid, keyhash, time, auth_time, ipaddr, user_agent) VALUES ($1, $2, $3, $4, $5, $6, $7);", m.SessionID, m.Userid, m.KeyHash, m.Time, m.AuthTime, m.IPAddr, m.UserAgent)
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
	args := make([]interface{}, 0, len(models)*7)
	for c, m := range models {
		n := c * 7
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4, n+5, n+6, n+7))
		args = append(args, m.SessionID, m.Userid, m.KeyHash, m.Time, m.AuthTime, m.IPAddr, m.UserAgent)
	}
	_, err := db.Exec("INSERT INTO usersessions (sessionid, userid, keyhash, time, auth_time, ipaddr, user_agent) VALUES "+strings.Join(placeholders, ", ")+conflictSQL+";", args...)
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

func sessionModelGetModelEqSessionID(db *sql.DB, sessionid string) (*Model, int, error) {
	m := &Model{}
	if err := db.QueryRow("SELECT sessionid, userid, keyhash, time, auth_time, ipaddr, user_agent FROM usersessions WHERE sessionid = $1;", sessionid).Scan(&m.SessionID, &m.Userid, &m.KeyHash, &m.Time, &m.AuthTime, &m.IPAddr, &m.UserAgent); err != nil {
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

func sessionModelUpdModelEqSessionID(db *sql.DB, m *Model, sessionid string) (int, error) {
	_, err := db.Exec("UPDATE usersessions SET (sessionid, userid, keyhash, time, auth_time, ipaddr, user_agent) = ROW($1, $2, $3, $4, $5, $6, $7) WHERE sessionid = $8;", m.SessionID, m.Userid, m.KeyHash, m.Time, m.AuthTime, m.IPAddr, m.UserAgent, sessionid)
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

func sessionModelDelEqSessionID(db *sql.DB, sessionid string) error {
	_, err := db.Exec("DELETE FROM usersessions WHERE sessionid = $1;", sessionid)
	return err
}

func sessionModelDelHasSessionID(db *sql.DB, sessionid []string) error {
	paramCount := 0
	args := make([]interface{}, 0, paramCount+len(sessionid))
	var placeholderssessionid string
	{
		placeholders := make([]string, 0, len(sessionid))
		for _, i := range sessionid {
			paramCount++
			placeholders = append(placeholders, fmt.Sprintf("($%d)", paramCount))
			args = append(args, i)
		}
		placeholderssessionid = strings.Join(placeholders, ", ")
	}
	_, err := db.Exec("DELETE FROM usersessions WHERE sessionid IN (VALUES "+placeholderssessionid+");", args...)
	return err
}

func sessionModelDelEqUserid(db *sql.DB, userid string) error {
	_, err := db.Exec("DELETE FROM usersessions WHERE userid = $1;", userid)
	return err
}

func sessionModelGetModelEqUseridOrdTime(db *sql.DB, userid string, orderasc bool, limit, offset int) ([]Model, error) {
	order := "DESC"
	if orderasc {
		order = "ASC"
	}
	res := make([]Model, 0, limit)
	rows, err := db.Query("SELECT sessionid, userid, keyhash, time, auth_time, ipaddr, user_agent FROM usersessions WHERE userid = $3 ORDER BY time "+order+" LIMIT $1 OFFSET $2;", limit, offset, userid)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := Model{}
		if err := rows.Scan(&m.SessionID, &m.Userid, &m.KeyHash, &m.Time, &m.AuthTime, &m.IPAddr, &m.UserAgent); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func sessionModelGetqIDEqUseridOrdSessionID(db *sql.DB, userid string, orderasc bool, limit, offset int) ([]qID, error) {
	order := "DESC"
	if orderasc {
		order = "ASC"
	}
	res := make([]qID, 0, limit)
	rows, err := db.Query("SELECT sessionid FROM usersessions WHERE userid = $3 ORDER BY sessionid "+order+" LIMIT $1 OFFSET $2;", limit, offset, userid)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := qID{}
		if err := rows.Scan(&m.SessionID); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
