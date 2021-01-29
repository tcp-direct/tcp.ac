package main

import (
	"fmt"
	"github.com/prologic/bitcask"
)

func dbInit() {
	opts := []bitcask.Option{
		bitcask.WithMaxValueSize(24 / 1024 / 1024),
	}

	keyDB, _ = bitcask.Open(dbDir+"key", opts...) // delete keys (maybe for all objects?)
	fmt.Println("Initializing key database")

	imgDB, _ = bitcask.Open(dbDir+"img", opts...) // literal image files
	fmt.Println("Initializing img database")

	hashDB, _ = bitcask.Open(dbDir+"hsh", opts...) // this will probably only be for images?
	fmt.Println("Initializing checksum database")

	txtDB, _ = bitcask.Open(dbDir + "txt") // pastebin
	fmt.Println("Initializing txt database")

	urlDB, _ = bitcask.Open(dbDir + "url") // url shortener entries
	fmt.Println("Initializing url database")
}
