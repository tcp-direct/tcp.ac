package main

import (
	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func errThrow(c *gin.Context, respcode int, thrown error, msg string) {
	log.Error().
		Str("IP", c.ClientIP()).
		Str("User-Agent", c.GetHeader("User-Agent")).
		Err(thrown).Msg(msg)
	c.String(respcode, msg)
}

func b64d(str string) string {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
