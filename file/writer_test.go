package file_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"path/filepath"
	"strings"

	"os"

	"fmt"

	. "github.com/pivotal-cf/aqueduct-courier/file"
	"github.com/pivotal-cf/aqueduct-courier/file/filefakes"
	"github.com/pkg/errors"
)


const (
	UnixTimestampRegexp = `[0-9]{10}`
)

var _ = Describe("Writer", func() {

	Describe("Write", func() {
		var (
			data *filefakes.FakeData
			dir  string
		)

		BeforeEach(func() {
			data = new(filefakes.FakeData)
			data.NameReturns("best-name-evar")
			data.ContentReturns(strings.NewReader("reader-of-things"))
			data.ContentTypeReturns("json")
			var err error
			dir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		It("writes data content to a file", func() {
			w := Writer{}
			w.Write(data, dir)
			content, err := ioutil.ReadFile(filepath.Join(dir, "best-name-evar.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("reader-of-things"))
			_, err = os.Stat(filepath.Join(dir, "best-name-evar.json"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error when reader errors", func() {
			reader := new(filefakes.FakeReader)
			reader.ReadReturns(0, errors.New("reading things is hard"))
			data.ContentReturns(reader)
			w := Writer{}
			err := w.Write(data, dir)
			Expect(err).To(MatchError(ContainSubstring(ContentWritingErrorFormat, data.Name())))
			Expect(err).To(MatchError(ContainSubstring("reading things is hard")))
		})

		It("errors if writing the file returns an error", func() {
			nonExistentDir := "dir/that/does/not/ever/exist/like/ever"
			w := Writer{}
			err := w.Write(data, nonExistentDir)
			Expect(err).To(MatchError(ContainSubstring(CreateErrorFormat, data.Name())))
		})

		It("writes to a correctly named file, when the extension is set", func() {
			data.ContentTypeReturns("xml")

			w := Writer{}
			w.Write(data, dir)
			content, err := ioutil.ReadFile(filepath.Join(dir, "best-name-evar.xml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("reader-of-things"))
		})
	})

	Describe("Mkdir", func() {
		It("makes a directory and returns the path", func() {
			dir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			w := Writer{}

			path, err := w.Mkdir(dir)
			Expect(err).NotTo(HaveOccurred())

			_, err = ioutil.ReadDir(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(MatchRegexp(escapeWindowsPathRegex(
				filepath.Join(dir, fmt.Sprintf(`%s%s$`, OutputDirPrefix, UnixTimestampRegexp)))),
			)
		})

		It("errors if the directory cannot be created", func() {
			w := Writer{}

			nonExistentDir := "/non-exists/ever/please/do/not"
			path, err := w.Mkdir(nonExistentDir)
			Expect(path).To(Equal(""))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(
				fmt.Sprintf(
					`Failed creating directory %s%s:`,
					escapeWindowsPathRegex(filepath.Join(nonExistentDir, OutputDirPrefix)),
					UnixTimestampRegexp,
				),
			))
		})
	})
})

//go:generate counterfeiter . reader
type reader interface {
	Read(p []byte) (n int, err error)
}


func escapeWindowsPathRegex(path string) string {
	return strings.Replace(path, `\`, `\\`, -1)
}
