package ops

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/pkg/errors"
)

const (
	AuthorizationHeaderKey = "Authorization"
	PostPath               = "/placeholder"

	RequestCreationFailureMessage = "Failed make request object"
	PostFailedMessage             = "Failed to do request"
	UnexpectedResponseCodeFormat  = "Unexpected response code %d, request failed"
	ReadMetadataFileError         = "Unable to read metadata file"
	ReadDataFileError             = "Unable to read data file "
	InvalidMetadataFileError      = "Metadata file is invalid"
)

type SendExecutor struct{}

//go:generate counterfeiter . tarReader
type tarReader interface {
	ReadFile(string) ([]byte, error)
}

func (s SendExecutor) Send(reader tarReader, dataLoaderURL, apiToken string) error {
	metadataContent, err := reader.ReadFile(MetadataFileName)
	if err != nil {
		return errors.Wrap(err, ReadMetadataFileError)
	}

	var metadata Metadata
	err = json.Unmarshal(metadataContent, &metadata)
	if err != nil {
		return errors.Wrap(err, InvalidMetadataFileError)
	}

	for _, digest := range metadata.FileDigests {
		metadataReader, err := constructFileMetadataReader(metadata, digest)
		if err != nil {
			return errors.Wrap(err, RequestCreationFailureMessage)
		}

		fileContents, err := reader.ReadFile(digest.Name)
		if err != nil {
			return errors.Wrap(err, ReadDataFileError)
		}
		req, err := makeFileUploadRequest(
			fileContents,
			digest.Name,
			apiToken,
			dataLoaderURL+PostPath,
			metadataReader,
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

func constructFileMetadataReader(metadata Metadata, digest FileDigest) (io.Reader, error) {
	metadataMap := map[string]string{
		"filename":        digest.Name,
		"fileContentType": digest.MimeType,
		"fileMd5Checksum": digest.MD5Checksum,
		"collectedAt":     metadata.CollectedAt,
		"envType":         metadata.EnvType,
		"collectionId":    metadata.CollectionId,
	}
	metadataJson, err := json.Marshal(metadataMap)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(metadataJson), nil
}

func makeFileUploadRequest(fileContent []byte, fileName, apiToken, uploadURL string, metadataReader io.Reader) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	metadataPart, err := writer.CreateFormField("metadata")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(metadataPart, metadataReader)
	if err != nil {
		return nil, err
	}

	dataPart, err := writer.CreateFormFile("data", fileName)
	if err != nil {
		return nil, err
	}

	if _, err := dataPart.Write(fileContent); err != nil {
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
