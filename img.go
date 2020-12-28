package main

import(
	valid "github.com/asaskevich/govalidator"
	exifremove "github.com/scottleedavis/go-exif-remove"
	"golang.org/x/crypto/blake2b"
	"github.com/twharmon/gouid"
	"github.com/gin-gonic/gin"
	"encoding/json"
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
		keyurl := "image_is_duplicate"
		if key != "nil" { keyurl = baseUrl + "d/i/" + string(key) }

		d := Post{
			Imgurl: imgurl,
			Delkey: keyurl,
		}


                var p []byte
                p, err := json.Marshal(d)
                if err != nil {
                        errThrow(c, http.StatusInternalServerError, err.Error(), "internal error")
                        return
                }

		fmt.Println("[imgPost]["+id+"] Success: " + imgurl + " " + keyurl)
                c.JSON(200, string(p))
                return
}

func imgDel(c *gin.Context) {
	fmt.Println("[imgDel] Received request")
	rKey := c.Param("key")
	if (len(rKey) != 16 || !valid.IsAlphanumeric(rKey)) {
		fmt.Println("[imgDel] delete request failed sanity check")
		errThrow(c, 400, "400", "Bad request")
		return
	}

	targetImg, _ := keyDB.Get([]byte(rKey))
	if (targetImg == nil || !strings.Contains(string(targetImg), "i.")) {
		fmt.Println("[imgDel] no img delete entry found with key: " + rKey)
		errThrow(c, 400, "400", "Bad request")
		return
	}

	finalTarget := strings.Split(string(targetImg), ".")

	if !imgDB.Has([]byte(finalTarget[1])) {
		fmt.Println("[imgDel]["+finalTarget[1]+"] corresponding image not found in database??")
		errThrow(c, 500, "500", "Internal error") // this shouldn't happen...?
		return
	}
	err := imgDB.Delete([]byte(finalTarget[1]))
	if err != nil {
		fmt.Println("[imgDel]["+finalTarget[1]+"] Delete failed!!")
		errThrow(c, 500, err.Error(), "Internal error")
		return
	}

	if imgDB.Has([]byte(finalTarget[1])) {
		fmt.Println("[imgDel]["+finalTarget[1]+"] Delete failed!?")
		errThrow(c, 500, err.Error(), "Internal error")
		return
	}

	fmt.Println("[imgDel]["+finalTarget[1]+"] Image file deleted successfully") 
	fmt.Println("[imgDel]["+finalTarget[1]+"] Removing delete key entry")
	err = keyDB.Delete([]byte(rKey))
	if err != nil {
		fmt.Println("[imgDel]["+finalTarget[1]+"] Couldn't delete key")    // it would be insane to try and delete the hash here
	}		    							  // if someone is uploading this image again after del
	c.JSON(200, "OK")							 // and the file corresponding to the hash no longer exists
										// we will delete the hash entry then and re-add then
}


func imgView(c *gin.Context) {				// the user can access their image with or without a file extension in URI
	fmt.Println("[imgView] Received request")	// however it must be a valid extension (more checks further down)
	sUid := strings.Split(c.Param("uid"), ".")
	rUid := sUid[0]
	if len(sUid) > 1 {
		fExt = strings.ToLower(sUid[1])
		fmt.Println("[imgView] Detected file extension: " + fExt)
		if (fExt != "png" && fExt != "jpg" && fExt != "jpeg" && fExt != "gif") {
			fmt.Println("[imgView] Bad file extension!")
			errThrow(c, 400, "400", "Bad request")
			return
		}
	} else { fExt = "nil" }


	if (!valid.IsAlphanumeric(rUid) || len(rUid) < 3 || len(rUid) > 16) {
		fmt.Println("[imgView] request discarded as invalid")    // these limits should be variables eventually
		errThrow(c,400,"400", "Bad request")
		return
	}

	fmt.Println("[imgView][" + rUid + "] Request validated")	 // now that we think its a valid request we will query

	fBytes, _ := imgDB.Get([]byte(rUid))
	if fBytes == nil {
		fmt.Println("[imgView] No data found for: " + rUid)
		errThrow(c, 404, "404", "File not found")
		return
	}

	fmt.Println("[imgView][" + rUid + "] Detecting image type")	// not sure how a non image would get uploaded
	file := bytes.NewReader(fBytes)					// however, better safe than sorry
	imageFormat, ok := checkImage(file)
        if !ok {
		errThrow(c, http.StatusBadRequest, "bad request", "content does not appear to be an image")
		return
	} else { fmt.Println("[imgView][" + rUid + "] " + imageFormat + " detected") }

	if (fExt != "nil" && fExt != imageFormat) {			// additional extension sanity check
		fmt.Println("[imgView][" + rUid + "] given file extension does not match filetype " + imageFormat)
		errThrow(c,400,"400", "Bad request")
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
		errThrow(c, http.StatusBadRequest, err.Error(), "no file detected within request\n")
	}

	fmt.Println("[imgPost] detected new upload: " + f.Filename)

	file, err := f.Open()
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err.Error(), "error processing file\n")
	}

	fmt.Println("[imgPost] verifying file is an image")
	imageFormat, ok := checkImage(file)
	if !ok {
		errThrow(c, http.StatusBadRequest, "400", "input does not appear to be an image")
		return
	} else { fmt.Println("[imgPost] " + imageFormat + " detected") }

	fmt.Println("[imgPost] dumping byte form of file")
	fbytes, err := ioutil.ReadAll(file)
	if imageFormat != "gif" {
		fmt.Println("[imgPost] scrubbing exif")
		Scrubbed, err = exifremove.Remove(fbytes)
		if err != nil {
			errThrow(c, http.StatusInternalServerError, err.Error(), "error scrubbing exif")
			return
		}
	} else {
		fmt.Println("[imgPost] skipping exif scrub for gif image")
		Scrubbed = fbytes
	}

	fmt.Println("[imgPost] calculating blake2b checksum")

	Hashr, _ := blake2b.New(64,nil)
	Hashr.Write(Scrubbed)
	hash := Hashr.Sum(nil)

	fmt.Println("[imgPost] Checking for duplicate's in database")

	imgRef, _ := hashDB.Get(hash)

	if imgRef != nil {
		fmt.Println("[imgPost][" + string(imgRef) + "] duplicate checksum in hash database, checking if file still exists...")
		if imgDB.Has(imgRef) {
			fmt.Println("[imgPost][" + string(imgRef) + "] duplicate file found! returning URL for uid: " + string(imgRef))
			postUpload(c,string(imgRef),"nil")					// they weren't the original uploader so they don't get a delete key
			return
		} else {
			fmt.Println("[imgPost][" + string(imgRef) + "] stale hash found, deleting entry...")
			hashDB.Delete(hash)
		}
	}

	fmt.Println("[imgPost] no duplicate images found, generating uid and delete key")

	uid := gouid.String(5)		// these should both be config directives eventually
	key := gouid.String(16) 		// generate delete key


	// lets make sure that we don't clash even though its highly unlikely
	for uidRef, _ := imgDB.Get([]byte(uid)); uidRef != nil; {
		fmt.Println("[imgPost] uid already exists! generating new...")
		uid = gouid.String(5)
	}
	for keyRef, _ := keyDB.Get([]byte(key)); keyRef != nil; {
		fmt.Println("[imgPost] delete key already exists! generating new...")
		key = gouid.String(16)
	}


	hashDB.Put([]byte(hash),[]byte(uid)) // save checksum to db to prevent dupes in the future

	fmt.Println("[imgPost][" + uid + "] saving file to database")

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

	fmt.Println("[imgPost][" + uid + "] saved to database successfully, returning JSON")

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
