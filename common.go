package main

import (
	"git.tcp.direct/kayos/common/hash"
	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/json-iterator/go"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type EntryType uint8

const (
	Image EntryType = iota
	Text
	URL
	// Custom is for user uploads of arbitrary formats.
	Custom
)

type UserID uint64

// Entry FIXME: not currently used
type Entry interface {
	TypeCode() string
	UID() string
	DelKey() string
	OwnerID() uint64
	Private() bool
	Sum() []byte
	Bytes() []byte
}

type Post struct {
	entryType EntryType
	b2sum     []byte
	data      []byte
	owner     UserID
	uid, key  string
	priv      bool
	log       *zerolog.Logger
}

func (p *Post) setLogger() {
	pl := log.Logger.With().Str("caller", p.TypeCode(true)+":"+p.UID()).
		Bool("private", p.priv).Logger()
	p.log = &pl
}

func newPost(entryType EntryType, data []byte, priv bool) *Post {
	p := &Post{
		entryType: entryType,
		priv:      priv,
		data:      data,
	}
	p.setLogger()
	return p
}

func typeToString(t EntryType, long bool) string {
	switch t {
	case Image:
		if long {
			return "img"
		}
		return "i"
	case Text:
		if long {
			return "txt"
		}
		return "t"
	case URL:
		if long {
			return "url"
		}
		return "u"
	case Custom:
		if long {
			return "custom"
		}
		return "c"
	default:
		panic("unknown entry type")
	}
}

func (p *Post) TypeCode(long bool) (code string) {
	return typeToString(p.entryType, long)
}

func (p *Post) UID() string {
	return p.uid
}

func (p *Post) DelKey() string {
	return p.key
}

func (p *Post) Private() bool {
	return p.priv
}

func (p *Post) Log() *zerolog.Logger {
	if p.log != nil {
		return p.log
	}
	p.setLogger()
	return p.log
}

func validateKey(rKey string) bool {
	if len(rKey) != config.DeleteKeySize || !valid.IsAlphanumeric(rKey) {
		return false
	}
	return true
}

func (p *Post) URLString() string {
	var keyurl = ""
	url := config.BaseURL + p.TypeCode(false) + "/" + p.UID()
	if p.DelKey() != "" {
		keyurl = config.BaseURL + "d/" + p.TypeCode(false) + "/" + p.DelKey()
	}

	p.Log().Info().Msg("success")

	if keyurl != "" {
		return url + "\nDelete: " + keyurl + "\n"
	}
	return url
}

func (p *Post) NewPostResponse(responder any) {
	var keyurl = ""
	url := config.BaseURL + p.TypeCode(false) + "/" + p.UID()
	if p.DelKey() != "" {
		keyurl = config.BaseURL + "d/" + p.TypeCode(false) + "/" + p.DelKey()
	}

	log.Info().
		Str("type", p.TypeCode(false)).
		Str("uid", p.UID()).Str("key", p.DelKey()).
		Bool("private", p.Private()).Msg("success")

	// for backwards compatibility with image scripts.
	urlString := p.TypeCode(true)
	delString := "del"
	if p.TypeCode(false) == "i" {
		urlString = "Imgurl"
		delString = "ToDelete"
	}

	if cg, ginok := responder.(*gin.Context); ginok {
		cg.JSON(201, gin.H{urlString: url, delString: keyurl})
	}
	if ct, tdok := responder.(*textValidator); tdok {
		js, err := json.Marshal(gin.H{urlString: url, delString: keyurl})
		if err != nil {
			log.Error().Interface("post", p).
				Err(err).Msg("json marshal failed")
			ct.out = []byte("{\"error\":\"json marshal failed\"}")
		}
		ct.out = js
	}
	return
}

func (p *Post) Sum() []byte {
	if p.b2sum == nil && p.data != nil {
		p.b2sum = hash.Sum(hash.TypeBlake2b, p.data)
	}
	return p.b2sum
}

func (p *Post) Bytes() []byte {
	return p.data
}
