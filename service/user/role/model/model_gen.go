// Code generated by go generate. DO NOT EDIT.
package rolemodel

import (
	"database/sql"
	"github.com/lib/pq"
)

const (
	roleModelTableName = "userroles"
)

func roleModelSetup(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS userroles (roleid VARCHAR(511) PRIMARY KEY, userid VARCHAR(31) NOT NULL, role VARCHAR(255) NOT NULL);")
	return err
}

func roleModelGet(db *sql.DB, key string) (*Model, int, error) {
	m := &Model{}
	if err := db.QueryRow("SELECT roleid, userid, role FROM userroles WHERE roleid = $1;", key).Scan(&m.roleid, &m.Userid, &m.Role); err != nil {
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

func roleModelInsert(db *sql.DB, m *Model) (int, error) {
	_, err := db.Exec("INSERT INTO userroles (roleid, userid, role) VALUES ($1, $2, $3);", m.roleid, m.Userid, m.Role)
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

func roleModelUpdate(db *sql.DB, m *Model) error {
	_, err := db.Exec("UPDATE userroles SET (roleid, userid, role) = ($1, $2, $3) WHERE roleid = $1;", m.roleid, m.Userid, m.Role)
	return err
}

func roleModelDelete(db *sql.DB, m *Model) error {
	_, err := db.Exec("DELETE FROM userroles WHERE roleid = $1;", m.roleid)
	return err
}

func roleModelGetqUseridEqRoleOrdUserid(db *sql.DB, role string, orderasc bool, limit, offset int) ([]qUserid, error) {
	order := "DESC"
	if orderasc {
		order = "ASC"
	}
	res := make([]qUserid, 0, limit)
	rows, err := db.Query("SELECT userid FROM userroles WHERE role = $3 ORDER BY userid "+order+" LIMIT $1 OFFSET $2;", limit, offset, role)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := qUserid{}
		if err := rows.Scan(&m.Userid); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func roleModelGetqRoleEqUseridOrdRole(db *sql.DB, userid string, orderasc bool, limit, offset int) ([]qRole, error) {
	order := "DESC"
	if orderasc {
		order = "ASC"
	}
	res := make([]qRole, 0, limit)
	rows, err := db.Query("SELECT role FROM userroles WHERE userid = $3 ORDER BY role "+order+" LIMIT $1 OFFSET $2;", limit, offset, userid)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := qRole{}
		if err := rows.Scan(&m.Role); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
