package config

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

const defaultConfigFile = "config.yaml"

type Server struct {
	JWT      JWT      `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Log      Log      `mapstructure:"log" json:"log" yaml:"log"`
	Email    Email    `mapstructure:"email" json:"email" yaml:"email"`
	HTTP     HTTP     `mapstructure:"http" json:"http" yaml:"http"`
	Postgres Postgres `mapstructure:"postgres" json:"postgres" yaml:"postgres"`
}

type JWT struct {
	SigningKey  string `mapstructure:"signing-key" json:"signing-key" yaml:"signing-key"`
	ExpiresTime string `mapstructure:"expires-time" json:"expires-time" yaml:"expires-time"`
	BufferTime  string `mapstructure:"buffer-time" json:"buffer-time" yaml:"buffer-time"`
	Issuer      string `mapstructure:"issuer" json:"issuer" yaml:"issuer"`
}

type Log struct {
	Level    string `mapstructure:"level" json:"level" yaml:"level"`
	Format   string `mapstructure:"format" json:"format" yaml:"format"`
	ShowLine bool   `mapstructure:"show-line" json:"show-line" yaml:"show-line"`
}

type Email struct {
	To          string `mapstructure:"to" json:"to" yaml:"to"`
	From        string `mapstructure:"from" json:"from" yaml:"from"`
	Host        string `mapstructure:"host" json:"host" yaml:"host"`
	Secret      string `mapstructure:"secret" json:"secret" yaml:"secret"`
	Nickname    string `mapstructure:"nickname" json:"nickname" yaml:"nickname"`
	Port        int    `mapstructure:"port" json:"port" yaml:"port"`
	IsSSL       bool   `mapstructure:"is-ssl" json:"is-ssl" yaml:"is-ssl"`
	IsLoginAuth bool   `mapstructure:"is-loginauth" json:"is-loginauth" yaml:"is-loginauth"`
}

type HTTP struct {
	Port int `mapstructure:"port" json:"port" yaml:"port"`
}

type Postgres struct {
	Host            string `mapstructure:"host" json:"host" yaml:"host"`
	Port            string `mapstructure:"port" json:"port" yaml:"port"`
	Database        string `mapstructure:"database" json:"database" yaml:"database"`
	Username        string `mapstructure:"username" json:"username" yaml:"username"`
	Password        string `mapstructure:"password" json:"password" yaml:"password"`
	Parameters      string `mapstructure:"parameters" json:"parameters" yaml:"parameters"`
	LogLevel        string `mapstructure:"log-level" json:"log-level" yaml:"log-level"`
	MaxIdleConns    int    `mapstructure:"max-idle-conns" json:"max-idle-conns" yaml:"max-idle-conns"`
	MaxOpenConns    int    `mapstructure:"max-open-conns" json:"max-open-conns" yaml:"max-open-conns"`
	ConnMaxLifetime int    `mapstructure:"conn-max-lifetime" json:"conn-max-lifetime" yaml:"conn-max-lifetime"`
}

func (p Postgres) DSN() string {
	return "host=" + p.Host + " user=" + p.Username + " password=" + p.Password + " dbname=" + p.Database + " port=" + p.Port + " " + p.Parameters
}

func Load() (Server, error) {
	v := viper.New()
	v.SetConfigFile(configPath())
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

func configPath() string {
	var path string
	flag.StringVar(&path, "c", "", "choose config file")
	flag.Parse()
	if path != "" {
		return path
	}
	if path = os.Getenv("XIANYU_CONFIG"); path != "" {
		return path
	}
	return defaultConfigFile
}
