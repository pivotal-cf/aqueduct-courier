package file_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/pivotal-cf/aqueduct-courier/file"
)

var _ = Describe("TarReader", func() {
	var (
		tempDir           string
		sourceTarFilePath string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		sourceTarFilePath = generateValidTarFile(tempDir)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	})

	Describe("ReadFile", func() {
		It("reads the specified file from the tar archive and returns its contents", func() {
			reader := NewTarReader(sourceTarFilePath)

			contents2, err := reader.ReadFile("file2")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents2)).To(Equal("contents2"))

			contents1, err := reader.ReadFile("file1")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents1)).To(Equal("contents1"))
		})

		It("it fails if the sourceTarFile does not exist", func() {
			reader := NewTarReader("path/to/not/real/file")

			contents, err := reader.ReadFile("file-doesnt-exist")
			Expect(string(contents)).To(Equal(""))
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(OpenTarFileFailureFormat, "path/to/not/real/file"))))
		})

		It("it errors if the file specified is not found in the archive", func() {
			reader := NewTarReader(sourceTarFilePath)

			contents, err := reader.ReadFile("file-doesnt-exist")
			Expect(string(contents)).To(Equal(""))
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UnableToFindFileFormat, "file-doesnt-exist"))))
		})

		It("it errors if the file specified does not have real tar headers", func() {
			invalidFilePath := filepath.Join(tempDir, "not-a-tarfile")
			Expect(ioutil.WriteFile(invalidFilePath, []byte("not-tar"), 0644)).To(Succeed())

			reader := NewTarReader(invalidFilePath)

			contents, err := reader.ReadFile("file1")
			Expect(string(contents)).To(Equal(""))
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UnexpectedFileTypeFormat, invalidFilePath))))
		})
	})
})

func generateValidTarFile(destinationDir string) string {
	tarFilePath := filepath.Join(destinationDir, "some-tar-file")

	writer, err := NewTarWriter(tarFilePath)
	Expect(err).NotTo(HaveOccurred())

	Expect(writer.AddFile([]byte("contents1"), "file1")).To(Succeed())
	Expect(writer.AddFile([]byte("contents2"), "file2")).To(Succeed())

	Expect(writer.Close()).To(Succeed())

	return tarFilePath
}
