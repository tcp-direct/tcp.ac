package main

import (
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

	router.MaxMultipartMemory = 16 << 20

	imgR := router.Group("/i")
	{
		imgR.POST("/put", imgPost)	// put looks nicer even though its actually POST
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

	router.Run(webIP+":"+webPort)
}
