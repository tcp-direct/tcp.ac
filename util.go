package main

import (
	"fmt"
	"image"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func errThrow(c *gin.Context, respcode int, thrown error, msg string) error {
	log.Error().
		Str("IP", c.ClientIP()).
		Str("User-Agent", c.GetHeader("User-Agent")).
		Err(thrown).Msg(msg)
	c.String(respcode, msg)
	var err error
	if thrown != nil {
		err = fmt.Errorf("%s: %s", msg, thrown)
	}
	return err
}

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

func checkImage(r io.ReadSeeker) (fmt string, err error) {
	// in theory this makes sure the file is an image via magic bytes
	_, fmt, err = image.Decode(r)
	if err != nil {
		return
	}
	_, err = r.Seek(0, 0)
	return
}
