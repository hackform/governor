package rolemodel

import (
	"database/sql"
	"github.com/hackform/governor"
	"github.com/hackform/governor/service/db"
	"net/http"
)

//go:generate forge model -m Model -t userroles -p role -o model_gen.go qUserid qRole

type (
	Repo interface {
		New(userid, role string) (*Model, error)
		GetByID(userid, role string) (*Model, error)
		GetByRole(role string, limit, offset int) ([]string, error)
		GetUserRoles(userid string, limit, offset int) ([]string, error)
		Insert(m *Model) error
		Delete(m *Model) error
		DeleteUserRoles(userid string) error
		Setup() error
	}

	repo struct {
		db *sql.DB
	}

	// Model is the db User role model
	Model struct {
		roleid string `model:"roleid,VARCHAR(511) PRIMARY KEY"`
		Userid string `model:"userid,VARCHAR(31) NOT NULL"`
		Role   string `model:"role,VARCHAR(255) NOT NULL"`
	}

	qUserid struct {
		Userid string `query:"userid,getgroupeq,role"`
	}

	qRole struct {
		Role string `query:"role,getgroupeq,userid"`
	}
)

func New(conf governor.Config, l governor.Logger, database db.Database) Repo {
	l.Info("initialize user role model", nil)
	return &repo{
		db: database.DB(),
	}
}

// New creates a new User role Model
func (r *repo) New(userid, role string) (*Model, error) {
	m := &Model{
		Userid: userid,
		Role:   role,
	}
	m.ensureRoleid()
	return m, nil
}

func (m *Model) ensureRoleid() string {
	r := m.Userid + "|" + m.Role
	m.roleid = r
	return r
}

// GetByID returns a user role model with the given id
func (r *repo) GetByID(userid, role string) (*Model, error) {
	var m *Model
	if mRole, code, err := roleModelGet(r.db, (&Model{Userid: userid, Role: role}).ensureRoleid()); err != nil {
		if code == 2 {
			return nil, governor.NewError("Role not found for user", http.StatusNotFound, err)
		}
		return nil, governor.NewError("Failed to get role", http.StatusInternalServerError, err)
	} else {
		m = mRole
	}
	return m, nil
}

// GetByRole returns a list of userids with the given role
func (r *repo) GetByRole(role string, limit, offset int) ([]string, error) {
	m, err := roleModelGetqUseridEqRoleOrdUserid(r.db, role, true, limit, offset)
	if err != nil {
		return nil, governor.NewError("Failed to get userids of role", http.StatusInternalServerError, err)
	}
	userids := make([]string, 0, len(m))
	for _, i := range m {
		userids = append(userids, i.Userid)
	}
	return userids, nil
}

// GetUserRoles returns a list of a user's roles
func (r *repo) GetUserRoles(userid string, limit, offset int) ([]string, error) {
	m, err := roleModelGetqRoleEqUseridOrdRole(r.db, userid, true, limit, offset)
	if err != nil {
		return nil, governor.NewError("Failed to get roles of userid", http.StatusInternalServerError, err)
	}
	roles := make([]string, 0, len(m))
	for _, i := range m {
		roles = append(roles, i.Role)
	}
	return roles, nil
}

// Insert inserts the model into the db
func (r *repo) Insert(m *Model) error {
	m.ensureRoleid()
	if code, err := roleModelInsert(r.db, m); err != nil {
		if code == 3 {
			return governor.NewErrorUser("Role id must be unique", http.StatusBadRequest, err)
		}
		return governor.NewError("Failed to insert role", http.StatusInternalServerError, err)
	}
	return nil
}

// Delete deletes the model in the db
func (r *repo) Delete(m *Model) error {
	m.ensureRoleid()
	if err := roleModelDelete(r.db, m); err != nil {
		return governor.NewError("Failed to delete role", http.StatusInternalServerError, err)
	}
	return nil
}

const (
	sqlDeleteItem = "DELETE FROM " + roleModelTableName + " WHERE userid=$1;"
)

// DeleteUserRoles deletes all the roles of a user
func (r *repo) DeleteUserRoles(userid string) error {
	_, err := r.db.Exec(sqlDeleteItem, userid)
	if err != nil {
		return governor.NewError("Failed to delete roles of userid", http.StatusInternalServerError, err)
	}
	return nil
}

// Setup creates a new User role table
func (r *repo) Setup() error {
	if err := roleModelSetup(r.db); err != nil {
		return governor.NewError("Failed to setup role model", http.StatusInternalServerError, err)
	}
	return nil
}
