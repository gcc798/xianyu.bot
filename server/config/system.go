package config

type System struct {
	DbType       string `mapstructure:"db-type" json:"db-type" yaml:"db-type"`
	RouterPrefix string `mapstructure:"router-prefix" json:"router-prefix" yaml:"router-prefix"`
	Addr         int    `mapstructure:"addr" json:"addr" yaml:"addr"`
}
