package main

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
)

func errThrow(c *gin.Context, respcode int, Error string, msg string) {
	//	log.Error().Str("IP",c.ClientIP()).Str("err",Error).Msg(msg)
	if debugBool {
		c.String(respcode, msg)
	}
}

func b64d(str string) string {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
