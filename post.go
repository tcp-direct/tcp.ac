package main

import (
	"errors"
	"net/http"

	"git.tcp.direct/kayos/common/entropy"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

func mustJson(v any) []byte {
	js, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return js
}

var (
	message500      = mustJson(map[string]string{"error": "internal server error"})
	message400      = mustJson(map[string]string{"error": "bad request"})
	message404      = mustJson(map[string]string{"error": "file not found"})
	messageAdmin404 = mustJson(map[string]string{"error": "post id does not exist"})
)

type validator interface {
	checkURL(c *gin.Context) error
	checkContent(c *gin.Context, data []byte) error
	checkAndScrubPost(c any) ([]byte, error)
	getContentType(c *gin.Context) (string, error)
	finalize(data []byte) ([]byte, error)
}

func post(c any, vas validator, t EntryType, priv bool) error {
	scrubbed, err := vas.checkAndScrubPost(c)
	if err != nil {
		switch c.(type) {
		case *gin.Context:
			return errThrow(c.(*gin.Context), http.StatusBadRequest, err, message400)
		default:
			return err
		}
	}

	p := newPost(t, scrubbed, priv)

	var exists bool
	// the keyspace (stored in memory) for db.With("hsh") are hashes
	// making it quick to find duplicates. the value is the uid
	if db.With("hsh").Has(p.Sum()) {
		p, err, exists = getOldRef(p)
		if err != nil {
			switch c.(type) {
			case *gin.Context:
				return errThrow(c.(*gin.Context), http.StatusInternalServerError, err, message500)
			default:
				return err
			}
		}
	}

	if exists {
		p.NewPostResponse(c)
		return nil
	}

	p = instantiateWithIDs(p)
	if p == nil {
		switch c.(type) {
		case *gin.Context:
			return errThrow(c.(*gin.Context), 500, err, message500)
		default:
			return errors.New("upload failed")
		}
	}

	err = savePost(p)
	if err != nil {
		switch c.(type) {
		case *gin.Context:
			return errThrow(c.(*gin.Context), http.StatusInternalServerError, err, message500)
		default:
			return err
		}
	}

	// good to go, send them to the finisher function
	p.Log().Trace().Msg("saved to database successfully, sending to NewPostResponse")

	p.NewPostResponse(c)
	return nil
}

func savePost(p *Post) error {
	// insert actual file to database
	p.Log().Trace().Msg("saving file to database")
	err := db.With(p.TypeCode(true)).Put([]byte(p.UID()), p.Bytes())
	if err != nil {
		return err
	}
	return db.
		With("key").
		Put(
			[]byte(p.DelKey()),
			[]byte(p.TypeCode(false)+"."+p.UID()),
		)
}

func instantiateWithIDs(p *Post) *Post {
	slog := log.With().Str("caller", "instantiateWithIDs").Logger()
	// generate new uid and delete key
	p.uid = entropy.RandStrWithUpper(config.UIDSize)
	p.key = entropy.RandStrWithUpper(config.DeleteKeySize)
	// lets make sure that we don't clash even though its highly unlikely
	for db.With(p.TypeCode(true)).Has([]byte(p.UID())) {
		slog.Warn().Msg(" uid already exists! generating new...")
		p.uid = entropy.RandStrWithUpper(config.UIDSize)
	}
	for db.With("key").Has([]byte(p.DelKey())) {
		slog.Warn().Msg(" delete key already exists! generating new...")
		p.key = entropy.RandStrWithUpper(config.DeleteKeySize)
	}
	// save checksum to db to prevent dupes in the future
	err := db.With("hsh").Put(p.Sum(), []byte(p.UID()))
	if err != nil {
		slog.Error().Err(err).Msg("failed to save checksum to db")
		return nil
	}
	return p
}
