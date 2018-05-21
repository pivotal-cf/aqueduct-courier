package file_test

import (
	"crypto/md5"
	"encoding/base64"
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

	Describe("TarFilePath", func() {
		It("returns the sourceTarFile path", func() {
			tarFilePath := "path/to/tarfile"
			tarReader := NewTarReader(tarFilePath)
			Expect(tarReader.TarFilePath()).To(Equal(tarFilePath))
		})
	})

	Describe("FileMd5s", func() {
		It("returns the list of fileNames in the tarfile", func() {
			tarFilePath := generateValidTarFile(tempDir)
			tarReader := NewTarReader(tarFilePath)

			file1Content, err := tarReader.ReadFile("file1")
			Expect(err).NotTo(HaveOccurred())
			file1Sum := md5.Sum(file1Content)
			file1Md5 := base64.StdEncoding.EncodeToString(file1Sum[:])

			file2Content, err := tarReader.ReadFile("file2")
			Expect(err).NotTo(HaveOccurred())
			file2Sum := md5.Sum(file2Content)
			file2Md5 := base64.StdEncoding.EncodeToString(file2Sum[:])

			fileMd5s, err := tarReader.FileMd5s()
			Expect(err).NotTo(HaveOccurred())
			Expect(fileMd5s).To(Equal(map[string]string{
				"file1": file1Md5,
				"file2": file2Md5,
			}))
		})

		It("it fails if the sourceTarFile does not exist", func() {
			reader := NewTarReader("path/to/not/real/file")

			fileNames, err := reader.FileMd5s()
			Expect(fileNames).To(Equal(map[string]string{}))
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(OpenTarFileFailureFormat, "path/to/not/real/file"))))
		})

		It("it errors if the file specified does not have real tar headers", func() {
			invalidFilePath := filepath.Join(tempDir, "not-a-tarfile")
			Expect(ioutil.WriteFile(invalidFilePath, []byte("not-tar"), 0644)).To(Succeed())

			reader := NewTarReader(invalidFilePath)

			fileNames, err := reader.FileMd5s()
			Expect(fileNames).To(Equal(map[string]string{}))
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UnexpectedFileTypeFormat, invalidFilePath))))
		})
	})
})

func generateValidTarFile(destinationDir string) string {
	tarFilePath := filepath.Join(destinationDir, "some-tar-file")

	writer, err := NewTarWriter(tarFilePath)
	Expect(err).NotTo(HaveOccurred())
	defer writer.Close()

	Expect(writer.AddFile([]byte("contents1"), "file1")).To(Succeed())
	Expect(writer.AddFile([]byte("contents2"), "file2")).To(Succeed())

	return tarFilePath
}
