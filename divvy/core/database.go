package core

import (

	//   "strconv"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	// "gorm.io/driver/mysql"
	// "gorm.io/gorm"
)

// DB obj
var DB *gorm.DB

func ConnectDB() {
	// db connection with gorm
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	// "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	// dbURL := os.Getenv("DATABASE_URL")
	dsn := "root:password@tcp(127.0.0.1:3306)/divvy?charset=utf8mb4&parseTime=True&loc=Local" //dbURL
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	// db.LogMode(true)
	if err != nil {
		fmt.Println("DB ERROR:")
		panic(err)
	}

	DB = db
}
