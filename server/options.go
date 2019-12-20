package server

import "flag"

const (
	DefaultServerPort = 5000
)

type ServerConfiguration struct {
	Port       int
	ConfigPath string
}

func ParseServerOptions() ServerConfiguration {
	var config ServerConfiguration

	flag.StringVar(&config.ConfigPath, "config", "./config/config.yaml", "path to config file")
	flag.IntVar(&config.Port, "port", DefaultServerPort, "server tcp port to listen on")
	flag.Parse()

	return config
}
