package main

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/prologic/bitcask"
	"github.com/gin-gonic/gin"
	"fmt"
	"log"
)

var imgDB *bitcask.Bitcask
var md5DB *bitcask.Bitcask
var keyDB *bitcask.Bitcask
var urlDB *bitcask.Bitcask
var txtDB *bitcask.Bitcask
var errLog *log.Logger
var baseUrl string = "http://127.0.0.1:8081/"
var debugBool bool = true

func errThrow(c *gin.Context, respcode int, Error string, msg string) {
	errLog.Println(c.ClientIP() + ": " + Error)
	if debugBool {
		c.String(respcode, msg)
	}
}

//////////////////////////////////////////////////////

func txtPost(c *gin.Context) {
	return
}

//////////////////////////////////////////////////////

func urlPost(c *gin.Context) {
	return
}

//////////////////////////////////////////////////////

func init() {

	fmt.Println("Initializing...")

	//////////// init logging ////////////
	fmt.Println("Starting error logger")
	Logger := &lumberjack.Logger{
		Filename:   "error.log",
		MaxSize:    50, // megabytes
		MaxBackups: 8,
		MaxAge:     28, // days
		Compress:   true,
	}

	errLog = log.New(Logger, "", log.Ldate|log.Ltime|log.Lshortfile)
	/////////////////////////////////////

	/////////// init databases //////////
	opts := []bitcask.Option {
		bitcask.WithMaxValueSize(24 / 1024 / 1024),
	}
	keyDB, _ = bitcask.Open("./data/key", opts...)
	fmt.Println("Initializing key database")

	imgDB, _ = bitcask.Open("./data/img", opts...)
	fmt.Println("Initializing img database")

	md5DB, _ = bitcask.Open("./data/md5", opts...) // this will probably only be for images
	fmt.Println("Initializing md5 database")

	txtDB, _ = bitcask.Open("./data/txt")
	fmt.Println("Initializing txt database")

	urlDB, _ = bitcask.Open("./data/url")
	fmt.Println("Initializing url database")
	////////////////////////////////////
}



func main() {
	router := gin.Default()

	router.MaxMultipartMemory = 16 << 20

	imgR := router.Group("/i")
	{
		imgR.POST("/put", imgPost)
		imgR.GET("/:uid", imgView)
	}

	txtR := router.Group("/t")
	{
		txtR.POST("/put", txtPost)
	}

	urlR := router.Group("/u")
	{
		urlR.POST("/put", urlPost)
	}

	router.Run(":8081")
}
