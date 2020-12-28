package main

import(
	valid "github.com/asaskevich/govalidator"
	exifremove "github.com/scottleedavis/go-exif-remove"
	"github.com/twharmon/gouid"
	"github.com/gin-gonic/gin"
	"encoding/json"
	_ "image/gif"
	"crypto/md5"
	"io/ioutil"
	"net/http"
	"image"
	"bytes"
	"fmt"
	"io"
)

type Post struct {
    Imgurl  string `json:"Imgurl"`
    Delkey string  `json:"Delkey"`
}

func postUpload(c *gin.Context, id string, key string) {
                imgurl := baseUrl + "i/" + string(id)
                keyurl := baseUrl + "i/d/" + string(key)

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

                c.JSON(200, string(p))                                  // they weren't the original uploader so they don't get a delete key
                return
}

func imgView(c *gin.Context) {
	rUid := c.Param("uid")
	fmt.Println("[imgView] Received request")
	if (valid.IsAlphanumeric(rUid)) {
		fmt.Println("[imgView][" + rUid + "] Request validated")
		fBytes, _ := imgDB.Get([]byte(rUid))
		if fBytes == nil {
			fmt.Println("[imgView] No data found for: " + rUid)
			errThrow(c, 404, "404", "File not found")
			return
		}

		fmt.Println("[imgView][" + rUid + "] Detecting image type")
		file := bytes.NewReader(fBytes)
		imageFormat, ok := checkImage(file)
	        if !ok {
			errThrow(c, http.StatusBadRequest, "bad request", "content does not appear to be an image")
			return
		} else { fmt.Println("[imgView][" + rUid + "] " + imageFormat + " detected") }

		contentType := "image/" + imageFormat

		c.Data(200, contentType, fBytes)
	}
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

	fmt.Println("[imgPost] calculating MD5 hash")

	Hashr := md5.New()
	Hashr.Write(Scrubbed)
	hash := Hashr.Sum(nil)

	fmt.Println("[imgPost] Checking for duplicate's in database")

	imgRef, _ := md5DB.Get(hash)

	if imgRef != nil {
		fmt.Println("[imgPost][" + string(imgRef) + "] duplicate file found in md5 database, returning URL for uid: " + string(imgRef))
		postUpload(c,string(imgRef),"nil")					// they weren't the original uploader so they don't get a delete key
		return
	}

	fmt.Println("[imgPost] no duplicate md5s found, generating uid and delete key")

	uid := gouid.String(5)
	key := gouid.String(16) 		// generate delete key


	// lets make sure that we don't clash even though its highly unlikely
	for uidRef, _ := imgDB.Get([]byte(uid)); uidRef != nil; {
		fmt.Println("[imgPost] uid already exists! generating new...")
		uid = gouid.String(5)
	}
	for keyRef, _ := keyDB.Get([]byte(uid)); keyRef != nil; {
		fmt.Println("[imgPost] delete key already exists! generating new...")
		key = gouid.String(16)
	}


	md5DB.Put([]byte(hash),[]byte(uid)) // save md5 to db to prevent dupes in the future

	fmt.Println("[imgPost][" + uid + "] saving file to database")

	err = imgDB.Put([]byte(uid), []byte(Scrubbed))
	if err != nil {
		errThrow(c, http.StatusInternalServerError, err.Error(), "upload failed")
		return
	}

	err = keyDB.Put([]byte("i." + uid), []byte(key)) // add delete key to database with image prefix
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
