package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Rest    Rest
	MongoDb MongoDb
}

type Rest struct {
	Port string
}

type MongoDb struct {
	Host             string
	Port             string
	Dbname           string
	CreditCollection string
	Username         string
	Password         string
}

func InitConfig() (*Config, error) {
	var cfg Config

	viper.SetConfigFile("config/local.yml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed:%s", err)
	}

	cfg = Config{
		Rest: Rest{
			Port: viper.GetString("rest.port"),
		},
		MongoDb: MongoDb{
			Host:             viper.GetString("mongodb.host"),
			Port:             viper.GetString("mongodb.port"),
			Dbname:           viper.GetString("mongodb.dbname"),
			CreditCollection: viper.GetString("mongodb.credit_collection"),
			Username:         viper.GetString("mongodb.username"),
			Password:         viper.GetString("mongodb.password"),
		},
	}
	return &cfg, nil
}
