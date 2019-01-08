package dbw_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDbw(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dbw Suite")
}
