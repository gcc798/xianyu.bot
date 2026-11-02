package config

type Server struct {
	JWT      JWT      `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Log      Log      `mapstructure:"log" json:"log" yaml:"log"`
	Email    Email    `mapstructure:"email" json:"email" yaml:"email"`
	HTTP     HTTP     `mapstructure:"http" json:"http" yaml:"http"`
	Postgres Postgres `mapstructure:"postgres" json:"postgres" yaml:"postgres"`
}
