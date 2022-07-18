package main

import (
	"bytes"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"

	_ "git.tcp.direct/kayos/common"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	exifremove "github.com/scottleedavis/go-exif-remove"
)

type imageValidator struct{}

func (i imageValidator) finalize(data []byte) ([]byte, error) {
	return data, nil
}

func readAndScrubImage(c any, readHead io.ReadSeeker) (scrubbed []byte, err error) {
	cg := c.(*gin.Context)
	imageFormat, err := checkImage(readHead)
	if err != nil {
		return
	}
	cg.Set("real.extension", imageFormat)
	// dump this into a byte object and scrub it
	// TO-DO: Write our own function for scrubbing exif
	fbytes, err := io.ReadAll(readHead)
	if err != nil {
		return
	}
	scrubbed = fbytes
	if imageFormat == "gif" {
		return
	}
	scrubbed, err = exifremove.Remove(fbytes)
	if err != nil {
		return
	}
	return
}

func (i imageValidator) getContentType(c *gin.Context) (string, error) {
	imageType, ok := c.Get("real.extension")
	if !ok {
		return "", errors.New("no filetype in context")
	}
	return "image/" + imageType.(string), nil
}

func (i imageValidator) checkURL(c *gin.Context) error {
	sUID := strings.Split(c.Param("uid"), ".")
	var fExt string
	if len(sUID) > 1 {
		fExt = strings.ToLower(sUID[1])
		log.Trace().Str("caller", c.Request.RequestURI).Str("ext", fExt).Msg("detected file extension")
		if fExt != "png" && fExt != "jpg" && fExt != "jpeg" && fExt != "gif" && fExt != "webm" {
			return errors.New("bad file extension")
		}
		c.Set("url.extension", fExt)
	}
	return nil
}

func (i imageValidator) checkContent(c *gin.Context, data []byte) error {
	readHead := bytes.NewReader(data)
	var err error
	_, err = readAndScrubImage(c, readHead)
	if err != nil {
		return err
	}
	urlExt, uExists := c.Get("url.extension")
	bytExt, bExists := c.Get("real.extension")
	if uExists && bExists && urlExt != bytExt {
		return errors.New("bad file extension")
	}
	return nil
}

func (i imageValidator) checkAndScrubPost(c any) ([]byte, error) {
	cg := c.(*gin.Context)
	slog := log.With().Str("caller", "imgPost").
		Str("User-Agent", cg.GetHeader("User-Agent")).
		Str("RemoteAddr", cg.ClientIP()).Logger()
	// check if incoming POST data is invalid
	f, err := cg.FormFile("upload")
	if err != nil || f == nil {
		return nil, err
	}
	slog.Debug().Str("filename", f.Filename).Msg("[+] New upload")
	// read the incoming file into an io file reader
	file, err := f.Open()
	if err != nil {
		errThrow(cg, http.StatusInternalServerError, err, message500)
		return nil, err
	}
	scrubbed, err := readAndScrubImage(c, file)
	if err != nil {
		errThrow(cg, http.StatusBadRequest, err, message400)
		return nil, err
	}
	return scrubbed, nil
}

func checkImage(r io.ReadSeeker) (fmt string, err error) {
	// in theory this makes sure the file is an image via magic bytes
	_, fmt, err = image.Decode(r)
	if err != nil {
		return
	}
	_, err = r.Seek(0, 0)
	return
}
