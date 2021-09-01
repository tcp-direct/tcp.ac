package main

import (
	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Post struct {
	Type string
	Uid  string
	Key  string
	Priv bool
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
	url := baseUrl + p.Type + "/" + string(p.Uid)
	if p.Key != "" {
		keyurl = baseUrl + "d/" + p.Type + "/" + p.Key
	}
	log.Info().
		Str("type", p.Type).
		Str("uid", p.Uid).Str("key", p.Key).
		Bool("private", p.Priv).Msg("success")
	if keyurl != "" {
		return url + "\nDelete: " + keyurl
	}
	return url
}

func (p *Post) Serve(c *gin.Context) {
	var keyurl string = ""
	url := baseUrl + p.Type + "/" + string(p.Uid)
	if p.Key != "" {
		keyurl = baseUrl + "d/" + p.Type + "/" + p.Key
	}

	log.Info().
		Str("type", p.Type).
		Str("uid", p.Uid).Str("key", p.Key).
		Bool("private", p.Priv).Msg("success")

	c.JSON(201, gin.H{"Imgurl": url, "ToDelete": keyurl})
	return
}
