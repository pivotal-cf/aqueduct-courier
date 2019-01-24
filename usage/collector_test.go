package usage_test

import (
	"encoding/base64"

	"github.com/pkg/errors"

	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	. "github.com/pivotal-cf/aqueduct-courier/usage"
	"github.com/pivotal-cf/aqueduct-courier/usage/usagefakes"
)

var _ = Describe("Collector", func() {
	var (
		collector    *Collector
		usageService *ghttp.Server
		uaaService   *ghttp.Server
		cfApiClient  *usagefakes.FakeCfApiClient
	)

	BeforeEach(func() {
		uaaService = ghttp.NewServer()
		uaaService.RouteToHandler(http.MethodPost, "/oauth/token", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			credentialBytes := []byte("best-usage-service-client-id:best-usage-service-client-secret")

			base64credentials := base64.StdEncoding.EncodeToString(credentialBytes)
			Expect(req.Header.Get("authorization")).To(Equal("Basic " + base64credentials))

			w.Write([]byte(`{
					"access_token": "some-uaa-token",
					"token_type": "bearer",
					"expires_in": 3600
					}`))
		})

		usageService = ghttp.NewServer()
		usageService.RouteToHandler(http.MethodGet, "/system_report/app_usages", func(w http.ResponseWriter, req *http.Request) {
			Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
			w.WriteHeader(http.StatusOK)
		})
		cfApiClient = &usagefakes.FakeCfApiClient{}
		cfApiClient.GetUAAURLReturns(uaaService.URL(), nil)

		collector = NewCollector(cfApiClient, http.DefaultClient, usageService.URL(), "best-usage-service-client-id", "best-usage-service-client-secret")
	})

	AfterEach(func() {
		uaaService.Close()
		usageService.Close()
	})

	Describe("collect", func() {
		It("accesses the usage service with an OAuth client configured appropriately, with the endpoint discovered from the CfApiClient", func() {
			err := collector.Collect()

			Expect(err).ToNot(HaveOccurred())
			Expect(len(usageService.ReceivedRequests())).To(Equal(1))
		})

		It("returns an error if the usage service URL is invalid", func() {
			collector = NewCollector(cfApiClient, http.DefaultClient, " bad://url", "best-usage-service-client-id", "best-usage-service-client-secret")
			err := collector.Collect()

			Expect(err).To(MatchError(ContainSubstring(UsageServiceURLParsingError)))
			Expect(err).To(MatchError(ContainSubstring("first path segment in URL cannot contain colon")))
		})

		It("returns an error if fetching the UAA token fails", func() {
			cfApiClient.GetUAAURLReturns("", errors.New("getting UAA URL is hard"))
			err := collector.Collect()

			Expect(err).To(MatchError(ContainSubstring(GetUAAURLError)))
			Expect(err).To(MatchError(ContainSubstring("getting UAA URL is hard")))
		})

		It("returns an error when the request to the usage service endpoint fails", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/app_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusMovedPermanently)
			})
			err := collector.Collect()

			Expect(err).To(MatchError(ContainSubstring("301 response missing Location header")))
			Expect(err).To(MatchError(ContainSubstring(UsageServiceRequestError)))
		})

		It("returns an error when the request receives an unsuccessful response", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/app_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusInternalServerError)
			})
			err := collector.Collect()

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UsageServiceUnexpectedResponseStatusErrorFormat, 500))))
		})

	})
})
