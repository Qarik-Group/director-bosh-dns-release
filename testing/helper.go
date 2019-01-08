package testing

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func WriteBlob(store string, data []byte) (string, error) {
	id := uuid.NewSHA1(uuid.Nil, data)
	dir := id.String()[0:2]
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
