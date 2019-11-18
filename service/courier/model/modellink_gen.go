// Code generated by go generate forge model v0.3; DO NOT EDIT.

package couriermodel

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"strings"
)

const (
	linkModelTableName = "courierlinks"
)

func linkModelSetup(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS courierlinks (linkid VARCHAR(63) PRIMARY KEY, url VARCHAR(2047) NOT NULL, creatorid VARCHAR(31) NOT NULL, creation_time BIGINT NOT NULL);")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS courierlinks_creatorid_index ON courierlinks (creatorid);")
	if err != nil {
		return err
	}
	return nil
}

func linkModelInsert(db *sql.DB, m *LinkModel) (int, error) {
	_, err := db.Exec("INSERT INTO courierlinks (linkid, url, creatorid, creation_time) VALUES ($1, $2, $3, $4);", m.LinkID, m.URL, m.CreatorID, m.CreationTime)
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

func linkModelInsertBulk(db *sql.DB, models []*LinkModel, allowConflict bool) (int, error) {
	conflictSQL := ""
	if allowConflict {
		conflictSQL = " ON CONFLICT DO NOTHING"
	}
	placeholders := make([]string, 0, len(models))
	args := make([]interface{}, 0, len(models)*4)
	for c, m := range models {
		n := c * 4
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4))
		args = append(args, m.LinkID, m.URL, m.CreatorID, m.CreationTime)
	}
	_, err := db.Exec("INSERT INTO courierlinks (linkid, url, creatorid, creation_time) VALUES "+strings.Join(placeholders, ", ")+conflictSQL+";", args...)
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

func linkModelGetLinkModelEqLinkID(db *sql.DB, linkid string) (*LinkModel, int, error) {
	m := &LinkModel{}
	if err := db.QueryRow("SELECT linkid, url, creatorid, creation_time FROM courierlinks WHERE linkid = $1;", linkid).Scan(&m.LinkID, &m.URL, &m.CreatorID, &m.CreationTime); err != nil {
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

func linkModelDelEqLinkID(db *sql.DB, linkid string) error {
	_, err := db.Exec("DELETE FROM courierlinks WHERE linkid = $1;", linkid)
	return err
}

func linkModelGetLinkModelOrdCreationTime(db *sql.DB, orderasc bool, limit, offset int) ([]LinkModel, error) {
	order := "DESC"
	if orderasc {
		order = "ASC"
	}
	res := make([]LinkModel, 0, limit)
	rows, err := db.Query("SELECT linkid, url, creatorid, creation_time FROM courierlinks ORDER BY creation_time "+order+" LIMIT $1 OFFSET $2;", limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := LinkModel{}
		if err := rows.Scan(&m.LinkID, &m.URL, &m.CreatorID, &m.CreationTime); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func linkModelGetLinkModelEqCreatorIDOrdCreationTime(db *sql.DB, creatorid string, orderasc bool, limit, offset int) ([]LinkModel, error) {
	order := "DESC"
	if orderasc {
		order = "ASC"
	}
	res := make([]LinkModel, 0, limit)
	rows, err := db.Query("SELECT linkid, url, creatorid, creation_time FROM courierlinks WHERE creatorid = $3 ORDER BY creation_time "+order+" LIMIT $1 OFFSET $2;", limit, offset, creatorid)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := LinkModel{}
		if err := rows.Scan(&m.LinkID, &m.URL, &m.CreatorID, &m.CreationTime); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
