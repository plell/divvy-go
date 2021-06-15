package core

import (
	"log"
	"os"

	"github.com/logdna/logdna-go/logger"
)

var Logger *logger.Logger

func StartDNALogger() {
	logDNAKey := os.Getenv("LOG_DNA_KEY")
	options := logger.Options{}
	// options.Level = "fatal"
	options.Hostname = "gotest"
	options.App = "myapp"
	// options.IPAddress = "10.0.1.101"
	// options.MacAddress = "C0:FF:EE:C0:FF:EE"
	options.Env = "test"
	options.Tags = "logging,golang"
	Logger, err := logger.NewLogger(options, logDNAKey)
	if err != nil {
		log.Println("error!")
		log.Println(err)
	}

	Logger.Debug("Start logging!")
	Logger.Close()
}

func LogInfo(content string) {
	logDNAKey := os.Getenv("LOG_DNA_KEY")
	options := logger.Options{}
	// options.Level = "fatal"
	options.Hostname = "gotest"
	options.App = "myapp"
	// options.IPAddress = "10.0.1.101"
	// options.MacAddress = "C0:FF:EE:C0:FF:EE"
	options.Env = "test"
	options.Tags = "logging,golang"
	Logger, err := logger.NewLogger(options, logDNAKey)
	if err != nil {
		log.Println("error!")
		log.Println(err)
	}

	Logger.Info(content)
	Logger.Close()
}
