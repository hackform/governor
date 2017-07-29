package governor

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
)

const (
	banner = `
  __ _  _____   _____ _ __ _ __   ___  _ __
 / _' |/ _ \ \ / / _ \ '__| '_ \ / _ \| '__|
| (_| | (_) \ V /  __/ |  | | | | (_) | |
 \__. |\___/ \_/ \___|_|  |_| |_|\___/|_|
  __/ |
 |___/  %s

 %s
`
)

type (
	// Server is an http gateway
	Server struct {
		i          *echo.Echo
		log        *logrus.Logger
		h          *health
		s          *setup
		config     Config
		showBanner bool
	}
)

// New creates a new Server
func New(config Config) (*Server, error) {
	l := newLogger(config)
	l.Info("initialized logger")
	i := echo.New()
	l.Info("initialized server instance")
	i.HideBanner = true
	i.HTTPErrorHandler = errorHandler(i, l)
	l.Info("initialized error handling")
	i.Pre(middleware.RemoveTrailingSlash())
	if config.IsDebug() {
		i.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "time=${time_rfc3339}, method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
		}))
	}
	i.Use(middleware.BodyLimit("2M"))
	i.Use(middleware.CORS())
	i.Use(middleware.Recover())
	i.Use(middleware.Gzip())
	l.Info("initialized middleware")

	healthService := newHealth()
	if err := healthService.Mount(config, i.Group(config.BaseURL+"/healthz"), l); err != nil {
		return nil, err
	}
	setupService := newSetup()
	if err := setupService.Mount(config, i.Group(config.BaseURL+"/setupz"), l); err != nil {
		return nil, err
	}

	s := &Server{
		i:          i,
		log:        l,
		config:     config,
		h:          healthService,
		s:          setupService,
		showBanner: true,
	}

	l.Info("server instance created")
	return s, nil
}

// Start starts the server at the specified port
func (s *Server) Start() error {
	if s.showBanner {
		fmt.Printf(color.BlueString(banner+"\n"), color.GreenString(s.config.Version), "http server on "+color.RedString(":"+s.config.Port))
	}
	s.i.Logger.Fatal(s.i.Start(":" + s.config.Port))
	return nil
}

// Logger returns an instance to the logger
func (s *Server) Logger() *logrus.Logger {
	return s.log
}
