package main

import(
	valid "github.com/asaskevich/govalidator"
	exifremove "github.com/scottleedavis/go-exif-remove"
	"golang.org/x/crypto/blake2b"
	"github.com/twharmon/gouid"
	"github.com/gin-gonic/gin"
	_ "image/gif"
	"io/ioutil"
	"net/http"
	"strings"
	"image"
	"bytes"
	"fmt"
	"io"
)

var fExt string

type Post struct {
    Imgurl  string `json:"Imgurl"`
    Delkey string  `json:"Delkey"`
}

func postUpload(c *gin.Context, id string, key string) {
	imgurl := baseUrl + "i/" + string(id)
	keyurl := "duplicate"
	if key != "nil" { keyurl = baseUrl + "d/i/" + string(key) }

	log.Info().Str("func","imgPost").Str("id",id).Str("status","201").Str("imgurl",imgurl).Str("keyurl",keyurl)
	c.JSON(201, gin.H{"delkey": keyurl,"imgurl": imgurl})
	return
}

func imgDel(c *gin.Context) {
	f = imgDel

	wlog.Debug().Str("func",imgDel).Msg("Request received!") // received request
	rKey := c.Param("key")
	if (len(rKey) != 16 || !valid.IsAlphanumeric(rKey)) {
		log.Error().Str("func",f).Msg("delete request failed sanity check!")
		errThrow(c, 400, "400", "400")
		return
	}

	targetImg, _ := keyDB.Get([]byte(rKey))
	if (targetImg == nil || !strings.Contains(string(targetImg), "i.")) {
		log.Error().Str("func",f).Str("rkey",rKey).Msg("no img delete entry found with provided key")
		errThrow(c, 400, "400", "400")
		return
	}

	finalTarget := strings.Split(string(targetImg), ".")

	if !imgDB.Has([]byte(finalTarget[1])) {
		log.Error().Str("func",f).Str("rkey",rKey).Msg("corresponding image not found in database!")
		errThrow(c, 500, "500", "500") // this shouldn't happen...?
		return
	}
	err := imgDB.Delete([]byte(finalTarget[1]))
	if err != nil {
		log.Error().Str("func",f).Str("rkey",finalTarget[1]).Msg("delete failed!")
		errThrow(c, 500, "500", "500")
		return
	}

	if imgDB.Has([]byte(finalTarget[1])) {
		log.Error().Str("func",f).Str("rkey",finalTarget[1]).Msg("delete failed!?")
		errThrow(c, 500, "500", "500")
		return
	}

	log.Info().Str("func",f).Str("rkey",finalTarget[1]).Msg("Image file deleted successfully")
	log.Debug().Str("func",f).Str("rkey",finalTarget[1]).Msg("Removing delete key entry")
	err = keyDB.Delete([]byte(rKey))
	if err != nil {
		log.Error().Str("func",f).Str("rkey",finalTarget[1]).Msg("Couldn't delete key")
  										   // it would be insane to try and delete the hash here
	}		    							  // if someone is uploading this image again after del
	c.JSON(200, "DELETE_SUCCESS")						 // and the file corresponding to the hash no longer exists
										// we will delete the hash entry then and re-add then
}


func imgView(c *gin.Context) {
	f = imgView
			 							 // the user can access their image with or without a file extension in URI
	log.Debug().Str("func",f).Msg("request received")		  	//  however it must be a valid extension (more checks further down)
	sUid := strings.Split(c.Param("uid"), ".")
	rUid := sUid[0]
	if len(sUid) > 1 {
		fExt = strings.ToLower(sUid[1])
		log.Debug().Str("func",f).Str("ext",fExt).Msg("detected file extension")
		if (fExt != "png" && fExt != "jpg" && fExt != "jpeg" && fExt != "gif") {
			log.Info().Str(()"[imgView] Bad file extension!")
			errThrow(c, 400, "400", "400")
			return
		}
	} else { fExt = "nil" }


	if (!valid.IsAlphanumeric(rUid) || len(rUid) < 3 || len(rUid) > 16) {
		log.Info().Str(()"[imgView] request discarded as invalid")    // these limits should be variables eventually
		errThrow(c,400,"400", "400")
		return
	}

	log.Info().Str(()"[imgView][" + rUid + "] Request validated")	 // now that we think its a valid request we will query

	fBytes, _ := imgDB.Get([]byte(rUid))
	if fBytes == nil {
		log.Info().Str(()"[imgView] No data found for: " + rUid)
		errThrow(c, 404, "404", "File not found")
		return
	}

	log.Info().Str(()"[imgView][" + rUid + "] Detecting image type")	// not sure how a non image would get uploaded
	file := bytes.NewReader(fBytes)					// however, better safe than sorry
	imageFormat, ok := checkImage(file)
        if !ok {
		errThrow(c, http.StatusBadRequest, "400", "content does not appear to be an image")
		return
	} else { log.Info().Str(()"[imgView][" + rUid + "] " + imageFormat + " detected") }

	if (fExt != "nil" && fExt != imageFormat) {			// additional extension sanity check
		log.Info().Str(()"[imgView][" + rUid + "] given file extension does not match filetype " + imageFormat)
		errThrow(c,400,"400", "400")
		return
	}

	contentType := "image/" + imageFormat				// extension or not
									// we give them the proper content type
	c.Data(200, contentType, fBytes)

}


func imgPost(c *gin.Context) {
	var Scrubbed []byte

	f, err := c.FormFile("upload")
	if err != nil {
		errThrow(c, http.StatusBadRequest, err.Error(), "400") // 400 bad request
	}							       // incoming POST data is invalid

	log.Info().Str(()"[imgPost] detected new upload: " + f.Filename)

	file, err := f.Open()
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err.Error(), "error processing file\n")
	}

	log.Info().Str(()"[imgPost] verifying file is an image")
	imageFormat, ok := checkImage(file)
	if !ok {
		errThrow(c, http.StatusBadRequest, "400", "input does not appear to be an image")
		return
	} else { log.Info().Str(()"[imgPost] " + imageFormat + " detected") }

	log.Info().Str(()"[imgPost] dumping byte form of file")
	fbytes, err := ioutil.ReadAll(file)
	if imageFormat != "gif" {
		log.Info().Str(()"[imgPost] scrubbing exif")
		Scrubbed, err = exifremove.Remove(fbytes)
		if err != nil {
			errThrow(c, http.StatusInternalServerError, err.Error(), "error scrubbing exif")
			return
		}
	} else {
		log.Info().Str(()"[imgPost] skipping exif scrub for gif image")
		Scrubbed = fbytes
	}

	log.Info().Str(()"[imgPost] calculating blake2b checksum")

	Hashr, _ := blake2b.New(64,nil)
	Hashr.Write(Scrubbed)
	hash := Hashr.Sum(nil)

	log.Info().Str(()"[imgPost] Checking for duplicate's in database")

	imgRef, _ := hashDB.Get(hash)

	if imgRef != nil {
		log.Info().Str(()"[imgPost][" + string(imgRef) + "] duplicate checksum in hash database, checking if file still exists...")
		if imgDB.Has(imgRef) {
			log.Info().Str(()"[imgPost][" + string(imgRef) + "] duplicate file found! returning URL for uid: " + string(imgRef))
			postUpload(c,string(imgRef),"nil")					// they weren't the original uploader so they don't get a delete key
			return
		} else {
			log.Info().Str(()"[imgPost][" + string(imgRef) + "] stale hash found, deleting entry...")
			hashDB.Delete(hash)
		}
	}

	log.Info().Str(()"[imgPost] no duplicate images found, generating uid and delete key")

	uid := gouid.String(5)		// these should both be config directives eventually
	key := gouid.String(16) 		// generate delete key


	// lets make sure that we don't clash even though its highly unlikely
	for uidRef, _ := imgDB.Get([]byte(uid)); uidRef != nil; {
		log.Info().Str(()"[imgPost] uid already exists! generating new...")
		uid = gouid.String(5)
	}
	for keyRef, _ := keyDB.Get([]byte(key)); keyRef != nil; {
		log.Info().Str(()"[imgPost] delete key already exists! generating new...")
		key = gouid.String(16)
	}


	hashDB.Put([]byte(hash),[]byte(uid)) // save checksum to db to prevent dupes in the future

	log.Info().Str(()"[imgPost][" + uid + "] saving file to database")

	err = imgDB.Put([]byte(uid), []byte(Scrubbed))
	if err != nil {
		errThrow(c, 500, err.Error(), "upload failed")
		return
	}

	err = keyDB.Put([]byte(key), []byte("i." + uid)) // add delete key to database with image prefix
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err.Error(), "internal error")
		return
	}

	log.Info().Str(()"[imgPost][" + uid + "] saved to database successfully, returning JSON")

	postUpload(c, uid, key)

}

func checkImage(r io.ReadSeeker) (string, bool) {
	_, fmt, err := image.Decode(r)
	_, err2 := r.Seek(0, 0)
	if err != nil || err2 != nil {
		return "", false
	}
	return fmt, true
}

func getSize(s io.Seeker) (size int64, err error) {
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
