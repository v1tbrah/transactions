package config

import "errors"

var errPgConnStringIsEmpty = errors.New("pg conn string is empty")

type Config struct {
	runAPIAddress string
	pgConnString  string
	ginMode       string
	logLvl        string
}

func New(options ...string) (*Config, error) {

	newConfig := &Config{}

	for _, opt := range options {
		switch opt {
		case WithFlag:
			newConfig.parseFromFlag()
		case WithEnv:
			if err := newConfig.parseFromEnv(); err != nil {
				return nil, err
			}
		}
	}

	newConfig.setDefaultIfNotConfigured()

	if newConfig.pgConnString == "" {
		return nil, errPgConnStringIsEmpty
	}

	return newConfig, nil
}

func (c *Config) setDefaultIfNotConfigured() {

	if c.runAPIAddress == "" {
		c.runAPIAddress = ":5555"
	}

	if c.ginMode == "" {
		c.ginMode = "release"
	}

	if c.logLvl == "" {
		c.logLvl = "info"
	}

}

func (c *Config) RunAPIAddress() string {
	return c.runAPIAddress
}

func (c *Config) PgConnString() string {
	return c.pgConnString
}

func (c *Config) GinMode() string {
	return c.ginMode
}

func (c *Config) LogLvl() string {
	return c.logLvl
}

func (c *Config) String() string {
	return "run API address :" + c.runAPIAddress +
		"Gin mode :" + c.ginMode +
		"Log lvl: " + c.logLvl
}
