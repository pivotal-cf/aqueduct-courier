package cf_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

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
			Expect(err).To(MatchError(ContainSubstring(CfApiURLParsingError)))
			Expect(err).To(MatchError(ContainSubstring("first path segment in URL cannot contain colon")))
		})

		It("returns an error when the request to the CF API endpoint fails", func() {
			cfApiService.RouteToHandler(http.MethodGet, "/v2/info", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusMovedPermanently)
			})
			uaaURL, err = client.GetUAAURL()

			Expect(err).To(MatchError(ContainSubstring(CfApiRequestError)))
			Expect(err).To(MatchError(ContainSubstring("301 response missing Location header")))
		})

		It("returns an error when unmarshaling the response fails", func() {
			v2InfoResponse := fmt.Sprintf(`{"messed-up,"}`)
			cfApiService.RouteToHandler(http.MethodGet, "/v2/info", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(v2InfoResponse))
			})

			uaaURL, err = client.GetUAAURL()
			Expect(err).To(MatchError(ContainSubstring(CFApiUnmarshalError)))
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
