package main

import (
	"bytes"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"

	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	exifremove "github.com/scottleedavis/go-exif-remove"
	"github.com/twharmon/gouid"
	"golang.org/x/crypto/blake2b"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

var fExt string

func imgDel(c *gin.Context) {
	slog := log.With().Str("caller", "imgView").Logger()

	rKey := c.Param("key")

	if !validateKey(rKey) {
		errThrow(c, 400, errors.New("failed to validate delete key"), "invalid request")
		return
	}

	targetImg, err := db.With("key").Get([]byte(rKey))
	if err != nil {
		errThrow(c, 400, err, "invalid request")
		return
	}

	if targetImg == nil || !strings.Contains(string(targetImg), "i.") {
		errThrow(c, 400, errors.New("no img delete entry found with provided key"), "invalid request")
		return
	}

	finalTarget := strings.Split(string(targetImg), ".")

	if !db.With("img").Has([]byte(finalTarget[1])) {
		// this shouldn't happen...?
		errThrow(c, 500, errors.New("corresponding image todelete not found in database"), "internal server error")
		return
	}
	err = db.With("img").Delete([]byte(finalTarget[1]))
	if err != nil {
		errThrow(c, 500, err, "internal server error")
		return
	}

	if db.With("img").Has([]byte(finalTarget[1])) {
		slog.Error().Str("rkey", finalTarget[1]).Msg("delete failed!?")
		errThrow(c, 500, errors.New("failed to delete entry"), "internal server error")
		return
	}

	slog.Info().Str("rkey", finalTarget[1]).Msg("Image file deleted successfully")
	slog.Trace().Str("rkey", finalTarget[1]).Msg("Removing delete key entry")
	err = db.With("key").Delete([]byte(rKey))
	if err != nil {
		slog.Error().Str("rkey", finalTarget[1]).Msg("Couldn't delete key")
		// it would be insane to try and delete the hash here
	} // if someone is uploading this image again after del
	c.JSON(200, "DELETE_SUCCESS") // and the file corresponding to the hash no longer exists
	// we will delete the hash entry then and re-add then
}

func imgView(c *gin.Context) {
	slog := log.With().Str("caller", "imgView").Logger()
	sUID := strings.Split(c.Param("uid"), ".")
	rUID := sUID[0]

	if len(sUID) > 1 {
		fExt = strings.ToLower(sUID[1])
		slog.Trace().Str("ext", fExt).Msg("detected file extension")
		if fExt != "png" && fExt != "jpg" && fExt != "jpeg" && fExt != "gif" {
			errThrow(c, 400, errors.New("bad file extension"), "invalid request")
			return
		}
	} else {
		fExt = "nil"
	}

	// if it doesn't match the key size or it isn't alphanumeric - throw it out
	if !valid.IsAlphanumeric(rUID) || len(rUID) != config.UIDSize {
		slog.Warn().
			Str("remoteaddr", c.ClientIP()).
			Msg("request discarded as invalid")

		errThrow(c, 400, errors.New("invalid request"), "invalid request")
		return
	}

	// now that we think its a valid request we will query
	slog.Trace().Str("rUid", rUID).Msg("request validated")

	// query bitcask for the id
	fBytes, _ := db.With("img").Get([]byte(rUID))
	if fBytes == nil {
		slog.Error().Str("rUid", rUID).Msg("no corresponding file for this id")
		errThrow(c, 404, errors.New("entry not found"), "File not found")
		return
	}

	// read the data from bitcask into a reader
	file := bytes.NewReader(fBytes)
	imageFormat, err := checkImage(file)
	if err != nil {
		// extra sanity check to make sure we don't serve non-image data that somehow got into the database
		errThrow(c, http.StatusBadRequest, errors.New("entry in datbase is not an image: "+err.Error()), "invalid request")
		return
	}

	// additional extension sanity check - if they're gonna use an extension it needs to be the right one
	if fExt != "nil" && fExt != imageFormat {
		errThrow(c, 400, errors.New("requested file extension does not match filetype"), "invalid request")
		return
	}

	// extension or not (they are optional)
	// we give them the proper content type
	contentType := "image/" + imageFormat
	c.Data(200, contentType, fBytes)

	slog.Info().Str("rUid", rUID).Msg("Successful upload")
}

func newUIDandKey() (uid string, key string) {
	slog := log.With().Str("caller", "newUIDandKey").Logger()
	// generate new uid and delete key
	uid = gouid.String(config.UIDSize, gouid.MixedCaseAlphaNum)
	key = gouid.String(config.DeleteKeySize, gouid.MixedCaseAlphaNum)

	// lets make sure that we don't clash even though its highly unlikely
	for db.With("img").Has([]byte(uid)) {
		slog.Warn().Msg(" uid already exists! generating new...")
		uid = gouid.String(config.UIDSize, gouid.MixedCaseAlphaNum)
	}
	for db.With("key").Has([]byte(key)) {
		slog.Warn().Msg(" delete key already exists! generating new...")
		key = gouid.String(config.DeleteKeySize, gouid.MixedCaseAlphaNum)
	}
	return
}

func readAndScrubImage(file io.ReadSeeker) (scrubbed []byte, err error) {
	imageFormat, err := checkImage(file)
	if err != nil {
		return
	}

	// dump this into a byte object and scrub it
	// TO-DO: Write our own function for scrubbing exif
	fbytes, err := io.ReadAll(file)
	if err != nil {
		return
	}

	scrubbed = fbytes

	if imageFormat == "gif" {
		return
	}

	scrubbed, err = exifremove.Remove(fbytes)
	if err != nil {
		return
	}

	return
}

func imgPost(c *gin.Context) {
	slog := log.With().Str("caller", "imgPost").
		Str("User-Agent", c.GetHeader("User-Agent")).
		Str("RemoteAddr", c.ClientIP()).Logger()

	var priv = false

	// check if incoming POST data is invalid
	f, err := c.FormFile("upload")
	if err != nil || f == nil {
		errThrow(c, http.StatusBadRequest, err, "invalid request")
		return
	}

	if f.Size < 128 {
		errThrow(c, http.StatusBadRequest, err, "invalid request")
		return
	}

	// read the incoming file into an io file reader
	file, err := f.Open()
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err, "error processing file\n")
		return
	}

	scrubbed, err := readAndScrubImage(file)
	if err != nil {
		errThrow(c, http.StatusBadRequest, err, "invalid request")
		return
	}

	slog.Debug().Str("filename", f.Filename).Msg("[+] New upload")

	Hashr, err := blake2b.New(64, nil)
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err, "internal server error")
	}
	Hashr.Write(scrubbed)
	hash := Hashr.Sum(nil)

	// the keys (stored in memory) for db.With("hsh") are hashes
	// making it quick to find duplicates. the value is the uid
	if db.With("hsh").Has(hash) {
		imgRef, err := db.With("hsh").Get(hash)
		if err != nil {
			errThrow(c, http.StatusInternalServerError, err, "internal server error")
			return
		}

		slog.Trace().Caller().Msg("duplicate checksum in hash database, checking if file still exists...")

		if db.With("img").Has(imgRef) {
			slog.Debug().Str("ogUid", string(imgRef)).Msg("duplicate file found! returning original URL")
			post := &Post{entryType: Image, uid: string(imgRef), key: "", priv: false}
			post.Serve(c)
			return
		}

		slog.Trace().
			Str("ogUid", string(imgRef)).
			Msg("stale hash found, deleting entry...")

		err = db.With("hsh").Delete(hash)
		if err != nil {
			slog.Warn().Err(err).Msg("failed to delete stale hash")
		}
	}

	uid, key := newUIDandKey()

	// save checksum to db to prevent dupes in the future
	err = db.With("hsh").Put(hash, []byte(uid))
	if err != nil {
		errThrow(c, 500, err, "upload failed")
		return
	}

	// insert actual file to database
	slog.Trace().Str("uid", uid).Msg("saving file to database")
	err = db.With("img").Put([]byte(uid), scrubbed)
	if err != nil {
		errThrow(c, 500, err, "upload failed")
		return
	}

	// add delete key to database with image prefix
	// there is a whole db for delete keys
	err = db.With("key").Put([]byte(key), []byte("i."+uid))
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err, "internal error")
		return
	}

	// good to go, send them to the finisher function
	slog.Trace().Str("uid", uid).Msg("saved to database successfully, sending to Serve")

	post := &Post{
		entryType: Image,
		uid:       uid,
		key:       key,
		priv:      priv,
	}

	post.Serve(c)

}

func checkImage(r io.ReadSeeker) (fmt string, err error) {
	// in theory this makes sure the file is an image via magic bytes
	_, fmt, err = image.Decode(r)
	if err != nil {
		return
	}
	_, err = r.Seek(0, 0)
	return
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
