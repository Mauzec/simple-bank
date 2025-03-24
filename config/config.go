package config

import "github.com/spf13/viper"

type Config struct {
	DBSource   string `mapstructure:"DB_SOURCE"`
	ServerAddr string `mapstructure:"SERVER_ADDR"`
}

func LoadConfig(name, ext string, paths ...string) (Config, error) {
	for _, path := range paths {
		viper.AddConfigPath(path)
	}
	viper.SetConfigName(name)
	viper.SetConfigType(ext)

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	config := Config{}

	if err != nil {
		return config, err
	}

	err = viper.Unmarshal(&config)
	return config, err
}
