package main

import (

	//   "strconv"

	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	DB "github.com/plell/divvygo/divvy/database"
	routes "github.com/plell/divvygo/divvy/routes"
	// "gorm.io/driver/mysql"
	// "gorm.io/gorm"
)

// var mySigningKey = os.Get("MY_JWT_TOKEN") // get this from env
var mySigningKey = []byte("mysecretphrase")

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
	routes.MakeRoutes(e)
	// DB connect
	DB.Connect()
	// DB Automigrate
	DB.MigrateUp()

	// Start server
	e.Logger.Fatal(e.Start(":8000"))
}
