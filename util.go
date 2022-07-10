package main

import (
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
