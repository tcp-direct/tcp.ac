package main

import (
	"errors"
	"net"
	"strings"

	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/twharmon/gouid"
	"golang.org/x/crypto/blake2b"

	termbin "git.tcp.direct/kayos/putxt"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

func init() {
	termbin.UseChannel = true
}

func incoming() {
	var msg termbin.Message
	select {
	case msg = <-termbin.Msg:
		switch msg.Type {
		case termbin.Error:
			log.Error().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg(msg.Content)
		case termbin.IncomingData:
			log.Debug().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg("termbin_data")
		case termbin.Finish:
			log.Debug().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg(msg.Content)
		case termbin.Debug:
			log.Debug().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg(msg.Content)
		case termbin.Final:
			log.Debug().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg(msg.Content)

			termPost(msg.Bytes)
		}
	}
}

func termPost(b []byte) {
	slog := log.With().Str("caller", "termPost").Logger()
	Hashr, _ := blake2b.New(64, nil)
	Hashr.Write(b)
	hash := Hashr.Sum(nil)
	if ogTxt, _ := db.With("hsh").Get(hash); ogTxt != nil {
		if db.With("txt").Has(ogTxt) {
			slog.Debug().Str("ogUid", string(ogTxt)).Msg("duplicate file found! returning original URL")
			post := &Post{
				entryType: Text,
				uid:       string(ogTxt),
				key:       "",
				priv:      false,
			}
			termbin.Reply <- termbin.Message{Type: termbin.ReturnURL, Content: post.URLString()}
			return
		}
	}

	// generate new uid and delete key
	uid := gouid.String(config.UIDSize, gouid.MixedCaseAlphaNum)
	key := gouid.String(config.DeleteKeySize, gouid.MixedCaseAlphaNum)

	// lets make sure that we don't clash even though its highly unlikely
	for uidRef, _ := db.With("txt").Get([]byte(uid)); uidRef != nil; {
		slog.Info().Msg(" uid already exists! generating new...")
		uid = gouid.String(config.UIDSize, gouid.MixedCaseAlphaNum)
	}
	for keyRef, _ := db.With("key").Get([]byte(key)); keyRef != nil; {
		slog.Info().Msg(" delete key already exists! generating new...")
		key = gouid.String(config.DeleteKeySize, gouid.MixedCaseAlphaNum)
	}

	db.With("hsh").Put(hash, []byte(uid))

	uid = gouid.String(config.UIDSize, gouid.MixedCaseAlphaNum)
	key = gouid.String(config.DeleteKeySize, gouid.MixedCaseAlphaNum)

	for uidRef, _ := db.With("txt").Get([]byte(uid)); uidRef != nil; {
		slog.Info().Msg(" uid already exists! generating new...")
		uid = gouid.String(config.UIDSize, gouid.MixedCaseAlphaNum)
	}
	for keyRef, _ := db.With("key").Get([]byte(key)); keyRef != nil; {
		slog.Info().Msg(" delete key already exists! generating new...")
		key = gouid.String(config.DeleteKeySize, gouid.MixedCaseAlphaNum)
	}

	db.With("hsh").Put([]byte(hash), []byte(uid))

	err := db.With("txt").Put([]byte(uid), b)
	if err != nil {
		slog.Error().Err(err).Msg("failed to save text!")
		termbin.Reply <- termbin.Message{Type: termbin.ReturnURL, Content: "internal server error"}
		return
	}
	err = db.With("key").Put([]byte(key), []byte("t."+uid))
	if err != nil {
		slog.Error().Msg("failed to save delete key!")
		termbin.Reply <- termbin.Message{Type: termbin.ReturnError, Content: "internal server error"}
		return
	}

	slog.Debug().Str("uid", uid).Msg("saved to database successfully, sending to Serve")

	post := &Post{
		entryType: Text,
		uid:       uid,
		key:       key,
		priv:      false,
	}

	termbin.Reply <- termbin.Message{Type: termbin.ReturnURL, Content: post.URLString()}
}

func txtView(c *gin.Context) {
	raddr := net.ParseIP(c.RemoteIP())
	if termbin.Rater.Check(&termbin.Identity{Actual: raddr}) {
		errThrow(c, 429, errors.New("ratelimitted"), "too many requests")
		return
	}

	sUid := strings.Split(c.Param("uid"), ".")
	rUid := sUid[0]
	fExt = ""

	if len(sUid) > 1 {
		fExt = strings.ToLower(sUid[1])
		if fExt != "txt" {
			errThrow(c, 400, errors.New("bad file extension"), "400")
			return
		}
	}

	// if it doesn't match the key size or it isn't alphanumeric - throw it out
	if !valid.IsAlphanumeric(rUid) || len(rUid) != config.UIDSize {
		errThrow(c, 400, errors.New("request discarded as invalid"), "400")
		return
	}

	// query bitcask for the id
	fBytes, _ := db.With("txt").Get([]byte(rUid))
	if fBytes == nil {
		errThrow(c, 404, errors.New("file not found"), "file not found")
		return
	}

	file, err := termbin.Deflate(fBytes)
	if err != nil {
		errThrow(c, 500, err, "internal server error")
	}
	c.Data(200, "text/plain", file)

}

func txtDel(c *gin.Context) {
	slog := log.With().
		Str("caller", "txtDel").Logger()
	slog.Debug().Msg("new_request")
	if !validateKey(c.Param("key")) {
		errThrow(c, 400, errors.New("bad key"), "400")
		return
	}

	rKey := c.Param("key")

	targetTxt, _ := db.With("key").Get([]byte(rKey))
	if targetTxt == nil || !strings.Contains(string(targetTxt), "t.") {
		errThrow(c, 400, errors.New("no txt delete entry found with provided key"), "400")
		return
	}

	t := strings.Split(string(targetTxt), ".")[1]

	if !db.With("txt").Has([]byte(t)) {

		errThrow(c, 500, errors.New("image not found in database"), "500") // this shouldn't happen...?
		return
	}
	if err := db.With("txt").Delete([]byte(t)); err != nil {
		errThrow(c, 500, errors.New("delete failed"), "500")
		return
	}

	if db.With("txt").Has([]byte(t)) {
		slog.Error().Str("rkey", t).Msg("delete failed!?")
		errThrow(c, 500, errors.New("delete failed, this shouldn't happen"), "500")
		return
	}

	slog.Info().Str("rkey", t).Msg("Text file deleted successfully")
	slog.Debug().Str("rkey", t).Msg("Removing delete key entry")
	err := db.With("key").Delete([]byte(rKey))
	if err != nil {
		slog.Error().Str("rkey", t).Msg("Couldn't delete key")
	}
	c.JSON(200, "DELETE_SUCCESS")
}

func serveTermbin() {
	go func() {
		for {
			incoming()
		}
	}()
	split := strings.Split(config.TermbinListen, ":")
	err := termbin.Listen(split[0], split[1])
	if err != nil {
		println(err.Error())
		return
	}
}
