package main

import (
	"fmt"
	"github.com/babon21/avito-url-shortened/config"
	"github.com/babon21/avito-url-shortened/handlers"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"net/http"
)

func main() {
	appConfig, err := config.ParseConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot parse config")
	}

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s", appConfig.DB.User, appConfig.DB.Password, appConfig.DB.Host, appConfig.DB.Port, appConfig.DB.DBName)
	db, err := sqlx.Open("pgx", connStr)

	if err != nil {
		log.Fatal().Err(err).Msg(err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	handler := &handlers.UrlShortHandler{
		BaseUrl: appConfig.HTTPHost + ":" + appConfig.HTTPPort,
		Db:      db,
	}

	http.Handle("/", handler)

	log.Info().Msg(fmt.Sprintf("Start http server on %s port", appConfig.HTTPPort))
	http.ListenAndServe(":"+appConfig.HTTPPort, nil)
}
