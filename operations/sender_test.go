package operations_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/aqueduct-courier/operations"
	"github.com/pivotal-cf/aqueduct-courier/operations/operationsfakes"
	"github.com/pkg/errors"
)

var _ = Describe("Sender", func() {
	var (
		client         *operationsfakes.FakeHttpClient
		tmpFile        *os.File
		tarContent     string
		doBodyContents []byte
		sender         SendExecutor
		err            error
	)

	BeforeEach(func() {
		sender = SendExecutor{}
		client = new(operationsfakes.FakeHttpClient)

		tmpFile, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

		tarContent = "tar-content"
		_, err = tmpFile.Write([]byte(tarContent))
		Expect(err).NotTo(HaveOccurred())
		Expect(tmpFile.Close()).To(Succeed())

		emptyBody := ioutil.NopCloser(strings.NewReader(""))

		client.DoStub = func(request *http.Request) (response *http.Response, e error) {
			var err error
			doBodyContents, err = ioutil.ReadAll(request.Body)
			Expect(err).NotTo(HaveOccurred())

			return &http.Response{StatusCode: http.StatusCreated, Body: emptyBody}, nil
		}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpFile.Name())).To(Succeed())
	})

	It("posts to the data loader with the file as content", func() {
		senderVersion := "best-sender-version"
		Expect(sender.Send(client, tmpFile.Name(), "http://example.com", "some-key", senderVersion)).To(Succeed(), "")

		Expect(client.DoCallCount()).To(Equal(1))
		req := client.DoArgsForCall(0)
		Expect(req.URL.String()).To(Equal(fmt.Sprintf("http://example.com%s", PostPath)))

		Expect(err).ToNot(HaveOccurred())
		Expect(string(doBodyContents)).To(Equal(tarContent))
	})

	It("posts to the data loader with the correct API key in the header", func() {
		Expect(sender.Send(client, tmpFile.Name(), "http://example.com", "some-key", "")).To(Succeed())
		req := client.DoArgsForCall(0)
		Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-key"))
	})

	It("fails if the request object cannot be created", func() {
		err := sender.Send(client, tmpFile.Name(), "127.0.0.1:a", "some-key", "")
		Expect(err).To(MatchError(ContainSubstring(RequestCreationFailureMessage)))
	})

	It("errors when the POST cannot be completed", func() {
		client.DoReturns(nil, errors.New("doing requests is hard"))
		err := sender.Send(client, tmpFile.Name(), "http://example.com", "some-key", "")
		Expect(err).To(MatchError(ContainSubstring("doing requests is hard")))
		Expect(err).To(MatchError(ContainSubstring(PostFailedMessage)))
	})

	It("errors when the response code is not StatusCreated", func() {
		emptyBody := ioutil.NopCloser(strings.NewReader(""))
		client.DoReturns(&http.Response{StatusCode: http.StatusUnauthorized, Body: emptyBody}, nil)

		err := sender.Send(client, tmpFile.Name(), "http://example.com", "invalid-key", "")
		Expect(err).To(MatchError(UnauthorizedErrorMessage))
	})

	It("errors if the error response cannot be read", func() {
		client.DoReturns(&http.Response{StatusCode: http.StatusExpectationFailed, Body: ioutil.NopCloser(&badReader{})}, nil)
		err := sender.Send(client, tmpFile.Name(), "http://example.com", "invalid-key", "")
		Expect(err).To(MatchError(fmt.Sprintf(UnexpectedServerErrorFormat, "unknown")))
	})

	It("errors if the error response cannot be read into the expected structure", func() {
		badBody := ioutil.NopCloser(strings.NewReader(`{not json`))
		client.DoReturns(&http.Response{StatusCode: http.StatusExpectationFailed, Body: badBody}, nil)
		err := sender.Send(client, tmpFile.Name(), "http://example.com", "invalid-key", "")
		Expect(err).To(MatchError(fmt.Sprintf(UnexpectedServerErrorFormat, "unknown")))
	})

	It("errors when the response code is not 201/401", func() {
		emptyBody := ioutil.NopCloser(strings.NewReader(`{"error": {"uuid": "error-uuid"}}`))
		client.DoReturns(&http.Response{StatusCode: http.StatusExpectationFailed, Body: emptyBody}, nil)

		err := sender.Send(client, tmpFile.Name(), "http://example.com", "invalid-key", "")
		Expect(err).To(MatchError(fmt.Sprintf(UnexpectedServerErrorFormat, "error-uuid")))
	})

	It("when the tarFile does not exist", func() {
		err := sender.Send(client, "path/to/not/the/tarFile", "http://example.com", "some-key", "")
		Expect(err).To(MatchError(ContainSubstring(ReadDataFileError)))
	})
})

type badReader struct{}

func (r *badReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("reading is hard")
}
