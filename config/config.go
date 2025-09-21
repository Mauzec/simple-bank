package config

import (
	"time"

	"github.com/spf13/viper"
)

type TokenType string

type Config struct {
	DBSource   string `mapstructure:"DB_SOURCE"`
	ServerAddr string `mapstructure:"SERVER_ADDR"`

	TokenType         TokenType `mapstructure:"TOKEN_TYPE"`
	TokenSymmetricKey string    `mapstructure:"TOKEN_SYMMETRIC_KEY"`

	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

func LoadConfig(name, ext string, paths ...string) (Config, error) {
	for _, path := range paths {
		viper.AddConfigPath(path)
	}
	viper.SetConfigName(name)
	viper.SetConfigType(ext)

	viper.AutomaticEnv()
	_ = viper.BindEnv("DB_SOURCE")
	_ = viper.BindEnv("SERVER_ADDR")
	_ = viper.BindEnv("TOKEN_TYPE")
	_ = viper.BindEnv("TOKEN_SYMMETRIC_KEY")
	_ = viper.BindEnv("ACCESS_TOKEN_DURATION")
	err := viper.ReadInConfig()
	config := Config{}

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, err
		}
	}

	err = viper.Unmarshal(&config)
	return config, err
}
