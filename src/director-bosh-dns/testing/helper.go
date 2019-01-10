package testing

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type BlobStore struct {
	Store       string
	lastModTime time.Time
}

func NewBlobStore(store string) *BlobStore {
	return &BlobStore{
		Store:       store,
		lastModTime: time.Now().Add(-time.Hour * 24),
	}
}

func (b *BlobStore) WriteBlobInDir(dir string, data []byte) (string, error) {
	id := uuid.NewSHA1(uuid.Nil, data)
	err := os.MkdirAll(filepath.Join(b.Store, dir), 0755)
	if err != nil {
		return "", err
	}

	path := filepath.Join(b.Store, dir, id.String())
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return "", err
	}

	b.lastModTime = b.lastModTime.Add(time.Second * 2)
	if err := os.Chtimes(path, b.lastModTime, b.lastModTime); err != nil {
		return "", err
	}

	return path, nil
}

func (b *BlobStore) WriteBlob(data []byte) (string, error) {
	id := uuid.NewSHA1(uuid.Nil, data)
	dir := id.String()[0:2]
	return b.WriteBlobInDir(dir, data)
}

func (b *BlobStore) WriteTarBlob(data string) (string, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	_, err := zw.Write([]byte(data))
	if err != nil {
		return "", err
	}

	if err := zw.Close(); err != nil {
		return "", err
	}
	return b.WriteBlob(buf.Bytes())
}
