package dbw_test

import (
	"io/ioutil"
	"log"
	"os"

	. "github.com/starkandwayne/bosh-dns-local-fsnotify/dbw"
	. "github.com/starkandwayne/bosh-dns-local-fsnotify/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DNSBlobWatcher", func() {
	var (
		store   string
		records string
		watcher Watcher
		quit    chan bool
	)

	BeforeEach(func() {
		r, err := ioutil.TempFile("", "records.*.json")
		Expect(err).NotTo(HaveOccurred())
		records = r.Name()

		store, err = ioutil.TempDir("", "store")
		Expect(err).NotTo(HaveOccurred())

		log := log.New(GinkgoWriter, "", 0)

		quit = make(chan bool)

		watcher = NewDNSBlobWatcher(Config{
			StorePath:   store,
			RecordsPath: records,
			Logger:      log,
		})
	})

	Context("given a blobstore without dns blobs", func() {
		BeforeEach(func() {
			// _, err := WriteBlob(store, []byte(`{"records":["first"]}`))
			// Expect(err).NotTo(HaveOccurred())
			// _, err = WriteBlob(store, []byte(`{"records":["second"]}`))
			// Expect(err).NotTo(HaveOccurred())
		})
		It("should create empty records file", func() {
			By("Starting the watcher")
			go watcher.Start(quit)
			Eventually(func() []byte {
				r, err := ioutil.ReadFile(records)
				if err != nil {
					return []byte{}
				}
				return r
			}).Should(MatchJSON(`{"records":[]}`))

		})
	})

	Context("given a blobstore with existing dns blobs", func() {
		var (
			lastDNSBlob string
		)

		BeforeEach(func() {
			_, err := WriteBlob(store, []byte(`{"records":["first"]}`))
			Expect(err).NotTo(HaveOccurred())
			_, err = WriteBlob(store, []byte(`{"records":["second"]}`))
			Expect(err).NotTo(HaveOccurred())
			lastDNSBlob, err = WriteBlob(store, []byte(`{"records":["last"]}`))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update records to match latest blob", func() {
			By("Starting the watcher")
			go watcher.Start(quit)
			rl, err := ioutil.ReadFile(lastDNSBlob)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() []byte {
				r, _ := ioutil.ReadFile(records)
				return r
			}).Should(MatchJSON(rl))
		})
	})

	AfterEach(func() {
		close(quit)
		os.Remove(records)
		os.RemoveAll(store)

	})
})
