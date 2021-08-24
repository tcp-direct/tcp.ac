package main

import (
	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/twharmon/gouid"
	"golang.org/x/crypto/blake2b"
	"strings"
	"tcp.ac/termbin"
)

func init() {
	termbin.UseChannel = true
}

func incoming() {
	var msg termbin.Message

	select {
	case msg = <-termbin.Msg:
		switch msg.Type {
		case "ERROR":
			log.Error().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg(msg.Content)
		case "INCOMING_DATA":
			log.Debug().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg("termbin_data")
		case "FINISH":
			log.Debug().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg(msg.Content)

		case "DEBUG":
			log.Debug().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg(msg.Content)

		case "FINAL":
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
	if ogTxt, _ := hashDB.Get(hash); ogTxt != nil {
		if txtDB.Has(ogTxt) {
			slog.Debug().Str("ogUid", string(ogTxt)).Msg("duplicate file found! returning original URL")
			post := &Post{
				Type: "t",
				Uid:  string(ogTxt),
				Key:  "",
				Priv: false,
			}
			termbin.Reply <- termbin.Message{Type: "URL", Content: post.URLString()}
			return
		}
	}

	// generate new uid and delete key
	uid := gouid.String(uidSize, gouid.MixedCaseAlphaNum)
	key := gouid.String(keySize, gouid.MixedCaseAlphaNum)

	// lets make sure that we don't clash even though its highly unlikely
	for uidRef, _ := txtDB.Get([]byte(uid)); uidRef != nil; {
		slog.Info().Msg(" uid already exists! generating new...")
		uid = gouid.String(uidSize, gouid.MixedCaseAlphaNum)
	}
	for keyRef, _ := keyDB.Get([]byte(key)); keyRef != nil; {
		slog.Info().Msg(" delete key already exists! generating new...")
		key = gouid.String(keySize, gouid.MixedCaseAlphaNum)
	}

	hashDB.Put(hash, []byte(uid))

	uid = gouid.String(uidSize, gouid.MixedCaseAlphaNum)
	key = gouid.String(keySize, gouid.MixedCaseAlphaNum)

	for uidRef, _ := txtDB.Get([]byte(uid)); uidRef != nil; {
		slog.Info().Msg(" uid already exists! generating new...")
		uid = gouid.String(uidSize, gouid.MixedCaseAlphaNum)
	}
	for keyRef, _ := keyDB.Get([]byte(key)); keyRef != nil; {
		slog.Info().Msg(" delete key already exists! generating new...")
		key = gouid.String(keySize, gouid.MixedCaseAlphaNum)
	}

	hashDB.Put([]byte(hash), []byte(uid))

	err = txtDB.Put([]byte(uid), b)
	if err != nil {
		slog.Error().Msg("failed to save text!")
		termbin.Reply <- termbin.Message{Type: "ERROR", Content: "internal server error"}
		return
	}
	err = keyDB.Put([]byte(key), []byte("t."+uid))
	if err != nil {
		slog.Error().Msg("failed to save delete key!")
		termbin.Reply <- termbin.Message{Type: "ERROR", Content: "internal server error"}
		return
	}

	slog.Debug().Str("uid", uid).Msg("saved to database successfully, sending to Serve")

	post := &Post{
		Type: "t",
		Uid:  uid,
		Key:  key,
		Priv: false,
	}

	termbin.Reply <- termbin.Message{Type: "URL", Content: post.URLString()}
}

func txtView(c *gin.Context) {
	slog := log.With().Str("caller", "txtView").Logger()

	raddr, _ := c.RemoteIP()
	if termbin.Rater.Check(&termbin.TermbinSource{Actual: raddr}) {
		errThrow(c, 429, "ratelimited", "too many requests")
		return
	}

	// the user can access their image with or without a file extension in URI
	slog.Debug().Msg("request received") //  however it must be a valid extension (more checks further down)

	sUid := strings.Split(c.Param("uid"), ".")
	rUid := sUid[0]
	fExt = ""

	if len(sUid) > 1 {
		fExt = strings.ToLower(sUid[1])
		if fExt != "txt" {
			slog.Error().Msg("bad file extension")
			errThrow(c, 400, "400", "400")
			return
		}
	}

	// if it doesn't match the key size or it isn't alphanumeric - throw it out
	if !valid.IsAlphanumeric(rUid) || len(rUid) != uidSize {
		slog.Error().Msg("request discarded as invalid") // these limits should be variables eventually
		errThrow(c, 400, "400", "400")
		return
	}

	// query bitcask for the id
	fBytes, _ := txtDB.Get([]byte(rUid))
	if fBytes == nil {
		slog.Error().Str("rUid", rUid).Msg("no corresponding file for this id")
		errThrow(c, 404, "404", "file not found")
		return
	}

	file, _ := termbin.Deflate(fBytes)
	c.Data(200, "text/plain", file)

	slog.Info().Str("rUid", rUid).Msg("success")
}

func txtDel(c *gin.Context) {
	slog := log.With().
		Str("caller", "txtDel").Logger()
	slog.Debug().Msg("new_request")
	if !validateKey(c.Param("key")) {
		errThrow(c, 400, "400", "400")
		return
	}

	rKey := c.Param("key")

	targetTxt, _ := keyDB.Get([]byte(rKey))
	if targetTxt == nil || !strings.Contains(string(targetTxt), "t.") {
		slog.Warn().Str("rkey", rKey).Msg("no txt delete entry found with provided key")
		errThrow(c, 400, "400", "400")
		return
	}

	t := strings.Split(string(targetTxt), ".")[1]

	if !txtDB.Has([]byte(t)) {
		slog.Warn().Str("rkey", rKey).Msg("corresponding image not found in database!")
		errThrow(c, 500, "500", "500") // this shouldn't happen...?
		return
	}
	if err := txtDB.Delete([]byte(t)); err != nil {
		slog.Error().Str("rkey", t).Msg("delete failed!")
		errThrow(c, 500, "500", "500")
		return
	}

	if txtDB.Has([]byte(t)) {
		slog.Error().Str("rkey", t).Msg("delete failed!?")
		errThrow(c, 500, "500", "500")
		return
	}

	slog.Info().Str("rkey", t).Msg("Text file deleted successfully")
	slog.Debug().Str("rkey", t).Msg("Removing delete key entry")
	err = keyDB.Delete([]byte(rKey))
	if err != nil {
		slog.Error().Str("rkey", t).Msg("Couldn't delete key")
	}
	c.JSON(200, "DELETE_SUCCESS")
}

//func serveTermbin() {
func serveTermbin() {
	go func() {
		for {
			incoming()
		}
	}()
	err := termbin.Listen("", txtPort)
	if err != nil {
		println(err.Error())
		return
	}
}
