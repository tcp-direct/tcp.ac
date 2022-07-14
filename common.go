package main

import (
	"git.tcp.direct/kayos/common/hash"
	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

type EntryType uint8

const (
	Image EntryType = iota
	Text
	URL
	// Custom is for user uploads of arbitrary formats.
	Custom
)

type UserID uint64

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

func NewImg(data []byte, priv bool) *Post {
	p := &Post{
		entryType: Image,
		priv:      priv,
		data:      data,
	}
	p.setLogger()
	return p
}

func NewTxt(data []byte, priv bool) *Post {
	p := &Post{
		entryType: Text,
		priv:      priv,
		data:      data,
	}
	p.setLogger()
	return p
}

func (p *Post) TypeCode(long bool) (code string) {
	switch p.entryType {
	case Image:
		code = "i"
		if long {
			code += "mage"
		}
	case Text:
		code = "t"
		if long {
			code += "ext"
		}
	case URL:
		code = "u"
		if long {
			code += "rl"
		}
	default:
		panic("not implemented")
	}
	return
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
	// if it doesn't match the key size or it isn't alphanumeric - throw it out
	if len(rKey) != config.DeleteKeySize || !valid.IsAlphanumeric(rKey) {
		log.Warn().Str("rKey", rKey).
			Msg("delete request failed sanity check!")
		return false
	}
	return true
}

func (p *Post) URLString() string {
	var keyurl string = ""
	url := config.BaseURL + p.TypeCode(false) + "/" + string(p.UID())
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
	var keyurl string = ""
	url := config.BaseURL + p.TypeCode(false) + "/" + string(p.UID())
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
