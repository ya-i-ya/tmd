package main

import (
	"github.com/rs/zerolog/log"
	"tmd/internal"
)

func main() {
	if err := internal.Run(); err != nil {
		log.Fatal().Err(err).Msg("Application failed")
	}
}
