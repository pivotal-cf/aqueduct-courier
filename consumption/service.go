package consumption

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
)

const (
	AppUsagesReportName     = "app_usages"
	ServiceUsagesReportName = "service_usages"
	TaskUsagesReportName    = "task_usages"

	CreateUsageServiceHTTPRequestError              = "error creating HTTP request to usage service endpoint"
	UsageServiceRequestError                        = "error accessing usage service"
	UsageServiceUnexpectedResponseStatusErrorFormat = "unexpected status %d when accessing usage service: %s"

	AppUsagesRequestError     = "error retrieving app usages data"
	ServiceUsagesRequestError = "error retrieving service usages data"
	TaskUsagesRequestError    = "error retrieving task usages data"
	UnmarshalResponseError    = "error unmarshalling response"
	ReadResponseError         = "error reading response"
)

//go:generate counterfeiter . httpClient
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Service struct {
	BaseURL *url.URL
	Client  httpClient
}

type usage struct {
	Month            int     `json:"month"`
	Year             int     `json:"year"`
	DurationInHours  float64 `json:"duration_in_hours"`
	AverageInstances float64 `json:"average_instances"`
	MaximumInstances float64 `json:"maximum_instances"`
}

type serviceReport struct {
	ReportTime            string `json:"report_time"`
	MonthlyServiceReports []struct {
		ServiceName string  `json:"service_name"`
		ServiceGUID string  `json:"service_guid"`
		Usages      []usage `json:"usages"`
		Plans       []struct {
			Usages          []usage `json:"usages"`
			ServicePlanGUID string  `json:"service_plan_guid"`
		} `json:"plans"`
	} `json:"monthly_service_reports"`
	YearlyServiceReport []struct {
		ServiceName      string  `json:"service_name"`
		ServiceGUID      string  `json:"service_guid"`
		Year             int     `json:"year"`
		DurationInHours  float64 `json:"duration_in_hours"`
		MaximumInstances float64 `json:"maximum_instances"`
		AverageInstances float64 `json:"average_instances"`
		Plans            []struct {
			Year             int     `json:"year"`
			ServicePlanGUID  string  `json:"service_plan_guid"`
			DurationInHours  float64 `json:"duration_in_hours"`
			MaximumInstances float64 `json:"maximum_instances"`
			AverageInstances float64 `json:"average_instances"`
		} `json:"plans"`
	} `json:"yearly_service_report"`
}

func (s *Service) AppUsages() (io.Reader, error) {
	contents, err := s.makeRequest(AppUsagesReportName)
	if err != nil {
		return nil, errors.Wrap(err, AppUsagesRequestError)
	}
	return bytes.NewReader(contents), nil
}

func (s *Service) ServiceUsages() (io.Reader, error) {
	contents, err := s.makeRequest(ServiceUsagesReportName)
	if err != nil {
		return nil, errors.Wrap(err, ServiceUsagesRequestError)
	}

	var sReport serviceReport
	if err := json.Unmarshal(contents, &sReport); err != nil {
		return nil, errors.Wrapf(err, UnmarshalResponseError)
	}

	redactedContent, err := json.Marshal(sReport)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(redactedContent), nil
}

func (s *Service) TaskUsages() (io.Reader, error) {
	respBody, err := s.makeRequest(TaskUsagesReportName)
	if err != nil {
		return nil, errors.Wrap(err, TaskUsagesRequestError)
	}
	return bytes.NewReader(respBody), nil
}

func (s *Service) makeRequest(reportName string) ([]byte, error) {
	targetURL, _ := url.Parse(s.BaseURL.String())
	targetURL.Path = path.Join(targetURL.Path, SystemReportPathPrefix, reportName)
	req, err := http.NewRequest(http.MethodGet, targetURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, CreateUsageServiceHTTPRequestError)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, UsageServiceRequestError)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf(UsageServiceUnexpectedResponseStatusErrorFormat, resp.StatusCode, reportName)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, ReadResponseError)
	}

	return contents, nil
}
