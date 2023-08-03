package cf

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
)

const (
	CfApiURLParsingError                     = "error parsing CF API URL: %s"
	CreateCfApiHTTPRequestError              = "error creating HTTP request for CF API endpoint: %s"
	CfApiRequestError                        = "error accessing CF API endpoint: %s"
	CFApiReadResponseError                   = "error reading response from CF API endpoint: %s"
	CFApiUnmarshalError                      = "error unmarshaling response from CF API endpoint: %s"
	CFApiUnexpectedResponseStatusErrorFormat = "unexpected status in CF API response: %d"
	UAAEndpointEmptyError                    = "UAA url is empty"
)

type Client struct {
	cfApiURL   string
	httpClient httpClient
}

//go:generate counterfeiter . httpClient
type httpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

func NewClient(cfApiURL string, httpClient httpClient) *Client {
	return &Client{cfApiURL: cfApiURL, httpClient: httpClient}
}

func (cl *Client) GetUAAURL() (string, error) {
	cfApiURL, err := url.Parse(cl.cfApiURL)
	if err != nil {
		return "", errors.Wrapf(err, CfApiURLParsingError, cl.cfApiURL)
	}

	cfApiURL.Path = path.Join(cfApiURL.Path, "/v2/info")
	req, err := http.NewRequest(http.MethodGet, cfApiURL.String(), nil)
	if err != nil {
		return "", errors.Wrap(err, CreateCfApiHTTPRequestError)
	}

	resp, err := cl.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, CfApiRequestError, cfApiURL.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf(CFApiUnexpectedResponseStatusErrorFormat, resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, CFApiReadResponseError, cfApiURL.String())
	}

	var cfResponse struct {
		TokenEndpoint string `json:"token_endpoint"`
	}
	err = json.Unmarshal(respBody, &cfResponse)
	if err != nil {
		return "", errors.Wrapf(err, CFApiUnmarshalError, cfApiURL.String())
	}

	if cfResponse.TokenEndpoint == "" {
		return "", errors.New(UAAEndpointEmptyError)
	}

	return cfResponse.TokenEndpoint, nil
}
