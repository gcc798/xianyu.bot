package config

type Log struct {
	Level    string `mapstructure:"level" json:"level" yaml:"level"`
	Format   string `mapstructure:"format" json:"format" yaml:"format"`
	ShowLine bool   `mapstructure:"show-line" json:"show-line" yaml:"show-line"`
}
