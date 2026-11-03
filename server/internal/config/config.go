package config

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

// defaultConfigFile 是未显式指定路径时读取的配置文件。
const defaultConfigFile = "config.yaml"

// Server 汇总服务启动所需的全部配置。
type Server struct {
	JWT      JWT      `mapstructure:"jwt" json:"jwt" yaml:"jwt"`                // JWT 控制登录令牌的签名与有效期。
	Log      Log      `mapstructure:"log" json:"log" yaml:"log"`                // Log 控制应用日志的级别与输出格式。
	Email    Email    `mapstructure:"email" json:"email" yaml:"email"`          // Email 配置邮件发送服务器与身份信息。
	HTTP     HTTP     `mapstructure:"http" json:"http" yaml:"http"`             // HTTP 配置 Web 服务监听参数。
	Postgres Postgres `mapstructure:"postgres" json:"postgres" yaml:"postgres"` // Postgres 配置业务数据库连接池。
}

// JWT 描述登录令牌的签发策略。
type JWT struct {
	SigningKey  string `mapstructure:"signing-key" json:"signing-key" yaml:"signing-key"`    // SigningKey 是 HMAC 令牌签名密钥。
	ExpiresTime string `mapstructure:"expires-time" json:"expires-time" yaml:"expires-time"` // ExpiresTime 是令牌有效期，支持 Go duration 与天数写法。
	BufferTime  string `mapstructure:"buffer-time" json:"buffer-time" yaml:"buffer-time"`    // BufferTime 是令牌到期前允许刷新的缓冲期。
	Issuer      string `mapstructure:"issuer" json:"issuer" yaml:"issuer"`                   // Issuer 写入令牌签发方声明。
}

// Log 描述 Zap 日志输出策略。
type Log struct {
	Level    string `mapstructure:"level" json:"level" yaml:"level"`             // Level 是最低输出级别。
	Format   string `mapstructure:"format" json:"format" yaml:"format"`          // Format 选择 json 或控制台格式。
	ShowLine bool   `mapstructure:"show-line" json:"show-line" yaml:"show-line"` // ShowLine 决定日志是否记录调用位置。
}

// Email 描述 SMTP 发信配置。
type Email struct {
	To          string `mapstructure:"to" json:"to" yaml:"to"`                               // To 是默认收件人列表，多个地址以逗号分隔。
	From        string `mapstructure:"from" json:"from" yaml:"from"`                         // From 是发件邮箱及认证用户名。
	Host        string `mapstructure:"host" json:"host" yaml:"host"`                         // Host 是 SMTP 服务器主机名。
	Secret      string `mapstructure:"secret" json:"secret" yaml:"secret"`                   // Secret 是 SMTP 认证凭据。
	Nickname    string `mapstructure:"nickname" json:"nickname" yaml:"nickname"`             // Nickname 是邮件中展示的发件人名称。
	Port        int    `mapstructure:"port" json:"port" yaml:"port"`                         // Port 是 SMTP 服务端口。
	IsSSL       bool   `mapstructure:"is-ssl" json:"is-ssl" yaml:"is-ssl"`                   // IsSSL 决定是否直接使用 TLS 连接。
	IsLoginAuth bool   `mapstructure:"is-loginauth" json:"is-loginauth" yaml:"is-loginauth"` // IsLoginAuth 决定是否使用 LOGIN 认证机制。
}

// HTTP 描述 Web 服务监听配置。
type HTTP struct {
	Port int `mapstructure:"port" json:"port" yaml:"port"` // Port 是 HTTP 监听端口。
}

// Postgres 描述 PostgreSQL 连接与连接池配置。
type Postgres struct {
	Host            string `mapstructure:"host" json:"host" yaml:"host"`                                        // Host 是数据库主机名。
	Port            string `mapstructure:"port" json:"port" yaml:"port"`                                        // Port 是数据库端口。
	Database        string `mapstructure:"database" json:"database" yaml:"database"`                            // Database 是业务数据库名称。
	Username        string `mapstructure:"username" json:"username" yaml:"username"`                            // Username 是数据库登录用户。
	Password        string `mapstructure:"password" json:"password" yaml:"password"`                            // Password 是数据库登录密码。
	Parameters      string `mapstructure:"parameters" json:"parameters" yaml:"parameters"`                      // Parameters 是附加到 DSN 的 PostgreSQL 参数。
	LogLevel        string `mapstructure:"log-level" json:"log-level" yaml:"log-level"`                         // LogLevel 是 GORM 的 SQL 日志级别。
	MaxIdleConns    int    `mapstructure:"max-idle-conns" json:"max-idle-conns" yaml:"max-idle-conns"`          // MaxIdleConns 是连接池最大空闲连接数。
	MaxOpenConns    int    `mapstructure:"max-open-conns" json:"max-open-conns" yaml:"max-open-conns"`          // MaxOpenConns 是连接池最大连接数。
	ConnMaxLifetime int    `mapstructure:"conn-max-lifetime" json:"conn-max-lifetime" yaml:"conn-max-lifetime"` // ConnMaxLifetime 是连接最大存活秒数。
}

// DSN 生成供 PostgreSQL 驱动使用的连接字符串。
func (p Postgres) DSN() string {
	return "host=" + p.Host + " user=" + p.Username + " password=" + p.Password + " dbname=" + p.Database + " port=" + p.Port + " " + p.Parameters
}

// Load 从命令行、环境变量或默认路径加载服务配置。
func Load() (Server, error) {
	v := viper.New()
	v.SetConfigFile(configPath())
	v.SetConfigType("yaml")
	if err := v.BindEnv("jwt.signing-key", "XIANYU_JWT_SIGNING_KEY"); err != nil {
		return Server{}, err
	}
	if err := v.BindEnv("postgres.password", "XIANYU_POSTGRES_PASSWORD"); err != nil {
		return Server{}, err
	}
	if err := v.ReadInConfig(); err != nil {
		return Server{}, err
	}
	var settings Server
	if err := v.Unmarshal(&settings); err != nil {
		return Server{}, err
	}
	return settings, nil
}

// configPath 按命令行、环境变量、默认值的顺序选择配置路径。
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
