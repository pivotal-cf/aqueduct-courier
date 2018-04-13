package file_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFileWriter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "File Suite")
}
