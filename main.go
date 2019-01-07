package main

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

const (
	store   = "/var/vcap/store/blobstore/store"
	records = "/var/vcap/instance/dns/records.json"
)

func findAndUpdateLatestRecords() error {
	var candidate os.FileInfo

	files, err := filepath.Glob(fmt.Sprintf("%s/*/*", store))
	if err != nil {
		return err
	}

	for _, f := range files {
		fi, err := os.Stat(f)
		if err != nil {
			continue
		}

		if candidate != nil && fi.ModTime().Before(candidate.ModTime()) {
			continue
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

		candidate = fi
		log.Println("candidate:", f)
	}
	return nil
}

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Create == fsnotify.Create {
					if len(strings.TrimPrefix(event.Name, store)) == 3 {
						err := watcher.Add(event.Name)
						if err != nil {
							log.Println("error:", err)
							continue
						}
					}
				}

				err := findAndUpdateLatestRecords()
				if err != nil {
					log.Println("error:", err)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(store)

	dirs, err := ioutil.ReadDir(store)
	if err != nil {
		log.Fatal(err)
	}

	for _, dir := range dirs {
		err = watcher.Add(filepath.Join(store, dir.Name()))
		if err != nil {
			log.Fatal(err)
		}
	}

	err = findAndUpdateLatestRecords()
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
