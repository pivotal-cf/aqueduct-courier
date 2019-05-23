package consumption_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"

	"github.com/pivotal-cf/aqueduct-courier/consumption/consumptionfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf/aqueduct-courier/consumption"
)

var _ = Describe("Service", func() {
	var (
		service    *Service
		fakeClient *consumptionfakes.FakeHttpClient
		usageURL   *url.URL
	)

	BeforeEach(func() {
		var err error
		fakeClient = new(consumptionfakes.FakeHttpClient)

		usageHost := "http://usage.example.com/supposedprefix/"
		usageURL, err = url.Parse(usageHost)
		Expect(err).To(Not(HaveOccurred()))

		service = &Service{
			BaseURL: usageURL,
			Client:  fakeClient,
		}
	})

	Describe("App Usages", func() {
		It("returns app usage content", func() {
			body := &readerCloser{reader: bytes.NewReader([]byte(`successful app usage content`))}
			appUsagesResponse := &http.Response{Body: body, StatusCode: http.StatusOK}
			fakeClient.DoReturns(appUsagesResponse, nil)

			expectedBody := []byte(`successful app usage content`)
			respBody, err := service.AppUsages()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeClient.DoCallCount()).To(Equal(1))
			req := fakeClient.DoArgsForCall(0)

			usageURL.Path = path.Join(usageURL.Path, SystemReportPathPrefix, AppUsagesReportName)
			Expect(req.URL).To(Equal(usageURL))

			Expect(body.isClosed).To(BeTrue())
			content, err := ioutil.ReadAll(respBody)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(Equal([]byte(expectedBody)))
		})

		It("errors when the request to the usage service fails", func() {
			fakeClient.DoReturns(nil, errors.New("requesting things is hard"))
			_, err := service.AppUsages()

			Expect(err).To(MatchError(ContainSubstring("requesting things is hard")))
			Expect(err).To(MatchError(ContainSubstring(UsageServiceRequestError)))
		})

		It("errors when the usage service returns an unexpected response", func() {
			body := &readerCloser{}
			badStatusResponse := &http.Response{Body: body, StatusCode: http.StatusInternalServerError}
			fakeClient.DoReturns(badStatusResponse, nil)
			_, err := service.AppUsages()

			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(ContainSubstring(AppUsagesRequestError)))
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UsageServiceUnexpectedResponseStatusErrorFormat, http.StatusInternalServerError, AppUsagesReportName))))
		})
	})

	Describe("Service Usages", func() {
		It("removes service plan names from monthly and yearly service usage contents and returns the rest", func() {
			reportsJson := []byte(`{
  "report_time": "2017-05-11",
  "monthly_service_reports": [
    {
      "service_name": "cool-monthly-service-name",
      "service_guid": "cool-monthly-service-guid",
      "usages": [
        {
          "month": 1,
          "year": 2019,
          "duration_in_hours": 20,
          "average_instances": 40,
          "maximum_instances": 65
        }
      ],
      "plans": [
        {
          "usages": [
            {
              "month": 5,
              "year": 2019,
              "duration_in_hours": 385.61,
              "average_instances": 1.5,
              "maximum_instances": 3
            }
          ],
          "service_plan_name": "cool-monthly-service-plan-name",
          "service_plan_guid": "cool-monthly-service-plan-guid"
        }
      ]
    }
  ],
  "yearly_service_report": [
    {
      "service_name": "cool-yearly-service-name",
      "service_guid": "cool-yearly-service-guid",
      "year": 2019,
      "duration_in_hours": 699,
      "maximum_instances": 5,
      "average_instances": 3.6,
      "plans": [{
        "service_plan_name": "cool-yearly-service-plan-name",
        "service_plan_guid": "cool-yearly-service-plan-guid",
        "year": 2019,
        "duration_in_hours": 69,
        "maximum_instances": 5,
        "average_instances": 3.6
      }]
    }
  ]
}`)
			body := &readerCloser{reader: bytes.NewReader(reportsJson)}
			serviceUsagesResponse := &http.Response{Body: body, StatusCode: http.StatusOK}
			fakeClient.DoReturns(serviceUsagesResponse, nil)

			respBody, err := service.ServiceUsages()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeClient.DoCallCount()).To(Equal(1))
			req := fakeClient.DoArgsForCall(0)

			usageURL.Path = path.Join(usageURL.Path, SystemReportPathPrefix, ServiceUsagesReportName)
			Expect(req.URL).To(Equal(usageURL))

			Expect(body.isClosed).To(BeTrue())
			content, err := ioutil.ReadAll(respBody)
			Expect(err).NotTo(HaveOccurred())

			var actualResults map[string]interface{}
			Expect(json.Unmarshal(content, &actualResults)).To(Succeed())

			expectedResultJson := []byte(`{
  "report_time": "2017-05-11",
  "monthly_service_reports": [
    {
      "service_name": "cool-monthly-service-name",
      "service_guid": "cool-monthly-service-guid",
      "usages": [
        {
          "month": 1,
          "year": 2019,
          "duration_in_hours": 20,
          "average_instances": 40,
          "maximum_instances": 65
        }
      ],
      "plans": [
        {
          "usages": [
            {
              "month": 5,
              "year": 2019,
              "duration_in_hours": 385.61,
              "average_instances": 1.5,
              "maximum_instances": 3
            }
          ],
          "service_plan_guid": "cool-monthly-service-plan-guid"
        }
      ]
    }
  ],
  "yearly_service_report": [
    {
      "service_name": "cool-yearly-service-name",
      "service_guid": "cool-yearly-service-guid",
      "year": 2019,
      "duration_in_hours": 699,
      "maximum_instances": 5,
      "average_instances": 3.6,
      "plans": [{
        "service_plan_guid": "cool-yearly-service-plan-guid",
        "year": 2019,
        "duration_in_hours": 69,
        "maximum_instances": 5,
        "average_instances": 3.6
      }]
    }
  ]
}`)
			var expectedResults map[string]interface{}
			Expect(json.Unmarshal(expectedResultJson, &expectedResults)).To(Succeed())

			Expect(actualResults).To(Equal(expectedResults))
		})

		It("errors when the request to the usage service fails", func() {
			fakeClient.DoReturns(nil, errors.New("requesting things is hard"))
			_, err := service.ServiceUsages()

			Expect(err).To(MatchError(ContainSubstring("requesting things is hard")))
			Expect(err).To(MatchError(ContainSubstring(UsageServiceRequestError)))
		})

		It("errors when the usage service returns an unexpected response", func() {
			body := &readerCloser{}
			badStatusResponse := &http.Response{Body: body, StatusCode: http.StatusInternalServerError}
			fakeClient.DoReturns(badStatusResponse, nil)
			_, err := service.ServiceUsages()

			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(ContainSubstring(ServiceUsagesRequestError)))
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UsageServiceUnexpectedResponseStatusErrorFormat, http.StatusInternalServerError, ServiceUsagesReportName))))
		})

		It("errors if the contents cannot be read from the response", func() {
			body := &readerCloser{reader: &badReader{}}
			badReaderResponse := &http.Response{Body: body, StatusCode: http.StatusOK}
			fakeClient.DoReturns(badReaderResponse, nil)

			_, err := service.ServiceUsages()
			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(ContainSubstring(ReadResponseError)))
			Expect(err).To(MatchError(ContainSubstring("bad-reader")))
		})

		It("errors if the contents are not json", func() {
			body := &readerCloser{reader: bytes.NewReader([]byte(`not-good-json`))}
			badJSONResponse := &http.Response{Body: body, StatusCode: http.StatusOK}
			fakeClient.DoReturns(badJSONResponse, nil)

			_, err := service.ServiceUsages()
			Expect(err).To(MatchError(ContainSubstring(UnmarshalResponseError)))
		})
	})

	Describe("Task Usages", func() {
		It("returns task usage content", func() {
			body := &readerCloser{reader: bytes.NewReader([]byte(`successful task usage content`))}
			appUsagesResponse := &http.Response{Body: body, StatusCode: http.StatusOK}
			fakeClient.DoReturns(appUsagesResponse, nil)

			expectedBody := []byte(`successful task usage content`)

			respBody, err := service.TaskUsages()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeClient.DoCallCount()).To(Equal(1))
			req := fakeClient.DoArgsForCall(0)

			usageURL.Path = path.Join(usageURL.Path, SystemReportPathPrefix, TaskUsagesReportName)
			Expect(req.URL).To(Equal(usageURL))

			Expect(body.isClosed).To(BeTrue())
			content, err := ioutil.ReadAll(respBody)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(Equal(expectedBody))
		})

		It("errors when the request to the usage service fails", func() {
			fakeClient.DoReturns(nil, errors.New("requesting things is hard"))
			_, err := service.TaskUsages()

			Expect(err).To(MatchError(ContainSubstring("requesting things is hard")))
			Expect(err).To(MatchError(ContainSubstring(UsageServiceRequestError)))
		})

		It("errors when the usage service returns an unexpected response", func() {
			body := &readerCloser{}
			badStatusResponse := &http.Response{Body: body, StatusCode: http.StatusInternalServerError}
			fakeClient.DoReturns(badStatusResponse, nil)
			_, err := service.TaskUsages()

			Expect(body.isClosed).To(BeTrue())
			Expect(err).To(MatchError(ContainSubstring(TaskUsagesRequestError)))
			Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf(UsageServiceUnexpectedResponseStatusErrorFormat, http.StatusInternalServerError, TaskUsagesReportName))))
		})
	})
})

type badReader struct{}

func (b *badReader) Read([]byte) (int, error) {
	return 0, errors.New("bad-reader")
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
