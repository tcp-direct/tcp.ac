package main

import(
	"github.com/gin-gonic/gin"
)

func errThrow(c *gin.Context, respcode int, Error string, msg string) {
//	log.Error().Str("IP",c.ClientIP()).Str("err",Error).Msg(msg)
	if debugBool {
		c.String(respcode, msg)
	}
}
