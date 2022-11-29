package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

func (c *Config) parseFromFlag() {

	flag.StringVar(&c.runAPIAddress, "a", "", "api server run address")

	flag.StringVar(&c.pgConnString, "p", "", "connection string to postgres db")

	flag.StringVar(&c.ginMode, "g", "", "gin mode")

	flag.StringVar(&c.logLvl, "l", "", "log lvl")

	flag.Parse()

}

func (c *Config) parseFromEnv() (err error) {

	envConfig := struct {
		RunAPIAddress string `env:"RUN_API_ADDRESS"`
		PgConnString  string `env:"PG_CONN_STRING"`
		GinMode       string `env:"GIN_MODE"`
		LogLevel      string `env:"LOG_LEVEL"`
	}{}

	if err = env.Parse(&envConfig); err != nil {
		return fmt.Errorf("parsing config from env: %w", err)
	}

	if envConfig.RunAPIAddress != "" {
		c.runAPIAddress = envConfig.RunAPIAddress
	}

	if envConfig.PgConnString != "" {
		c.pgConnString = envConfig.PgConnString
	}

	if envConfig.GinMode != "" {
		c.ginMode = envConfig.GinMode
	}

	if envConfig.LogLevel != "" {
		c.ginMode = envConfig.LogLevel
	}

	return nil
}
