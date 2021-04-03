package core

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func Pong(c echo.Context) error {
	return c.String(http.StatusOK, "Pong")
}

func AbstractError(c echo.Context) error {
	return c.String(http.StatusInternalServerError, "")
}

var pool = "abcdefghijklmnopqrstuvwxyzABCEFGHIJKLMNOPQRSTUVWXYZ:|?$%@][{}#&/()*1234567890"

func MakeSelector(tableName string) string {
	rand.Seed(time.Now().UnixNano())
	l := 24
	bytes := make([]byte, l)

	randomSelector := ""
	// enter while loop, exited when n = 2
	n := 0
	for n < 1 {
		// create random string
		for i := 0; i < l; i++ {
			bytes[i] = pool[rand.Intn(len(pool))]
		}

		randomSelector = string(bytes)
		selector := Selector{}

		// create record in selectors to make sure only unique selector are made
		result := DB.Table(tableName).Where("selector = ?", randomSelector).First(&selector)
		if result.Error != nil {
			// good, this is a unique selector
			selector := Selector{
				Selector: randomSelector,
				Type:     tableName,
			}
			result := DB.Create(&selector) // pass pointer of data to Create
			if result.Error != nil {
				// db create failed
			}
			// leave loop
			log.Println("Made unique selector")
			n++
		} else {
			log.Println("Made duplicate selector, retry")
		}
	}

	return randomSelector
}
