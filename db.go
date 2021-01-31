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
	log.Debug().Msg("Initializing checksum database")

	keyDB, _ = bitcask.Open(dbDir+"key", opts...) // delete keys (maybe for all objects?)
	log.Debug().Msg("Initializing key database")

	imgDB, _ = bitcask.Open(dbDir+"img", opts...) // literal image files
	log.Debug().Msg("Initializing img database")

	txtDB, _ = bitcask.Open(dbDir + "txt") // pastebin
	log.Debug().Msg("Initializing txt database")

	urlDB, _ = bitcask.Open(dbDir + "url") // url shortener entries
	log.Debug().Msg("Initializing url database")
}
