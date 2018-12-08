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
	_, err := db.Exec("CREATE TABLE userroles (roleid VARCHAR(512) PRIMARY KEY, userid VARCHAR(255) NOT NULL, role VARCHAR(255) NOT NULL);")
	return err
}

func roleModelGet(db *sql.DB, key string) (*Model, int, error) {
	m := &Model{}
	if err := db.QueryRow("SELECT roleid, userid, role FROM userroles WHERE roleid = $1;", key).Scan(&m.roleid, &m.Userid, &m.Role); err != nil {
		if err == sql.ErrNoRows {
			return nil, 2, err
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

func roleModelGetuseridByRoleGroupByroleid(db *sql.DB, key string, limit, offset int) ([]useridByRole, error) {
	res := make([]useridByRole, 0, limit)
	rows, err := db.Query("SELECT roleid, userid FROM userroles WHERE role = $1 ORDER BY roleid ASC LIMIT $2 OFFSET $3;", key, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := useridByRole{}
		if err := rows.Scan(&m.roleid, &m.Userid); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func roleModelGetroleByUseridGroupByroleid(db *sql.DB, key string, limit, offset int) ([]roleByUserid, error) {
	res := make([]roleByUserid, 0, limit)
	rows, err := db.Query("SELECT roleid, role FROM userroles WHERE userid = $1 ORDER BY roleid ASC LIMIT $2 OFFSET $3;", key, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := roleByUserid{}
		if err := rows.Scan(&m.roleid, &m.Role); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
