package config

import (
	"flag"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type MemcachedConfig struct {
	ServerList []string
}

type CacheConfig struct {
	CacheType string
	Memcached MemcachedConfig
}

type ApplicationConfig struct {
	Server   ServerConfig
	Database DatabaseConfig
	Cache    CacheConfig
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

func ReadConfig(configFilePath string) (*ApplicationConfig, error) {

	appConfig := &ApplicationConfig{}

	cfg := viper.New()

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	_ = cfg.BindPFlags(pflag.CommandLine)

	cfg.SetDefault("Database.Host", "localhost")
	cfg.SetDefault("Database.Port", 5432)
	cfg.SetDefault("Database.User", "shortly_user")
	cfg.SetDefault("Database.Password", "1")
	cfg.SetDefault("Database.Database", "shortly")
	cfg.SetDefault("Database.SSLMode", "disable")

	cfg.SetConfigFile(configFilePath)

	err := cfg.ReadInConfig()
	if err != nil {
		return nil, err
	}

	if err := cfg.Unmarshal(appConfig); err != nil {
		return nil, err
	}

	return appConfig, nil
}
