package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	GRPC     GRPC
	Postgres Postgres
	Auth     Auth
	Kafka    Kafka
}

type GRPC struct {
	Port string
}

type Postgres struct {
	Host     string
	Port     string
	User     string
	DbName   string
	Password string
	Sslmode  string
}

type Auth struct {
	SecretKey       string
	AccessTokenTTL  string
	RefreshTokenTTL string
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
		return nil, fmt.Errorf("read config file failed:%s", err)
	}

	cfg = Config{
		GRPC: GRPC{
			Port: viper.GetString("grpc.port"),
		},
		Postgres: Postgres{
			Host:     viper.GetString("postgres.host"),
			Port:     viper.GetString("postgres.port"),
			User:     viper.GetString("postgres.user"),
			DbName:   viper.GetString("postgres.dbName"),
			Password: viper.GetString("postgres.password"),
			Sslmode:  viper.GetString("postgres.sslmode"),
		},
		Auth: Auth{
			SecretKey:       viper.GetString("auth.secretKey"),
			AccessTokenTTL:  viper.GetString("auth.accessTokenTTl"),
			RefreshTokenTTL: viper.GetString("auth.refreshTokenTTl"),
		},
		Kafka: Kafka{
			Brokers: viper.GetString("kafka.brokers"),
			Topic:   viper.GetString("kafka.topic"),
		},
	}

	return &cfg, nil

}
