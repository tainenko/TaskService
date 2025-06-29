package conf

import (
	"fmt"
	"github.com/spf13/viper"
)

type Database struct {
	DBType   string `mapstructure:"DBType"`
	Username string `mapstructure:"Username"`
	Password string `mapstructure:"Password"`
	Host     string `mapstructure:"Host"`
	DBName   string `mapstructure:"DBName"`
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

type Config struct {
	Mode     string   `mapstructure:"mode"`
	Server   Server   `mapstructure:"Server"`
	App      App      `mapstructure:"App"`
	Database Database `mapstructure:"Database"`
}

func LoadConfig(env string) (*Config, error) {
	vp := viper.New()
	vp.SetConfigName(fmt.Sprintf("config.%s", env))
	vp.AddConfigPath("conf/")
	vp.SetConfigType("yaml")

	err := vp.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("Fatal error config file: %s \n", err)
	}
	fmt.Printf("Using config:%+v\n", vp.AllSettings())

	config := &Config{}
	err = vp.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("Fatal error Unmarshal config file: %s \n", err))
	}
	return config, nil
}
