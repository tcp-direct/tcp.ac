package main

import (
	"git.tcp.direct/kayos/common/hash"
	"git.tcp.direct/kayos/common/pool"
	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/json-iterator/go"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var str = pool.NewStringFactory()

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
	s := str.Get()
	defer str.MustPut(s)
	switch t {
	case Image:
		s.WriteString("i")
		if long {
			s.WriteString("mg")
		}
	case Text:
		s.WriteString("t")
		if long {
			s.WriteString("xt")
		}
	case URL:
		s.WriteString("u")
		if long {
			s.WriteString("rl")
		}
	case Custom:
		s.WriteString("e")
		if long {
			s.WriteString("tc")
		}
	default:
		panic("unknown entry type")
	}
	return s.String()
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

type Response struct {
	View   string `json:"view,omitempty"`
	Delete string `json:"delete,omitempty"`
	Error  string `json:"error,omitempty"`
}

func (p *Post) viewURL() string {
	s := str.Get()
	defer str.MustPut(s)
	s.WriteString(config.BaseURL)
	s.WriteString(p.TypeCode(false))
	s.WriteString("/")
	s.WriteString(p.UID())
	return s.String()
}

func (p *Post) delURL() string {
	if p.key == "" {
		return ""
	}
	s := str.Get()
	defer str.MustPut(s)
	s.WriteString(config.BaseURL)
	s.WriteString("d/")
	s.WriteString(p.TypeCode(false))
	s.WriteString("/")
	s.WriteString(p.DelKey())
	return s.String()
}

func (p *Post) String() string {
	resp := new(Response)
	resp.View = p.viewURL()
	resp.Delete = p.delURL()
	str, _ := json.MarshalToString(resp)
	return str
}

func (p *Post) NewPostResponse(responder any) {
	view := p.viewURL()
	log.Info().
		Str("type", p.TypeCode(false)).
		Str("uid", p.UID()).Str("key", p.DelKey()).
		Bool("private", p.Private()).Msgf("success: %s", view)

	urlString := p.TypeCode(true)
	delString := "del"
	// for backwards compatibility with image scripts.
	if p.TypeCode(false) == "i" {
		urlString = "Imgurl"
		delString = "ToDelete"
	}

	if cg, ginok := responder.(*gin.Context); ginok {
		cg.JSON(201, gin.H{urlString: view, delString: p.delURL()})
	}
	if ct, tdok := responder.(*textValidator); tdok {
		js, err := json.Marshal(gin.H{urlString: view, delString: p.delURL()})
		if err != nil {
			log.Error().Interface("post", p).
				Err(err).Msg("json marshal failed")
			ct.out = []byte("{\"error\":\"internal server error\"}")
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
