package config

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

const defaultConfigFile = "config.yaml"

func Load() (Server, error) {
	path := getConfigPath()

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return Server{}, err
	}
	var settings Server
	if err := v.Unmarshal(&settings); err != nil {
		return Server{}, err
	}
	return settings, nil
}

func getConfigPath() string {
	var config string
	flag.StringVar(&config, "c", "", "choose config file.")
	flag.Parse()
	if config != "" {
		return config
	}
	if config = os.Getenv("XIANYU_CONFIG"); config != "" {
		return config
	}
	return defaultConfigFile
}
