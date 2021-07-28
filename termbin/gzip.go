package termbin

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

func gzipCompress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return data, err
	}
	if err := gz.Close(); err != nil {
		return data, err
	}
	return b.Bytes(), nil
}

// Deflate is an exported function to deflate gzip'd data - usable from other packages
func Deflate(data []byte) ([]byte, error) {
	var (
		gz  *gzip.Reader
		out []byte
		err error
	)
	r := bytes.NewReader(data)
	if gz, err = gzip.NewReader(r); err != nil {
		return data, err
	}
	if out, err = ioutil.ReadAll(gz); err != nil {
		return data, err
	}
	return out, nil
}
