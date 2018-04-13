package file_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"path/filepath"
	"strings"

	"os"

	. "github.com/pivotal-cf/aqueduct-courier/file"
	"github.com/pivotal-cf/aqueduct-courier/file/filefakes"
	"github.com/pkg/errors"
)

var _ = Describe("Writer", func() {

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
		fileInfo, err := os.Stat(filepath.Join(dir, "best-name-evar.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(fileInfo.Mode()).To(Equal(os.FileMode(0644)))
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

//go:generate counterfeiter . reader
type reader interface {
	Read(p []byte) (n int, err error)
}
