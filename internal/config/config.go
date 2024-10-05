package config

import "github.com/spf13/viper"

type Config struct {
	DebugLevel string `mapstructure:"debugLevel"`
	ServerPort string `mapstructure:"serverPort"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return &Config{}, err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return &Config{}, err
	}

	return &cfg, err
}
