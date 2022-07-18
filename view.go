package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

func view(c *gin.Context, validate validator, t EntryType) {
	slog := log.With().Str("caller", "view").Logger()
	if err := validate.checkURL(c); err != nil {
		errThrow(c, 400, err, message400)
		return
	}
	sUID := strings.Split(c.Param("uid"), ".")
	rUID := sUID[0]
	// if it doesn't match the key size or it isn't alphanumeric - throw it out
	if !govalidator.IsAlphanumeric(rUID) || len(rUID) != config.UIDSize {
		slog.Warn().
			Str("remoteaddr", c.ClientIP()).
			Msg("request discarded as invalid")
		errThrow(c, 400, errors.New(string(message400)), message400)
		return
	}
	slog.Trace().Str("rUid", rUID).Msg("request validated")
	// query bitcask for the id
	fBytes, err := db.With(typeToString(t, true)).Get([]byte(rUID))
	if fBytes == nil || err != nil {
		slog.Error().Str("rUid", rUID).Msg("no corresponding file for this id")
		realErr := err
		if err == nil {
			realErr = errors.New("no corresponding file for this id")
		}
		errThrow(c, 404, realErr, message404)
		return
	}
	err = validate.checkContent(c, fBytes)
	if err != nil {
		errThrow(c, http.StatusBadRequest, err, message400)
		return
	}
	var contentType string
	contentType, err = validate.getContentType(c)
	if err != nil {
		errThrow(c, http.StatusBadRequest, err, message400)
		return
	}

	fBytes, err = validate.finalize(fBytes)
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err, message500)
		return
	}

	c.Data(200, contentType, fBytes)
}
