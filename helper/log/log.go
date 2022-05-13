package log

import (
	"github.com/crazygit/BinanceMarketMonitor/helper"
	"github.com/sirupsen/logrus"
	"os"
)

var log = logrus.New()

func init() {
	if helper.IsProductionEnvironment() {
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetLevel(logrus.DebugLevel)
	}
	log.SetOutput(os.Stdout)
}

func GetLog() *logrus.Logger {
	return log
}
