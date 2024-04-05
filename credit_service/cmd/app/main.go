package main

import (
	"bank/credit_service/internal/app"
	"bank/credit_service/internal/config"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	cfg, err := config.InitConfig()
	if err != nil {
		logger.Fatalf("init config failed:%s", err)
	}

	app.RunRest(cfg, logger)
}

//test
