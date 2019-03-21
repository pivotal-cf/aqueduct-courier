package consumption

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/pivotal-cf/aqueduct-utils/data"

	"github.com/pivotal-cf/aqueduct-courier/cf"

	"github.com/pkg/errors"
)

const (
	UsageServiceURLParsingError                     = "error parsing Usage Service URL"
	GetUAAURLError                                  = "error getting UAA URL"
	CreateUsageServiceHTTPRequestError              = "error creating HTTP request to usage service endpoint"
	UsageServiceRequestError                        = "error accessing usage service"
	UsageServiceUnexpectedResponseStatusErrorFormat = "unexpected status %d when accessing usage service: %s"

	AppUsagesReportName     = "app_usages"
	ServiceUsagesReportName = "service_usages"
	TaskUsagesReportName    = "task_usages"
	SystemReportPathPrefix  = "system_report"
)

//go:generate counterfeiter . cfApiClient
type cfApiClient interface {
	GetUAAURL() (string, error)
}

type httpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

type DataCollector struct {
	logger log.Logger
	cfApiClient     cfApiClient
	httpClient      httpClient
	usageServiceURL string
	clientID        string
	clientSecret    string
}

func NewDataCollector(logger log.Logger, cfClient cfApiClient, httpClient httpClient, usageServiceURL, clientID, clientSecret string) *DataCollector {
	return &DataCollector{
		logger: logger,
		cfApiClient:     cfClient,
		usageServiceURL: usageServiceURL,
		httpClient:      httpClient,
		clientID:        clientID,
		clientSecret:    clientSecret,
	}
}

func (dc *DataCollector) Collect() ([]Data, error) {
	dc.logger.Printf("Collecting data from Usage Service at %s", dc.usageServiceURL)

	usageURL, err := url.Parse(dc.usageServiceURL)
	if err != nil {
		return []Data{}, errors.Wrapf(err, UsageServiceURLParsingError)
	}

	uaaURL, err := dc.cfApiClient.GetUAAURL()
	if err != nil {
		return []Data{}, errors.Wrap(err, GetUAAURLError)
	}

	authedClient := cf.NewOAuthClient(
		uaaURL,
		dc.clientID,
		dc.clientSecret,
		5*time.Second,
		dc.httpClient,
	)

	appUsageBody, err := getSystemReportBody(AppUsagesReportName, *usageURL, authedClient)
	if err != nil {
		return []Data{}, err
	}

	serviceUsageBody, err := getSystemReportBody(ServiceUsagesReportName, *usageURL, authedClient)
	if err != nil {
		return []Data{}, err
	}

	taskUsageBody, err := getSystemReportBody(TaskUsagesReportName, *usageURL, authedClient)
	if err != nil {
		return []Data{}, err
	}

	return []Data{
		NewData(appUsageBody, data.AppUsageDataType),
		NewData(serviceUsageBody, data.ServiceUsageDataType),
		NewData(taskUsageBody, data.TaskUsageDataType),
	}, nil
}

func getSystemReportBody(reportName string, baseURL url.URL, client httpClient) (io.Reader, error) {
	baseURL.Path = path.Join(SystemReportPathPrefix, reportName)

	req, err := http.NewRequest(http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, CreateUsageServiceHTTPRequestError)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, UsageServiceRequestError)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf(UsageServiceUnexpectedResponseStatusErrorFormat, resp.StatusCode, reportName)
	}
	return resp.Body, nil
}
