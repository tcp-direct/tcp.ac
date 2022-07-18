package main

import (
	"errors"
	"strings"

	"git.tcp.direct/kayos/common/squish"
	termdumpster "git.tcp.direct/kayos/putxt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

type textValidator struct {
	in  []byte
	out []byte
}

type textIngestor struct{}

func (i textIngestor) Ingest(data []byte) ([]byte, error) {
	tp := &textValidator{in: data}
	err := post(tp, tp, Text, false)
	if err != nil {
		return nil, err
	}
	return tp.out, nil
}

func (i textValidator) finalize(data []byte) ([]byte, error) {
	return squish.Gunzip(data)
}

func (i textValidator) getContentType(c *gin.Context) (string, error) {
	return "text/plain", nil
}

func (i textValidator) checkURL(c *gin.Context) error {
	sUID := strings.Split(c.Param("uid"), ".")
	var fExt string
	if len(sUID) > 1 {
		fExt = strings.ToLower(sUID[1])
		log.Trace().Str("caller", c.Request.RequestURI).Str("ext", fExt).Msg("detected file extension")
		if fExt != "txt" {
			return errors.New("bad file extension")
		}
		c.Set("url.extension", fExt)
	}
	return nil
}

func (i textValidator) checkContent(c *gin.Context, data []byte) error {
	return nil
}

func (i textValidator) checkAndScrubPost(c any) ([]byte, error) {
	if i.in == nil {
		return nil, errors.New("no data")
	}
	return i.in, nil
}

func serveTermbin() error {
	td := termdumpster.NewTermDumpster(textIngestor{}).WithGzip().WithLogger(&log.Logger).
		WithMaxSize(int64(config.KVMaxValueSizeMB * 1024 * 1024))
	split := strings.Split(config.TermbinListen, ":")
	log.Info().Str("listen", config.TermbinListen).Msg("starting termbin")
	err := td.Listen(split[0], split[1])
	if err != nil {
		return err
	}
	return nil
}
