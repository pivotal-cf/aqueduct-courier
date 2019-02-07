package consumption

import (
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
	UsageServiceUnexpectedResponseStatusErrorFormat = "unexpected status in usage service response: %d"
)

//go:generate counterfeiter . cfApiClient
type cfApiClient interface {
	GetUAAURL() (string, error)
}

type httpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

type Collector struct {
	cfApiClient     cfApiClient
	httpClient      httpClient
	usageServiceURL string
	clientID        string
	clientSecret    string
}

func NewCollector(cfClient cfApiClient, httpClient httpClient, usageServiceURL, clientID, clientSecret string) *Collector {
	return &Collector{
		cfApiClient:     cfClient,
		usageServiceURL: usageServiceURL,
		httpClient:      httpClient,
		clientID:        clientID,
		clientSecret:    clientSecret,
	}
}

func (c *Collector) Collect() (Data, error) {
	usageURL, err := url.Parse(c.usageServiceURL)
	if err != nil {
		return Data{}, errors.Wrapf(err, UsageServiceURLParsingError)
	}

	uaaURL, err := c.cfApiClient.GetUAAURL()
	if err != nil {
		return Data{}, errors.Wrap(err, GetUAAURLError)
	}

	authedClient := cf.NewOAuthClient(
		uaaURL,
		c.clientID,
		c.clientSecret,
		5*time.Second,
		c.httpClient,
	)

	usageURL.Path = path.Join(usageURL.Path, "system_report", "app_usages")

	req, err := http.NewRequest(http.MethodGet, usageURL.String(), nil)
	if err != nil {
		return Data{}, errors.Wrap(err, CreateUsageServiceHTTPRequestError)
	}

	resp, err := authedClient.Do(req)
	if err != nil {
		return Data{}, errors.Wrap(err, UsageServiceRequestError)
	}
	if resp.StatusCode != http.StatusOK {
		return Data{}, errors.Errorf(UsageServiceUnexpectedResponseStatusErrorFormat, resp.StatusCode)
	}

	return NewData(resp.Body, data.AppUsageDataType), nil
}
