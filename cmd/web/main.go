package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pochtalion"
	"pochtalion/mailgun"
	"pochtalion/web"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

var addrFlag = flag.String("addr", ":80", "liste http addr")
var dbFlag = flag.String("db", "_database.db", "database file")
var loginFlag = flag.String("uname", "dev", "http auth")
var pwdFlag = flag.String("upwd", "PFC", "http auth")
var tplsFlag = flag.String("tpls", "./", "templates")

var MAILGUN_DOMAIN = os.Getenv("MAILGUN_DOMAIN")
var MAILGUN_APIKEY = os.Getenv("MAILGUN_APIKEY")
var MAILGUN_APIPKEY = os.Getenv("MAILGUN_APIPKEY")

func main() {
	flag.Parse()

	log.Println("main", "Start pochtalion, flags:")
	log.Println("main", "Addr", *addrFlag)
	log.Println("main", "DB", *dbFlag)
	log.Println("main", "Login", *loginFlag)
	log.Println("main", "PWD", *pwdFlag)
	log.Println("main", "Sender Domain", MAILGUN_DOMAIN)

	e := echo.New()
	e.Use(middleware.BasicAuth(func(username, password string) bool {
		if username == *loginFlag && password == *pwdFlag {
			return true
		}
		return false
	}))

	web.SetupRouting(e)
	pochtalion.Sender = mailgun.Newpochtalion(
		MAILGUN_DOMAIN,
		MAILGUN_APIKEY,
		MAILGUN_APIPKEY)

	go func() {
		log.Println("main", "running http server at", *addrFlag)
		e.Run(standard.New(*addrFlag))
	}()

	// ---------------------------
	// run listener of OS
	// ---------------------------

	osSignal := make(chan os.Signal, 2)
	close := make(chan struct{})
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-osSignal
		log.Println("main", "signal completion of the process")

		// TODO: shutdown...

		close <- struct{}{}
	}()
	<-close

	os.Exit(0)
}
