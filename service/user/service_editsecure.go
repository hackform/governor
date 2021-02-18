package user

import (
	"bytes"
	htmlTemplate "html/template"
	"net/http"
	"net/url"
	"time"

	"xorkevin.dev/governor"
	usermodel "xorkevin.dev/governor/service/user/model"
)

const (
	kindResetEmail = "email"
	kindResetPass  = "pass"
)

type (
	emailEmailChange struct {
		Userid    string
		Key       string
		URL       string
		FirstName string
		LastName  string
		Username  string
	}

	queryEmailEmailChange struct {
		Userid    string
		Key       string
		FirstName string
		LastName  string
		Username  string
	}

	emailEmailChangeNotify struct {
		FirstName string
		LastName  string
		Username  string
	}
)

const (
	emailChangeTemplate       = "emailchange"
	emailChangeNotifyTemplate = "emailchangenotify"
)

func (e *emailEmailChange) Query() queryEmailEmailChange {
	return queryEmailEmailChange{
		Userid:    url.QueryEscape(e.Userid),
		Key:       url.QueryEscape(e.Key),
		FirstName: url.QueryEscape(e.FirstName),
		LastName:  url.QueryEscape(e.LastName),
		Username:  url.QueryEscape(e.Username),
	}
}

func (e *emailEmailChange) computeURL(base string, tpl *htmlTemplate.Template) error {
	b := &bytes.Buffer{}
	if err := tpl.Execute(b, e.Query()); err != nil {
		return governor.NewError("Failed executing email change url template", http.StatusInternalServerError, err)
	}
	e.URL = base + b.String()
	return nil
}

// UpdateEmail creates a pending user email update
func (s *service) UpdateEmail(userid string, newEmail string, password string) error {
	if _, err := s.users.GetByEmail(newEmail); err != nil {
		if governor.ErrorStatus(err) != http.StatusNotFound {
			return governor.NewErrorUser("", 0, err)
		}
	} else {
		return governor.NewErrorUser("Email is already in use", http.StatusBadRequest, err)
	}
	m, err := s.users.GetByID(userid)
	if err != nil {
		if governor.ErrorStatus(err) == http.StatusNotFound {
			return governor.NewErrorUser("", 0, err)
		}
		return err
	}
	if m.Email == newEmail {
		return governor.NewErrorUser("Emails cannot be the same", http.StatusBadRequest, err)
	}
	if ok, err := s.users.ValidatePass(password, m); err != nil {
		return err
	} else if !ok {
		return governor.NewErrorUser("Incorrect password", http.StatusForbidden, nil)
	}

	needInsert := false
	mr, err := s.resets.GetByID(m.Userid, kindResetEmail)
	if err != nil {
		if governor.ErrorStatus(err) != http.StatusNotFound {
			return governor.NewErrorUser("", 0, err)
		}
		needInsert = true
		mr = s.resets.New(m.Userid, kindResetEmail)
	}
	code, err := s.resets.RehashCode(mr)
	if err != nil {
		return err
	}
	mr.Params = newEmail
	if needInsert {
		if err := s.resets.Insert(mr); err != nil {
			if governor.ErrorStatus(err) == http.StatusBadRequest {
				return governor.NewErrorUser("", 0, err)
			}
			return err
		}
	} else {
		if err := s.resets.Update(mr); err != nil {
			return err
		}
	}

	emdata := emailEmailChange{
		Userid:    userid,
		Key:       code,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Username:  m.Username,
	}
	if err := emdata.computeURL(s.emailurlbase, s.tplemailchange); err != nil {
		return err
	}
	if err := s.mailer.Send("", "", []string{newEmail}, emailChangeTemplate, emdata); err != nil {
		return governor.NewError("Failed to send new email verification", http.StatusInternalServerError, err)
	}
	return nil
}

// CommitEmail commits an email update from the cache
func (s *service) CommitEmail(userid string, key string, password string) error {
	mr, err := s.resets.GetByID(userid, kindResetEmail)
	if err != nil {
		if governor.ErrorStatus(err) == http.StatusNotFound {
			return governor.NewErrorUser("New email verification expired", http.StatusBadRequest, err)
		}
		return governor.NewError("Failed to update email", http.StatusInternalServerError, err)
	}

	if time.Now().Round(0).Unix() > mr.CodeTime+s.passwordResetTime {
		return governor.NewErrorUser("New email verification expired", http.StatusBadRequest, err)
	}
	if ok, err := s.resets.ValidateCode(key, mr); err != nil {
		return governor.NewError("Failed to update email", http.StatusInternalServerError, err)
	} else if !ok {
		return governor.NewErrorUser("Invalid code", http.StatusForbidden, nil)
	}

	m, err := s.users.GetByID(userid)
	if err != nil {
		if governor.ErrorStatus(err) == http.StatusNotFound {
			return governor.NewErrorUser("", 0, err)
		}
		return err
	}

	if ok, err := s.users.ValidatePass(password, m); err != nil {
		return err
	} else if !ok {
		return governor.NewErrorUser("Incorrect password", http.StatusForbidden, nil)
	}

	m.Email = mr.Params

	if err := s.resets.Delete(userid, kindResetEmail); err != nil {
		return err
	}

	if err = s.users.Update(m); err != nil {
		if governor.ErrorStatus(err) == http.StatusBadRequest {
			return governor.NewErrorUser("Email is already in use by another account", 0, err)
		}
		return err
	}

	emdatanotify := emailEmailChangeNotify{
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Username:  m.Username,
	}
	if err := s.mailer.Send("", "", []string{m.Email}, emailChangeNotifyTemplate, emdatanotify); err != nil {
		s.logger.Error("Failed to send old email change notification", map[string]string{
			"error":      err.Error(),
			"actiontype": "commitemailoldmail",
		})
	}
	return nil
}

type (
	emailPassChange struct {
		FirstName string
		LastName  string
		Username  string
	}

	emailForgotPass struct {
		Userid    string
		Key       string
		URL       string
		FirstName string
		LastName  string
		Username  string
	}

	queryEmailForgotPass struct {
		Userid    string
		Key       string
		FirstName string
		LastName  string
		Username  string
	}

	emailPassReset struct {
		FirstName string
		LastName  string
		Username  string
	}
)

func (e *emailForgotPass) Query() queryEmailForgotPass {
	return queryEmailForgotPass{
		Userid:    url.QueryEscape(e.Userid),
		Key:       url.QueryEscape(e.Key),
		FirstName: url.QueryEscape(e.FirstName),
		LastName:  url.QueryEscape(e.LastName),
		Username:  url.QueryEscape(e.Username),
	}
}

func (e *emailForgotPass) computeURL(base string, tpl *htmlTemplate.Template) error {
	b := &bytes.Buffer{}
	if err := tpl.Execute(b, e.Query()); err != nil {
		return governor.NewError("Failed executing forgot pass url template", http.StatusInternalServerError, err)
	}
	e.URL = base + b.String()
	return nil
}

const (
	passChangeTemplate = "passchange"
	forgotPassTemplate = "forgotpass"
	passResetTemplate  = "passreset"
)

// UpdatePassword updates the password
func (s *service) UpdatePassword(userid string, newPassword string, oldPassword string) error {
	m, err := s.users.GetByID(userid)
	if err != nil {
		if governor.ErrorStatus(err) == http.StatusNotFound {
			return governor.NewErrorUser("", 0, err)
		}
		return err
	}
	if ok, err := s.users.ValidatePass(oldPassword, m); err != nil {
		return err
	} else if !ok {
		return governor.NewErrorUser("Incorrect password", http.StatusForbidden, err)
	}
	if err := s.users.RehashPass(m, newPassword); err != nil {
		return err
	}

	if err = s.users.Update(m); err != nil {
		return err
	}

	emdata := emailPassChange{
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Username:  m.Username,
	}
	if err := s.mailer.Send("", "", []string{m.Email}, passChangeTemplate, emdata); err != nil {
		s.logger.Error("Failed to send password change notification email", map[string]string{
			"error":      err.Error(),
			"actiontype": "updatepasswordmail",
		})
	}
	return nil
}

// ForgotPassword invokes the forgot password reset procedure
func (s *service) ForgotPassword(useroremail string) error {
	var m *usermodel.Model
	if isEmail(useroremail) {
		mu, err := s.users.GetByEmail(useroremail)
		if err != nil {
			if governor.ErrorStatus(err) == http.StatusNotFound {
				return nil
			}
			return err
		}
		m = mu
	} else {
		mu, err := s.users.GetByUsername(useroremail)
		if err != nil {
			if governor.ErrorStatus(err) == http.StatusNotFound {
				return nil
			}
			return err
		}
		m = mu
	}

	needInsert := false
	mr, err := s.resets.GetByID(m.Userid, kindResetPass)
	if err != nil {
		if governor.ErrorStatus(err) != http.StatusNotFound {
			return governor.NewErrorUser("", 0, err)
		}
		needInsert = true
		mr = s.resets.New(m.Userid, kindResetPass)
	}
	code, err := s.resets.RehashCode(mr)
	if err != nil {
		return err
	}
	if needInsert {
		if err := s.resets.Insert(mr); err != nil {
			if governor.ErrorStatus(err) == http.StatusBadRequest {
				return governor.NewErrorUser("", 0, err)
			}
			return err
		}
	} else {
		if err := s.resets.Update(mr); err != nil {
			return err
		}
	}

	emdata := emailForgotPass{
		Userid:    m.Userid,
		Key:       code,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Username:  m.Username,
	}
	if err := emdata.computeURL(s.emailurlbase, s.tplforgotpass); err != nil {
		return err
	}
	if err := s.mailer.Send("", "", []string{m.Email}, forgotPassTemplate, emdata); err != nil {
		return governor.NewError("Failed to send password reset email", http.StatusInternalServerError, err)
	}
	return nil
}

// ResetPassword completes the forgot password procedure
func (s *service) ResetPassword(userid string, key string, newPassword string) error {
	mr, err := s.resets.GetByID(userid, kindResetPass)
	if err != nil {
		if governor.ErrorStatus(err) == http.StatusNotFound {
			return governor.NewErrorUser("Password reset expired", http.StatusBadRequest, err)
		}
		return governor.NewError("Failed to reset password", http.StatusInternalServerError, err)
	}

	if time.Now().Round(0).Unix() > mr.CodeTime+s.passwordResetTime {
		return governor.NewErrorUser("Password reset expired", http.StatusBadRequest, err)
	}
	if ok, err := s.resets.ValidateCode(key, mr); err != nil {
		return governor.NewError("Failed to reset password", http.StatusInternalServerError, err)
	} else if !ok {
		return governor.NewErrorUser("Invalid code", http.StatusForbidden, nil)
	}

	m, err := s.users.GetByID(userid)
	if err != nil {
		if governor.ErrorStatus(err) == http.StatusNotFound {
			return governor.NewErrorUser("", 0, err)
		}
		return err
	}

	if err := s.users.RehashPass(m, newPassword); err != nil {
		return err
	}

	if err := s.resets.Delete(userid, kindResetPass); err != nil {
		return err
	}

	if err := s.users.Update(m); err != nil {
		return err
	}

	emdata := emailPassReset{
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Username:  m.Username,
	}
	if err := s.mailer.Send("", "", []string{m.Email}, passResetTemplate, emdata); err != nil {
		s.logger.Error("Failed to send password change notification email", map[string]string{
			"error":      err.Error(),
			"actiontype": "resetpasswordmail",
		})
	}
	return nil
}
