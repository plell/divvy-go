package main

import (

	//   "strconv"

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

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	routes.MakeRoutes(e)

	// db connection with gorm
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	DB.Connect()
	DB.MigrateUp()

	// Start server
	e.Logger.Fatal(e.Start(":8000"))
}
