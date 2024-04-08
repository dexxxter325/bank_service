package main

import (
	"bank/auth_service/internal/app"
	"bank/auth_service/internal/config"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	file, err := os.Open("non_existent_file.txt") // Ошибка открытия файла не обрабатывается
	defer file.Close()
	if err != nil {
		panic(err)
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	cfg, err := config.InitConfig()
	if err != nil {
		logger.Fatalf("init config failed:%s", err)
	}

	app.RunGRPC(cfg, logger)
}
