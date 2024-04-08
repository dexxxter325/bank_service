package main

import (
	"bank/auth_service/internal/app"
	"bank/auth_service/internal/config"
	"github.com/sirupsen/logrus"
)

func main() {
	x := 10
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	cfg, err := config.InitConfig()
	if err != nil {
		logger.Fatalf("init config failed:%s", err)
	}

	app.RunGRPC(cfg, logger)
}
