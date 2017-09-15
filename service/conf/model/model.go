package confmodel

import (
	"database/sql"
	"fmt"
	"github.com/hackform/governor"
	"net/http"
	"time"
)

const (
	tableName     = "config"
	rowID         = 0
	moduleID      = "confmodel"
	moduleIDModel = moduleID + ".Model"
)

type (
	// Model is the db Config model
	Model struct {
		Props
	}

	// Props stores Config info
	Props struct {
		Orgname      string `json:"orgname"`
		CreationTime int64  `json:"creation_time"`
	}
)

// New creates a new Conf Model
func New(orgname string) (*Model, *governor.Error) {
	return &Model{
		Props: Props{
			Orgname:      orgname,
			CreationTime: time.Now().Unix(),
		},
	}, nil
}

const (
	moduleIDModGet = moduleIDModel + ".Get"
)

var (
	sqlGet = fmt.Sprintf("SELECT orgname, creation_time FROM %s WHERE config=$1;", tableName)
)

// Get returns the conf model
func Get(db *sql.DB) (*Model, *governor.Error) {
	mConf := &Model{}
	if err := db.QueryRow(sqlGet, rowID).Scan(&mConf.Orgname, &mConf.CreationTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, governor.NewError(moduleIDModGet, "no conf found with that id", 2, http.StatusNotFound)
		}
		return nil, governor.NewError(moduleIDModGet, err.Error(), 0, http.StatusInternalServerError)
	}
	return mConf, nil
}

const (
	moduleIDModIns = moduleIDModel + ".Insert"
)

// Insert inserts the model into the db
func (m *Model) Insert(db *sql.DB) *governor.Error {
	_, err := db.Exec(fmt.Sprintf("INSERT INTO %s (config, orgname, creation_time) VALUES ($1, $2, $3);", tableName), rowID, m.Orgname, m.CreationTime)
	if err != nil {
		return governor.NewError(moduleIDModIns, err.Error(), 0, http.StatusInternalServerError)
	}
	return nil
}

const (
	moduleIDModUp = moduleIDModel + ".Update"
)

// Update updates the model in the db
func (m *Model) Update(db *sql.DB) *governor.Error {
	_, err := db.Exec(fmt.Sprintf("UPDATE %s SET (orgname, creation_time) = ($2, $3) WHERE config = $1;", tableName), rowID, m.Orgname, m.CreationTime)
	if err != nil {
		return governor.NewError(moduleIDModUp, err.Error(), 0, http.StatusInternalServerError)
	}
	return nil
}

const (
	moduleIDSetup = moduleID + ".Setup"
)

// Setup creates a new Config table
func Setup(db *sql.DB) *governor.Error {
	_, err := db.Exec(fmt.Sprintf("CREATE TABLE %s (config INT PRIMARY KEY, orgname VARCHAR(255) NOT NULL, creation_time BIGINT NOT NULL);", tableName))
	if err != nil {
		return governor.NewError(moduleIDSetup, err.Error(), 0, http.StatusInternalServerError)
	}
	return nil
}
