package main

import (
	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/twharmon/gouid"
	"golang.org/x/crypto/blake2b"
	"net/http"
	"strings"
)

var hash []byte
var uid string
var key string

func txtFin(c *gin.Context, id string, key string) {
	txturl := baseUrl + "t/" + string(id)
	keyurl := "duplicate"
	if key != "nil" {
		keyurl = baseUrl + "d/t/" + string(key)
	}

	log.Info().Str("func", "txtPost").Str("id", id).Str("status", "201").Str("txturl", txturl).Str("keyurl", keyurl).Msg("success")
	c.JSON(201, gin.H{"delkey": keyurl, "txturl": txturl})
	return
}

func txtDel(c *gin.Context) {
	fn = "txtDel"

	// make sure its the proper amount of characters and is alphanumeric
	log.Info().Str("func", fn).Msg("delete request received")
	rKey := c.Param("key")

	if len(rKey) != keySize || !valid.IsAlphanumeric(rKey) {
		log.Error().Err(err).Str("func", fn).Msg("delete request failed sanity check!")
		errThrow(c, 400, "400", "400")
		return
	}

	// first we see if the key even exists, and if it does, is it a txt key (vs txt or url)
	targettxt, _ := keyDB.Get([]byte(rKey))
	if targettxt == nil || !strings.Contains(string(targettxt), "t.") {
		log.Error().Err(err).Str("func", fn).Str("rkey", rKey).Msg("no txt delete entry found with provided key")
		errThrow(c, 400, "400", "400")
		return
	}

	// seperate the key from the indicator to get the actual txt delete key
	finalTarget := strings.Split(string(targettxt), ".")

	// somehow we have a key that doesn't correspond with with an actual entry
	if !txtDB.Has([]byte(finalTarget[1])) {
		log.Error().Err(err).Str("func", fn).Str("rkey", rKey).Msg("corresponding text entry not found in database!")
		errThrow(c, 500, "500", "500")
		return
	}

	// if we passed all those checks, delete the txt entry from the database
	err := txtDB.Delete([]byte(finalTarget[1]))

	// failed to delete it? shouldn't happen
	if err != nil {
		log.Error().Err(err).Str("func", fn).Str("rkey", finalTarget[1]).Msg("delete failed!")
		errThrow(c, 500, "500", "500")
		return
	}

	// make sure its actually gone
	if txtDB.Has([]byte(finalTarget[1])) {
		log.Error().Err(err).Str("func", fn).Str("rkey", finalTarget[1]).Msg("delete failed!?")
		errThrow(c, 500, "500", "500")
		return
	}

	// remove the delete key entry but not the hash (see below)
	log.Info().Str("func", fn).Str("rkey", finalTarget[1]).Msg("text deleted successfully")
	log.Debug().Str("func", fn).Str("rkey", finalTarget[1]).Msg("Removing delete key entry")
	err = keyDB.Delete([]byte(rKey))
	if err != nil {
		log.Error().Err(err).Str("func", fn).Str("rkey", finalTarget[1]).Msg("Couldn't delete key")
	}
	c.JSON(200, "DELETE_SUCCESS")


	// it would be insane to try and delete the hash here
	// if someone is uploading this text again after del
        // and the file corresponding to the hash no longer exists
 	// we will delete the hash entry then and re-add then
}

func txtView(c *gin.Context) {
	fn = "txtView"
	// the user can access their text with or without a file extension in URI
	// however it must be a valid extension (more checks further down)
	log.Info().Str("func", fn).Msg("request received")
	sUid := strings.Split(c.Param("uid"), ".")
	rUid := sUid[0]
	if len(sUid) > 1 {
		fExt = strings.ToLower(sUid[1])
		log.Debug().Str("func", fn).Str("ext", fExt).Msg("detected file extension")
		if fExt != "txt" {
			log.Error().Err(err).Str("func", fn).Msg("Bad file extension!")
			errThrow(c, 400, "400", "400")
			return
		}
	} else {
		fExt = "nil"
	}

	if !valid.IsAlphanumeric(rUid) || len(rUid) != uidSize {
		log.Error().Err(err).Str("func", fn).Msg("request discarded as invalid")
		errThrow(c, 400, "400", "400")
		return
	}

	// now that we think its a valid request we will query
	log.Debug().Str("func", fn).Str("rUid", rUid).Msg("request validated")
	fBytes, _ := txtDB.Get([]byte(rUid))
	if fBytes == nil {
		log.Error().Err(err).Str("func", fn).Str("rUid", rUid).Msg("no corresponding file for this id")
		errThrow(c, 404, "404", "File not found")
		return
	}

	c.Data(200, "text/plain", fBytes)

	log.Info().Str("func", fn).Str("rUid", rUid).Msg("Success")
}

func txtPost(c *gin.Context) {
	fn = "txtPost"

	t := c.PostForm("txt")
	priv := c.PostForm("priv")


	tbyte := []byte(t)

	if err != nil {
		log.Error().Err(err).Str("fn",fn).Msg("Oh?")
		errThrow(c, 500, "500", "500")
		return
	}

	if len(t) == 0 {
		log.Warn().Str("fn",fn).Msg("received an empty request")
		errThrow(c, 400, "400", "400")
		return
	}

	if c.ContentType() != "text/plain" {
		log.Warn().Str("fn",fn).Str("ContentType", c.ContentType()).Msg("received a non-text content-type")
		errThrow(c, 400, "400", "400")
		return
	}


	// an optional switch for a privnote style burn after read
	// priv := c.GetBool("private")

	if err != nil {
		// incoming POST data is invalid
		errThrow(c, http.StatusBadRequest, err.Error(), "400") // 400 bad request
	}

	log.Debug().Str("func", fn).Msg("New paste")

	if priv == "on" {
		// check for dupes
		log.Debug().Str("func", fn).Msg("calculating blake2b checksum")
		Hashr, _ := blake2b.New(64, nil)
		Hashr.Write(tbyte)
		hash := Hashr.Sum(nil)
		log.Debug().Str("func", fn).Msg("Checking for duplicate's in database")
		txtRef, _ := txtDB.Get(hash)
		ogUid := string(txtRef)

		if txtRef != nil {
			log.Debug().Str("func", fn).Str("ogUid", ogUid).Msg("duplicate checksum in hash database, checking if file still exists...")
			if txtDB.Has(txtRef) {
				log.Debug().Str("func", fn).Str("ogUid", ogUid).Msg("duplicate file found! returning original URL")
				// they weren't the original uploader so they don't get a delete key
				txtFin(c, ogUid, "nil")
				return
			} else {
				log.Debug().Str("func", fn).Str("ogUid", ogUid).Msg("stale hash found, deleting entry...")
				hashDB.Delete(hash)
			}
		}
		log.Info().Str("func", fn).Msg("no duplicate txts found, generating uid and delete key")

		// generate identifier and delete key based on configured sizes
		uid := gouid.String(uidSize)
		key := gouid.String(keySize)

		// lets make sure that we don't clash even though its highly unlikely
		for uidRef, _ := txtDB.Get([]byte(uid)); uidRef != nil; {
			log.Info().Str("func", fn).Msg("uid already exists! generating another...")
			uid = gouid.String(uidSize)
		}
		for keyRef, _ := keyDB.Get([]byte(key)); keyRef != nil; {
			log.Info().Str("func", fn).Msg(" delete key already exists! generating another...")
			key = gouid.String(keySize)
		}

		// save checksum to db to prevent dupes in the future
		hashDB.Put([]byte(hash), []byte(uid))

		log.Debug().Str("func", fn).Str("uid", uid).Msg("saving file to database")

		err = txtDB.Put([]byte(uid), []byte(tbyte))
		if err != nil {
			errThrow(c, 500, err.Error(), "upload failed")
			return
		}

		// add delete key to database with txt prefix
		err = keyDB.Put([]byte(key), []byte("t."+uid))
		if err != nil {
			errThrow(c, http.StatusInternalServerError, err.Error(), "internal error")
			return
		}
	}


	log.Debug().Str("func", fn).Str("uid", uid).Msg("saved to database successfully, sending to txtFin")

	txtFin(c, uid, key)

}
