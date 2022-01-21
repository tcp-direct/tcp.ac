package main

import (
	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
}

type Post struct {
	entryType EntryType
	owner     UserID
	uid, key  string
	priv      bool
	log       *zerolog.Logger
}

func (p *Post) TypeCode() (code string) {
	switch p.entryType {
	case Image:
		code = "i"
	case Text:
		code = "t"
	case URL:
		code = "u"
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

func (p *Post) Log(l zerolog.Logger) *zerolog.Logger {
	if p.log != nil {
		return p.log
	}
	nl := l.With().
		Str("type", p.TypeCode()).
		Str("uid", p.UID()).Str("key", p.DelKey()).
		Bool("private", p.Private()).Logger()
	p.log = &nl
	return p.log
}

func validateKey(rKey string) bool {
	// if it doesn't match the key size or it isn't alphanumeric - throw it out
	if len(rKey) != keySize || !valid.IsAlphanumeric(rKey) {
		log.Warn().Str("rKey", rKey).
			Msg("delete request failed sanity check!")
		return false
	}
	return true
}

func (p *Post) URLString() string {
	var keyurl string = ""
	url := baseURL + p.TypeCode() + "/" + string(p.UID())
	if p.DelKey() != "" {
		keyurl = baseURL + "d/" + p.TypeCode() + "/" + p.DelKey()
	}

	p.Log(log.Logger).Info().Msg("success")

	if keyurl != "" {
		return url + "\nDelete: " + keyurl + "\n"
	}
	return url
}

func (p *Post) Serve(c *gin.Context) {
	var keyurl string = ""
	url := baseURL + p.TypeCode() + "/" + string(p.UID())
	if p.DelKey() != "" {
		keyurl = baseURL + "d/" + p.TypeCode() + "/" + p.DelKey()
	}

	log.Info().
		Str("type", p.TypeCode()).
		Str("uid", p.UID()).Str("key", p.DelKey()).
		Bool("private", p.Private()).Msg("success")

	c.JSON(201, gin.H{"Imgurl": url, "ToDelete": keyurl})
	return
}
