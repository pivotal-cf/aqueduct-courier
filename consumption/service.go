package consumption

import (
	"github.com/pivotal-cf/aqueduct-courier/cf"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"path"
)

const (
	AppUsagesReportName     = "app_usages"
	ServiceUsagesReportName = "service_usages"
	TaskUsagesReportName    = "task_usages"

	CreateUsageServiceHTTPRequestError              = "error creating HTTP request to usage service endpoint"
	UsageServiceRequestError                        = "error accessing usage service"
	UsageServiceUnexpectedResponseStatusErrorFormat = "unexpected status %d when accessing usage service: %s"

	AppUsagesRequestError = "error retrieving app usages data"
	ServiceUsagesRequestError = "error retrieving service usages data"
	TaskUsagesRequestError ="error retrieving task usages data"
)

type Service struct {
	BaseURL *url.URL
	Client cf.OAuthClient
}

func(s *Service) AppUsages() (io.Reader, error) {
	respBody, err := s.makeRequest(AppUsagesReportName)
	if err != nil {
		return nil, errors.Wrap(err, AppUsagesRequestError)
	}
	return respBody, nil
}

func(s *Service) ServiceUsages() (io.Reader, error) {
	respBody, err := s.makeRequest(ServiceUsagesReportName)
	if err != nil {
		return nil, errors.Wrap(err, ServiceUsagesRequestError)
	}
	return respBody, nil
}

func(s *Service) TaskUsages() (io.Reader, error) {
	respBody, err := s.makeRequest(TaskUsagesReportName)
	if err != nil {
		return nil, errors.Wrap(err, TaskUsagesRequestError)
	}
	return respBody, nil
}

func (s *Service) makeRequest(reportName string) (io.Reader, error) {
	s.BaseURL.Path = path.Join(SystemReportPathPrefix, reportName)
	req, err := http.NewRequest(http.MethodGet, s.BaseURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, CreateUsageServiceHTTPRequestError)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, UsageServiceRequestError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf(UsageServiceUnexpectedResponseStatusErrorFormat, resp.StatusCode, reportName)
	}
	return resp.Body, nil
}
