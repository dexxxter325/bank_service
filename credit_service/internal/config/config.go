package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Rest    Rest
	MongoDb MongoDb
	Kafka   Kafka
}

type Rest struct {
	Port string
}

type MongoDb struct {
	Host             string
	Port             string
	Dbname           string
	CreditCollection string
	UserIDCollection string
	Username         string
	Password         string
}

type Kafka struct {
	Brokers string
	Topic   string
}

func InitConfig() (*Config, error) {
	return InitConfigByPath("config/local.yml")
}

func InitConfigByPath(configPath string) (*Config, error) {
	var cfg Config

	viper.SetConfigFile(configPath)

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
			UserIDCollection: viper.GetString("mongodb.userid_collection"),
			Username:         viper.GetString("mongodb.username"),
			Password:         viper.GetString("mongodb.password"),
		},
		Kafka: Kafka{
			Brokers: viper.GetString("kafka.brokers"),
			Topic:   viper.GetString("kafka.topic"),
		},
	}
	return &cfg, nil
}
