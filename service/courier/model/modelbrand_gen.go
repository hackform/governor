// Code generated by go generate forge model v0.3; DO NOT EDIT.

package model

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

const (
	brandModelTableName = "courierbrands"
)

func brandModelSetup(db *sql.DB) (int, error) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS courierbrands (creatorid VARCHAR(31), brandid VARCHAR(63), PRIMARY KEY (creatorid, brandid), creation_time BIGINT NOT NULL);")
	if err != nil {
		return 0, err
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS courierbrands_creation_time_index ON courierbrands (creation_time);")
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

func brandModelInsert(db *sql.DB, m *BrandModel) (int, error) {
	_, err := db.Exec("INSERT INTO courierbrands (creatorid, brandid, creation_time) VALUES ($1, $2, $3);", m.CreatorID, m.BrandID, m.CreationTime)
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

func brandModelInsertBulk(db *sql.DB, models []*BrandModel, allowConflict bool) (int, error) {
	conflictSQL := ""
	if allowConflict {
		conflictSQL = " ON CONFLICT DO NOTHING"
	}
	placeholders := make([]string, 0, len(models))
	args := make([]interface{}, 0, len(models)*3)
	for c, m := range models {
		n := c * 3
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", n+1, n+2, n+3))
		args = append(args, m.CreatorID, m.BrandID, m.CreationTime)
	}
	_, err := db.Exec("INSERT INTO courierbrands (creatorid, brandid, creation_time) VALUES "+strings.Join(placeholders, ", ")+conflictSQL+";", args...)
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

func brandModelGetBrandModelEqCreatorIDEqBrandID(db *sql.DB, creatorid string, brandid string) (*BrandModel, int, error) {
	m := &BrandModel{}
	if err := db.QueryRow("SELECT creatorid, brandid, creation_time FROM courierbrands WHERE creatorid = $1 AND brandid = $2;", creatorid, brandid).Scan(&m.CreatorID, &m.BrandID, &m.CreationTime); err != nil {
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

func brandModelDelEqCreatorIDEqBrandID(db *sql.DB, creatorid string, brandid string) error {
	_, err := db.Exec("DELETE FROM courierbrands WHERE creatorid = $1 AND brandid = $2;", creatorid, brandid)
	return err
}

func brandModelGetBrandModelOrdCreationTime(db *sql.DB, orderasc bool, limit, offset int) ([]BrandModel, error) {
	order := "DESC"
	if orderasc {
		order = "ASC"
	}
	res := make([]BrandModel, 0, limit)
	rows, err := db.Query("SELECT creatorid, brandid, creation_time FROM courierbrands ORDER BY creation_time "+order+" LIMIT $1 OFFSET $2;", limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := BrandModel{}
		if err := rows.Scan(&m.CreatorID, &m.BrandID, &m.CreationTime); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func brandModelGetBrandModelEqCreatorIDOrdCreationTime(db *sql.DB, creatorid string, orderasc bool, limit, offset int) ([]BrandModel, error) {
	order := "DESC"
	if orderasc {
		order = "ASC"
	}
	res := make([]BrandModel, 0, limit)
	rows, err := db.Query("SELECT creatorid, brandid, creation_time FROM courierbrands WHERE creatorid = $3 ORDER BY creation_time "+order+" LIMIT $1 OFFSET $2;", limit, offset, creatorid)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
		}
	}()
	for rows.Next() {
		m := BrandModel{}
		if err := rows.Scan(&m.CreatorID, &m.BrandID, &m.CreationTime); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
