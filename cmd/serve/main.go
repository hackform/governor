package main

import (
	"github.com/hackform/governor"
	"github.com/hackform/governor/service/conf"
	"github.com/hackform/governor/service/health"
	"github.com/hackform/governor/service/user"
	_ "github.com/lib/pq"
)

func main() {
	g := governor.New(governor.NewConfig())

	hS := health.New()
	cS := conf.New()
	uS := user.New()

	g.MountRoute("/api/health", hS)
	g.MountRoute("/api/conf", cS)
	g.MountRoute("/api/u", uS)
	g.Start(8080)
}
