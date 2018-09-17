package user

import (
	"bytes"
	"encoding/gob"
	"github.com/hackform/governor"
	"github.com/hackform/governor/service/user/model"
	"github.com/hackform/governor/service/user/session"
	"github.com/hackform/governor/util/rank"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

func (u *userService) UpdateUser(userid string, ruser reqUserPut) *governor.Error {
	m, err := usermodel.GetByIDB64(u.db.DB(), userid)
	if err != nil {
		if err.Code() == 2 {
			err.SetErrorUser()
		}
		err.AddTrace(moduleIDUser)
		return err
	}
	m.Username = ruser.Username
	m.FirstName = ruser.FirstName
	m.LastName = ruser.LastName
	if err = m.Update(u.db.DB()); err != nil {
		err.AddTrace(moduleIDUser)
		return err
	}
	return nil
}

func (u *userService) UpdateRank(userid string, updaterid string, updaterRank rank.Rank, editAddRank rank.Rank, editRemoveRank rank.Rank) *governor.Error {
	if err := canUpdateRank(editAddRank, updaterRank, userid, updaterid, updaterRank.Has(rank.TagAdmin)); err != nil {
		return err
	}
	if err := canUpdateRank(editRemoveRank, updaterRank, userid, updaterid, updaterRank.Has(rank.TagAdmin)); err != nil {
		return err
	}

	m, err := usermodel.GetByIDB64(u.db.DB(), userid)
	if err != nil {
		if err.Code() == 2 {
			err.SetErrorUser()
		}
		err.AddTrace(moduleIDUser)
		return err
	}

	if editAddRank.Has("admin") {
		t, _ := time.Now().MarshalText()
		u.logger.WithFields(logrus.Fields{
			"time":     string(t),
			"origin":   moduleIDUser,
			"userid":   userid,
			"username": m.Username,
		}).Info("admin status added")
	}
	if editRemoveRank.Has("admin") {
		t, _ := time.Now().MarshalText()
		u.logger.WithFields(logrus.Fields{
			"time":     string(t),
			"origin":   moduleIDUser,
			"userid":   userid,
			"username": m.Username,
		}).Info("admin status removed")
	}

	diff := make(map[string]int)
	for k, v := range editAddRank {
		if v {
			diff[k] = usermodel.RoleAdd
		}
	}
	for k, v := range editRemoveRank {
		if v {
			diff[k] = usermodel.RoleRemove
		}
	}

	s := session.Session{
		Userid: userid,
	}

	var sarr []string
	if sgobs, err := u.cache.Cache().HGetAll(s.UserKey()).Result(); err == nil {
		sarr = make([]string, 0, len(sgobs))
		for _, v := range sgobs {
			s := session.Session{}
			if err = gob.NewDecoder(bytes.NewBufferString(v)).Decode(&s); err != nil {
				return governor.NewError(moduleIDUser, err.Error(), 0, http.StatusInternalServerError)
			}
			sarr = append(sarr, s.SessionID)
		}
	} else {
		return governor.NewError(moduleIDUser, err.Error(), 0, http.StatusInternalServerError)
	}

	if len(sarr) > 0 {
		if err := u.cache.Cache().Del(sarr...).Err(); err != nil {
			return governor.NewError(moduleIDUser, err.Error(), 0, http.StatusInternalServerError)
		}
		if err := u.cache.Cache().HDel(s.UserKey(), sarr...).Err(); err != nil {
			return governor.NewError(moduleIDUser, err.Error(), 0, http.StatusInternalServerError)
		}
	}

	if err := m.UpdateRoles(u.db.DB(), diff); err != nil {
		err.AddTrace(moduleIDUser)
		return err
	}
	return nil
}

func canUpdateRank(edit, updater rank.Rank, editid, updaterid string, isAdmin bool) *governor.Error {
	for key := range edit {
		k := strings.SplitN(key, "_", 2)
		if len(k) == 1 {
			switch k[0] {
			case rank.TagAdmin:
				// updater cannot change one's own admin status nor change another's admin status if he is not admin
				if editid == updaterid || !isAdmin {
					return governor.NewErrorUser(moduleIDUser, "forbidden rank edit", 0, http.StatusForbidden)
				}
			case rank.TagSystem:
				// no one can change the system status
				return governor.NewErrorUser(moduleIDUser, "forbidden rank edit", 0, http.StatusForbidden)
			case rank.TagUser:
				// only admins can change the user status
				if !isAdmin {
					return governor.NewErrorUser(moduleIDUser, "forbidden rank edit", 0, http.StatusForbidden)
				}
			default:
				// other tags cannot be edited
				return governor.NewErrorUser(moduleIDUser, "forbidden rank edit", 0, http.StatusBadRequest)
			}
		} else {
			// cannot edit group rank if not an admin or a moderator of that group
			if !isAdmin && updater.HasMod(k[1]) {
				return governor.NewErrorUser(moduleIDUser, "forbidden rank edit", 0, http.StatusForbidden)
			}
		}
	}
	return nil
}
