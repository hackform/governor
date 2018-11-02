package governor

import (
	"github.com/labstack/echo"
	"net/http"
	"regexp"
)

const (
	moduleIDSetup = "govsetup"
)

type (
	setup struct {
		services []Service
	}
)

// New creates a new Setup service
func newSetup() *setup {
	return &setup{
		services: []Service{},
	}
}

type (
	// ReqSetupPost is the http post request for the setup
	ReqSetupPost struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		Email     string `json:"email"`
		Firstname string `json:"first_name"`
		Lastname  string `json:"last_name"`
		Orgname   string `json:"orgname"`
	}
)

var (
	emailRegex = regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]+$`)
)

func (r *ReqSetupPost) valid() *Error {
	if len(r.Username) < 3 {
		return NewErrorUser(moduleIDSetup, "username must be longer than 2 chars", 0, http.StatusBadRequest)
	}
	if len(r.Password) < 10 {
		return NewErrorUser(moduleIDSetup, "password must be longer than 9 chars", 0, http.StatusBadRequest)
	}
	if !emailRegex.MatchString(r.Email) {
		return NewErrorUser(moduleIDSetup, "email is invalid", 0, http.StatusBadRequest)
	}
	if len(r.Orgname) == 0 {
		return NewErrorUser(moduleIDSetup, "organization name must be provided", 0, http.StatusBadRequest)
	}
	return nil
}

type (
	responseSetupPost struct {
		Username  string `json:"username"`
		Firstname string `json:"first_name"`
		Lastname  string `json:"last_name"`
		Version   string `json:"version"`
		Orgname   string `json:"orgname"`
	}
)

var (
	setupRun = false
)

// Mount is a collection of routes
func (s *setup) Mount(conf Config, l Logger, r *echo.Group) error {
	r.POST("", func(c echo.Context) error {
		if setupRun {
			return NewError(moduleIDSetup, "setup already run", 128, http.StatusForbidden)
		}

		rsetup := &ReqSetupPost{}
		if err := c.Bind(rsetup); err != nil {
			return NewErrorUser(moduleIDSetup, err.Error(), 0, http.StatusBadRequest)
		}
		if err := rsetup.valid(); err != nil {
			return err
		}

		for _, service := range s.services {
			if err := service.Setup(conf, l, *rsetup); err != nil {
				if err.Code() == 128 {
					setupRun = true
				}
				err.AddTrace(moduleIDSetup)
				return err
			}
		}

		l.Info("successfully setup services", moduleIDSetup, "services setup", 0, nil)

		return c.JSON(http.StatusCreated, &responseSetupPost{
			Username:  rsetup.Username,
			Firstname: rsetup.Firstname,
			Lastname:  rsetup.Lastname,
			Version:   conf.Version,
			Orgname:   rsetup.Orgname,
		})
	})

	l.Info("mounted setup service", moduleIDSetup, "mount setup service", 0, nil)
	return nil
}

func (s *setup) addService(service Service) {
	s.services = append(s.services, service)
}
