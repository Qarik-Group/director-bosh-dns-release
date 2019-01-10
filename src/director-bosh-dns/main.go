package main

import (
	"log"
	"os"

	"github.com/starkandwayne/director-bosh-dns-release/src/director-bosh-dns/dbw"
)

const (
	store   = "/var/vcap/store/blobstore/store"
	records = "/var/vcap/data/director-bosh-dns/records.json"
)

func main() {
	log := log.New(os.Stdout, "", 0)

	w := dbw.NewDNSBlobWatcher(dbw.Config{
		StorePath:   store,
		RecordsPath: records,
		Logger:      log,
	})

	quit := make(chan bool)
	done := make(chan bool)
	w.Start(quit, done)
	<-done
}
