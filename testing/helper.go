package testing

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func WriteBlobInDir(store string, dir string, data []byte) (string, error) {
	id := uuid.NewSHA1(uuid.Nil, data)
	err := os.MkdirAll(filepath.Join(store, dir), 0755)
	if err != nil {
		return "", err
	}

	path := filepath.Join(store, dir, id.String())
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return "", err
	}

	return path, nil
}

func WriteBlob(store string, data []byte) (string, error) {
	id := uuid.NewSHA1(uuid.Nil, data)
	dir := id.String()[0:2]
	return WriteBlobInDir(store, dir, data)
}

func WriteTarBlob(store string, data string) (string, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	_, err := zw.Write([]byte(data))
	if err != nil {
		return "", err
	}

	if err := zw.Close(); err != nil {
		return "", err
	}
	return WriteBlob(store, buf.Bytes())
}
