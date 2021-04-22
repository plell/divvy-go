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
	// "gorm.io/driver/mysql"
	// "gorm.io/gorm"
)

func main() {

	// Echo instance
	e := echo.New()

	// Load .env
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

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

	// set stripe webhooks
	// stripe listen --forward-to http://localhost:8000/webhook
	e.GET("/webhook", echo.HandlerFunc(core.HandleStripeWebhook))

	c := cron.New()
	// c.AddFunc("0 30 * * * *", func() { fmt.Println("Every hour on the half hour") })
	// c.AddFunc("@hourly", func() { fmt.Println("Every hour") })
	c.AddFunc("@every 1m", func() {
		core.DoChargeTransfersAndRefundsCron()
		fmt.Println("Every minute!")
	})
	c.Start()

	// Start server
	e.Logger.Fatal(e.Start(":8000"))
}
