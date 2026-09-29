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
}

type App struct {
	LogSavePath string `mapstructure:"LogSavePath"`
	LogFileName string `mapstructure:"LogFileName"`
	LogFileExt  string `mapstructure:"LogFileExt"`
}

type Auth struct {
	JWTSecret       string `mapstructure:"JWTSecret"`
	TokenTTLMinutes int    `mapstructure:"TokenTTLMinutes"`
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
	vp.SetDefault("Auth.TokenTTLMinutes", 60)
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
	return config, nil
}
