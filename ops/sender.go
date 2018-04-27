package ops

import (
	"net/http"

	"io/ioutil"

	"os"
	"path/filepath"

	"bytes"
	"io"
	"mime/multipart"

	"github.com/pkg/errors"
)

const (
	AuthorizationHeaderKey = "Authorization"
	PostPath               = "/placeholder"

	RequestCreationFailureMessage = "Failed make request object"
	PostFailedMessage             = "Failed to do request"
	UnexpectedResponseCodeFormat  = "Unexpected response code %d, request failed"
	ReadDirectoryErrorFormat      = "Error reading %s"
	NoDataErrorFormat             = "Cannot find data in %s"
)

type SendExecutor struct{}

func (s SendExecutor) Send(directoryPath, dataLoaderURL, apiToken string) error {
	fileInfos, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		return errors.Wrapf(err, ReadDirectoryErrorFormat, directoryPath)
	}

	dataSent := false
	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}

		req, err := makeFileUploadRequest(filepath.Join(directoryPath, info.Name()), apiToken, dataLoaderURL+PostPath)
		if err != nil {
			return errors.Wrap(err, RequestCreationFailureMessage)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrap(err, PostFailedMessage)
		}
		if resp.StatusCode != http.StatusCreated {
			return errors.Errorf(UnexpectedResponseCodeFormat, resp.StatusCode)
		}
		dataSent = true
	}

	if !dataSent {
		return errors.Errorf(NoDataErrorFormat, directoryPath)
	}

	return nil
}

func makeFileUploadRequest(filePath, apiToken, uploadURL string) (*http.Request, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("data", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, uploadURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(AuthorizationHeaderKey, "Token "+apiToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}
