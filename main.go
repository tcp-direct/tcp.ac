package main

import (
	valid "github.com/asaskevich/govalidator"
	"github.com/scottleedavis/go-exif-remove"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/prologic/bitcask"
	"github.com/twharmon/gouid"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"image"
	"bytes"
	"fmt"
	"log"
	"io"
)

var imgDB *bitcask.Bitcask
var urlDB *bitcask.Bitcask
var txtDB *bitcask.Bitcask
var errLog *log.Logger
var debugBool bool = true

func errThrow(c *gin.Context, respcode int, Error string, msg string) {
	errLog.Println(c.ClientIP() + ": " + Error)
	if debugBool {
		c.String(respcode, msg)
	}
}

func imgView(c *gin.Context) {
	rUid := c.Param("uid")
	fmt.Println("[imgView] Received request")
	if (valid.IsAlphanumeric(rUid)) {
		fmt.Println("[imgView][" + rUid + "] Request validated")
		fBytes, _ := imgDB.Get([]byte(rUid))
		if fBytes == nil {
			fmt.Println("[imgView] No data found for: " + rUid)
			errThrow(c, 404, "404", "File not found")
			return
		}

		fmt.Println("[imgView][" + rUid + "] Detecting image type")
		file := bytes.NewReader(fBytes)
		imageFormat, ok := checkImage(file)
	        if !ok {
			errThrow(c, http.StatusBadRequest, "bad request", "content does not appear to be an image")
			return
		} else { fmt.Println("[imgView][" + rUid + "] " + imageFormat + " detected") }

		contentType := "image/" + imageFormat

		c.Data(200, contentType, fBytes)
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
	} else { fmt.Println("[imgPost] " + imageFormat + " detected") }

	fmt.Println("[imgPost] generating uid")
	uid := gouid.String(8)

	fmt.Println("[imgPost][" + uid + "] dumping byte form of file and scrubbing exif")
	fbytes, err := ioutil.ReadAll(file)
	Scrubbed, err := exifremove.Remove(fbytes)
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err.Error(), "error scrubbing exif")
		return
	}

	fmt.Println("[imgPost][" + uid + "] saving file to database (fin)")

	err = imgDB.Put([]byte(uid), []byte(Scrubbed))

	if err != nil {
		errThrow(c, http.StatusInternalServerError, err.Error(), "error saving file")
		fmt.Println(err.Error())
		return
	} else { fmt.Println("[imgPost][" + uid + "] saved to database successfully") }
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
		bitcask.WithMaxValueSize(16 / 1024 / 1024),
	}
	imgDB, _ = bitcask.Open("img.db", opts...)
	fmt.Println("Opening img.db")

	txtDB, _ = bitcask.Open("txt.db")
	fmt.Println("Opening txt.db")

	urlDB, _ = bitcask.Open("url.db")
	fmt.Println("Opening url.db")
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
