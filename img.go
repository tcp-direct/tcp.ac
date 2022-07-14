package main

import (
	"bytes"
	"errors"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"

	_ "git.tcp.direct/kayos/common"
	"git.tcp.direct/kayos/common/entropy"
	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	exifremove "github.com/scottleedavis/go-exif-remove"

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

func instantiateWithIDs(p *Post) *Post {
	slog := log.With().Str("caller", "instantiateWithIDs").Logger()
	// generate new uid and delete key
	p.uid = entropy.RandStrWithUpper(config.UIDSize)
	p.key = entropy.RandStrWithUpper(config.DeleteKeySize)
	// lets make sure that we don't clash even though its highly unlikely
	for db.With(p.TypeCode(true)).Has([]byte(p.UID())) {
		slog.Warn().Msg(" uid already exists! generating new...")
		p.uid = entropy.RandStrWithUpper(config.UIDSize)
	}
	for db.With("key").Has([]byte(p.DelKey())) {
		slog.Warn().Msg(" delete key already exists! generating new...")
		p.key = entropy.RandStrWithUpper(config.DeleteKeySize)
	}
	// save checksum to db to prevent dupes in the future
	err := db.With("hsh").Put(p.Sum(), []byte(p.UID()))
	if err != nil {
		return nil
	}
	return p
}

func savePost(p *Post) error {
	// insert actual file to database
	p.Log().Trace().Msg("saving file to database")
	err := db.With(p.TypeCode(true)).Put([]byte(p.UID()), p.Bytes())
	if err != nil {
		return err
	}
	return db.
		With("key").
		Put(
			[]byte(p.DelKey()),
			[]byte(p.TypeCode(false)+"."+p.UID()),
		)
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

type validatingScrubber func(c *gin.Context) ([]byte, error)

func imgValidateAndScrub(c *gin.Context) ([]byte, error) {
	slog := log.With().Str("caller", "imgPost").
		Str("User-Agent", c.GetHeader("User-Agent")).
		Str("RemoteAddr", c.ClientIP()).Logger()
	// check if incoming POST data is invalid
	f, err := c.FormFile("upload")
	if err != nil || f == nil {
		return nil, err
	}

	slog.Debug().Str("filename", f.Filename).Msg("[+] New upload")

	// read the incoming file into an io file reader
	file, err := f.Open()
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err, "error processing file\n")
		return nil, err
	}

	scrubbed, err := readAndScrubImage(file)
	if err != nil {
		errThrow(c, http.StatusBadRequest, err, "invalid request")
		return nil, err
	}
	return scrubbed, nil
}

func getOldRef(p *Post) (*Post, error) {
	var oldRef []byte
	oldRef, err := db.With("hsh").Get(p.Sum())
	if err != nil {
		return nil, err
	}
	p.Log().Trace().Caller().Msg("duplicate checksum in hash database, checking if file still exists...")
	if db.With(p.TypeCode(true)).Has(oldRef) {
		p.Log().Debug().Str("ogUid", string(oldRef)).
			Msg("duplicate file found! returning original URL")
		p.uid = string(oldRef)
		p.key = ""
		p.priv = false
		return p, nil
	}
	p.Log().Trace().
		Str("ogUid", string(oldRef)).
		Msg("stale hash found, deleting entry...")
	err = db.With("hsh").Delete(p.Sum())
	if err != nil {
		p.Log().Error().Err(err).Msg("failed to delete stale hash")
		p = nil
	}
	return p, err
}

func post(c *gin.Context, vas validatingScrubber, t EntryType) error {
	scrubbed, err := vas(c)
	if err != nil {
		if c != nil {
			return errThrow(c, http.StatusBadRequest, err, "invalid request")
		}
		return err
	}
	var p *Post
	switch t {
	case Image:
		p = NewImg(scrubbed, false)
	case Text:
		p = NewTxt(scrubbed, false)
	default:
		return errors.New("invalid entry type")
	}

	// the keys (stored in memory) for db.With("hsh") are hashes
	// making it quick to find duplicates. the value is the uid
	if db.With("hsh").Has(p.Sum()) {
		p, err = getOldRef(p)
		if err != nil {
			if c != nil {
				return errThrow(c, http.StatusInternalServerError, err, "internal server error")
			}
			return err
		}
	}

	p = instantiateWithIDs(p)
	if p == nil {
		if c != nil {
			return errThrow(c, 500, err, "upload failed")
		}
		return err
	}

	err = savePost(p)
	if err != nil {
		if c != nil {
			return errThrow(c, http.StatusInternalServerError, err, "internal error")
		}
		return err
	}

	// good to go, send them to the finisher function
	p.Log().Trace().Msg("saved to database successfully, sending to NewPostResponse")

	p.NewPostResponse(c)
	return nil
}
