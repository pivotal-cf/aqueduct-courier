package cf

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
)

const (
	CfApiURLParsingError                     = "error parsing CF API URL"
	CreateCfApiTTPRequestError               = "error creating HTTP request to CF API endpoint"
	CfApiRequestError                        = "error accessing CF API ENDPOINT"
	CFApiUnmarshalError                      = "error unmarshaling CF API response"
	CFApiUnexpectedResponseStatusErrorFormat = "unexpected status in CF API response: %d"
	UAAEndpointEmptyError                    = "UAA url is empty"
)

type Client struct {
	cfApiURL   string
	httpClient httpClient
}

type httpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

func NewClient(cfApiURL string, httpClient httpClient) *Client {
	return &Client{cfApiURL: cfApiURL, httpClient: httpClient}
}

func (cl *Client) GetUAAURL() (string, error) {
	cfApiURL, err := url.Parse(cl.cfApiURL)
	if err != nil {
		return "", errors.Wrap(err, CfApiURLParsingError)
	}

	cfApiURL.Path = path.Join(cfApiURL.Path, "/v2/info")
	req, err := http.NewRequest(http.MethodGet, cfApiURL.String(), nil)
	if err != nil {
		return "", errors.Wrap(err, CreateCfApiTTPRequestError)
	}

	resp, err := cl.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, CfApiRequestError)
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf(CFApiUnexpectedResponseStatusErrorFormat, resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	var cfResponse struct {
		TokenEndpoint string `json:"token_endpoint"`
	}
	err = json.Unmarshal(respBody, &cfResponse)
	if err != nil {
		return "", errors.Wrap(err, CFApiUnmarshalError)
	}

	if cfResponse.TokenEndpoint == "" {
		return "", errors.New(UAAEndpointEmptyError)
	}

	return cfResponse.TokenEndpoint, nil
}
