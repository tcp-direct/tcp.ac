package main

import (
	"github.com/scottleedavis/go-exif-remove"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/twharmon/gouid"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"image"
	"fmt"
	"log"
	"io"
)

var errLog *log.Logger

var debugBool bool = true

func errThrow(c *gin.Context, respcode int, Error string, msg string) {
	errLog.Println(c.ClientIP() + ": " + Error)
	if debugBool {
		c.String(respcode, msg)
	}
}

func imgPost(c *gin.Context) {
	f, err := c.FormFile("upload")
	if err != nil {
		errThrow(c, http.StatusBadRequest, err.Error(), "no file detected within request")
	}

	fmt.Println("[imgPost] detected new upload: " + f.Filename)

	file, err := f.Open()
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err.Error(), "error processing file")
	}

	fmt.Println("[imgPost] verifying file is an image")
	imageFormat, ok := checkImage(file)
	if !ok {
		errThrow(c, http.StatusBadRequest, err.Error(), "input does not appear to be an image")
		return
	}

	fmt.Println("[imgPost] generating uid")
	uid := gouid.String(8)

	fmt.Println("[imgPost][" + uid + "] dumping byte form of file and scrubbing exif")
	fbytes, err := ioutil.ReadAll(file)
	Scrubbed, err := exifremove.Remove(fbytes)
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err.Error(), "error scrubbing exif")
		return
	}

	fmt.Println("[imgPost][" + uid + "] saving file (fin)")

//	contentType := "image/" + imageFormat

	err = ioutil.WriteFile("./live/img/" + uid + "." + imageFormat, Scrubbed, 755)
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err.Error(), "error saving file")
		return
	}
}

func checkImage(r io.ReadSeeker) (string, bool) {
	_, fmt, err := image.Decode(r)
	_, err2 := r.Seek(0, 0)
	if err != nil || err2 != nil {
		return "", false
	}

	return fmt, true
}

func getSize(s io.Seeker) (size int64, err error) {
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

//////////////////////////////////////////////////////

func txtPost(c *gin.Context) {
	return
}

//////////////////////////////////////////////////////

func urlPost(c *gin.Context) {
	return
}


func init() {
	err := &lumberjack.Logger{
		Filename:   "error.log",
		MaxSize:    50, // megabytes
		MaxBackups: 8,
		MaxAge:     28, // days
		Compress:   true,
	}

	errLog = log.New(err, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	router := gin.Default()

	router.MaxMultipartMemory = 12 << 20

	router.StaticFile("/", "./resources/index.html")
	router.StaticFile("/min.css", "./resources/spectre.min.css")
	router.StaticFile("/icons.css", "./resources/spectre-iconf.min.css")

	imgR := router.Group("/i")
	{
		imgR.POST("/put", imgPost)
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
