package main

import (
	"github.com/scottleedavis/go-exif-remove"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/twharmon/gouid"
	"mime/multipart"
	"io/ioutil"
	"net/http"
	"errors"
	"image"
	"fmt"
	"log"
)

var errLog *log.Logger

var debugBool bool = true

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

func errThrow(w http.ResponseWriter, r *http.Request,  Error, msg string) {
	errLog.Println(remoteAddr(r) + ": " + Error)
	if debugBool {
		fmt.Fprintf(w, msg)
	}
}

func imgPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[imgPost] detected new upload")

	err := r.ParseMultipartForm(12 << 20)
	if err != nil {
		errThrow(w, r, http.StatusBadRequest, err.Error(), "error parsing upload data")
		return
	}

	fmt.Println("[imgPost] creating file handler and checking size")
	f, size, err := extractFile(r, "upload")
	if err != nil {
		errThrow(r, http.StatusBadRequest, err.Error(), "upload data is invalid")
		return
	}
	defer f.Close()

	fmt.Println("[imgPost] verifying file is an image")
	imageFormat, ok := checkImage(f)
	if !ok {
		errThrow(r, http.StatusBadRequest, "not an image", "input does not appear to be an image")
		return
	}

	fmt.Println("[imgPost] generating uid")
	uid := gouid.String(8)

	fmt.Println("[imgPost][" + uid + "] dumping byte form of file and scrubbing exif")
	fbytes, err := ioutil.ReadAll(f)
	Scrubbed, err := exifremove.Remove(fbytes)
	if _, err:= io.Copy(buf, f); err != nil {
		errThrow(r, http.StatusInternalServerError, err.Error(), "error scrubbing exif")
		return
	}

	fmt.Println("[imgPost][" + uid + "] saving file (fin)")

//	contentType := "image/" + imageFormat

	err := ioutil.WriteFile("./live/img/" + uid + "." + imageFormat, Scrubbed)
	if err != nil {
		errThrow(r, http.StatusInternalServerError, err.Error(), "error saving file")
		return
	}
}

func remoteAddr(r *http.Request) (ip string) {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

///// Stolen functions below //////////////////////////////////////

func checkImage(r io.ReadSeeker) (string, bool) {
	_, fmt, err := image.Decode(r)
	_, err2 := r.Seek(0, 0)
	if err != nil || err2 != nil {
		return "", false
	}

	return fmt, true
}

func extractFile(r *http.Request, field string) (multipart.File, int64, error) {
	files, found := r.MultipartForm.file[field]

	if !found || len(files) < 1 {
		return nil, "", 0, fmt.Errorf("'%s' not found", field)
	}
	file := files[0]

	fmt.Printf("Uploaded File: %+v\n", r.Filename)
	fmt.Printf("File Size: %+v\n", r.Size)
	fmt.Printf("MIME Header: %+v\n", r.Header)

	f, err := file.Open()

	if err != nil {
		return nil, "", 0, errors.New("could not open multipart file")
	}

	size, err := getSize(f)
	if err != nil {
		return nil, "", 0, errors.New("could not find size of file")
	}

	return f, size, nil
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

////////////////////////////////////////////////////////////////////////////


func main() {

	router := goji.NewMux()

	router.HandleFunc(pat.Post("/i/put"), imgPost)
	router.HandleFunc(pat.Post("/t/put"), txtPost)
	router.HandleFunc(pat.Post("/u/put"), urlPost)

	http.ListenAndServe("0.0.0.0:8080", router)
}
