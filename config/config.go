package config

import (
	"flag"
	"strings"

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

type JWTConfig struct {
	Secret string
}

type BillingConfig struct {
	Dir string
}

type ApplicationConfig struct {
	Server   ServerConfig
	Database DatabaseConfig
	Cache    CacheConfig
	Auth     JWTConfig

	Billing  BillingConfig
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

	cfg.SetEnvPrefix("shortly")
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cfg.AutomaticEnv()

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	_ = cfg.BindPFlags(pflag.CommandLine)

	// database default settings
	cfg.SetDefault("Database.Host", "localhost")
	cfg.SetDefault("Database.Port", 5432)
	cfg.SetDefault("Database.User", "shortly_user")
	cfg.SetDefault("Database.Password", "1")
	cfg.SetDefault("Database.Database", "shortly")
	cfg.SetDefault("Database.SSLMode", "disable")

	cfg.SetDefault("Billing.Dir", ".")

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
