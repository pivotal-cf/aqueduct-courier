package ops_test

import (
	"net/http"

	"fmt"

	"io/ioutil"

	"path/filepath"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	. "github.com/pivotal-cf/aqueduct-courier/ops"
)

var _ = Describe("Sender", func() {
	var (
		dataLoader    *ghttp.Server
		sender        SendExecutor
		dataDirectory string
	)

	BeforeEach(func() {
		var err error
		dataLoader = ghttp.NewServer()
		sender = SendExecutor{}

		dataDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		Expect(ioutil.WriteFile(filepath.Join(dataDirectory, "some-file"), []byte(""), 0644)).To(Succeed())
	})

	AfterEach(func() {
		dataLoader.Close()
	})

	It("posts to the data loader with the correct API key in the header and the file as content", func() {
		dir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		Expect(ioutil.WriteFile(filepath.Join(dir, "data-file1"), []byte("data-file1-contents"), 0644)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(dir, "data-file2"), []byte("data-file2-contents"), 0644)).To(Succeed())

		dataLoader.RouteToHandler(http.MethodPost, PostPath, ghttp.CombineHandlers(
			ghttp.VerifyHeader(http.Header{
				"Authorization": []string{"Token some-key"},
			}),
			func(w http.ResponseWriter, req *http.Request) {
				file, fileHeaders, err := req.FormFile("data")
				Expect(err).ToNot(HaveOccurred())
				contents, err := ioutil.ReadAll(file)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal(fileHeaders.Filename + "-contents"))
			},
			ghttp.RespondWith(http.StatusCreated, ""),
		))

		Expect(sender.Send(dir, dataLoader.URL(), "some-key")).To(Succeed())
		reqs := dataLoader.ReceivedRequests()
		Expect(len(reqs)).To(Equal(2))

		verifyFileSentInRequest("data-file1", reqs[0])
		verifyFileSentInRequest("data-file2", reqs[1])
	})

	It("does not post data files in subdirectories of the data dir", func() {
		Expect(os.Mkdir(filepath.Join(dataDirectory, "dir-not-to-read"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(dataDirectory, "dir-not-to-read", "file-not-to-send"), []byte(""), 0644)).To(Succeed())

		dataLoader.RouteToHandler(http.MethodPost, PostPath, ghttp.CombineHandlers(
			ghttp.RespondWith(http.StatusCreated, ""),
			func(w http.ResponseWriter, req *http.Request) {
				_, fileHeaders, err := req.FormFile("data")
				Expect(err).ToNot(HaveOccurred())
				Expect(fileHeaders.Filename).NotTo(Equal("file-not-to-send"))
			},
		))

		Expect(sender.Send(dataDirectory, dataLoader.URL(), "some-key")).To(Succeed())
	})

	It("fails if the request object cannot be created", func() {
		err := sender.Send(dataDirectory, "127.0.0.1:a", "some-key")
		Expect(err).To(MatchError(ContainSubstring(RequestCreationFailureMessage)))
	})

	It("errors when the POST cannot be completed", func() {
		err := sender.Send(dataDirectory, "http://127.0.0.1:999999", "some-key")
		Expect(err).To(MatchError(ContainSubstring(PostFailedMessage)))
	})

	It("errors when the response code is not StatusCreated", func() {
		dataLoader.AppendHandlers(
			ghttp.RespondWith(http.StatusUnauthorized, ""),
		)

		err := sender.Send(dataDirectory, dataLoader.URL(), "invalid-key")
		Expect(err).To(MatchError(fmt.Sprintf(UnexpectedResponseCodeFormat, http.StatusUnauthorized)))
	})

	It("errors when the data directory does not exist", func() {
		err := sender.Send("/does/not/exist", dataLoader.URL(), "some-key")
		Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(ReadDirectoryErrorFormat, "/does/not/exist"))))
	})

	It("errors when the data directory does not not contain any data", func() {
		dirWithNoData, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		err = sender.Send(dirWithNoData, dataLoader.URL(), "some-key")
		Expect(err).To(MatchError(fmt.Sprintf(NoDataErrorFormat, dirWithNoData)))
	})
})

func verifyFileSentInRequest(filename string, req *http.Request) {
	_, fileHeader, err := req.FormFile("data")
	Expect(err).NotTo(HaveOccurred())
	Expect(fileHeader.Filename).To(Equal(filename))
}
