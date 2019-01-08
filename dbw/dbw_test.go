package dbw_test

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	. "github.com/starkandwayne/bosh-dns-local-fsnotify/dbw"
	. "github.com/starkandwayne/bosh-dns-local-fsnotify/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DNSBlobWatcher", func() {
	var (
		store   string
		records string
		watcher func() Watcher
		quit    chan bool
		done    chan bool
	)

	BeforeEach(func() {
		r, err := ioutil.TempFile("", "records.*.json")
		Expect(err).NotTo(HaveOccurred())
		records = r.Name()

		store, err = ioutil.TempDir("", "store")
		Expect(err).NotTo(HaveOccurred())

		log := log.New(GinkgoWriter, "", 0)

		quit = make(chan bool)
		done = make(chan bool)

		watcher = func() Watcher {
			return NewDNSBlobWatcher(Config{
				StorePath:   store,
				RecordsPath: records,
				Logger:      log,
			})
		}
	})

	Context("given a blobstore without dns blobs", func() {
		BeforeEach(func() {
			_, err := WriteTarBlob(store, "blob1")
			Expect(err).NotTo(HaveOccurred())
			_, err = WriteTarBlob(store, "blob2")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create empty records file", func() {
			By("Starting the watcher")
			go watcher().Start(quit, done)
			Eventually(func() ([]byte, error) {
				return ioutil.ReadFile(records)
			}).Should(MatchJSON(`{"records":[]}`))

		})

		It("should detect new dns record blobs", func() {
			By("Starting the watcher")
			go watcher().Start(quit, done)
			<-done
			blob := []byte(`{"records":["last"]}`)
			_, err := WriteBlob(store, blob)
			By("blob written")
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() ([]byte, error) {
				return ioutil.ReadFile(records)
			}, "5s", "1s").Should(MatchJSON(blob))

		})
	})

	Context("when the records file and its directory do not exist", func() {
		BeforeEach(func() {
			rd, err := ioutil.TempDir("", "instance")
			Expect(err).NotTo(HaveOccurred())
			records = filepath.Join(rd, "dns", "records.json")
		})

		It("should create records file and directory", func() {
			By("Starting the watcher")
			go watcher().Start(quit, done)
			Eventually(func() ([]byte, error) {
				return ioutil.ReadFile(records)
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
			_, err = WriteTarBlob(store, "blob1")
			Expect(err).NotTo(HaveOccurred())
			_, err = WriteTarBlob(store, "blob2")
			Expect(err).NotTo(HaveOccurred())

		})

		It("should update records to match latest blob", func() {
			By("Starting the watcher")
			go watcher().Start(quit, done)
			rl, err := ioutil.ReadFile(lastDNSBlob)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() ([]byte, error) {
				return ioutil.ReadFile(records)
			}).Should(MatchJSON(rl))
		})

		It("should watch for dns blobs in existing directories", func() {
			_, err := WriteBlobInDir(store, "00", []byte(`{"records":["first"]}`))
			By("Starting the watcher")
			go watcher().Start(quit, done)
			<-done
			lastDNSBlob, err = WriteBlobInDir(store, "00", []byte(`{"records":["new"]}`))
			rl, err := ioutil.ReadFile(lastDNSBlob)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() ([]byte, error) {
				return ioutil.ReadFile(records)
			}).Should(MatchJSON(rl))
		})

		It("should watch for dns blobs in new directories", func() {
			By("Starting the watcher")
			go watcher().Start(quit, done)
			<-done
			first := []byte(`{"records":["first"]}`)
			_, err := WriteBlobInDir(store, "00", first)
			Eventually(func() ([]byte, error) {
				return ioutil.ReadFile(records)
			}).Should(MatchJSON(first))

			new := []byte(`{"records":["new"]}`)
			_, err = WriteBlobInDir(store, "00", new)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() ([]byte, error) {
				return ioutil.ReadFile(records)
			}).Should(MatchJSON(new))
		})

	})

	AfterEach(func() {
		close(quit)
		os.Remove(records)
		os.RemoveAll(store)

	})
})
