package user

import (
	"net/http"
	"strings"
	"time"
	"xorkevin.dev/governor"
	"xorkevin.dev/governor/service/user/session"
	"xorkevin.dev/governor/service/user/token"
	"xorkevin.dev/governor/util/uid"
)

type (
	emailNewLogin struct {
		FirstName string
		Username  string
		SessionID string
		IP        string
		UserAgent string
		Time      string
	}
)

const (
	newLoginTemplate = "newlogin"
	newLoginSubject  = "newlogin_subject"
)

type (
	resUserAuth struct {
		Valid        bool          `json:"valid"`
		AccessToken  string        `json:"access_token,omitempty"`
		RefreshToken string        `json:"refresh_token,omitempty"`
		SessionToken string        `json:"session_token,omitempty"`
		Claims       *token.Claims `json:"claims,omitempty"`
	}
)

// Login authenticates a user and returns auth tokens
func (u *userService) Login(userid, password, sessionToken, ipAddress, userAgent string) (bool, *resUserAuth, error) {
	m, err := u.repo.GetByID(userid)
	if err != nil {
		if governor.ErrorStatus(err) == http.StatusNotFound {
			return false, nil, governor.NewErrorUser("", 0, err)
		}
		return false, nil, err
	}
	if ok, err := u.repo.ValidatePass(password, m); err != nil {
		return false, nil, err
	} else if !ok {
		return false, &resUserAuth{
			Valid: false,
		}, nil
	}

	sessionID := ""
	isMember := false
	// If claims userid matches model, session_id is provided, is in list of user
	// sessions, set it as the sessionID. The session can be expired by time.
	if ok, claims := u.tokenizer.GetClaims(sessionToken, sessionSubject); ok {
		if userid == claims.Userid {
			if isM, err := u.SessionExists(userid, claims.Id); err == nil && isM {
				sessionID = claims.Id
				isMember = isM
			}
		}
	}

	var s *session.Session
	if sessionID == "" {
		// otherwise, create a new sessionID
		if s, err = session.New(m, ipAddress, userAgent); err != nil {
			return false, nil, err
		}
	} else {
		if s, err = session.FromSessionID(sessionID, userid, ipAddress, userAgent); err != nil {
			return false, nil, err
		}
	}

	// generate an access token
	accessToken, claims, err := u.tokenizer.Generate(m, u.accessTime, authenticationSubject, "")
	if err != nil {
		return false, nil, err
	}
	// generate a refresh token with the sessionKey
	refreshToken, _, err := u.tokenizer.Generate(m, u.refreshTime, refreshSubject, s.SessionID+":"+s.SessionKey)
	if err != nil {
		return false, nil, err
	}
	// generate a session token
	newSessionToken, _, err := u.tokenizer.Generate(m, u.refreshTime, sessionSubject, s.SessionID)
	if err != nil {
		return false, nil, err
	}

	if u.newLoginEmail && !isMember {
		emdata := emailNewLogin{
			FirstName: m.FirstName,
			Username:  m.Username,
			SessionID: s.SessionID,
			IP:        s.IP,
			Time:      time.Unix(s.Time, 0).String(),
			UserAgent: s.UserAgent,
		}
		if err := u.mailer.Send(m.Email, newLoginSubject, newLoginTemplate, emdata); err != nil {
			u.logger.Error("fail send new login email", map[string]string{
				"err": err.Error(),
			})
		}
	}

	if err := u.AddSession(s, time.Duration(u.refreshTime*b1)); err != nil {
		return false, nil, err
	}

	return true, &resUserAuth{
		Valid:        true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionToken: newSessionToken,
		Claims:       claims,
	}, nil
}

// ExchangeToken validates a refresh token and returns an auth token
func (u *userService) ExchangeToken(refreshToken, ipAddress, userAgent string) (bool, *resUserAuth, error) {
	sessionID := ""
	sessionKey := ""
	userid := ""
	// if session_id is provided, is in cache, and is valid, set it as the sessionID
	// the session cannot be expired
	if ok, claims := u.tokenizer.GetClaims(refreshToken, refreshSubject); ok {
		if s := strings.Split(claims.Id, ":"); len(s) == 2 {
			if key, err := u.GetSessionKey(s[0]); err == nil {
				sessionID = s[0]
				sessionKey = key
				userid = claims.Userid
			}
		}
	}

	if sessionID == "" {
		return false, nil, governor.NewErrorUser("malformed refresh token", http.StatusUnauthorized, nil)
	}

	// check the refresh token
	validToken, claims := u.tokenizer.Validate(refreshToken, refreshSubject, sessionID+":"+sessionKey)
	if !validToken {
		return false, &resUserAuth{
			Valid: false,
		}, nil
	}

	// update the user session with a new latest time
	s, err := session.FromSessionID(sessionID, userid, ipAddress, userAgent)
	if err != nil {
		return false, nil, err
	}
	if err := u.UpdateUserSession(s); err != nil {
		return false, nil, err
	}

	// generate a new accessToken from the refreshToken claims
	accessToken, err := u.tokenizer.GenerateFromClaims(claims, u.accessTime, authenticationSubject, "")
	if err != nil {
		return false, nil, err
	}

	return true, &resUserAuth{
		Valid:       true,
		AccessToken: accessToken,
		Claims:      claims,
	}, nil
}

// RefreshToken invalidates the previous refresh token and creates a new one
func (u *userService) RefreshToken(refreshToken string) (bool, *resUserAuth, error) {
	sessionID := ""
	sessionKey := ""
	// If session_id is provided, is in cache, and is valid, set it as the
	// sessionID. The session cannot be expired.
	if ok, claims := u.tokenizer.GetClaims(refreshToken, refreshSubject); ok {
		if s := strings.Split(claims.Id, ":"); len(s) == 2 {
			if key, err := u.GetSessionKey(s[0]); err == nil {
				sessionID = s[0]
				sessionKey = key
			}
		}
	}

	if sessionID == "" {
		return false, nil, governor.NewErrorUser("malformed refresh token", http.StatusUnauthorized, nil)
	}

	// check the refresh token
	validToken, claims := u.tokenizer.Validate(refreshToken, refreshSubject, sessionID+":"+sessionKey)
	if !validToken {
		return false, &resUserAuth{
			Valid: false,
		}, nil
	}

	// create a new key for the session
	key, err := uid.NewU(0, 16)
	if err != nil {
		return false, nil, governor.NewError("Failed to create new uid", http.StatusInternalServerError, err)
	}
	sessionKey = key.Base64()

	// generate a new refreshToken from the refreshToken claims
	newRefreshToken, err := u.tokenizer.GenerateFromClaims(claims, u.refreshTime, refreshSubject, sessionID+":"+sessionKey)
	if err != nil {
		return false, nil, err
	}

	// generate a new sessionToken from the refreshToken claims
	sessionToken, err := u.tokenizer.GenerateFromClaims(claims, u.refreshTime, sessionSubject, sessionID)
	if err != nil {
		return false, nil, err
	}

	// set the session id and key into cache
	if err := u.UpdateSessionKey(sessionID, sessionKey, time.Duration(u.refreshTime*b1)); err != nil {
		return false, nil, err
	}

	return true, &resUserAuth{
		Valid:        true,
		RefreshToken: newRefreshToken,
		SessionToken: sessionToken,
		Claims:       claims,
	}, nil
}

// Logout removes the user session in cache
func (u *userService) Logout(refreshToken string) (bool, error) {
	sessionID := ""
	sessionKey := ""
	// if session_id is provided, is in cache, and is valid, set it as the sessionID
	// the session can be expired by time
	if ok, claims := u.tokenizer.GetClaims(refreshToken, refreshSubject); ok {
		if s := strings.Split(claims.Id, ":"); len(s) == 2 {
			if key, err := u.GetSessionKey(s[0]); err == nil {
				sessionID = s[0]
				sessionKey = key
			}
		}
	}

	if sessionID == "" {
		return false, governor.NewErrorUser("malformed refresh token", http.StatusUnauthorized, nil)
	}

	// check the refresh token
	validToken, _ := u.tokenizer.ValidateSkipTime(refreshToken, refreshSubject, sessionID+":"+sessionKey)
	if !validToken {
		return false, nil
	}

	// delete the session in cache
	if err := u.EndSession(sessionID); err != nil {
		return false, err
	}
	return true, nil
}
