package main

import (
	"github.com/prologic/bitcask"
	"github.com/rs/zerolog/log"
)

func dbInit() {
	opts := []bitcask.Option{
		bitcask.WithMaxValueSize(24 / 1024 / 1024),
	}

	hashDB, _ = bitcask.Open(dbDir+"hsh", opts...) // this will probably only be for images?
	log.Info().Msg("Initializing checksum database")

	keyDB, _ = bitcask.Open(dbDir+"key", opts...) // delete keys (maybe for all objects?)
	log.Info().Msg("Initializing key database")

	imgDB, _ = bitcask.Open(dbDir+"img", opts...) // literal image files
	log.Info().Msg("Initializing img database")

	txtDB, _ = bitcask.Open(dbDir + "txt") // pastebin
	log.Info().Msg("Initializing txt database")

	urlDB, _ = bitcask.Open(dbDir + "url") // url shortener entries
	log.Info().Msg("Initializing url database")
}

func dbSync() {
	hashDB.Sync()
	keyDB.Sync()
	imgDB.Sync()
	urlDB.Sync()
}
