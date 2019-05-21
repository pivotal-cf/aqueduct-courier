package cf_test

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pivotal-cf/aqueduct-courier/consumption/consumptionfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pkg/errors"

	. "github.com/pivotal-cf/aqueduct-courier/cf"
)

var _ = Describe("Client", func() {
	var (
		client       *Client
		cfApiService *ghttp.Server
		err          error
		uaaURL       string
	)

	BeforeEach(func() {
		cfApiService = ghttp.NewServer()
		v2InfoResponse := fmt.Sprintf(`{"token_endpoint":"http://api.funstuff.com/uaa"}`)

		cfApiService.RouteToHandler(http.MethodGet, "/v2/info", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(v2InfoResponse))
		})
		client = NewClient(cfApiService.URL(), http.DefaultClient)
	})

	AfterEach(func() {
		cfApiService.Close()
	})

	Describe("GetUAAURL", func() {
		It("makes a request to UAA and retrieves the UAA url", func() {
			uaaURL, err = client.GetUAAURL()
			Expect(err).NotTo(HaveOccurred())
			Expect(uaaURL).To(Equal("http://api.funstuff.com/uaa"))
			Expect(len(cfApiService.ReceivedRequests())).To(Equal(1))
		})

		It("returns an error when the CF API URL is invalid", func() {
			client = NewClient(" bad://url", nil)
			uaaURL, err = client.GetUAAURL()
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(CfApiURLParsingError, " bad://url"))))
			Expect(err).To(MatchError(ContainSubstring("first path segment in URL cannot contain colon")))
		})

		It("returns an error when the request to the CF API endpoint fails", func() {
			cfApiService.RouteToHandler(http.MethodGet, "/v2/info", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusMovedPermanently)
			})
			uaaURL, err = client.GetUAAURL()

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(CfApiRequestError, cfApiService.URL()+"/v2/info"))))
			Expect(err).To(MatchError(ContainSubstring("301 response missing Location header")))
		})

		It("returns an error when reading the response fails", func() {
			fakeHTTPClient := consumptionfakes.FakeHttpClient{}
			fakeHTTPClient.DoReturns(&http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(&badReader{})}, nil)
			client = NewClient("http://some-cf-api-url", &fakeHTTPClient)

			uaaURL, err = client.GetUAAURL()
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(CFApiReadResponseError, "http://some-cf-api-url/v2/info"))))
			Expect(err).To(MatchError(ContainSubstring("Reading is hard")))
		})

		It("returns an error when unmarshaling the response fails", func() {
			v2InfoResponse := fmt.Sprintf(`{"messed-up,"}`)
			cfApiService.RouteToHandler(http.MethodGet, "/v2/info", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(v2InfoResponse))
			})

			uaaURL, err = client.GetUAAURL()
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(CFApiUnmarshalError, cfApiService.URL()+"/v2/info"))))
			Expect(err).To(MatchError(ContainSubstring("invalid character '}' after object key")))
		})

		It("returns an error when the response is not 200", func() {
			cfApiService.RouteToHandler(http.MethodGet, "/v2/info", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})

			uaaURL, err = client.GetUAAURL()
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(CFApiUnexpectedResponseStatusErrorFormat, 500))))
		})

		It("returns an error if the UAA endpoint is empty", func() {
			v2InfoResponse := fmt.Sprintf(`{"token_endpoint":"", "other_non_string_prop": true}`)
			cfApiService.RouteToHandler(http.MethodGet, "/v2/info", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(v2InfoResponse))
			})

			uaaURL, err = client.GetUAAURL()
			Expect(err).To(MatchError(ContainSubstring(UAAEndpointEmptyError)))
		})

	})
})

type badReader struct{}

func (r *badReader) Read(b []byte) (n int, err error) {
	return 0, errors.New("Reading is hard")
}
