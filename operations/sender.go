package operations

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

const (
	AuthorizationHeaderKey         = "Authorization"
	PostPath                       = "/collections/batch"
	TarMimeType                    = "application/tar"
	HTTPSenderVersionRequestHeader = "Pivotal-Telemetry-Sender-Version"

	RequestCreationFailureMessage = "Failed make request object"
	PostFailedMessage             = "Failed to do request"
	ReadDataFileError             = "Unable to read data file"
	UnauthorizedErrorMessage      = "User is not authorized to perform this action"
	UnexpectedServerErrorFormat   = "There was an issue sending data. Please try again or contact your Pivotal field team if this error persists. Error ID %s"
)

type SendExecutor struct{}

//go:generate counterfeiter . httpClient
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

func (s SendExecutor) Send(client httpClient, tarFilePath, dataLoaderURL, apiToken, senderVersion string) error {
	file, err := os.Open(tarFilePath)
	if err != nil {
		return errors.Wrap(err, ReadDataFileError)
	}
	defer file.Close()

	req, err := makeFileUploadRequest(file, apiToken, dataLoaderURL+PostPath, senderVersion)
	if err != nil {
		return errors.Wrap(err, RequestCreationFailureMessage)
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, PostFailedMessage)
	}

	return checkStatusCode(resp)
}

func makeFileUploadRequest(bodyReader io.Reader, apiToken, uploadURL, senderVersion string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, uploadURL, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set(AuthorizationHeaderKey, "Token "+apiToken)
	req.Header.Set(HTTPSenderVersionRequestHeader, senderVersion)
	req.Header.Set("Content-Type", TarMimeType)

	return req, nil
}

func checkStatusCode(resp *http.Response) error {
	switch statusCode := resp.StatusCode; statusCode {
	case http.StatusCreated:
		return nil
	case http.StatusUnauthorized:
		return errors.New(UnauthorizedErrorMessage)
	default:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Errorf(UnexpectedServerErrorFormat, "unknown")
		}

		var errResp map[string]map[string]string
		err = json.Unmarshal(body, &errResp)
		if err != nil {
			return errors.Errorf(UnexpectedServerErrorFormat, "unknown")
		}
		return errors.Errorf(UnexpectedServerErrorFormat, errResp["error"]["uuid"])
	}
}
