package user

import (
	"context"
	"encoding/json"
	htmlTemplate "html/template"
	"net/http"
	"strconv"
	"time"
	"xorkevin.dev/governor"
	"xorkevin.dev/governor/service/kvstore"
	"xorkevin.dev/governor/service/mail"
	"xorkevin.dev/governor/service/msgqueue"
	"xorkevin.dev/governor/service/user/apikey"
	"xorkevin.dev/governor/service/user/approval/model"
	"xorkevin.dev/governor/service/user/gate"
	"xorkevin.dev/governor/service/user/model"
	"xorkevin.dev/governor/service/user/reset/model"
	"xorkevin.dev/governor/service/user/role"
	"xorkevin.dev/governor/service/user/role/invitation/model"
	"xorkevin.dev/governor/service/user/session/model"
	"xorkevin.dev/governor/service/user/token"
	"xorkevin.dev/governor/util/rank"
)

const (
	authRoutePrefix = "/auth"
)

const (
	// NewUserQueueID is emitted when a new user is created
	NewUserQueueID = "gov.user.new"
	// DeleteUserQueueID is emitted when a user is deleted
	DeleteUserQueueID = "gov.user.delete"
)

const (
	time5m     int64 = int64(5 * time.Minute / time.Second)
	time24h    int64 = int64(24 * time.Hour / time.Second)
	time6month int64 = time24h * 365 / 2
)

type (
	// User is a user management service
	User interface {
		GetByID(userid string) (*ResUserGet, error)
		CheckUserExists(userid string) (bool, error)
	}

	Service interface {
		governor.Service
		User
	}

	service struct {
		users             usermodel.Repo
		sessions          sessionmodel.Repo
		approvals         approvalmodel.Repo
		invitations       invitationmodel.Repo
		resets            resetmodel.Repo
		roles             role.Role
		apikeys           apikey.Apikey
		kvusers           kvstore.KVStore
		kvsessions        kvstore.KVStore
		queue             msgqueue.Msgqueue
		mailer            mail.Mail
		gate              gate.Gate
		tokenizer         token.Tokenizer
		logger            governor.Logger
		baseURL           string
		authURL           string
		accessTime        int64
		refreshTime       int64
		refreshCacheTime  int64
		confirmTime       int64
		passwordResetTime int64
		invitationTime    int64
		userCacheTime     int64
		newLoginEmail     bool
		passwordMinSize   int
		userApproval      bool
		rolesummary       rank.Rank
		emailurlbase      string
		tplemailchange    *htmlTemplate.Template
		tplforgotpass     *htmlTemplate.Template
		tplnewuser        *htmlTemplate.Template
	}

	router struct {
		s service
	}

	// NewUserProps are properties of a newly created user
	NewUserProps struct {
		Userid       string `json:"userid"`
		Username     string `json:"username"`
		Email        string `json:"email"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		CreationTime int64  `json:"creation_time"`
	}

	// DeleteUserProps are properties of a deleted user
	DeleteUserProps struct {
		Userid string `json:"userid"`
	}

	ctxKeyUser struct{}
)

// GetCtxUser returns a User service from the context
func GetCtxUser(inj governor.Injector) User {
	v := inj.Get(ctxKeyUser{})
	if v == nil {
		return nil
	}
	return v.(User)
}

// setCtxUser sets a User service in the context
func setCtxUser(inj governor.Injector, u User) {
	inj.Set(ctxKeyUser{}, u)
}

// NewCtx creates a new User service from a context
func NewCtx(inj governor.Injector) Service {
	users := usermodel.GetCtxRepo(inj)
	sessions := sessionmodel.GetCtxRepo(inj)
	approvals := approvalmodel.GetCtxRepo(inj)
	invitations := invitationmodel.GetCtxRepo(inj)
	resets := resetmodel.GetCtxRepo(inj)
	roles := role.GetCtxRole(inj)
	apikeys := apikey.GetCtxApikey(inj)
	kv := kvstore.GetCtxKVStore(inj)
	queue := msgqueue.GetCtxMsgqueue(inj)
	mailer := mail.GetCtxMail(inj)
	tokenizer := token.GetCtxTokenizer(inj)
	g := gate.GetCtxGate(inj)

	return New(
		users,
		sessions,
		approvals,
		invitations,
		resets,
		roles,
		apikeys,
		kv,
		queue,
		mailer,
		tokenizer,
		g,
	)
}

// New creates a new User service
func New(
	users usermodel.Repo,
	sessions sessionmodel.Repo,
	approvals approvalmodel.Repo,
	invitations invitationmodel.Repo,
	resets resetmodel.Repo,
	roles role.Role,
	apikeys apikey.Apikey,
	kv kvstore.KVStore,
	queue msgqueue.Msgqueue,
	mailer mail.Mail,
	tokenizer token.Tokenizer,
	g gate.Gate,
) Service {
	return &service{
		users:             users,
		sessions:          sessions,
		approvals:         approvals,
		invitations:       invitations,
		resets:            resets,
		roles:             roles,
		apikeys:           apikeys,
		kvusers:           kv.Subtree("users"),
		kvsessions:        kv.Subtree("sessions"),
		queue:             queue,
		mailer:            mailer,
		gate:              g,
		tokenizer:         tokenizer,
		accessTime:        time5m,
		refreshTime:       time6month,
		refreshCacheTime:  time24h,
		confirmTime:       time24h,
		passwordResetTime: time24h,
		invitationTime:    time24h,
		userCacheTime:     time24h,
	}
}

func (s *service) Register(inj governor.Injector, r governor.ConfigRegistrar, jr governor.JobRegistrar) {
	setCtxUser(inj, s)

	r.SetDefault("accesstime", "5m")
	r.SetDefault("refreshtime", "4380h")
	r.SetDefault("refreshcache", "24h")
	r.SetDefault("confirmtime", "24h")
	r.SetDefault("passwordresettime", "24h")
	r.SetDefault("invitationtime", "24h")
	r.SetDefault("usercachetime", "24h")
	r.SetDefault("newloginemail", true)
	r.SetDefault("passwordminsize", 8)
	r.SetDefault("userapproval", false)
	r.SetDefault("rolesummary", []string{rank.TagUser, rank.TagAdmin})
	r.SetDefault("email.url.base", "http://localhost:8080")
	r.SetDefault("email.url.emailchange", "/a/confirm/email?key={{.Userid}}.{{.Key}}")
	r.SetDefault("email.url.forgotpass", "/x/resetpass?key={{.Userid}}.{{.Key}}")
	r.SetDefault("email.url.newuser", "/x/confirm?userid={{.Userid}}&key={{.Key}}")
}

func (s *service) router() *router {
	return &router{
		s: *s,
	}
}

func (s *service) Init(ctx context.Context, c governor.Config, r governor.ConfigReader, l governor.Logger, m governor.Router) error {
	s.logger = l
	l = s.logger.WithData(map[string]string{
		"phase": "init",
	})

	s.baseURL = c.BaseURL
	s.authURL = c.BaseURL + r.URL() + authRoutePrefix
	if t, err := time.ParseDuration(r.GetStr("accesstime")); err != nil {
		return governor.NewError("Failed to parse access time", http.StatusBadRequest, err)
	} else {
		s.accessTime = int64(t / time.Second)
	}
	if t, err := time.ParseDuration(r.GetStr("refreshtime")); err != nil {
		return governor.NewError("Failed to parse refresh time", http.StatusBadRequest, err)
	} else {
		s.refreshTime = int64(t / time.Second)
	}
	if t, err := time.ParseDuration(r.GetStr("refreshcache")); err != nil {
		return governor.NewError("Failed to parse refresh cache", http.StatusBadRequest, err)
	} else {
		s.refreshCacheTime = int64(t / time.Second)
	}
	if t, err := time.ParseDuration(r.GetStr("confirmtime")); err != nil {
		return governor.NewError("Failed to parse confirm time", http.StatusBadRequest, err)
	} else {
		s.confirmTime = int64(t / time.Second)
	}
	if t, err := time.ParseDuration(r.GetStr("passwordresettime")); err != nil {
		return governor.NewError("Failed to parse password reset time", http.StatusBadRequest, err)
	} else {
		s.passwordResetTime = int64(t / time.Second)
	}
	if t, err := time.ParseDuration(r.GetStr("invitationtime")); err != nil {
		return governor.NewError("Failed to parse role invitation time", http.StatusBadRequest, err)
	} else {
		s.invitationTime = int64(t / time.Second)
	}
	if t, err := time.ParseDuration(r.GetStr("usercachetime")); err != nil {
		return governor.NewError("Failed to parse user cache time", http.StatusBadRequest, err)
	} else {
		s.userCacheTime = int64(t / time.Second)
	}
	s.newLoginEmail = r.GetBool("newloginemail")
	s.passwordMinSize = r.GetInt("passwordminsize")
	s.userApproval = r.GetBool("userapproval")
	s.rolesummary = rank.FromSlice(r.GetStrSlice("rolesummary"))

	s.emailurlbase = r.GetStr("email.url.base")
	if t, err := htmlTemplate.New("email.url.emailchange").Parse(r.GetStr("email.url.emailchange")); err != nil {
		return governor.NewError("Failed to parse email change url template", http.StatusBadRequest, err)
	} else {
		s.tplemailchange = t
	}
	if t, err := htmlTemplate.New("email.url.forgotpass").Parse(r.GetStr("email.url.forgotpass")); err != nil {
		return governor.NewError("Failed to parse forgot pass url template", http.StatusBadRequest, err)
	} else {
		s.tplforgotpass = t
	}
	if t, err := htmlTemplate.New("email.url.newuser").Parse(r.GetStr("email.url.newuser")); err != nil {
		return governor.NewError("Failed to parse new user url template", http.StatusBadRequest, err)
	} else {
		s.tplnewuser = t
	}

	l.Info("loaded config", map[string]string{
		"accesstime (s)":        strconv.FormatInt(s.accessTime, 10),
		"refreshtime (s)":       strconv.FormatInt(s.refreshTime, 10),
		"refreshcache (s)":      strconv.FormatInt(s.refreshCacheTime, 10),
		"confirmtime (s)":       strconv.FormatInt(s.confirmTime, 10),
		"passwordresettime (s)": strconv.FormatInt(s.passwordResetTime, 10),
		"invitationtime (s)":    strconv.FormatInt(s.invitationTime, 10),
		"usercachetime (s)":     strconv.FormatInt(s.userCacheTime, 10),
		"newloginemail":         strconv.FormatBool(s.newLoginEmail),
		"passwordminsize":       strconv.Itoa(s.passwordMinSize),
		"issuer":                r.GetStr("issuer"),
		"userapproval":          strconv.FormatBool(s.userApproval),
		"rolesummary":           s.rolesummary.String(),
		"tplemailchange":        r.GetStr("email.url.emailchange"),
		"tplforgotpass":         r.GetStr("email.url.forgotpass"),
		"tplnewuser":            r.GetStr("email.url.newuser"),
	})

	sr := s.router()
	sr.mountRoute(m.Group("/user"))
	sr.mountAuth(m.Group(authRoutePrefix))
	sr.mountApikey(m.Group("/apikey"))
	l.Info("mounted http routes", nil)
	return nil
}

func (s *service) Setup(req governor.ReqSetup) error {
	l := s.logger.WithData(map[string]string{
		"phase": "setup",
	})

	madmin, err := s.users.New(req.Username, req.Password, req.Email, req.Firstname, req.Lastname)
	if err != nil {
		return err
	}

	if err := s.users.Setup(); err != nil {
		return err
	}
	l.Info("created user table", nil)

	if err := s.sessions.Setup(); err != nil {
		return err
	}
	l.Info("created usersessions table", nil)

	if err := s.approvals.Setup(); err != nil {
		return err
	}
	l.Info("created userapprovals table", nil)

	if err := s.invitations.Setup(); err != nil {
		return err
	}
	l.Info("created userroleinvitations table", nil)

	if err := s.resets.Setup(); err != nil {
		return err
	}
	l.Info("created userresets table", nil)

	b, err := json.Marshal(NewUserProps{
		Userid:       madmin.Userid,
		Username:     madmin.Username,
		Email:        madmin.Email,
		FirstName:    madmin.FirstName,
		LastName:     madmin.LastName,
		CreationTime: madmin.CreationTime,
	})
	if err != nil {
		return governor.NewError("Failed to encode admin user props to json", http.StatusInternalServerError, err)
	}

	if err := s.users.Insert(madmin); err != nil {
		return err
	}
	if err := s.roles.InsertRoles(madmin.Userid, rank.Admin()); err != nil {
		return err
	}

	if err := s.queue.Publish(NewUserQueueID, b); err != nil {
		s.logger.Error("Failed to publish new user", map[string]string{
			"error":      err.Error(),
			"actiontype": "publishadminuser",
		})
	}

	l.Info("inserted new setup admin", map[string]string{
		"username": madmin.Username,
		"userid":   madmin.Userid,
	})
	return nil
}

func (s *service) Start(ctx context.Context) error {
	return nil
}

func (s *service) Stop(ctx context.Context) {
}

func (s *service) Health() error {
	return nil
}
