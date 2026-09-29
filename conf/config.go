package conf

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

type Database struct {
	DBType   string `mapstructure:"DBType"`
	Username string `mapstructure:"Username"`
	Password string `mapstructure:"Password"`
	Host     string `mapstructure:"Host"`
	DBName   string `mapstructure:"DBName"`
	SSLMode  string `mapstructure:"SSLMode"`

	MaxOpenConns    int `mapstructure:"MaxOpenConns"`
	MaxIdleConns    int `mapstructure:"MaxIdleConns"`
	ConnMaxLifetime int `mapstructure:"ConnMaxLifetimeSeconds"`
}

type Server struct {
	RunMode  string `mapstructure:"RunMode"`
	HttpPort int    `mapstructure:"HttpPort"`
	// TrustedProxies lists proxy IPs/CIDRs whose X-Forwarded-For header is trusted
	// for the client IP. Empty (the default) trusts none, so clients cannot spoof
	// their IP to dodge rate limits.
	TrustedProxies []string `mapstructure:"TrustedProxies"`
	// MetricsToken, if set, is required as a Bearer token to read /metrics.
	MetricsToken string `mapstructure:"MetricsToken"`
}

type App struct {
	LogSavePath string `mapstructure:"LogSavePath"`
	LogFileName string `mapstructure:"LogFileName"`
	LogFileExt  string `mapstructure:"LogFileExt"`
}

type Auth struct {
	JWTSecret       string `mapstructure:"JWTSecret"`
	TokenTTLMinutes int    `mapstructure:"TokenTTLMinutes"`
	// RefreshTTLHours is the lifetime of a refresh token (sliding: each refresh issues a new one).
	RefreshTTLHours int `mapstructure:"RefreshTTLHours"`

	// Per-IP limit on /auth/register and /auth/login.
	IPRatePerMinute float64 `mapstructure:"IPRatePerMinute"`
	IPBurst         int     `mapstructure:"IPBurst"`
	// Failed logins allowed per account within the window before it is throttled.
	LoginMaxFailures          int `mapstructure:"LoginMaxFailures"`
	LoginFailureWindowMinutes int `mapstructure:"LoginFailureWindowMinutes"`
}

type Config struct {
	Mode     string   `mapstructure:"mode"`
	Server   Server   `mapstructure:"Server"`
	App      App      `mapstructure:"App"`
	Database Database `mapstructure:"Database"`
	Auth     Auth     `mapstructure:"Auth"`
}

// LoadConfig reads conf/config.<env>.yaml. Any value can be overridden by an
// environment variable named TASK_<SECTION>_<KEY>, e.g. TASK_DATABASE_PASSWORD.
func LoadConfig(env string) (*Config, error) {
	vp := viper.New()
	vp.SetConfigName(fmt.Sprintf("config.%s", env))
	vp.AddConfigPath("conf/")
	vp.SetConfigType("yaml")

	// Defaults double as the key registry so AutomaticEnv works with Unmarshal.
	vp.SetDefault("mode", env)
	vp.SetDefault("Server.RunMode", "debug")
	vp.SetDefault("Server.HttpPort", 8080)
	vp.SetDefault("Database.DBType", "postgres")
	vp.SetDefault("Database.Username", "")
	vp.SetDefault("Database.Password", "")
	vp.SetDefault("Database.Host", "")
	vp.SetDefault("Database.DBName", "")
	vp.SetDefault("Database.SSLMode", "disable")
	vp.SetDefault("Database.MaxOpenConns", 25)
	vp.SetDefault("Database.MaxIdleConns", 5)
	vp.SetDefault("Database.ConnMaxLifetimeSeconds", 300)
	vp.SetDefault("Auth.JWTSecret", "")
	vp.SetDefault("Auth.TokenTTLMinutes", 15)
	vp.SetDefault("Auth.RefreshTTLHours", 720)
	vp.SetDefault("Auth.IPRatePerMinute", 20)
	vp.SetDefault("Auth.IPBurst", 10)
	vp.SetDefault("Auth.LoginMaxFailures", 5)
	vp.SetDefault("Auth.LoginFailureWindowMinutes", 15)
	vp.SetDefault("Server.TrustedProxies", []string{})
	vp.SetDefault("Server.MetricsToken", "")
	vp.SetDefault("App.LogSavePath", "")
	vp.SetDefault("App.LogFileName", "")
	vp.SetDefault("App.LogFileExt", "")

	vp.SetEnvPrefix("TASK")
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vp.AutomaticEnv()

	if err := vp.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	config := &Config{}
	if err := vp.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := config.Auth.validateRateLimits(); err != nil {
		return nil, err
	}
	return config, nil
}

func (a Auth) validateRateLimits() error {
	if a.RefreshTTLHours < 1 {
		return fmt.Errorf("Auth.RefreshTTLHours must be positive")
	}
	if a.IPRatePerMinute <= 0 || a.IPBurst < 1 {
		return fmt.Errorf("Auth.IPRatePerMinute and Auth.IPBurst must be positive")
	}
	if a.LoginMaxFailures < 1 || a.LoginFailureWindowMinutes < 1 {
		return fmt.Errorf("Auth.LoginMaxFailures and Auth.LoginFailureWindowMinutes must be positive")
	}
	return nil
}
