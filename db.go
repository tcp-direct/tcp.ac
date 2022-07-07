package main

import (
	prologic "git.tcp.direct/Mirrors/bitcask-mirror"
	"git.tcp.direct/tcp.direct/database/bitcask"
)

var stores = []string{"hsh", "key", "img", "txt", "url"}
var megabyte = float64(1024 * 1024)

var db *bitcask.DB

func dbInit() error {
	db = bitcask.OpenDB(dbDir)
	var bitcaskopts = []prologic.Option{prologic.WithMaxValueSize(uint64(maxSize) * uint64(megabyte))}
	for _, st := range stores {
		if err := db.Init(st, bitcaskopts...); err != nil {
			return err
		}
	}
	return nil
}
