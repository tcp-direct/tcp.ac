package main

import (
	"bytes"
	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	exifremove "github.com/scottleedavis/go-exif-remove"
	"github.com/twharmon/gouid"
	"golang.org/x/crypto/blake2b"
	"image"
	_ "image/gif"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var fExt string

func postUpload(c *gin.Context, id string, key string) {
	imgurl := baseUrl + "i/" + string(id)
	keyurl := "duplicate"
	if key != "nil" {
		keyurl = baseUrl + "d/i/" + string(key)
	}

	log.Info().Str("func", "imgPost").Str("id", id).Str("status", "201").Str("imgurl", imgurl).Str("keyurl", keyurl)
	c.JSON(201, gin.H{"delkey": keyurl, "imgurl": imgurl})
	return
}

func imgDel(c *gin.Context) {
	fn = "imgDel"

	log.Debug().Str("func", fn).Msg("Request received!") // received request
	rKey := c.Param("key")

	// if it doesn't match the key size or it isn't alphanumeric - throw it out
	if len(rKey) != keySize || !valid.IsAlphanumeric(rKey) {
		log.Error().Str("func", fn).Msg("delete request failed sanity check!")
		errThrow(c, 400, "400", "400")
		return
	}

	targetImg, _ := keyDB.Get([]byte(rKey))
	if targetImg == nil || !strings.Contains(string(targetImg), "i.") {
		log.Error().Str("func", fn).Str("rkey", rKey).Msg("no img delete entry found with provided key")
		errThrow(c, 400, "400", "400")
		return
	}

	finalTarget := strings.Split(string(targetImg), ".")

	if !imgDB.Has([]byte(finalTarget[1])) {
		log.Error().Str("func", fn).Str("rkey", rKey).Msg("corresponding image not found in database!")
		errThrow(c, 500, "500", "500") // this shouldn't happen...?
		return
	}
	err := imgDB.Delete([]byte(finalTarget[1]))
	if err != nil {
		log.Error().Str("func", fn).Str("rkey", finalTarget[1]).Msg("delete failed!")
		errThrow(c, 500, "500", "500")
		return
	}

	if imgDB.Has([]byte(finalTarget[1])) {
		log.Error().Str("func", fn).Str("rkey", finalTarget[1]).Msg("delete failed!?")
		errThrow(c, 500, "500", "500")
		return
	}

	log.Info().Str("func", fn).Str("rkey", finalTarget[1]).Msg("Image file deleted successfully")
	log.Debug().Str("func", fn).Str("rkey", finalTarget[1]).Msg("Removing delete key entry")
	err = keyDB.Delete([]byte(rKey))
	if err != nil {
		log.Error().Str("func", fn).Str("rkey", finalTarget[1]).Msg("Couldn't delete key")
		// it would be insane to try and delete the hash here
	} // if someone is uploading this image again after del
	c.JSON(200, "DELETE_SUCCESS") // and the file corresponding to the hash no longer exists
	// we will delete the hash entry then and re-add then
}

func imgView(c *gin.Context) {
	fn = "imgView"

	// the user can access their image with or without a file extension in URI
	log.Debug().Str("func", fn).Msg("request received") //  however it must be a valid extension (more checks further down)
	sUid := strings.Split(c.Param("uid"), ".")
	rUid := sUid[0]

	if len(sUid) > 1 {
		fExt = strings.ToLower(sUid[1])
		log.Debug().Str("func", fn).Str("ext", fExt).Msg("detected file extension")
		if fExt != "png" && fExt != "jpg" && fExt != "jpeg" && fExt != "gif" {
			log.Error().Str("func", fn).Msg("Bad file extension!")
			errThrow(c, 400, "400", "400")
			return
		}
	} else {
		fExt = "nil"
	}

	// if it doesn't match the key size or it isn't alphanumeric - throw it out
	if (!valid.IsAlphanumeric(rUid) || len(rUid) != uidSize) {
		log.Error().Str("func", fn).Msg("request discarded as invalid") // these limits should be variables eventually
		errThrow(c, 400, "400", "400")
		return
	}

	// now that we think its a valid request we will query
	log.Debug().Str("func", fn).Str("rUid", rUid).Msg("request validated")

	// query bitcask for the id
	fBytes, _ := imgDB.Get([]byte(rUid))
	if fBytes == nil {
		log.Error().Str("func", fn).Str("rUid", rUid).Msg("no corresponding file for this id")
		errThrow(c, 404, "404", "File not found")
		return
	}

	// read the data from bitcask into a reader
	file := bytes.NewReader(fBytes)
	imageFormat, ok := checkImage(file)
	if !ok {
		// extra sanity check to make sure we don't serve non-image data that somehow got into the database
		errThrow(c, http.StatusBadRequest, "400", "400")
		log.Error().Str("func", fn).Str("rUid", rUid).Msg("the file we grabbed is not an image!?")
		return
	} else {
		// seems like its an image, we will proceed
		log.Debug().Str("func", fn).Str("rUid", rUid).Str("imageFormat", imageFormat).Msg("Image format detected")
	}

	// additional extension sanity check - if they're gonna use an extension it needs to be the right one
	if fExt != "nil" && fExt != imageFormat {
		log.Error().Str("func", fn).Str("rUid", rUid).Msg("requested file extension does not match filetype")
		errThrow(c, 400, "400", "400")
		return
	}

	// extension or not (they are optional)
	// we give them the proper content type
	contentType := "image/" + imageFormat
	c.Data(200, contentType, fBytes)

	log.Info().Str("func", fn).Str("rUid", rUid).Msg("Successful upload")
}

func imgPost(c *gin.Context) {
	fn = "imgPost"
	var Scrubbed []byte

	// check if incoming POST data is invalid
	f, err := c.FormFile("upload")
	if err != nil {
		errThrow(c, http.StatusBadRequest, err.Error(), "400")
	}

	log.Debug().Str("func", fn).Str("filename", f.Filename).Msg("[+] New upload")

	// read the incoming file into an io file reader
	file, err := f.Open()
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err.Error(), "error processing file\n")
	}

	// verify the incoming file is an image
	log.Debug().Str("func", fn).Msg("verifying file is an image")
	imageFormat, ok := checkImage(file)
	if !ok {
		errThrow(c, http.StatusBadRequest, "400", "input does not appear to be an image")
		return
	} else {
		log.Debug().Str("func", fn).Msg("image file type detected")
	}

	// dump this into a byte object and scrub it
	// TO-DO: Write our own function for scrubbing exif
	log.Debug().Str("func", fn).Msg("dumping byte form of file")
	fbytes, err := ioutil.ReadAll(file)
	if imageFormat != "gif" {
		log.Debug().Str("func", fn).Err(err).Msg("error scrubbing exif")
		Scrubbed, err = exifremove.Remove(fbytes)
		if err != nil {
			errThrow(c, http.StatusInternalServerError, err.Error(), "error scrubbing exif")
			return
		}
	} else {
		log.Debug().Str("func", fn).Msg("skipping exif scrub for gif image")
		Scrubbed = fbytes
	}

	// time to start checking for duplicates
	log.Debug().Str("func", fn).Msg("calculating blake2b checksum")
	Hashr, _ := blake2b.New(64, nil)
	Hashr.Write(Scrubbed)
	hash := Hashr.Sum(nil)

	log.Debug().Str("func", fn).Msg("checking for duplicate files")

	// the keys (stored in memory) for hashDb are hashes
	// making it quick to find duplicates. the value is the uid
	imgRef, _ := hashDB.Get(hash)
	ogUid := string(imgRef)

	// if we get a hit we redirect them to the original image file with no delete key
	if imgRef != nil {
		log.Debug().Str("func", fn).Str("ogUid", ogUid).Msg("duplicate checksum in hash database, checking if file still exists...")
		if imgDB.Has(imgRef) {
			log.Debug().Str("func", fn).Str("ogUid", ogUid).Msg("duplicate file found! returning original URL")
			postUpload(c, ogUid, "nil")
			return
		} else {
			log.Debug().Str("func", fn).Str("ogUid", ogUid).Msg("stale hash found, deleting entry...")
			hashDB.Delete(hash)
		}
	}

	log.Info().Str("func", fn).Msg("no duplicate images found, generating uid and delete key")

	// generate new uid and delete key
	uid := gouid.String(uidSize)
	key := gouid.String(keySize)

	// lets make sure that we don't clash even though its highly unlikely
	for uidRef, _ := imgDB.Get([]byte(uid)); uidRef != nil; {
		log.Info().Str("func", fn).Msg(" uid already exists! generating new...")
		uid = gouid.String(uidSize)
	}
	for keyRef, _ := keyDB.Get([]byte(key)); keyRef != nil; {
		log.Info().Str("func", fn).Msg(" delete key already exists! generating new...")
		key = gouid.String(keySize)
	}

	// save checksum to db to prevent dupes in the future
	hashDB.Put([]byte(hash), []byte(uid))

	// insert actual file to database
	log.Debug().Str("func", fn).Str("uid", uid).Msg("saving file to database")
	err = imgDB.Put([]byte(uid), []byte(Scrubbed))
	if err != nil {
		errThrow(c, 500, err.Error(), "upload failed")
		return
	}

	// add delete key to database with image prefix
	// there is a whole db for delete keys
	err = keyDB.Put([]byte(key), []byte("i."+uid))
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err.Error(), "internal error")
		return
	}

	// good to go, send them to the finisher function
	log.Debug().Str("func", fn).Str("uid", uid).Msg("saved to database successfully, sending to postUpload")
	postUpload(c, uid, key)

}

func checkImage(r io.ReadSeeker) (string, bool) {
	// in theory this makes sure the file is an image via magic bytes
	_, fmt, err := image.Decode(r)
	_, err2 := r.Seek(0, 0)
	if err != nil || err2 != nil {
		return "", false
	}
	return fmt, true
}

func getSize(s io.Seeker) (size int64, err error) {
	// get size of file
	if _, err = s.Seek(0, 0); err != nil {
		return
	}

	// 2 == from the end of the file
	if size, err = s.Seek(0, 2); err != nil {
		return
	}

	_, err = s.Seek(0, 0)
	return
}
