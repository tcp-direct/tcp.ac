package main

import (
	"git.tcp.direct/tcp.direct/database/bitcask"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

var stores = []string{"hsh", "key", "img", "txt", "url"}
var megabyte = float64(1024 * 1024)

var db *bitcask.DB

func init() {
	bitcask.SetDefaultBitcaskOptions(
		bitcask.WithMaxKeySize(uint32(config.KVMaxKeySizeMB*int(megabyte))),
		bitcask.WithMaxValueSize(uint64(config.KVMaxValueSizeMB*int(megabyte))),
	)
}
func dbInit() error {
	db = bitcask.OpenDB(config.DBDir)
	for _, st := range stores {
		if err := db.Init(st); err != nil {
			return err
		}
	}
	return nil
}
