package consumption_test

import (
	"encoding/base64"
	"io/ioutil"
	"log"

	"github.com/pivotal-cf/aqueduct-utils/data"

	"github.com/pkg/errors"

	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	. "github.com/pivotal-cf/aqueduct-courier/consumption"
	"github.com/pivotal-cf/aqueduct-courier/consumption/consumptionfakes"
)

var _ = Describe("Collector", func() {
	var (
		logger *log.Logger
		collector    *Collector
		usageService *ghttp.Server
		uaaService   *ghttp.Server
		cfApiClient  *consumptionfakes.FakeCfApiClient
	)

	BeforeEach(func() {
		logger = log.New(GinkgoWriter, "", 0)
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
			w.Write([]byte(`successful app usage content`))
		})
		usageService.RouteToHandler(http.MethodGet, "/system_report/service_usages", func(w http.ResponseWriter, req *http.Request) {
			Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`successful service usage content`))
		})
		usageService.RouteToHandler(http.MethodGet, "/system_report/task_usages", func(w http.ResponseWriter, req *http.Request) {
			Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`successful task usage content`))
		})
		cfApiClient = &consumptionfakes.FakeCfApiClient{}
		cfApiClient.GetUAAURLReturns(uaaService.URL(), nil)

		collector = NewCollector(*logger, cfApiClient, http.DefaultClient, usageService.URL(), "best-usage-service-client-id", "best-usage-service-client-secret")
	})

	AfterEach(func() {
		uaaService.Close()
		usageService.Close()
	})

	Describe("collect", func() {
		It("accesses the usage service with an OAuth client configured appropriately, with the endpoint discovered from the CfApiClient", func() {
			usageData, err := collector.Collect()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(usageData)).To(Equal(3))

			appUsageContent, err := ioutil.ReadAll(usageData[0].Content())
			Expect(err).ToNot(HaveOccurred())
			serviceUsageContent, err := ioutil.ReadAll(usageData[1].Content())
			Expect(err).ToNot(HaveOccurred())
			taskUsageContent, err := ioutil.ReadAll(usageData[2].Content())
			Expect(err).ToNot(HaveOccurred())

			Expect(usageData[0].DataType()).To(Equal(data.AppUsageDataType))
			Expect(usageData[1].DataType()).To(Equal(data.ServiceUsageDataType))
			Expect(usageData[2].DataType()).To(Equal(data.TaskUsageDataType))

			Expect(len(usageService.ReceivedRequests())).To(Equal(3))
			Expect(appUsageContent).To(Equal([]byte("successful app usage content")))
			Expect(serviceUsageContent).To(Equal([]byte("successful service usage content")))
			Expect(taskUsageContent).To(Equal([]byte("successful task usage content")))
		})

		It("returns an error if the usage service URL is invalid", func() {
			collector = NewCollector(*logger, cfApiClient, http.DefaultClient, " bad://url", "best-usage-service-client-id", "best-usage-service-client-secret")
			_, err := collector.Collect()

			Expect(err).To(MatchError(ContainSubstring(UsageServiceURLParsingError)))
			Expect(err).To(MatchError(ContainSubstring("first path segment in URL cannot contain colon")))
		})

		It("returns an error if fetching the UAA token fails", func() {
			cfApiClient.GetUAAURLReturns("", errors.New("getting UAA URL is hard"))
			_, err := collector.Collect()

			Expect(err).To(MatchError(ContainSubstring(GetUAAURLError)))
			Expect(err).To(MatchError(ContainSubstring("getting UAA URL is hard")))
		})

		It("returns an error when the request to the app usage service endpoint fails", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/app_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusMovedPermanently)
			})
			_, err := collector.Collect()

			Expect(err).To(MatchError(ContainSubstring("301 response missing Location header")))
			Expect(err).To(MatchError(ContainSubstring(UsageServiceRequestError)))
		})

		It("returns an error when the request to the app usage endpoint receives an unsuccessful response", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/app_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusInternalServerError)
			})
			_, err := collector.Collect()

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UsageServiceUnexpectedResponseStatusErrorFormat, 500, AppUsagesReportName))))
		})

		It("returns an error when the request to the service usage service endpoint fails", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/service_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusMovedPermanently)
			})
			_, err := collector.Collect()

			Expect(err).To(MatchError(ContainSubstring("301 response missing Location header")))
			Expect(err).To(MatchError(ContainSubstring(UsageServiceRequestError)))
		})

		It("returns an error when the request to the service usage endpoint receives an unsuccessful response", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/service_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusInternalServerError)
			})
			_, err := collector.Collect()

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UsageServiceUnexpectedResponseStatusErrorFormat, 500, ServiceUsagesReportName))))
		})

		It("returns an error when the request to the task usage task endpoint fails", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/task_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusMovedPermanently)
			})
			_, err := collector.Collect()

			Expect(err).To(MatchError(ContainSubstring("301 response missing Location header")))
			Expect(err).To(MatchError(ContainSubstring(UsageServiceRequestError)))
		})

		It("returns an error when the request to the task usage endpoint receives an unsuccessful response", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/task_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusInternalServerError)
			})
			_, err := collector.Collect()

			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UsageServiceUnexpectedResponseStatusErrorFormat, 500, TaskUsagesReportName))))
		})
	})
})
