package main

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"zadanie-6105/internal/api"
	"zadanie-6105/internal/postgres"
	"zadanie-6105/internal/repository"
	"zadanie-6105/internal/service"
)

const defaultAddr = ":8080"

func main() {
	pool, err := postgres.Pool()
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}

	a := api.New(service.NewService(repository.NewRepository(pool)))

	err = http.ListenAndServe(defaultAddr, a)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}
}
