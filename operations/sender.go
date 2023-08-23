package operations

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

const (
	AuthorizationHeaderKey         = "Authorization"
	PostPath                       = "/collections/batch"
	TarContentType                 = "application/tar"
	GzipContentEncoding            = "gzip"
	HTTPSenderVersionRequestHeader = "Pivotal-Telemetry-Sender-Version"

	RequestCreationFailureMessage = "Failed make request object"
	PostFailedMessage             = "Failed to do request"
	ReadDataFileError             = "Unable to read data file"
	UnauthorizedErrorMessage      = "User is not authorized to perform this action"
	UnexpectedServerErrorFormat   = "There was an issue sending collector_tar. Please try again or contact your VMware field team if this error persists. Error ID %s"
)

type SendExecutor struct{}

//go:generate counterfeiter . httpClient
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

func (s SendExecutor) Send(client httpClient, tarFilePath, dataLoaderURL, apiToken, senderVersion string) error {
	fileReader, err := gzipOnTheFlyFileReader(tarFilePath)
	if err != nil {
		return errors.Wrap(err, ReadDataFileError)
	}
	defer fileReader.Close()

	req, err := makeFileUploadRequest(fileReader, apiToken, dataLoaderURL+PostPath, senderVersion)
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
	req.Header.Set(AuthorizationHeaderKey, "Bearer "+apiToken)
	req.Header.Set(HTTPSenderVersionRequestHeader, senderVersion)
	req.Header.Set("Content-Type", TarContentType)
	req.Header.Set("Content-Encoding", GzipContentEncoding)

	return req, nil
}

func gzipOnTheFlyFileReader(fname string) (io.ReadCloser, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	// Use io.Pipe and a goroutine to create reader
	r, w := io.Pipe()
	go func() {
		// Always close the file.
		defer f.Close()

		// Copy file through gzip to pipe writer.
		gzw := gzip.NewWriter(w)
		_, err := io.Copy(gzw, f)

		// Use CloseWithError to propagate errors back to
		// the main goroutine.
		if err != nil {
			w.CloseWithError(err)
			return
		}

		// Flush the gzip writer.
		w.CloseWithError(gzw.Close())
	}()
	return r, nil
}

func checkStatusCode(resp *http.Response) error {
	switch statusCode := resp.StatusCode; statusCode {
	case http.StatusCreated:
		return nil
	case http.StatusUnauthorized:
		return errors.New(UnauthorizedErrorMessage)
	default:
		body, err := io.ReadAll(resp.Body)
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
