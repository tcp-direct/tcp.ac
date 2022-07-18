package main

import (
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func errThrow(c *gin.Context, respcode int, thrown error, msg []byte) error {
	log.Error().
		Str("IP", c.ClientIP()).
		Str("User-Agent", c.GetHeader("User-Agent")).
		Err(thrown).Msg(string(msg))
	c.Data(respcode, "application/json", msg)
	var err error
	if thrown != nil {
		err = fmt.Errorf("%s: %w", msg, thrown)
	}
	return err
}

// TODO: do we need this?
func getSize(s io.Seeker) (size int64, err error) {
	// get size of file
	if _, err = s.Seek(0, 0); err != nil {
		return
	}
	// 2 == from the end of the file
	if size, err = s.Seek(0, 2); err != nil {
		return
	}
	_, err = s.Seek(0, 0)
	return
}

func getOldRef(p *Post) (*Post, error, bool) {
	var oldRef []byte
	oldRef, err := db.With("hsh").Get(p.Sum())
	if err != nil {
		return nil, err, false
	}
	p.Log().Trace().Caller().Msg("duplicate checksum in hash database, checking if file still exists...")
	if db.With(p.TypeCode(true)).Has(oldRef) {
		p.Log().Debug().Str("ogUid", string(oldRef)).
			Msg("duplicate file found! returning original URL")
		p.uid = string(oldRef)
		p.key = ""
		p.priv = false
		return p, nil, true
	}
	p.Log().Trace().
		Str("ogUid", string(oldRef)).
		Msg("stale hash found, deleting entry...")
	err = db.With("hsh").Delete(p.Sum())
	if err != nil {
		p.Log().Error().Err(err).Msg("failed to delete stale hash")
		p = nil
	}
	return p, err, false
}
