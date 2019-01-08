package dbw

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	StorePath   string
	RecordsPath string
	Logger      *log.Logger
}

type Watcher struct {
	store   string
	records string
	log     *log.Logger
}

func NewDNSBlobWatcher(c Config) Watcher {
	return Watcher{
		store:   filepath.Clean(c.StorePath),
		records: c.RecordsPath,
		log:     c.Logger,
	}
}

func (w Watcher) Start(quit chan bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				w.log.Println("event:", event)
				if event.Op&fsnotify.Create == fsnotify.Create {
					if len(strings.TrimPrefix(event.Name, w.store)) == 3 {
						err := watcher.Add(event.Name)
						if err != nil {
							w.log.Println("error:", err)
							continue
						}
					}
				}

				err := w.findAndUpdateLatestRecords()
				if err != nil {
					w.log.Println("error:", err)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				w.log.Println("error:", err)
			case <-quit:
				return
			}
		}
	}()

	err = watcher.Add(w.store)

	dirs, err := ioutil.ReadDir(w.store)
	if err != nil {
		w.log.Fatal(err)
	}

	for _, dir := range dirs {
		err = watcher.Add(filepath.Join(w.store, dir.Name()))
		if err != nil {
			w.log.Fatal(err)
		}
	}

	err = w.findAndUpdateLatestRecords()
	if err != nil {
		w.log.Fatal(err)
	}
}

func (w Watcher) findAndUpdateLatestRecords() error {
	var candidate *os.File

	files, err := filepath.Glob(fmt.Sprintf("%s/*/*", w.store))
	if err != nil {
		return err
	}
	fmt.Println()

	for _, f := range files {
		fi, err := os.Stat(f)
		if err != nil {
			continue
		}

		if candidate != nil {
			cfi, err := candidate.Stat()
			if err != nil {
				continue
			}
			if fi.ModTime().Before(cfi.ModTime()) {
				continue
			}
		}

		fr, err := os.Open(f)
		if err != nil {
			continue
		}
		defer fr.Close()
		buffer := make([]byte, 512)
		n, err := fr.Read(buffer)
		if err != nil && err != io.EOF {
			continue
		}

		fr.Seek(0, 0)
		contentType := http.DetectContentType(buffer[:n])
		if !strings.Contains(contentType, "text/plain") {
			continue
		}

		candidate, _ = os.Open(f)
	}

	if candidate == nil {
		w.log.Println("no candidate found, creating empty records file")
		data := []byte(`{"records":[]}`)
		err = ioutil.WriteFile(w.records, data, 0644)
		if err != nil {
			return err
		}
		return nil
	}

	w.log.Println("candidate:", candidate.Name())
	data, err := ioutil.ReadFile(candidate.Name())
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(w.records, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
