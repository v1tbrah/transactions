package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"transactions/internal/api"
	"transactions/internal/config"
	"transactions/internal/pg"
)

func main() {

	gin.SetMode(gin.ReleaseMode)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	newCfg, err := config.New(config.WithFlag, config.WithEnv)
	if err != nil {
		log.Error().Err(err).Msg("creating config")
		os.Exit(1)
	}
	log.Info().Msg("config created")

	gin.SetMode(newCfg.GinMode())

	logLevel, err := zerolog.ParseLevel(newCfg.LogLvl())
	if err != nil {
		log.Error().Err(err).Str("log lvl", newCfg.LogLvl()).Msg("parsing log level")
		os.Exit(1)
	}
	zerolog.SetGlobalLevel(logLevel)

	newStorage, err := pg.New(newCfg.PgConnString())
	if err != nil {
		log.Error().Err(err).Msg("creating storage")
		os.Exit(1)
	}
	log.Info().Msg("storage created")

	if err = newStorage.InitFirstFiveUsersIfNotExistsForTestingApp(); err != nil {
		log.Error().Err(err).Msg("initializing first five users for testing app")
		os.Exit(1)
	}
	log.Info().Msg("there are 5 users in the storage: id[1, 2, 3, 4, 5]")

	newAPI, err := api.New(newStorage, newCfg)
	if err != nil {
		log.Error().Err(err).Msg("creating API")
		os.Exit(1)
	}
	log.Info().Msg("api created")

	newAPI.Run()

}
