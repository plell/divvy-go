package main

import (

	//   "strconv"

	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	core "github.com/plell/divvygo/divvy/core"
	"github.com/robfig/cron/v3"
)

func main() {

	// Echo instance
	e := echo.New()

	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// e.AutoTLSManager.HostPolicy = autocert.HostWhitelist("api.jamwallet.app")
	// e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.IPExtractor = echo.ExtractIPFromXFFHeader()

	// Make Routes
	core.MakeRoutes(e)

	// DB connect
	core.ConnectDB()
	// DB Automigrate
	core.MigrateUp()

	c := cron.New()
	c.AddFunc("@every 24h", func() {
		core.DoChargeTransfersAndRefundsCron()
		core.DoFeeTransferToJamWalletCron()
		core.DoPodDeletionCron()
		fmt.Println("cron ran!")
	})
	c.Start()

	go core.RunWebsocketBroker()
	core.StartDNALogger()

	// Start server
	// fmt.Println("start http 8000 server!")
	e.Logger.Fatal(e.Start(":443"))

	// fmt.Println("start tls 443 server!")
	// e.Logger.Fatal(e.StartAutoTLS(":443"))

}
