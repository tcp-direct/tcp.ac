package main

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

const favicon = "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"
const staticIndex = "PCFET0NUWVBFIGh0bWw+CjxodG1sPgo8aGVhZD4KCTxtZXRhIGNoYXJzZXQ9IlVURi04Ij4KCTxtZXRhIG5hbWU9ImRlc2NyaXB0aW9uIiBjb250ZW50PSJ0Y3AuYWMgLSB1bmRlciBjb25zdHJ1Y3Rpb24gLSBpbWFnZSB1cGxvYWRpbmcsIHVybCBzaG9ydGVuaW5nLCB0ZXh0IGJpbiI+Cgk8bWV0YSBuYW1lPSJhdXRob3IiIGNvbnRlbnQ9InRjcC5kaXJlY3QiPgoJPHRpdGxlPnRjcC5hYzwvdGl0bGU+CgoJPHN0eWxlIHR5cGU9InRleHQvY3NzIj4KCQlib2R5IHsKCQkJYmFja2dyb3VuZC1jb2xvcjojMTAxMDEwOwoJCQljb2xvcjojOTQ4REI4OwoJCQlmb250LWZhbWlseTptb25vc3BhY2U7CgkJCXRleHQtYWxpZ246Y2VudGVyOwoJCX0KCQkuaGVsbG8gewoJCQlwb3NpdGlvbjogZml4ZWQ7CgkJCXRvcDogNTAlOwoJCQlsZWZ0OiA1MCU7CgkJCXRyYW5zZm9ybTogdHJhbnNsYXRlKC01MCUsIC01MCUpOwoJCX0KCQkuZmFkaW5nIHsKCQkJYW5pbWF0aW9uOmZhZGluZyA1cyBpbmZpbml0ZQoJCX0KCQlAa2V5ZnJhbWVzIGZhZGluZ3sKCQkJMCV7b3BhY2l0eTowfQoJCQk1MCV7b3BhY2l0eToxfQoJCQkxMDAle29wYWNpdHk6MH0KCQl9Cgk8L3N0eWxlPgo8L2hlYWQ+Cjxib2R5Pgo8ZGl2IGNsYXNzPSJoZWxsbyI+CjxwcmUgY2xhc3M9ImZhZGluZyI+CgogICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAsZCAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgODggICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIApNTTg4TU1NICxhZFBQWWJhLCA4YixkUFBZYmEsICAgICAgLGFkUFBZWWJhLCAgLGFkUFBZYmEsICAKICA4OCAgIGE4IiAgICAgIiIgODhQJyAgICAiOGEgICAgICIiICAgICBgWTggYTgiICAgICAiIiAgCiAgODggICA4YiAgICAgICAgIDg4ICAgICAgIGQ4ICAgICAsYWRQUFBQUDg4IDhiICAgICAgICAgIAogIDg4LCAgIjhhLCAgICxhYSA4OGIsICAgLGE4IiA4ODggODgsICAgICw4OCAiOGEsICAgLGFhICAKICAiWTg4OCBgIlliYmQ4IicgODhgWWJiZFAiJyAgODg4IGAiOGJiZFAiWTggIGAiWWJiZDgiJyAgCiAgICAgICAgICAgICAgICAgIDg4ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICA4OCAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKCjwvcHJlPgo8L2Rpdj4KCjxhIGhyZWY9Imh0dHA6Ly9hZG1pbi50Y3AuYWMvIiBzdHlsZT0iZGlzcGxheTpub25lOyI+YWRtaW4gbG9naW48L2E+Cgo8L2JvZHk+Cgo8L2h0bWw+Cg=="

func favIcon(c *gin.Context) {
	ico, _ := base64.StdEncoding.DecodeString(favicon)
	c.Data(200, "image/ico", ico)
}

func placeHolder(c *gin.Context) {
	html, _ := base64.StdEncoding.DecodeString(staticIndex)
	c.Data(200, "text/html", html)
}

func httpRouter() *http.Server {
	if !config.Trace {
		log.Debug().Caller().Msg("running gin in release mode, enable trace to run gin in debug mode")
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.SetTrustedProxies([]string{"127.0.0.1", "10.8.0.1"})

	router.Use(logger.SetLogger(
		logger.WithLogger(
			func(c *gin.Context, l zerolog.Logger) zerolog.Logger {
				if zerolog.GlobalLevel() > zerolog.DebugLevel {
					// because this spams the logs
					if c.Request.URL.String() == "/ip" {
						return zerolog.Nop()
					}
				}
				return log.With().
					Str("caller", c.ClientIP()).
					Str("url", c.Request.URL.String()).
					Str("uagent", c.Request.Header.Get("User-Agent")).
					Logger()
			},
		),
	))

	router.MaxMultipartMemory = 16 << 20 // crude POST limit (fix this)

	// use gzip compression unless someone requests something with an explicit extension
	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPathsRegexs([]string{"(?i)\\.[a-z]*"})))

	router.GET("/favicon.ico", favIcon)
	router.GET("/", placeHolder)

	router.GET("/ip", func(c *gin.Context) { c.String(200, c.ClientIP()) })

	imgR := router.Group("/i")
	imgR.GET("/", placeHolder)
	imgR.POST("/put", func(c *gin.Context) { post(c, imageValidator{}, Image, false) })
	imgR.GET("/:uid", func(c *gin.Context) { view(c, imageValidator{}, Image) })

	txtR := router.Group("/t")
	txtR.GET("/", placeHolder)
	txtR.GET("/:uid", func(c *gin.Context) { view(c, textValidator{}, Text) })

	delR := router.Group("/d")
	delR.GET("/i/:key", func(c *gin.Context) { del(c, Image) })
	delR.GET("/t/:key", func(c *gin.Context) { del(c, Text) })

	log.Info().Str("Host", config.HTTPBind).
		Str("Port", config.HTTPPort).
		Msg("done; tcp.ac is live.")

	router.SetTrustedProxies(config.TrustedProxies)

	srv := &http.Server{
		Addr:    config.HTTPBind + ":" + config.HTTPPort,
		Handler: router,
	}

	go srv.ListenAndServe()

	return srv
}
