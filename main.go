package main

import (
	"log"
	"os"

	"github.com/starkandwayne/bosh-dns-local-fsnotify/dbw"
)

const (
	store   = "/var/vcap/store/blobstore/store"
	records = "/var/vcap/instance/dns/records.json"
)

func main() {
	log := log.New(os.Stdout, "", 0)

	w := dbw.NewDNSBlobWatcher(dbw.Config{
		StorePath:   store,
		RecordsPath: records,
		Logger:      log,
	})

	quit := make(chan bool)
	w.Start(quit)
}
