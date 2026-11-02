package config

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestServerConfigFiles(t *testing.T) {
	for _, name := range []string{"config.yaml", "config.docker.yaml"} {
		t.Run(name, func(t *testing.T) {
			reader := viper.New()
			reader.SetConfigFile(filepath.Join("..", "..", name))
			require.NoError(t, reader.ReadInConfig())

			var settings Server
			require.NoError(t, reader.Unmarshal(&settings))
			require.NotEmpty(t, settings.Postgres.Host)
			require.NotEmpty(t, settings.Postgres.Database)
			require.NotEmpty(t, settings.Postgres.DSN())
			require.NotEmpty(t, settings.Log.Level)
			require.NotZero(t, settings.HTTP.Port)
		})
	}
}
