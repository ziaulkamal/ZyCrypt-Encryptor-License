package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Security  SecurityConfig  `mapstructure:"security"`
	Storage   StorageConfig   `mapstructure:"storage"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

type ServerConfig struct {
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	BaseURL string `mapstructure:"base_url"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type SecurityConfig struct {
	SharedSecret     string `mapstructure:"shared_secret"`
	TokenTTLMinutes  int    `mapstructure:"token_ttl_minutes"`
	GracePeriodHours int    `mapstructure:"grace_period_hours"`
	MinPkgVersion    string `mapstructure:"min_pkg_version"`
}

type StorageConfig struct {
	ThemesPath string `mapstructure:"themes_path"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
}

var C Config

func Load(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("zycrypt")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/zycrypt/")
		viper.AddConfigPath("$HOME/.zycrypt")
	}

	viper.SetEnvPrefix("ZYCRYPT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("config file error: %w", err)
		}
	}

	return viper.Unmarshal(&C)
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8743)
	viper.SetDefault("security.token_ttl_minutes", 10)
	viper.SetDefault("security.grace_period_hours", 24)
	viper.SetDefault("security.min_pkg_version", "1.0.0")
	viper.SetDefault("rate_limit.requests_per_minute", 60)
	viper.SetDefault("storage.themes_path", "/var/zycrypt/themes")
}

func Addr() string {
	return fmt.Sprintf("%s:%d", C.Server.Host, C.Server.Port)
}
