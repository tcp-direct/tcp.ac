package main

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

func adminBypass(c *gin.Context, adminKey string, t EntryType) {
	slog := log.With().Str("caller", "admin_del").Str("type", typeToString(t, true)).Logger()
	slog.Trace().Msg("admin key attempt")
	if config.AdminKey == "" {
		errThrow(c, 404, errors.New("admin key not configured"), message404)
		return
	}
	rKey := c.Param("key")
	if adminKey != config.AdminKey {
		if zerolog.GlobalLevel() == zerolog.TraceLevel {
			slog.Warn().Str("wanted", config.AdminKey).Str("got", adminKey).Msg("bad admin key!")
		} else {
			slog.Warn().Msg("bad admin key!")
		}
		errThrow(c, 404, errors.New("bad key"), message404)
		return
	}

	slog.Trace().Msg("admin key accepted")
	if !db.With(typeToString(t, true)).Has([]byte(rKey)) {
		errThrow(c, 404, errors.New("failed to delete entry"), messageAdmin404)
		return
	}
	err := db.With(typeToString(t, true)).Delete([]byte(rKey))
	if err != nil {
		errThrow(c, 500, err, mustJson(map[string]string{"error": err.Error()}))
		return
	}
	slog.Info().Msg("admin deleted entry")
	c.JSON(200, "DELETE_SUCCESS")
}

func del(c *gin.Context, t EntryType) {
	slog := log.With().Str("caller", "del").Str("type", typeToString(t, true)).Logger()
	rKey := c.Param("key")
	adminKey, adminAttempt := c.GetQuery("admin")
	if adminAttempt {
		adminBypass(c, adminKey, t)
		return
	}

	if !validateKey(rKey) {
		errThrow(c, 400, errors.New("failed to validate delete key"), message404)
		return
	}

	target, err := db.With("key").Get([]byte(rKey))
	if err != nil {
		errThrow(c, 400, err, message400)
		return
	}
	if target == nil || !strings.HasPrefix(string(target), typeToString(t, false)+".") {
		errThrow(c, 400, errors.New("no delete entry found with provided key"), message404)
		return
	}
	finalTarget := strings.Split(string(target), ".")
	if !db.With(typeToString(t, true)).Has([]byte(finalTarget[1])) {
		// this shouldn't happen...?
		errThrow(c, 500, errors.New("corresponding image to delete not found in database"), message500)
		return
	}
	err = db.With(typeToString(t, true)).Delete([]byte(finalTarget[1]))
	if err != nil {
		errThrow(c, 500, err, message500)
		return
	}
	if db.With(typeToString(t, true)).Has([]byte(finalTarget[1])) {
		errThrow(c, 500, errors.New("failed to delete entry"), message500)
		return
	}
	slog.Info().Str("rkey", finalTarget[1]).Msg("Image file deleted successfully")
	slog.Trace().Str("rkey", finalTarget[1]).Msg("Removing delete key entry")
	err = db.With("key").Delete([]byte(rKey))
	if err != nil {
		slog.Error().Str("rkey", finalTarget[1]).Msg("Couldn't delete key")
	}
	c.JSON(200, "DELETE_SUCCESS")
}
