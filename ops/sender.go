package ops

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pivotal-cf/aqueduct-courier/file"
	"github.com/pkg/errors"
)

const (
	AuthorizationHeaderKey = "Authorization"
	PostPath               = "/placeholder"

	RequestCreationFailureMessage  = "Failed make request object"
	PostFailedMessage              = "Failed to do request"
	UnexpectedResponseCodeFormat   = "Unexpected response code %d, request failed"
	ReadMetadataFileErrorFormat    = "Error reading metadata from file %s"
	InvalidMetadataFileErrorFormat = "Metadata file %s is invalid"
)

type SendExecutor struct{}

func (s SendExecutor) Send(directoryPath, dataLoaderURL, apiToken string) error {
	metadataPath := filepath.Join(directoryPath, file.MetadataFileName)
	metadataContent, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return errors.Wrapf(err, ReadMetadataFileErrorFormat, metadataPath)
	}

	var metadata file.Metadata
	err = json.Unmarshal(metadataContent, &metadata)
	if err != nil {
		return errors.Wrapf(err, InvalidMetadataFileErrorFormat, metadataPath)
	}

	for _, digest := range metadata.FileDigests {
		req, err := makeFileUploadRequest(
			filepath.Join(directoryPath, digest.Name),
			apiToken,
			dataLoaderURL+PostPath,
			metadata.CollectedAt,
			digest,
		)
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
	}

	return nil
}

func makeFileUploadRequest(filePath, apiToken, uploadURL, collectedAt string, fileDigest file.Digest) (*http.Request, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	metadataPart, err := writer.CreateFormField("metadata")
	if err != nil {
		return nil, err
	}

	metadata := map[string]string{
		"filename":        fileDigest.Name,
		"fileContentType": fileDigest.MimeType,
		"fileMd5Checksum": fileDigest.MD5Checksum,
		"collectedAt":     collectedAt,
	}
	metadataJson, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(metadataPart, bytes.NewReader(metadataJson))
	if err != nil {
		return nil, err
	}

	dataPart, err := writer.CreateFormFile("data", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(dataPart, file)
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
