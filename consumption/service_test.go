package consumption_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/aqueduct-courier/cf"
	. "github.com/pivotal-cf/aqueduct-courier/consumption"
)

var _ = Describe("Service", func() {
	var (
		service      *Service
		usageService *ghttp.Server
		uaaService   *ghttp.Server
		authClient   cf.OAuthClient
	)

	BeforeEach(func() {
		usageService = ghttp.NewServer()
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

		authClient = cf.NewOAuthClient(
			uaaService.URL(),
			"best-usage-service-client-id",
			"best-usage-service-client-secret",
			5*time.Second,
			http.DefaultClient,
		)

		usageURL, err := url.Parse(usageService.URL())
		Expect(err).To(Not(HaveOccurred()))
		service = &Service{
			BaseURL: usageURL,
			Client:  authClient,
		}
	})

	AfterEach(func() {
		uaaService.Close()
		usageService.Close()
	})

	Describe("App Usages", func() {
		BeforeEach(func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/app_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`successful app usage content`))
			})
		})

		It("returns app usage content", func() {
			expectedBody := []byte(`successful app usage content`)

			respBody, err := service.AppUsages()
			Expect(err).NotTo(HaveOccurred())
			actualBytes, err := ioutil.ReadAll(respBody)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualBytes).To(Equal(expectedBody))
		})

		It("returns an error when the request to the usage service fails", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/app_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusMovedPermanently)
			})
			_, err := service.AppUsages()
			Expect(err).To(MatchError(ContainSubstring("301 response missing Location header")))
			Expect(err).To(MatchError(ContainSubstring(UsageServiceRequestError)))
		})

		It("returns an error when the usage service returns an unexpected response", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/app_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusInternalServerError)
			})
			_, err := service.AppUsages()
			Expect(err).To(MatchError(ContainSubstring(AppUsagesRequestError)))
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UsageServiceUnexpectedResponseStatusErrorFormat, http.StatusInternalServerError, AppUsagesReportName))))
		})
	})

	Describe("Service Usage", func() {
		BeforeEach(func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/service_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`successful service usage content`))
			})
		})

		It("returns service usage content", func() {
			expectedBody := []byte(`successful service usage content`)

			respBody, err := service.ServiceUsages()
			Expect(err).NotTo(HaveOccurred())
			actualBytes, err := ioutil.ReadAll(respBody)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualBytes).To(Equal(expectedBody))
		})

		It("returns an error when the request to the usage service fails", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/service_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusMovedPermanently)
			})
			_, err := service.ServiceUsages()
			Expect(err).To(MatchError(ContainSubstring("301 response missing Location header")))
			Expect(err).To(MatchError(ContainSubstring(UsageServiceRequestError)))
		})

		It("returns an error when the usage service returns an unexpected response", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/service_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusInternalServerError)
			})
			_, err := service.ServiceUsages()
			Expect(err).To(MatchError(ContainSubstring(ServiceUsagesRequestError)))
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UsageServiceUnexpectedResponseStatusErrorFormat, http.StatusInternalServerError, ServiceUsagesReportName))))
		})
	})

	Describe("Task Usage", func() {
		BeforeEach(func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/task_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`successful task usage content`))
			})
		})

		It("returns task usage content", func() {
			expectedBody := []byte(`successful task usage content`)

			respBody, err := service.TaskUsages()
			Expect(err).NotTo(HaveOccurred())

			actualBytes, err := ioutil.ReadAll(respBody)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualBytes).To(Equal(expectedBody))
		})

		It("returns an error when the request to the usage service fails", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/task_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusMovedPermanently)
			})
			_, err := service.TaskUsages()
			Expect(err).To(MatchError(ContainSubstring("301 response missing Location header")))
			Expect(err).To(MatchError(ContainSubstring(UsageServiceRequestError)))
		})

		It("returns an error when the usage service returns an unexpected response", func() {
			usageService.RouteToHandler(http.MethodGet, "/system_report/task_usages", func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
				w.WriteHeader(http.StatusInternalServerError)
			})
			_, err := service.TaskUsages()
			Expect(err).To(MatchError(ContainSubstring(TaskUsagesRequestError)))
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UsageServiceUnexpectedResponseStatusErrorFormat, http.StatusInternalServerError, TaskUsagesReportName))))
		})
	})
})
