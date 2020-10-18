// Code generated by go generate forge model v0.3; DO NOT EDIT.

package connectionmodel

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"strings"
)

const (
	connectionModelTableName = "oauthconnections"
)

func connectionModelSetup(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS oauthconnections (userid VARCHAR(31), clientid VARCHAR(31), PRIMARY KEY (userid, clientid), scope VARCHAR(4095) NOT NULL, codehash VARCHAR(31) NOT NULL, time BIGINT NOT NULL, creation_time BIGINT NOT NULL);")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS oauthconnections_userid_index ON oauthconnections (userid);")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS oauthconnections_clientid_index ON oauthconnections (clientid);")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS oauthconnections_time_index ON oauthconnections (time);")
	if err != nil {
		return err
	}
	return nil
}

func connectionModelInsert(db *sql.DB, m *Model) (int, error) {
	_, err := db.Exec("INSERT INTO oauthconnections (userid, clientid, scope, codehash, time, creation_time) VALUES ($1, $2, $3, $4, $5, $6);", m.Userid, m.ClientID, m.Scope, m.CodeHash, m.Time, m.CreationTime)
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

func connectionModelInsertBulk(db *sql.DB, models []*Model, allowConflict bool) (int, error) {
	conflictSQL := ""
	if allowConflict {
		conflictSQL = " ON CONFLICT DO NOTHING"
	}
	placeholders := make([]string, 0, len(models))
	args := make([]interface{}, 0, len(models)*6)
	for c, m := range models {
		n := c * 6
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4, n+5, n+6))
		args = append(args, m.Userid, m.ClientID, m.Scope, m.CodeHash, m.Time, m.CreationTime)
	}
	_, err := db.Exec("INSERT INTO oauthconnections (userid, clientid, scope, codehash, time, creation_time) VALUES "+strings.Join(placeholders, ", ")+conflictSQL+";", args...)
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

func connectionModelDelEqUserid(db *sql.DB, userid string) error {
	_, err := db.Exec("DELETE FROM oauthconnections WHERE userid = $1;", userid)
	return err
}

func connectionModelGetModelEqUseridEqClientID(db *sql.DB, userid string, clientid string) (*Model, int, error) {
	m := &Model{}
	if err := db.QueryRow("SELECT userid, clientid, scope, codehash, time, creation_time FROM oauthconnections WHERE userid = $1 AND clientid = $2;", userid, clientid).Scan(&m.Userid, &m.ClientID, &m.Scope, &m.CodeHash, &m.Time, &m.CreationTime); err != nil {
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

func connectionModelUpdModelEqUseridEqClientID(db *sql.DB, m *Model, userid string, clientid string) (int, error) {
	_, err := db.Exec("UPDATE oauthconnections SET (userid, clientid, scope, codehash, time, creation_time) = ROW($1, $2, $3, $4, $5, $6) WHERE userid = $7 AND clientid = $8;", m.Userid, m.ClientID, m.Scope, m.CodeHash, m.Time, m.CreationTime, userid, clientid)
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

func connectionModelDelEqUseridHasClientID(db *sql.DB, userid string, clientid []string) error {
	paramCount := 1
	args := make([]interface{}, 0, paramCount+len(clientid))
	args = append(args, userid)
	var placeholdersclientid string
	{
		placeholders := make([]string, 0, len(clientid))
		for _, i := range clientid {
			paramCount++
			placeholders = append(placeholders, fmt.Sprintf("($%d)", paramCount))
			args = append(args, i)
		}
		placeholdersclientid = strings.Join(placeholders, ", ")
	}
	_, err := db.Exec("DELETE FROM oauthconnections WHERE userid = $1 AND clientid IN (VALUES "+placeholdersclientid+");", args...)
	return err
}

func connectionModelGetModelEqUseridOrdTime(db *sql.DB, userid string, orderasc bool, limit, offset int) ([]Model, error) {
	order := "DESC"
	if orderasc {
		order = "ASC"
	}
	res := make([]Model, 0, limit)
	rows, err := db.Query("SELECT userid, clientid, scope, codehash, time, creation_time FROM oauthconnections WHERE userid = $3 ORDER BY time "+order+" LIMIT $1 OFFSET $2;", limit, offset, userid)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := Model{}
		if err := rows.Scan(&m.Userid, &m.ClientID, &m.Scope, &m.CodeHash, &m.Time, &m.CreationTime); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
