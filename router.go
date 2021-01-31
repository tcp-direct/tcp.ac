
package main

import (
	"github.com/gin-contrib/logger"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func txtPost(c *gin.Context) {
	return
}

func urlPost(c *gin.Context) {
	return
}

func httpRouter() {
	router := gin.New()

	router.MaxMultipartMemory = 16 << 20 // crude POST limit (fix this)

	// use gzip compression unless someone requests something with an explicit extension
	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPathsRegexs([]string{".*"})))

	router.Use(logger.SetLogger()) // use our own logger

	// static html and such
	// workaround the issue where the router tries to handle /*
	router.Static("/h", "./public")
	router.StaticFile("/favicon.ico", "./public/favicon.ico")
	router.GET("/", func(c *gin.Context) { c.Redirect(301, "h/") })


	imgR := router.Group("/i")
	{
		imgR.GET("/", func(c *gin.Context) { c.String(200,"") }) // javascript wants something here idk
		imgR.POST("/put", imgPost) // put looks nicer even though its actually POST
		imgR.GET("/:uid", imgView)
	}

	delR := router.Group("/d")
	{
		delR.GET("/i/:key", imgDel)
	}

	txtR := router.Group("/t")
	{
		txtR.POST("/put", txtPost)
	}

	urlR := router.Group("/u")
	{
		urlR.POST("/put", urlPost)
	}

	router.Run(webIP + ":" + webPort)
}
