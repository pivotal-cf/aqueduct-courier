package cf_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/pivotal-cf/aqueduct-courier/cf/cffakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	. "github.com/pivotal-cf/aqueduct-courier/cf"
)

var _ = Describe("Client", func() {
	const (
		cfURL = "https://example.com/whatever"
	)
	var (
		client         *Client
		responseReader *readerCloser
		fakeHTTPClient *cffakes.FakeHttpClient
	)

	BeforeEach(func() {
		v2InfoResponse := `{"token_endpoint":"http://api.funstuff.com/uaa"}`
		responseReader = &readerCloser{reader: bytes.NewReader([]byte(v2InfoResponse))}

		fakeHTTPClient = &cffakes.FakeHttpClient{}
		fakeHTTPClient.DoReturns(&http.Response{Body: responseReader, StatusCode: http.StatusOK}, nil)
		client = NewClient(cfURL, fakeHTTPClient)
	})

	Describe("GetUAAURL", func() {
		It("makes a request to UAA and retrieves the UAA url", func() {
			uaaURL, err := client.GetUAAURL()
			Expect(err).NotTo(HaveOccurred())
			Expect(uaaURL).To(Equal("http://api.funstuff.com/uaa"))
			Expect(responseReader.isClosed).To(BeTrue())
		})

		It("returns an error when the CF API URL is invalid", func() {
			client = NewClient(" bad://url", nil)
			_, err := client.GetUAAURL()
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(CfApiURLParsingError, " bad://url"))))
			Expect(err).To(MatchError(ContainSubstring("first path segment in URL cannot contain colon")))
		})

		It("returns an error when the request to the CF API endpoint fails", func() {
			fakeHTTPClient.DoReturns(nil, errors.New("Requesting stuff is hard"))

			_, err := client.GetUAAURL()

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(CfApiRequestError, cfURL+"/v2/info"))))
			Expect(err).To(MatchError(ContainSubstring("Requesting stuff is hard")))
		})

		It("returns an error when reading the response fails", func() {
			responseReader := &readerCloser{reader: &badReader{}}
			fakeHTTPClient.DoReturns(&http.Response{StatusCode: http.StatusOK, Body: responseReader}, nil)

			_, err := client.GetUAAURL()
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(CFApiReadResponseError, cfURL+"/v2/info"))))
			Expect(err).To(MatchError(ContainSubstring("Reading is hard")))
			Expect(responseReader.isClosed).To(BeTrue())
		})

		It("returns an error when unmarshaling the response fails", func() {
			responseReader = &readerCloser{reader: bytes.NewReader([]byte(`{"messed-up,"}`))}
			fakeHTTPClient.DoReturns(&http.Response{Body: responseReader, StatusCode: http.StatusOK}, nil)

			_, err := client.GetUAAURL()
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(CFApiUnmarshalError, cfURL+"/v2/info"))))
			Expect(err).To(MatchError(ContainSubstring("invalid character '}' after object key")))
			Expect(responseReader.isClosed).To(BeTrue())
		})

		It("returns an error when the response is not 200", func() {
			fakeHTTPClient.DoReturns(&http.Response{Body: responseReader, StatusCode: http.StatusInternalServerError}, nil)
			_, err := client.GetUAAURL()
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(CFApiUnexpectedResponseStatusErrorFormat, 500))))
			Expect(responseReader.isClosed).To(BeTrue())
		})

		It("returns an error if the UAA endpoint is empty", func() {
			responseReader = &readerCloser{reader: bytes.NewReader([]byte(`{"token_endpoint":"", "other_non_string_prop": true}`))}
			fakeHTTPClient.DoReturns(&http.Response{Body: responseReader, StatusCode: http.StatusOK}, nil)

			_, err := client.GetUAAURL()
			Expect(err).To(MatchError(ContainSubstring(UAAEndpointEmptyError)))
			Expect(responseReader.isClosed).To(BeTrue())
		})

	})
})

type badReader struct{}

func (r *badReader) Read(b []byte) (n int, err error) {
	return 0, errors.New("Reading is hard")
}

type readerCloser struct {
	reader   io.Reader
	isClosed bool
}

func (rc *readerCloser) Read(p []byte) (n int, err error) {
	return rc.reader.Read(p)
}

func (rc *readerCloser) Close() error {
	rc.isClosed = true
	return nil
}
