package config

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// TestServerConfigFiles 验证对应测试场景。
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
			require.Equal(t, 60, settings.MCP.HealthCheckIntervalSeconds)
			require.Positive(t, settings.MCP.CatalogRefreshConcurrency)
			require.NotEmpty(t, settings.RabbitMQ.URL)
			require.Positive(t, settings.RabbitMQ.Prefetch)
			require.Positive(t, settings.RabbitMQ.PublishTimeoutSeconds)
			require.Positive(t, settings.Worker.HealthPort)
		})
	}
}

// TestLoadSensitiveValuesFromEnvironment 验证敏感配置可以脱离仓库文件注入。
func TestLoadSensitiveValuesFromEnvironment(t *testing.T) {
	t.Setenv("XIANYU_CONFIG", filepath.Join("..", "..", "config.yaml"))
	t.Setenv("XIANYU_JWT_SIGNING_KEY", "test-signing-key")
	t.Setenv("XIANYU_POSTGRES_PASSWORD", "test-database-password")
	t.Setenv("XIANYU_RABBITMQ_URL", "amqp://test:test@127.0.0.1:5672/")

	settings, err := Load()
	require.NoError(t, err)
	require.Equal(t, "test-signing-key", settings.JWT.SigningKey)
	require.Equal(t, "test-database-password", settings.Postgres.Password)
	require.Equal(t, "amqp://test:test@127.0.0.1:5672/", settings.RabbitMQ.URL)
}
