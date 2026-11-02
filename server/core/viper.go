package core

import (
	"flag"
	"fmt"
	"os"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/spf13/viper"
)

const defaultConfigFile = "config.yaml"

// Viper 配置
func Viper() {
	config := getConfigPath()

	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	if err = v.Unmarshal(&global.GVA_CONFIG); err != nil {
		panic(fmt.Errorf("fatal error unmarshal config: %w", err))
	}
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
