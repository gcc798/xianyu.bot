package config

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
