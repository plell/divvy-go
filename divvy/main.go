package main

import (

	//   "strconv"

	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	core "github.com/plell/divvygo/divvy/core"
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

	// Make Routes
	core.MakeRoutes(e)
	// DB connect
	core.ConnectDB()
	// DB Automigrate
	core.MigrateUp()

	// Start server
	e.Logger.Fatal(e.Start(":8000"))
}
