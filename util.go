package main

import(
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func errThrow(c *gin.Context, respcode int, Error string, msg string) {
//	log.Error().Str("IP",c.ClientIP()).Str("err",Error).Msg(msg)
	if debugBool {
		c.String(respcode, msg)
	}
}
