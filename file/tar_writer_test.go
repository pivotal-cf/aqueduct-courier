package file_test

import (
	"archive/tar"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mholt/archiver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/pivotal-cf/aqueduct-courier/file"
)

var _ = Describe("TarWriter", func() {
	var (
		tempDir             string
		expectedTarFilePath string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		expectedTarFilePath = filepath.Join(tempDir, "some-tar-file")
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	})

	Describe("NewTarWriter", func() {
		It("creates the specified writer and tar file", func() {
			writer, err := NewTarWriter(expectedTarFilePath)
			Expect(err).NotTo(HaveOccurred())
			defer writer.Close() // for windows

			Expect(writer).ToNot(BeNil())
			Expect(writer).To(BeAssignableToTypeOf(&TarWriter{}))
			Expect(expectedTarFilePath).To(BeAnExistingFile())
		})

		It("errors when the tar file specified cannot be created", func() {
			nonexistentFile := "path/that/does/not/exist"
			writer, err := NewTarWriter(nonexistentFile)
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(CreateTarFileFailureFormat, nonexistentFile))))
			Expect(writer).To(BeNil())
		})
	})

	Describe("AddFile", func() {
		It("adds file contents to specified tar archive", func() {
			writer, err := NewTarWriter(expectedTarFilePath)
			Expect(err).NotTo(HaveOccurred())
			defer writer.Close() // for windows

			contents1 := []byte("best-contents1")
			contentsName1 := "contents-name1"
			err = writer.AddFile(contents1, contentsName1)
			Expect(err).NotTo(HaveOccurred())

			contents2 := []byte("best-contents2")
			contentsName2 := "contents-name2"
			err = writer.AddFile(contents2, contentsName2)
			Expect(err).NotTo(HaveOccurred())

			tarContentsDir, err := ioutil.TempDir(tempDir, "")
			Expect(err).NotTo(HaveOccurred())

			Expect(archiver.Tar.Open(expectedTarFilePath, tarContentsDir)).To(Succeed())

			outputContents, err := ioutil.ReadFile(filepath.Join(tarContentsDir, contentsName1))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(outputContents)).To(Equal(string(contents1)))

			outputContents, err = ioutil.ReadFile(filepath.Join(tarContentsDir, contentsName2))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(outputContents)).To(Equal(string(contents2)))
		})
	})

	Describe("Close", func() {
		It("causes errors when adding additional files after Close", func() {
			writer, err := NewTarWriter(expectedTarFilePath)
			Expect(err).NotTo(HaveOccurred())

			Expect(writer.Close()).To(Succeed())

			err = writer.AddFile([]byte{}, "some-file")
			Expect(err).To(MatchError(ContainSubstring(tar.ErrWriteAfterClose.Error())))
		})

		It("errors with subsequent calls to Close", func() {
			writer, err := NewTarWriter(expectedTarFilePath)
			Expect(err).NotTo(HaveOccurred())

			Expect(writer.Close()).To(Succeed())

			Expect(writer.Close()).To(MatchError(ContainSubstring(CloseWriterFailureMessage)))
		})
	})
})
